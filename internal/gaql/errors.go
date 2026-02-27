package gaql

import "fmt"

// ParseError represents a GAQL parsing error.
type ParseError struct {
	Message string
	Line    int
	Column  int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("gaql: %s at line %d, column %d", e.Message, e.Line, e.Column)
}

// ValidationError represents a GAQL semantic validation error.
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("gaql: validation error on %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("gaql: validation error: %s", e.Message)
}
