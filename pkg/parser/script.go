package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// Position represents a location in a script
type Position struct {
	Line   int
	Column int
	Offset int
}

// String returns a string representation of the position
func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}

// ScriptError represents an error with location information
type ScriptError struct {
	Pos  Position
	Msg  string
	Text string
}

// Error implements the error interface
func (e *ScriptError) Error() string {
	if e.Text != "" {
		return fmt.Sprintf("%s: %s\n%s\n%s^", e.Pos.String(), e.Msg, e.Text, strings.Repeat(" ", e.Pos.Column-1))
	}
	return fmt.Sprintf("%s: %s", e.Pos.String(), e.Msg)
}

// Statement represents a parsed statement with location info
type Statement struct {
	Text     string
	Position Position
	Type     StatementType
}

// StatementType represents the type of statement
type StatementType int

const (
	StatementTypeSQL StatementType = iota
	StatementTypeTQL
	StatementTypeComment
	StatementTypeEmpty
)

// Script represents a parsed script
type Script struct {
	Statements []Statement
	Source     string
}

// Parser parses TQL scripts with error location tracking
type Parser struct {
	scanner    *bufio.Scanner
	lineNum    int
	lineOffset int
	source     string
}

// NewParser creates a new script parser
func NewParser(source string) *Parser {
	return &Parser{
		scanner:    bufio.NewScanner(strings.NewReader(source)),
		lineNum:    0,
		lineOffset: 0,
		source:     source,
	}
}

// Parse parses the entire script
func (p *Parser) Parse() (*Script, error) {
	var statements []Statement
	var fullSource strings.Builder

	for p.scanner.Scan() {
		p.lineNum++
		line := p.scanner.Text()
		fullSource.WriteString(line)
		fullSource.WriteString("\n")

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			statements = append(statements, Statement{
				Text:     line,
				Position: Position{Line: p.lineNum, Column: 1, Offset: p.lineOffset},
				Type:     StatementTypeEmpty,
			})
			p.lineOffset += len(line) + 1
			continue
		}

		// Handle comment lines
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
			statements = append(statements, Statement{
				Text:     line,
				Position: Position{Line: p.lineNum, Column: 1, Offset: p.lineOffset},
				Type:     StatementTypeComment,
			})
			p.lineOffset += len(line) + 1
			continue
		}

		// Handle multi-line statements
		if strings.HasSuffix(trimmed, ";") {
			// Single line statement
			stmtType := p.determineStatementType(line)
			statements = append(statements, Statement{
				Text:     line,
				Position: Position{Line: p.lineNum, Column: 1, Offset: p.lineOffset},
				Type:     stmtType,
			})
			p.lineOffset += len(line) + 1
		} else {
			// Multi-line statement - collect until semicolon
			multiLine := line
			startLine := p.lineNum
			startOffset := p.lineOffset
			startColumn := 1

			for p.scanner.Scan() {
				p.lineNum++
				nextLine := p.scanner.Text()
				fullSource.WriteString(nextLine)
				fullSource.WriteString("\n")

				multiLine += "\n" + nextLine
				p.lineOffset += len(nextLine) + 1

				if strings.HasSuffix(strings.TrimSpace(nextLine), ";") {
					break
				}
			}

			stmtType := p.determineStatementType(multiLine)
			statements = append(statements, Statement{
				Text:     multiLine,
				Position: Position{Line: startLine, Column: startColumn, Offset: startOffset},
				Type:     stmtType,
			})
		}
	}

	if err := p.scanner.Err(); err != nil {
		return nil, &ScriptError{
			Pos: Position{Line: p.lineNum, Column: 1, Offset: p.lineOffset},
			Msg: fmt.Sprintf("IO error: %v", err),
		}
	}

	return &Script{
		Statements: statements,
		Source:     fullSource.String(),
	}, nil
}

// determineStatementType determines if a statement is SQL, TQL, or comment
func (p *Parser) determineStatementType(text string) StatementType {
	trimmed := strings.TrimSpace(text)

	// Empty lines
	if trimmed == "" {
		return StatementTypeEmpty
	}

	// Comments
	if strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
		return StatementTypeComment
	}

	upperText := strings.ToUpper(trimmed)

	// TQL-specific keywords (unambiguous)
	tqlKeywords := []string{
		"CREATE TENSOR", "DROP TENSOR", "ALTER TENSOR",
		"SHOW TENSORS", "DESCRIBE TENSOR",
		"COSINE_SIMILARITY", "EUCLIDEAN_DISTANCE",
		"TENSOR_SLICE", "TENSOR_RESHAPE",
		// Unambiguous tensor operations
		"TRANSPOSE", "MATRIX_MULTIPLY",
		"RELU", "SIGMOID", "TANH",
		"SVD", "EIGENVALUES",
		"CONV1D", "CONV2D",
		"ADD", "MULTIPLY",
	}

	// Check for unambiguous TQL keywords first
	for _, keyword := range tqlKeywords {
		if strings.HasPrefix(upperText, keyword) {
			return StatementTypeTQL
		}
	}

	// Check if statement contains TQL functions (even if it starts with SELECT)
	if strings.Contains(upperText, "COSINE_SIMILARITY") ||
		strings.Contains(upperText, "EUCLIDEAN_DISTANCE") ||
		strings.Contains(upperText, "TENSOR_SLICE") ||
		strings.Contains(upperText, "TENSOR_RESHAPE") {
		return StatementTypeTQL
	}

	// Special handling for ambiguous operations (SUM, MEAN, MAX, MIN)
	// These are TQL only if they appear as standalone operations
	if p.isStandaloneTensorOperation(trimmed) {
		return StatementTypeTQL
	}

	// Default to SQL
	return StatementTypeSQL
}

// ValidateStatement performs basic validation on a statement
func (p *Parser) ValidateStatement(stmt Statement) error {
	trimmed := strings.TrimSpace(stmt.Text)

	// Skip empty and comment statements
	if stmt.Type == StatementTypeEmpty || stmt.Type == StatementTypeComment {
		return nil
	}

	// Check for basic SQL syntax
	if !strings.HasSuffix(trimmed, ";") {
		return &ScriptError{
			Pos:  stmt.Position,
			Msg:  "Statement must end with semicolon",
			Text: stmt.Text,
		}
	}

	// Check for balanced parentheses
	if err := p.checkBalancedParentheses(stmt); err != nil {
		return err
	}

	// Check for TQL-specific syntax
	if stmt.Type == StatementTypeTQL {
		return p.validateTQLStatement(stmt)
	}

	return nil
}

// checkBalancedParentheses checks for balanced parentheses in a statement
func (p *Parser) checkBalancedParentheses(stmt Statement) error {
	text := stmt.Text
	stack := make([]rune, 0)

	for i, char := range text {
		switch char {
		case '(':
			stack = append(stack, char)
		case ')':
			if len(stack) == 0 {
				line, col := p.getLineColumn(stmt.Position.Offset + i)
				return &ScriptError{
					Pos:  Position{Line: line, Column: col, Offset: stmt.Position.Offset + i},
					Msg:  "Unmatched closing parenthesis",
					Text: stmt.Text,
				}
			}
			stack = stack[:len(stack)-1]
		}
	}

	if len(stack) > 0 {
		return &ScriptError{
			Pos:  stmt.Position,
			Msg:  "Unmatched opening parenthesis",
			Text: stmt.Text,
		}
	}

	return nil
}

// validateTQLStatement performs TQL-specific validation
func (p *Parser) validateTQLStatement(stmt Statement) error {
	text := strings.ToUpper(stmt.Text)

	// Validate CREATE TENSOR syntax
	if strings.Contains(text, "CREATE TENSOR") {
		// Support both single-line and multi-line CREATE TENSOR with optional chunk_size
		re := regexp.MustCompile(`(?i)CREATE\s+TENSOR\s+(\w+)\s*\(\s*shape\s*\[([^\]]+)\]\s*,\s*dtype\s+(\w+)(?:\s*,\s*chunk_size\s*\[([^\]]+)\])?\s*\)\s*;`)
		matches := re.FindStringSubmatch(text)
		if matches == nil {
			return &ScriptError{
				Pos:  stmt.Position,
				Msg:  "Invalid CREATE TENSOR syntax. Expected: CREATE TENSOR name (shape [dims], dtype type[, chunk_size [dims]])",
				Text: stmt.Text,
			}
		}

		// Validate shape format
		shapeStr := matches[2]
		if !regexp.MustCompile(`^\s*\d+(\s*,\s*\d+)*\s*$`).MatchString(shapeStr) {
			return &ScriptError{
				Pos:  stmt.Position,
				Msg:  "Invalid tensor shape format. Expected comma-separated integers",
				Text: stmt.Text,
			}
		}

		// Validate chunk_size format if present
		if len(matches) > 4 && matches[4] != "" {
			chunkStr := matches[4]
			if !regexp.MustCompile(`^\s*\d+(\s*,\s*\d+)*\s*$`).MatchString(chunkStr) {
				return &ScriptError{
					Pos:  stmt.Position,
					Msg:  "Invalid chunk_size format. Expected comma-separated integers",
					Text: stmt.Text,
				}
			}
		}
	}

	// Validate tensor operations
	if err := p.validateTensorOperation(stmt); err != nil {
		return err
	}

	return nil
}

// isStandaloneTensorOperation checks if this is a standalone tensor operation
func (p *Parser) isStandaloneTensorOperation(text string) bool {
	trimmed := strings.TrimSpace(text)

	// Remove trailing semicolon for checking
	if strings.HasSuffix(trimmed, ";") {
		trimmed = strings.TrimSpace(trimmed[:len(trimmed)-1])
	}

	// Simple pattern: OPERATION(tensor_name) or OPERATION(tensor_name, params)
	patterns := []string{
		`^SUM\(\w+\s*(,\s*axis\s*=\s*\d+)?\s*\)$`,
		`^MEAN\(\w+\s*(,\s*axis\s*=\s*\d+)?\s*\)$`,
		`^MAX\(\w+\s*(,\s*axis\s*=\s*\d+)?\s*\)$`,
		`^MIN\(\w+\s*(,\s*axis\s*=\s*\d+)?\s*\)$`,
	}

	for _, pattern := range patterns {
		if regexp.MustCompile(`(?i)` + pattern).MatchString(trimmed) {
			return true
		}
	}

	return false
}

// validateTensorOperation validates tensor operation syntax
func (p *Parser) validateTensorOperation(stmt Statement) error {
	text := strings.ToUpper(stmt.Text)

	// Pattern for tensor operations: OPERATION(tensor_name, parameters...)
	// Order matters - check longer operations first to avoid partial matches
	operationPatterns := map[string]string{
		"COSINE_SIMILARITY":  `(?i)COSINE_SIMILARITY\s*\(\s*(\w+)\s*,\s*(\w+)\s*\)`,
		"EUCLIDEAN_DISTANCE": `(?i)EUCLIDEAN_DISTANCE\s*\(\s*(\w+)\s*,\s*(\w+)\s*\)`,
		"MATRIX_MULTIPLY":    `(?i)MATRIX_MULTIPLY\s*\(\s*(\w+)\s*,\s*(\w+)\s*\)`,
		"EIGENVALUES":        `(?i)EIGENVALUES\s*\(\s*(\w+)\s*\)`,
		"CONV2D":             `(?i)CONV2D\s*\(\s*(\w+)\s*,\s*(\w+)\s*(?:,\s*stride\s*=\s*\[(\d+,\s*\d+)\])?\s*(?:,\s*padding\s*=\s*\[(\d+,\s*\d+)\])?\s*\)`,
		"CONV1D":             `(?i)CONV1D\s*\(\s*(\w+)\s*,\s*(\w+)\s*(?:,\s*stride\s*=\s*(\d+))?\s*(?:,\s*padding\s*=\s*(\d+))?\s*\)`,
		"TRANSPOSE":          `(?i)TRANSPOSE\s*\(\s*(\w+)\s*\)`,
		"SIGMOID":            `(?i)SIGMOID\s*\(\s*(\w+)\s*\)`,
		"RELU":               `(?i)RELU\s*\(\s*(\w+)\s*\)`,
		"TANH":               `(?i)TANH\s*\(\s*(\w+)\s*\)`,
		"SVD":                `(?i)SVD\s*\(\s*(\w+)\s*\)`,
		"MULTIPLY":           `(?i)MULTIPLY\s*\(\s*(\w+)\s*,\s*(\w+)\s*\)`,
		"ADD":                `(?i)ADD\s*\(\s*(\w+)\s*,\s*(\w+)\s*\)`,
		"SUM":                `(?i)SUM\s*\(\s*(\w+)\s*(?:,\s*axis\s*=\s*(\d+))?\s*\)`,
		"MEAN":               `(?i)MEAN\s*\(\s*(\w+)\s*(?:,\s*axis\s*=\s*(\d+))?\s*\)`,
		"MAX":                `(?i)MAX\s*\(\s*(\w+)\s*(?:,\s*axis\s*=\s*(\d+))?\s*\)`,
		"MIN":                `(?i)MIN\s*\(\s*(\w+)\s*(?:,\s*axis\s*=\s*(\d+))?\s*\)`,
	}

	// Check for exact operation match at the beginning of the statement (after whitespace)
	trimmed := strings.TrimSpace(text)
	for operation, pattern := range operationPatterns {
		if strings.HasPrefix(trimmed, operation) {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(text)
			if matches == nil {
				return &ScriptError{
					Pos:  stmt.Position,
					Msg:  fmt.Sprintf("Invalid %s syntax", operation),
					Text: stmt.Text,
				}
			}
			return nil // Found matching operation
		}
	}

	return nil
}

// getLineColumn converts an offset to line and column
func (p *Parser) getLineColumn(offset int) (int, int) {
	lines := strings.Split(p.source[:offset], "\n")
	line := len(lines)
	column := len(lines[len(lines)-1]) + 1
	return line, column
}

// ParseScript is a convenience function to parse a script
func ParseScript(source string) (*Script, error) {
	parser := NewParser(source)
	return parser.Parse()
}

// ValidateScript validates all statements in a script
func ValidateScript(script *Script) []error {
	var errors []error
	parser := &Parser{}

	for _, stmt := range script.Statements {
		if err := parser.ValidateStatement(stmt); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
