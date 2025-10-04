package runtime

import (
	"fmt"
)

// Error represents a runtime error.
type Error struct {
	Message string
	Line    int
	Column  int
}

func (e *Error) Error() string {
	if e == nil {
		return "Runtime error: unknown" 
	}
	if e.Line > 0 && e.Column > 0 {
		return fmt.Sprintf("Runtime error at %d:%d: %s", e.Line, e.Column, e.Message) // include location
	}
	return fmt.Sprintf("Runtime error: %s", e.Message)
}

// NewError creates a new runtime error.
func NewError(message string, line int, column int) *Error {
	return &Error{Message: message, Line: line, Column: column}
}
