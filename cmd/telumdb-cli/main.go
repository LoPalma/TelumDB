package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/telumdb/telumdb/internal/cli"
	"github.com/telumdb/telumdb/internal/client"
	"github.com/telumdb/telumdb/pkg/parser"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var (
		serverURL   = flag.String("url", "telumdb://localhost:5432", "Server URL")
		database    = flag.String("db", "", "Database name")
		username    = flag.String("user", "", "Username")
		password    = flag.String("password", "", "Password")
		command     = flag.String("c", "", "Execute command and exit")
		file        = flag.String("f", "", "Execute commands from file")
		batch       = flag.Bool("batch", false, "Run in batch mode (continue on errors)")
		verbose     = flag.Bool("v", false, "Verbose output")
		showHelp    = flag.Bool("help", false, "Show help message")
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	if *showVersion {
		printVersion()
		return
	}

	// Create client configuration
	clientConfig := &client.Config{
		ServerURL: *serverURL,
		Database:  *database,
		Username:  *username,
		Password:  *password,
	}

	// Initialize client
	dbClient, err := client.New(clientConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer dbClient.Close()

	ctx := context.Background()

	// Execute single command if provided
	if *command != "" {
		if err := executeCommand(ctx, dbClient, *command); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Execute commands from file if provided
	if *file != "" {
		if err := executeFileBatch(ctx, dbClient, *file, *batch, *verbose); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Interactive mode
	runInteractive(ctx, dbClient)
}

func executeCommand(ctx context.Context, cli *client.Client, command string) error {
	result, err := cli.Execute(ctx, command)
	if err != nil {
		return err
	}
	return printResult(result)
}

func executeFile(ctx context.Context, cli *client.Client, filename string) error {
	return executeFileBatch(ctx, cli, filename, false, false)
}

func executeFileBatch(ctx context.Context, cli *client.Client, filename string, batchMode, verbose bool) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Parse the script with error location tracking
	script, err := parser.ParseScript(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse script %s: %w", filename, err)
	}

	// Validate the script
	if errors := parser.ValidateScript(script); len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "Script validation errors in %s:\n", filename)
		for _, parseErr := range errors {
			fmt.Fprintf(os.Stderr, "  %v\n", parseErr)
		}
		if !batchMode {
			return fmt.Errorf("script validation failed")
		}
		fmt.Fprintf(os.Stderr, "Continuing in batch mode...\n")
	}

	// Execute statements
	var errorCount int
	for i, stmt := range script.Statements {
		if stmt.Type == parser.StatementTypeEmpty || stmt.Type == parser.StatementTypeComment {
			continue
		}

		if verbose {
			fmt.Printf("Executing statement %d/%d at %s...\n", i+1, len(script.Statements), stmt.Position.String())
		}

		if err := executeCommand(ctx, cli, stmt.Text); err != nil {
			errorCount++
			fmt.Fprintf(os.Stderr, "Error at %s: %v\n", stmt.Position.String(), err)
			if !batchMode {
				return fmt.Errorf("execution stopped due to error")
			}
			fmt.Fprintf(os.Stderr, "Continuing in batch mode...\n")
		} else if verbose {
			fmt.Printf("Statement %d executed successfully\n", i+1)
		}
	}

	if errorCount > 0 {
		fmt.Fprintf(os.Stderr, "Completed with %d errors\n", errorCount)
	} else if verbose {
		fmt.Printf("All %d statements executed successfully\n", len(script.Statements))
	}

	return nil
}

func runInteractive(ctx context.Context, dbClient *client.Client) {
	// Create REPL configuration
	homeDir, _ := os.UserHomeDir()
	historyFile := filepath.Join(homeDir, ".telumdb_history")

	config := &cli.Config{
		HistoryFile: historyFile,
		Prompt:      "telumdb> ",
		Continue:    "... ",
		MaxHistory:  1000,
	}

	// Create and run REPL
	repl, err := cli.NewREPL(ctx, dbClient, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create REPL: %v\n", err)
		os.Exit(1)
	}
	defer repl.Close()

	if err := repl.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: REPL failed: %v\n", err)
		os.Exit(1)
	}
}

func handleMetaCommand(command string, cli *client.Client) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "\\h", "\\help":
		printHelp()
	case "\\q", "\\quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	case "\\l", "\\list":
		return listDatabases(cli)
	case "\\dt", "\\tables":
		return listTables(cli)
	case "\\dT", "\\tensors":
		return listTensors(cli)
	case "\\d", "\\describe":
		if len(parts) < 2 {
			return fmt.Errorf("usage: \\d <table|tensor_name>")
		}
		return describeObject(cli, parts[1])
	case "\\c", "\\connect":
		if len(parts) < 2 {
			return fmt.Errorf("usage: \\c <database_name>")
		}
		return connectToDatabase(cli, parts[1])
	default:
		return fmt.Errorf("unknown command: %s", parts[0])
	}

	return nil
}

func listDatabases(_ *client.Client) error {
	// TODO: Implement database listing
	fmt.Println("Databases:")
	fmt.Println("  telumdb")
	return nil
}

func listTables(cli *client.Client) error {
	result, err := cli.Execute(context.Background(), "SHOW TABLES")
	if err != nil {
		return err
	}
	return printResult(result)
}

func listTensors(cli *client.Client) error {
	result, err := cli.Execute(context.Background(), "SHOW TENSORS")
	if err != nil {
		return err
	}
	return printResult(result)
}

func describeObject(cli *client.Client, name string) error {
	result, err := cli.Execute(context.Background(), fmt.Sprintf("DESCRIBE %s", name))
	if err != nil {
		return err
	}
	return printResult(result)
}

func connectToDatabase(_ *client.Client, dbname string) error {
	// TODO: Implement database switching
	fmt.Printf("Connected to database %s\n", dbname)
	return nil
}

func printResult(result *client.Result) error {
	if result == nil {
		return nil
	}

	// Print column headers
	if len(result.Columns) > 0 {
		fmt.Println(strings.Join(result.Columns, " | "))
		separator := make([]string, len(result.Columns))
		for i := range separator {
			separator[i] = "---"
		}
		fmt.Println(strings.Join(separator, " | "))
	}

	// Print rows
	for _, row := range result.Rows {
		rowStrings := make([]string, len(row))
		for i, col := range row {
			rowStrings[i] = fmt.Sprintf("%v", col)
		}
		fmt.Println(strings.Join(rowStrings, " | "))
	}

	// Print affected rows if applicable
	if result.Affected > 0 {
		fmt.Printf("%d rows affected.\n", result.Affected)
	}

	return nil
}

func printHelp() {
	fmt.Printf(`TelumDB CLI - Interactive Database Client

Usage:
  telumdb-cli [options]

Options:
  -url string        Server URL (default: "telumdb://localhost:5432")
  -db string          Database name
  -user string        Username
  -password string    Password
  -c string           Execute command and exit
  -f string           Execute commands from file
  -batch              Run in batch mode (continue on errors)
  -v                  Verbose output
  -help               Show this help message
  -version            Show version information

Meta Commands:
  \h, \help           Show this help message
  \q, \quit           Quit the CLI
  \l, \list           List databases
  \dt, \tables        List tables
  \dT, \tensors       List tensors
  \d <name>           Describe table or tensor
  \c <database>       Connect to database

SQL/TQL Commands:
  Standard SQL commands for traditional data
  Extended TQL commands for tensor operations

Examples:
  # Interactive mode
  telumdb-cli

  # Single command
  telumdb-cli -c "SHOW TABLES"

  # Execute script with error reporting
  telumdb-cli -f script.tql

  # Batch mode (continue on errors)
  telumdb-cli -f script.tql -batch

  # Verbose execution
  telumdb-cli -f script.tql -v

Script Format:
  -- Comments start with --
  CREATE TABLE users (id INTEGER, name VARCHAR(255));
  CREATE TENSOR embeddings (shape [1000, 768], dtype float32);
  SELECT * FROM users;

For more information, visit: https://github.com/telumdb/telumdb
`)
}

func printVersion() {
	fmt.Printf(`TelumDB CLI %s
Commit: %s
Built: %s

Copyright 2024 TelumDB Contributors
License: Apache License 2.0
`,
		version,
		commit,
		date,
	)
}
