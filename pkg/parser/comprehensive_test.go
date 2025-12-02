package parser

import (
	"testing"
)

func TestComprehensiveSQLParsing(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected int // Number of statements expected
	}{
		{
			name: "Basic SQL Operations",
			source: `SELECT * FROM users;
INSERT INTO users (name, email) VALUES ('John', 'john@example.com');
UPDATE users SET email = 'new@example.com' WHERE id = 1;
DELETE FROM users WHERE id = 1;`,
			expected: 4,
		},
		{
			name: "Advanced SQL Operations",
			source: `CREATE TABLE users (
	id INTEGER PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	email VARCHAR(255) UNIQUE,
	age INTEGER CHECK (age >= 0)
);
CREATE INDEX idx_users_email ON users(email);
CREATE VIEW active_users AS SELECT * FROM users WHERE active = 1;
DROP VIEW IF EXISTS active_users;`,
			expected: 4,
		},
		{
			name: "SQL Joins and Aggregates",
			source: `SELECT u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.created_at >= '2023-01-01'
GROUP BY u.id, u.name
HAVING COUNT(o.id) > 5
ORDER BY order_count DESC
LIMIT 10;`,
			expected: 1,
		},
		{
			name: "SQL Transactions",
			source: `BEGIN;
INSERT INTO audit_log (action) VALUES ('start');
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
INSERT INTO audit_log (action) VALUES ('transfer');
COMMIT;`,
			expected: 5,
		},
		{
			name: "SQL Functions and Case Statements",
			source: `SELECT 
	id,
	CASE 
		WHEN age < 18 THEN 'minor'
		WHEN age BETWEEN 18 AND 65 THEN 'adult'
		ELSE 'senior'
	END as age_group,
	UPPER(name) as uppercase_name,
	COALESCE(phone, 'N/A') as phone
FROM users
WHERE name LIKE '%john%' 
   AND email IS NOT NULL;`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			script, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != tt.expected {
				t.Errorf("Expected %d statements, got %d", tt.expected, len(script.Statements))
			}

			// Validate all statements
			for _, stmt := range script.Statements {
				if err := parser.ValidateStatement(stmt); err != nil {
					t.Errorf("Statement validation failed: %v\nStatement: %s", err, stmt.Text)
				}
			}
		})
	}
}

func TestComprehensiveTensorParsing(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected int // Number of statements expected
	}{
		{
			name: "Basic Tensor Operations",
			source: `CREATE TENSOR embeddings (shape [1000, 768], dtype float32);
CREATE TENSOR weights (shape [768, 256], dtype float32);
TRANSPOSE(embeddings);
SUM(embeddings, axis=0);
MEAN(embeddings, axis=1);
MAX(embeddings);
MIN(embeddings);`,
			expected: 6,
		},
		{
			name: "Advanced Tensor Operations",
			source: `MATRIX_MULTIPLY(embeddings, weights);
ADD(embeddings, bias);
MULTIPLY(embeddings, scale);
RELU(hidden_layer);
SIGMOID(output_layer);
TANH(activation_layer);`,
			expected: 6,
		},
		{
			name: "Convolution Operations",
			source: `CONV1D(input_signal, kernel_1d, stride=2, padding=1);
CONV2D(input_image, kernel_2d, stride=[2,2], padding=[1,1]);
CREATE TENSOR conv_filter (shape [3, 3, 1, 32], dtype float32);`,
			expected: 3,
		},
		{
			name: "Linear Algebra Operations",
			source: `SVD(covariance_matrix);
EIGENVALUES(correlation_matrix);
COSINE_SIMILARITY(vector_a, vector_b);
EUCLIDEAN_DISTANCE(point_a, point_b);`,
			expected: 4,
		},
		{
			name: "Mixed SQL and TQL",
			source: `-- Create tables
CREATE TABLE documents (id INTEGER, content TEXT, embedding_vector BLOB);
CREATE TABLE queries (id INTEGER, text TEXT, tensor_name VARCHAR);

-- Create tensors
CREATE TENSOR doc_embeddings (shape [10000, 768], dtype float32);
CREATE TENSOR query_embeddings (shape [100, 768], dtype float32);

-- Process embeddings
TRANSPOSE(doc_embeddings);
RELU(doc_embeddings);

-- SQL with tensor functions
SELECT d.id, d.content, 
       COSINE_SIMILARITY(d.embedding_vector, q.tensor_name) as similarity
FROM documents d, queries q
WHERE d.id > 100
ORDER BY similarity DESC
LIMIT 10;`,
			expected: 8,
		},
		{
			name: "Complex Multi-line Statements",
			source: `SELECT 
	u.id,
	u.name,
	COUNT(o.id) as total_orders,
	SUM(o.amount) as total_spent,
	AVG(o.amount) as avg_order_value
FROM users u
INNER JOIN orders o ON u.id = o.user_id
WHERE u.created_at >= '2023-01-01'
  AND u.status = 'active'
GROUP BY u.id, u.name
HAVING COUNT(o.id) >= 5
ORDER BY total_spent DESC, total_orders DESC
LIMIT 20;`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			script, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(script.Statements) != tt.expected {
				t.Errorf("Expected %d statements, got %d", tt.expected, len(script.Statements))
			}

			// Validate all statements
			for _, stmt := range script.Statements {
				if err := parser.ValidateStatement(stmt); err != nil {
					t.Errorf("Statement validation failed: %v\nStatement: %s", err, stmt.Text)
				}
			}
		})
	}
}

func TestStatementTypeDetection(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected map[string]StatementType
	}{
		{
			name: "Mixed Statement Types",
			source: `-- This is a comment
SELECT * FROM users;
CREATE TENSOR test (shape [10, 20], dtype float32);
SHOW TENSORS;
INSERT INTO logs (message) VALUES ('test');

-- Another comment
UPDATE users SET active = TRUE;`,
			expected: map[string]StatementType{
				"-- This is a comment":                                StatementTypeComment,
				"SELECT * FROM users;":                                StatementTypeSQL,
				"CREATE TENSOR test (shape [10, 20], dtype float32);": StatementTypeTQL,
				"SHOW TENSORS;":                                       StatementTypeTQL,
				"INSERT INTO logs (message) VALUES ('test');":         StatementTypeSQL,
				"-- Another comment":                                  StatementTypeComment,
				"UPDATE users SET active = TRUE;":                     StatementTypeSQL,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			script, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			for i, stmt := range script.Statements {
				expectedType, exists := tt.expected[stmt.Text]
				if !exists {
					t.Errorf("Unexpected statement: %s", stmt.Text)
					continue
				}
				if stmt.Type != expectedType {
					t.Errorf("Statement %d: expected type %v, got %v\nStatement: %s",
						i, expectedType, stmt.Type, stmt.Text)
				}
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Unmatched Parentheses",
			source:      "SELECT * FROM users WHERE id = (SELECT id FROM orders;",
			expectError: true,
			errorMsg:    "Unmatched opening parenthesis",
		},
		{
			name:        "Missing Semicolon",
			source:      "SELECT * FROM users",
			expectError: true,
			errorMsg:    "Statement must end with semicolon",
		},
		{
			name:        "Invalid CREATE TENSOR Syntax",
			source:      "CREATE TENSOR test (invalid syntax);",
			expectError: true,
			errorMsg:    "Invalid CREATE TENSOR syntax",
		},
		{
			name:        "Invalid Tensor Operation Syntax",
			source:      "TRANSPOSE();",
			expectError: true,
			errorMsg:    "Invalid TRANSPOSE syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			script, err := parser.Parse()

			// Parse might succeed but validation should fail
			if err != nil && !tt.expectError {
				t.Fatalf("Unexpected parse error: %v", err)
			}

			if script != nil {
				for _, stmt := range script.Statements {
					validationErr := parser.ValidateStatement(stmt)
					if tt.expectError {
						if validationErr == nil {
							t.Errorf("Expected validation error but got none")
						} else if !contains(validationErr.Error(), tt.errorMsg) {
							t.Errorf("Expected error containing '%s', got: %s", tt.errorMsg, validationErr.Error())
						}
					} else if validationErr != nil {
						t.Errorf("Unexpected validation error: %v", validationErr)
					}
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
