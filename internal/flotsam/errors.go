// Package flotsam provides Unix interop functionality for flotsam notes.
// This file contains error handling for flotsam operations.
package flotsam

import "fmt"

// Error represents errors from flotsam operations.
// This is simplified from the repository.Error type.
//revive:disable-next-line:exported Error is descriptive enough in flotsam package context
type Error struct {
	Operation string
	Context   string
	Err       error
}

func (e *Error) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("flotsam error in %s for context '%s': %v", e.Operation, e.Context, e.Err)
	}
	return fmt.Sprintf("flotsam error in %s: %v", e.Operation, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// NewError creates a new Error with operation and context.
func NewError(operation, context string, err error) *Error {
	return &Error{
		Operation: operation,
		Context:   context,
		Err:       err,
	}
}

// NewErrorNoContext creates a new Error without context.
func NewErrorNoContext(operation string, err error) *Error {
	return &Error{
		Operation: operation,
		Err:       err,
	}
}
