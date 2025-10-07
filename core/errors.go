package core

import "fmt"

// Error types for the GoAgent framework.

// ErrInvalidArgument indicates an argument was invalid.
type ErrInvalidArgument struct {
	Argument string
	Reason   string
}

func (e *ErrInvalidArgument) Error() string {
	return fmt.Sprintf("invalid argument %q: %s", e.Argument, e.Reason)
}

// ErrToolNotFound indicates a requested tool doesn't exist.
type ErrToolNotFound struct {
	ToolName string
}

func (e *ErrToolNotFound) Error() string {
	return fmt.Sprintf("tool not found: %s", e.ToolName)
}

// ErrToolExecution indicates a tool failed to execute.
type ErrToolExecution struct {
	ToolName string
	Err      error
}

func (e *ErrToolExecution) Error() string {
	return fmt.Sprintf("tool %q execution failed: %v", e.ToolName, e.Err)
}

func (e *ErrToolExecution) Unwrap() error {
	return e.Err
}

// ErrLLMFailure indicates an LLM request failed.
type ErrLLMFailure struct {
	Provider string
	Err      error
}

func (e *ErrLLMFailure) Error() string {
	return fmt.Sprintf("LLM %q request failed: %v", e.Provider, e.Err)
}

func (e *ErrLLMFailure) Unwrap() error {
	return e.Err
}

// ErrTimeout indicates an operation timed out.
type ErrTimeout struct {
	Operation string
}

func (e *ErrTimeout) Error() string {
	return fmt.Sprintf("operation timed out: %s", e.Operation)
}
