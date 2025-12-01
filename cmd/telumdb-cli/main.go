package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/telumdb/telumdb/internal/client"
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
	cli, err := client.New(clientConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	ctx := context.Background()

	// Execute single command if provided
	if *command != "" {
		if err := executeCommand(ctx, cli, *command); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Execute commands from file if provided
	if *file != "" {
		if err := executeFile(ctx, cli, *file); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Interactive mode
	runInteractive(ctx, cli)
}

func executeCommand(ctx context.Context, cli *client.Client, command string) error {
	result, err := cli.Execute(ctx, command)
	if err != nil {
		return err
	}
	return printResult(result)
}

func executeFile(ctx context.Context, cli *client.Client, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		if err := executeCommand(ctx, cli, line); err != nil {
			return fmt.Errorf("error executing command '%s': %w", line, err)
		}
	}

	return scanner.Err()
}

func runInteractive(ctx context.Context, cli *client.Client) {
	fmt.Printf("TelumDB CLI %s\n", version)
	fmt.Printf("Connected to %s\n", cli.Config().ServerURL)
	fmt.Println("Type '\\h' for help, '\\q' to quit.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("telumdb> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Handle meta commands
		if strings.HasPrefix(line, "\\") {
			if err := handleMetaCommand(line, cli); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			continue
		}

		// Execute SQL/TQL command
		if err := executeCommand(ctx, cli, line); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
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

func listDatabases(cli *client.Client) error {
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

func connectToDatabase(cli *client.Client, dbname string) error {
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
  CREATE TABLE users (id INTEGER, name VARCHAR(255));
  CREATE TENSOR embeddings (shape [768], dtype float32);
  SELECT * FROM users;
  SELECT * FROM embeddings WHERE slice = [0:100];

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
