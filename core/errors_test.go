package core

import (
	"errors"
	"testing"
)

// TestErrInvalidArgument tests the ErrInvalidArgument error type
func TestErrInvalidArgument(t *testing.T) {
	err := &ErrInvalidArgument{
		Argument: "temperature",
		Reason:   "must be between 0 and 2",
	}

	expected := `invalid argument "temperature": must be between 0 and 2`
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

// TestErrInvalidArgument_EmptyFields tests error with empty fields
func TestErrInvalidArgument_EmptyFields(t *testing.T) {
	err := &ErrInvalidArgument{
		Argument: "",
		Reason:   "",
	}

	// Should still produce a valid error message
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

// TestErrToolNotFound tests the ErrToolNotFound error type
func TestErrToolNotFound(t *testing.T) {
	err := &ErrToolNotFound{
		ToolName: "calculator",
	}

	expected := "tool not found: calculator"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

// TestErrToolNotFound_EmptyName tests error with empty tool name
func TestErrToolNotFound_EmptyName(t *testing.T) {
	err := &ErrToolNotFound{
		ToolName: "",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}

	// Should still contain the prefix
	if msg != "tool not found: " {
		t.Errorf("expected 'tool not found: ', got %q", msg)
	}
}

// TestErrToolExecution tests the ErrToolExecution error type
func TestErrToolExecution(t *testing.T) {
	innerErr := errors.New("division by zero")
	err := &ErrToolExecution{
		ToolName: "calculator",
		Err:      innerErr,
	}

	expected := `tool "calculator" execution failed: division by zero`
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

// TestErrToolExecution_Unwrap tests error unwrapping
func TestErrToolExecution_Unwrap(t *testing.T) {
	innerErr := errors.New("connection timeout")
	err := &ErrToolExecution{
		ToolName: "http_client",
		Err:      innerErr,
	}

	// Test Unwrap returns the inner error
	unwrapped := err.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("expected unwrapped error to be %v, got %v", innerErr, unwrapped)
	}

	// Test errors.Is works with unwrapping
	if !errors.Is(err, innerErr) {
		t.Error("errors.Is should find the inner error")
	}
}

// TestErrToolExecution_NilInnerError tests with nil inner error
func TestErrToolExecution_NilInnerError(t *testing.T) {
	err := &ErrToolExecution{
		ToolName: "tool",
		Err:      nil,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}

	// Unwrap should return nil
	if err.Unwrap() != nil {
		t.Errorf("expected Unwrap to return nil, got %v", err.Unwrap())
	}
}

// TestErrLLMFailure tests the ErrLLMFailure error type
func TestErrLLMFailure(t *testing.T) {
	innerErr := errors.New("rate limit exceeded")
	err := &ErrLLMFailure{
		Provider: "openai",
		Err:      innerErr,
	}

	expected := `LLM "openai" request failed: rate limit exceeded`
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

// TestErrLLMFailure_Unwrap tests error unwrapping
func TestErrLLMFailure_Unwrap(t *testing.T) {
	innerErr := errors.New("API key invalid")
	err := &ErrLLMFailure{
		Provider: "anthropic",
		Err:      innerErr,
	}

	// Test Unwrap returns the inner error
	unwrapped := err.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("expected unwrapped error to be %v, got %v", innerErr, unwrapped)
	}

	// Test errors.Is works with unwrapping
	if !errors.Is(err, innerErr) {
		t.Error("errors.Is should find the inner error")
	}
}

// TestErrLLMFailure_NilInnerError tests with nil inner error
func TestErrLLMFailure_NilInnerError(t *testing.T) {
	err := &ErrLLMFailure{
		Provider: "provider",
		Err:      nil,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}

	// Unwrap should return nil
	if err.Unwrap() != nil {
		t.Errorf("expected Unwrap to return nil, got %v", err.Unwrap())
	}
}

// TestErrTimeout tests the ErrTimeout error type
func TestErrTimeout(t *testing.T) {
	err := &ErrTimeout{
		Operation: "chat completion",
	}

	expected := "operation timed out: chat completion"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

// TestErrTimeout_EmptyOperation tests error with empty operation
func TestErrTimeout_EmptyOperation(t *testing.T) {
	err := &ErrTimeout{
		Operation: "",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}

	if msg != "operation timed out: " {
		t.Errorf("expected 'operation timed out: ', got %q", msg)
	}
}

// TestErrorTypes_AsInterface tests that errors implement error interface
func TestErrorTypes_AsInterface(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "ErrInvalidArgument",
			err:  &ErrInvalidArgument{Argument: "test", Reason: "test"},
		},
		{
			name: "ErrToolNotFound",
			err:  &ErrToolNotFound{ToolName: "test"},
		},
		{
			name: "ErrToolExecution",
			err:  &ErrToolExecution{ToolName: "test", Err: errors.New("test")},
		},
		{
			name: "ErrLLMFailure",
			err:  &ErrLLMFailure{Provider: "test", Err: errors.New("test")},
		},
		{
			name: "ErrTimeout",
			err:  &ErrTimeout{Operation: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each error should implement error interface
			msg := tt.err.Error()
			if msg == "" {
				t.Errorf("%s.Error() returned empty string", tt.name)
			}
		})
	}
}

// TestErrorWrapping tests error wrapping with errors.Is and errors.As
func TestErrorWrapping(t *testing.T) {
	// Create a sentinel error
	sentinelErr := errors.New("sentinel error")

	// Wrap it in ErrToolExecution
	toolErr := &ErrToolExecution{
		ToolName: "test_tool",
		Err:      sentinelErr,
	}

	// errors.Is should find the sentinel error
	if !errors.Is(toolErr, sentinelErr) {
		t.Error("errors.Is failed to find wrapped sentinel error")
	}

	// errors.As should extract ErrToolExecution
	var extractedErr *ErrToolExecution
	if !errors.As(toolErr, &extractedErr) {
		t.Error("errors.As failed to extract ErrToolExecution")
	}

	if extractedErr.ToolName != "test_tool" {
		t.Errorf("expected ToolName 'test_tool', got %q", extractedErr.ToolName)
	}
}

// TestErrorWrapping_LLMFailure tests error wrapping for LLM failures
func TestErrorWrapping_LLMFailure(t *testing.T) {
	// Create a chain of wrapped errors
	baseErr := errors.New("network error")
	llmErr := &ErrLLMFailure{
		Provider: "openai",
		Err:      baseErr,
	}

	// Should be able to find the base error
	if !errors.Is(llmErr, baseErr) {
		t.Error("errors.Is failed to find base error")
	}

	// Should be able to extract the LLM error
	var extractedErr *ErrLLMFailure
	if !errors.As(llmErr, &extractedErr) {
		t.Error("errors.As failed to extract ErrLLMFailure")
	}

	if extractedErr.Provider != "openai" {
		t.Errorf("expected Provider 'openai', got %q", extractedErr.Provider)
	}
}

// TestErrorTypes_DifferentProviders tests different provider names
func TestErrorTypes_DifferentProviders(t *testing.T) {
	providers := []string{"openai", "anthropic", "ollama", "custom"}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			err := &ErrLLMFailure{
				Provider: provider,
				Err:      errors.New("test error"),
			}

			msg := err.Error()
			// Message should contain the provider name
			if msg == "" {
				t.Error("expected non-empty error message")
			}

			// Use errors.As to verify the error is correct type
			var llmErr *ErrLLMFailure
			if !errors.As(err, &llmErr) {
				t.Error("errors.As failed")
			}

			if llmErr.Provider != provider {
				t.Errorf("expected provider %q, got %q", provider, llmErr.Provider)
			}
		})
	}
}

// TestErrorTypes_DifferentOperations tests different timeout operations
func TestErrorTypes_DifferentOperations(t *testing.T) {
	operations := []string{
		"chat completion",
		"embedding generation",
		"tool execution",
		"agent run",
	}

	for _, op := range operations {
		t.Run(op, func(t *testing.T) {
			err := &ErrTimeout{Operation: op}

			msg := err.Error()
			if msg == "" {
				t.Error("expected non-empty error message")
			}

			// Use errors.As to verify the error is correct type
			var timeoutErr *ErrTimeout
			if !errors.As(err, &timeoutErr) {
				t.Error("errors.As failed")
			}

			if timeoutErr.Operation != op {
				t.Errorf("expected operation %q, got %q", op, timeoutErr.Operation)
			}
		})
	}
}

// Benchmark tests

func BenchmarkErrInvalidArgument(b *testing.B) {
	err := &ErrInvalidArgument{
		Argument: "temperature",
		Reason:   "must be between 0 and 2",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkErrToolNotFound(b *testing.B) {
	err := &ErrToolNotFound{ToolName: "calculator"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkErrToolExecution(b *testing.B) {
	err := &ErrToolExecution{
		ToolName: "calculator",
		Err:      errors.New("division by zero"),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkErrLLMFailure(b *testing.B) {
	err := &ErrLLMFailure{
		Provider: "openai",
		Err:      errors.New("rate limit"),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkErrTimeout(b *testing.B) {
	err := &ErrTimeout{Operation: "chat completion"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
