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
	// Be nil-safe to avoid panics when formatting errors through interfaces
	if e == nil {
		return "Runtime error: unknown"
	}
	// Make error messages more user-friendly: only include location when available
	if e.Line > 0 && e.Column > 0 {
		return fmt.Sprintf("Runtime error at %d:%d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("Runtime error: %s", e.Message)
}

func NewError(message string, line int, column int) *Error {
	return &Error{Message: message, Line: line, Column: column}
}
