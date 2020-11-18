package cmdutil

import (
	"errors"
	"fmt"
)

var (
	// ErrSilent is an error that triggers a non zero exit code without any
	// error message.
	ErrSilent = errors.New("ErrSilent")
	// ErrNoPromptArgRequired is raised when the application is not running
	// interactively and thus requires an argument on the command-line instead
	// of prompting for input.
	ErrNoPromptArgRequired = errors.New("argument required when not running interactively")
)

// A FlagError is raised when flag processing fails.
type FlagError struct {
	err error
}

// NewFlagError returns a new *FlagError wrapping the given error.
func NewFlagError(err error) *FlagError {
	return &FlagError{err: err}
}

// NewFlagErrorf returns a new, formatted *FlagError from the given format and
// arguments.
func NewFlagErrorf(format string, a ...interface{}) *FlagError {
	return NewFlagError(fmt.Errorf(format, a...))
}

// Error implements the error interface.
func (e FlagError) Error() string {
	return e.err.Error()
}

// Unwrap implements unwrapping functionality.
func (e FlagError) Unwrap() error {
	return e.err
}
