// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/yashrahurikar23/goagents/core"
)

// MockTool is a mock implementation of core.Tool for testing.
//
// WHY THIS EXISTS:
// Real tools may have side effects (HTTP calls, file operations, etc.)
// that we don't want in unit tests. Mocks provide predictable behavior
// without external dependencies.
//
// DESIGN DECISIONS:
// - Function field for Execute allows per-test customization
// - Call tracking enables assertion of tool usage
// - Error injection simulates tool failures
// - Thread-safe for concurrent test execution
type MockTool struct {
	// NameValue is returned by Name()
	NameValue string

	// DescriptionValue is returned by Description()
	DescriptionValue string

	// SchemaValue is returned by Schema()
	SchemaValue *core.ToolSchema

	// ExecuteFunc is called when Execute() is invoked.
	// If nil, returns a default successful result.
	ExecuteFunc func(ctx context.Context, args map[string]interface{}) (interface{}, error)

	// CallHistory tracks all executions
	mu    sync.RWMutex
	calls []ExecuteCall
}

// ExecuteCall records a single Execute() invocation
type ExecuteCall struct {
	Args   map[string]interface{}
	Result interface{}
	Error  error
}

// NewMockTool creates a new mock tool with sensible defaults.
//
// WHY THIS WAY:
// Default values make simple tests easier to write.
// Tests can override specific fields as needed.
func NewMockTool(name, description string) *MockTool {
	return &MockTool{
		NameValue:        name,
		DescriptionValue: description,
		SchemaValue: &core.ToolSchema{
			Name:        name,
			Description: description,
			Parameters:  []core.Parameter{},
		},
	}
}

// Name implements core.Tool.Name()
func (m *MockTool) Name() string {
	return m.NameValue
}

// Description implements core.Tool.Description()
func (m *MockTool) Description() string {
	return m.DescriptionValue
}

// Schema implements core.Tool.Schema()
func (m *MockTool) Schema() *core.ToolSchema {
	return m.SchemaValue
}

// Execute implements core.Tool.Execute()
//
// WHY THIS WAY:
// - If ExecuteFunc is set, delegates to custom test behavior
// - Otherwise returns a generic success result
// - Records call for later assertion
func (m *MockTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result interface{}
	var err error

	if m.ExecuteFunc != nil {
		result, err = m.ExecuteFunc(ctx, args)
	} else {
		// Default behavior: return success with args echoed
		result = map[string]interface{}{
			"status": "success",
			"args":   args,
		}
	}

	// Record the call
	m.calls = append(m.calls, ExecuteCall{
		Args:   args,
		Result: result,
		Error:  err,
	})

	return result, err
}

// GetCalls returns all recorded Execute() calls.
// Useful for asserting the tool was called correctly.
func (m *MockTool) GetCalls() []ExecuteCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent race conditions
	calls := make([]ExecuteCall, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// CallCount returns the number of times Execute() was called.
func (m *MockTool) CallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.calls)
}

// Reset clears all recorded calls.
func (m *MockTool) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
}

// WithExecuteResult configures the mock to return a specific result.
func (m *MockTool) WithExecuteResult(result interface{}) *MockTool {
	m.ExecuteFunc = func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return result, nil
	}
	return m
}

// WithExecuteError configures the mock to return an error.
func (m *MockTool) WithExecuteError(err error) *MockTool {
	m.ExecuteFunc = func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return nil, err
	}
	return m
}

// WithSchema sets a custom schema.
// Useful for testing schema validation.
func (m *MockTool) WithSchema(schema *core.ToolSchema) *MockTool {
	m.SchemaValue = schema
	return m
}

// WithSequentialResults configures the mock to return different results
// for each call. Useful for testing retry logic or multiple tool invocations.
func (m *MockTool) WithSequentialResults(results []interface{}, errors []error) *MockTool {
	callCount := 0
	m.ExecuteFunc = func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		if callCount >= len(results) {
			return nil, fmt.Errorf("mock: no more results configured (call %d)", callCount+1)
		}

		result := results[callCount]
		var err error
		if callCount < len(errors) {
			err = errors[callCount]
		}
		callCount++

		return result, err
	}
	return m
}

// Example usage in tests:
//
// Basic usage:
//   tool := mocks.NewMockTool("calculator", "Performs calculations")
//   result, err := tool.Execute(ctx, map[string]interface{}{"a": 5, "b": 3})
//   // result is a success map with the args
//
// Custom result:
//   tool := mocks.NewMockTool("calculator", "Performs calculations").
//       WithExecuteResult(8)
//   result, err := tool.Execute(ctx, map[string]interface{}{"a": 5, "b": 3})
//   // result == 8
//
// Error simulation:
//   tool := mocks.NewMockTool("calculator", "Performs calculations").
//       WithExecuteError(errors.New("division by zero"))
//   result, err := tool.Execute(ctx, args)
//   // err != nil
//
// Call tracking:
//   tool := mocks.NewMockTool("calculator", "Performs calculations")
//   tool.Execute(ctx, map[string]interface{}{"a": 5})
//   calls := tool.GetCalls()
//   // calls[0].Args["a"] == 5
//
// Sequential results:
//   tool := mocks.NewMockTool("api", "API client").WithSequentialResults(
//       []interface{}{"first", "second"},
//       []error{nil, nil},
//   )
//   result1, _ := tool.Execute(ctx, nil)  // Returns "first"
//   result2, _ := tool.Execute(ctx, nil)  // Returns "second"
