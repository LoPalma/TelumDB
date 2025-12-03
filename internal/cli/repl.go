package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/telumdb/telumdb/internal/client"
)

// REPL represents an interactive read-eval-print loop
type REPL struct {
	client    *client.Client
	rl        *readline.Instance
	ctx       context.Context
	prompt    string
	continues bool
}

// Config holds REPL configuration
type Config struct {
	HistoryFile string
	Prompt      string
	Continue    string
	MaxHistory  int
}

// NewREPL creates a new REPL instance
func NewREPL(ctx context.Context, cli *client.Client, config *Config) (*REPL, error) {
	if config == nil {
		config = &Config{
			HistoryFile: "",
			Prompt:      "telumdb> ",
			Continue:    "... ",
			MaxHistory:  1000,
		}
	}

	// Create readline instance
	rl, err := readline.NewEx(&readline.Config{
		Prompt:              config.Prompt,
		HistoryFile:         config.HistoryFile,
		AutoComplete:        &completer{},
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create readline: %w", err)
	}

	repl := &REPL{
		client:    cli,
		rl:        rl,
		ctx:       ctx,
		prompt:    config.Prompt,
		continues: false,
	}

	return repl, nil
}

// Run starts the REPL loop
func (r *REPL) Run() error {
	fmt.Printf("TelumDB CLI - Interactive Mode\n")
	fmt.Printf("Connected to %s\n", r.client.Config().ServerURL)
	fmt.Printf("Type '\\h' for help, '\\q' to quit.\n\n")

	for {
		// Set appropriate prompt
		if r.continues {
			r.rl.SetPrompt("... ")
		} else {
			r.rl.SetPrompt(r.prompt)
		}

		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if r.continues {
					// Cancel multi-line input
					r.continues = false
					fmt.Println("^C")
					continue
				}
				continue
			} else if err == io.EOF {
				fmt.Println("\nGoodbye!")
				break
			}
			return fmt.Errorf("read error: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle meta commands
		if strings.HasPrefix(line, "\\") {
			if err := r.handleMetaCommand(line); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			r.continues = false
			continue
		}

		// Check if this is a complete statement or needs continuation
		if r.needsContinuation(line) {
			if !r.continues {
				// Start collecting multi-line input
				r.continues = true
				r.rl.SetPrompt("... ")
			}
			continue
		}

		// Execute the command
		if err := r.executeCommand(line); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		r.continues = false
	}

	return nil
}

// Close closes the REPL and cleans up resources
func (r *REPL) Close() error {
	if r.rl != nil {
		return r.rl.Close()
	}
	return nil
}

// needsContinuation determines if a line needs continuation
func (r *REPL) needsContinuation(line string) bool {
	line = strings.TrimSpace(line)

	// Empty lines don't need continuation
	if line == "" {
		return false
	}

	// Check for unclosed quotes
	inQuote := false
	quoteChar := rune(0)

	for _, char := range line {
		if !inQuote && (char == '\'' || char == '"') {
			inQuote = true
			quoteChar = char
		} else if inQuote && char == quoteChar {
			inQuote = false
			quoteChar = 0
		}
	}

	if inQuote {
		return true
	}

	// Check for unclosed parentheses, brackets, braces
	parenCount := 0
	bracketCount := 0
	braceCount := 0

	for _, char := range line {
		switch char {
		case '(':
			parenCount++
		case ')':
			parenCount--
		case '[':
			bracketCount++
		case ']':
			bracketCount--
		case '{':
			braceCount++
		case '}':
			braceCount--
		}
	}

	if parenCount > 0 || bracketCount > 0 || braceCount > 0 {
		return true
	}

	// Check for line continuation character
	if strings.HasSuffix(line, "\\") {
		return true
	}

	// Check for keywords that typically require multi-line
	lowerLine := strings.ToLower(line)
	multiLineKeywords := []string{
		"create", "insert", "update", "delete", "select",
		"begin", "start transaction", "case",
	}

	for _, keyword := range multiLineKeywords {
		if strings.HasPrefix(lowerLine, keyword) && !strings.Contains(line, ";") {
			return true
		}
	}

	// Check for tensor operations that might span multiple lines
	tensorKeywords := []string{
		"create tensor", "tensor_slice", "tensor_reshape",
		"conv1d", "conv2d", "matrix_multiply",
	}

	for _, keyword := range tensorKeywords {
		if strings.Contains(lowerLine, keyword) && !strings.Contains(line, ";") {
			return true
		}
	}

	return false
}

// executeCommand executes a single command
func (r *REPL) executeCommand(command string) error {
	result, err := r.client.Execute(r.ctx, command)
	if err != nil {
		return err
	}
	return r.printResult(result)
}

// handleMetaCommand handles REPL meta commands
func (r *REPL) handleMetaCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "\\h", "\\help":
		r.printHelp()
	case "\\q", "\\quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	case "\\l", "\\list":
		return r.listDatabases()
	case "\\dt", "\\tables":
		return r.listTables()
	case "\\dT", "\\tensors":
		return r.listTensors()
	case "\\d", "\\describe":
		if len(parts) < 2 {
			return fmt.Errorf("usage: \\d <table|tensor_name>")
		}
		return r.describeObject(parts[1])
	case "\\c", "\\connect":
		if len(parts) < 2 {
			return fmt.Errorf("usage: \\c <database_name>")
		}
		return r.connectToDatabase(parts[1])
	case "\\history":
		return r.showHistory()
	case "\\clear":
		return r.clearHistory()
	default:
		return fmt.Errorf("unknown command: %s", parts[0])
	}

	return nil
}

// printResult prints query results
func (r *REPL) printResult(result *client.Result) error {
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

// Helper methods for meta commands
func (r *REPL) printHelp() {
	fmt.Printf(`TelumDB CLI - Interactive Database Client

Meta Commands:
  \h, \help           Show this help message
  \q, \quit           Quit the CLI
  \l, \list           List databases
  \dt, \tables        List tables
  \dT, \tensors       List tensors
  \d <name>           Describe table or tensor
  \c <database>       Connect to database
  \history            Show command history
  \clear              Clear command history

SQL/TQL Commands:
  Standard SQL commands for traditional data
  Extended TQL commands for tensor operations

Features:
  - Command history with up/down arrows
  - Multi-line input with automatic indentation
  - Tab completion support
  - Ctrl+C to cancel multi-line input
  - Ctrl+R to search history

`)
}

func (r *REPL) listDatabases() error {
	// TODO: Implement database listing
	fmt.Println("Databases:")
	fmt.Println("  telumdb")
	return nil
}

func (r *REPL) listTables() error {
	result, err := r.client.Execute(r.ctx, "SHOW TABLES")
	if err != nil {
		return err
	}
	return r.printResult(result)
}

func (r *REPL) listTensors() error {
	result, err := r.client.Execute(r.ctx, "SHOW TENSORS")
	if err != nil {
		return err
	}
	return r.printResult(result)
}

func (r *REPL) describeObject(name string) error {
	result, err := r.client.Execute(r.ctx, fmt.Sprintf("DESCRIBE %s", name))
	if err != nil {
		return err
	}
	return r.printResult(result)
}

func (r *REPL) connectToDatabase(dbname string) error {
	// TODO: Implement database switching
	fmt.Printf("Connected to database %s\n", dbname)
	return nil
}

func (r *REPL) showHistory() error {
	// Note: readline library doesn't expose GetHistory method directly
	// This is a placeholder - history management would need custom implementation
	fmt.Println("Command history is available with up/down arrows.")
	fmt.Println("Use Ctrl+R to search history.")
	return nil
}

func (r *REPL) clearHistory() error {
	// Note: readline library doesn't expose ClearHistory method directly
	// This is a placeholder - history management would need custom implementation
	fmt.Println("History clearing not yet implemented.")
	return nil
}

// filterInput filters input characters for readline
func filterInput(r rune) (rune, bool) {
	switch r {
	// Allow most characters, filter out problematic ones
	case 0, 1, 2, 3, 4, 5, 6, 7, 8, 11, 12, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31:
		return r, false
	}
	return r, true
}

// completer implements readline.AutoCompleter
type completer struct{}

// Do implements tab completion
func (c *completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	// Basic completion for SQL keywords and meta commands
	lineStr := string(line)

	// Meta commands
	if strings.HasPrefix(lineStr, "\\") {
		metaCommands := []string{
			"\\help", "\\h", "\\quit", "\\q", "\\list", "\\l",
			"\\tables", "\\dt", "\\tensors", "\\dT", "\\describe", "\\d",
			"\\connect", "\\c", "\\history", "\\clear",
		}

		var matches [][]rune
		for _, cmd := range metaCommands {
			if strings.HasPrefix(cmd, lineStr) {
				matches = append(matches, []rune(cmd))
			}
		}
		return matches, len(line)
	}

	// SQL keywords
	sqlKeywords := []string{
		"SELECT", "FROM", "WHERE", "INSERT", "INTO", "VALUES",
		"UPDATE", "SET", "DELETE", "CREATE", "TABLE", "DROP",
		"ALTER", "INDEX", "VIEW", "BEGIN", "COMMIT", "ROLLBACK",
		"SHOW", "DESCRIBE", "AND", "OR", "NOT", "NULL", "LIKE",
		"ORDER", "BY", "GROUP", "HAVING", "LIMIT", "OFFSET",
		"JOIN", "INNER", "LEFT", "RIGHT", "FULL", "OUTER",
		"UNION", "DISTINCT", "COUNT", "SUM", "AVG", "MAX", "MIN",
	}

	// TQL keywords
	tqlKeywords := []string{
		"TENSOR", "CREATE", "DROP", "ADD", "MULTIPLY", "MATRIX_MULTIPLY",
		"RELU", "SIGMOID", "TANH", "SUM", "MEAN", "MAX", "MIN",
		"TRANSPOSE", "SVD", "EIGENVALUES", "COSINE_SIMILARITY",
		"EUCLIDEAN_DISTANCE", "CONV1D", "CONV2D", "TENSOR_SLICE",
		"TENSOR_RESHAPE", "SHAPE", "DTYPE", "CHUNK_SIZE",
	}

	var matches [][]rune
	upperLine := strings.ToUpper(lineStr)

	for _, keyword := range append(sqlKeywords, tqlKeywords...) {
		if strings.HasPrefix(keyword, upperLine) {
			matches = append(matches, []rune(keyword))
		}
	}

	return matches, len(line)
}
