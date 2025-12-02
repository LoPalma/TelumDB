package storage

import (
	"fmt"
	"strings"
)

// SimpleCondition implements the Condition interface for basic filtering
type SimpleCondition struct {
	field    string
	operator string
	value    interface{}
}

// NewSimpleCondition creates a new simple condition
func NewSimpleCondition(field, operator string, value interface{}) Condition {
	return &SimpleCondition{
		field:    field,
		operator: operator,
		value:    value,
	}
}

// String returns the string representation of the condition
func (c *SimpleCondition) String() string {
	return fmt.Sprintf("%s %s %v", c.field, c.operator, c.value)
}

// Field returns the field name
func (c *SimpleCondition) Field() string {
	return c.field
}

// Operator returns the operator
func (c *SimpleCondition) Operator() string {
	return c.operator
}

// Value returns the value
func (c *SimpleCondition) Value() interface{} {
	return c.value
}

// MapCondition converts a map[string]interface{} to a Condition
func MapCondition(m map[string]interface{}) Condition {
	if len(m) == 0 {
		return nil
	}

	var conditions []string
	for field, value := range m {
		conditions = append(conditions, fmt.Sprintf("%s = %v", field, value))
	}

	return &SimpleCondition{
		field:    "combined",
		operator: "AND",
		value:    strings.Join(conditions, " AND "),
	}
}
