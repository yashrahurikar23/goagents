// Package mocks provides mock implementations for testing.
// This package is NOT part of the production goagent package.
// It is only used during development for testing purposes.
package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/yashrahurikar/goagents/core"
)

// MockLLM is a mock implementation of core.LLM for testing.
//
// WHY THIS EXISTS:
// Real LLM calls are expensive, slow, and non-deterministic.
// Tests need predictable, fast, offline behavior to verify logic
// without external dependencies.
//
// DESIGN DECISIONS:
// - Function fields instead of interface methods allow per-test customization
// - Call tracking enables assertion of LLM usage patterns
// - Error injection simulates failure scenarios
// - Thread-safe for concurrent test execution
type MockLLM struct {
	// ChatFunc is called when Chat() is invoked.
	// If nil, returns a default successful response.
	ChatFunc func(ctx context.Context, messages []core.Message) (*core.Response, error)

	// CompleteFunc is called when Complete() is invoked.
	// If nil, returns the prompt as the response.
	CompleteFunc func(ctx context.Context, prompt string) (string, error)

	// CallHistory tracks all calls made to this mock
	mu            sync.RWMutex
	chatCalls     []ChatCall
	completeCalls []CompleteCall
}

// ChatCall records a single Chat() invocation
type ChatCall struct {
	Messages []core.Message
	Response *core.Response
	Error    error
}

// CompleteCall records a single Complete() invocation
type CompleteCall struct {
	Prompt   string
	Response string
	Error    error
}

// NewMockLLM creates a new mock LLM with default behavior.
//
// Default behavior:
// - Chat() returns "Mock response" with no tool calls
// - Complete() echoes the prompt back
func NewMockLLM() *MockLLM {
	return &MockLLM{}
}

// Chat implements core.LLM.Chat().
//
// WHY THIS WAY:
// - If ChatFunc is set, delegates to custom test behavior
// - Otherwise returns sensible default for basic tests
// - Records call for later assertion
func (m *MockLLM) Chat(ctx context.Context, messages []core.Message) (*core.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var resp *core.Response
	var err error

	if m.ChatFunc != nil {
		resp, err = m.ChatFunc(ctx, messages)
	} else {
		// Default behavior: return simple response
		resp = &core.Response{
			Content: "Mock response",
		}
	}

	// Record the call
	m.chatCalls = append(m.chatCalls, ChatCall{
		Messages: messages,
		Response: resp,
		Error:    err,
	})

	return resp, err
}

// Complete implements core.LLM.Complete().
//
// WHY THIS WAY:
// - If CompleteFunc is set, delegates to custom test behavior
// - Otherwise echoes prompt for predictable testing
// - Records call for later assertion
func (m *MockLLM) Complete(ctx context.Context, prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var resp string
	var err error

	if m.CompleteFunc != nil {
		resp, err = m.CompleteFunc(ctx, prompt)
	} else {
		// Default behavior: echo prompt
		resp = fmt.Sprintf("Mock completion: %s", prompt)
	}

	// Record the call
	m.completeCalls = append(m.completeCalls, CompleteCall{
		Prompt:   prompt,
		Response: resp,
		Error:    err,
	})

	return resp, err
}

// GetChatCalls returns all recorded Chat() calls.
// Useful for asserting the LLM was called correctly.
func (m *MockLLM) GetChatCalls() []ChatCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent race conditions
	calls := make([]ChatCall, len(m.chatCalls))
	copy(calls, m.chatCalls)
	return calls
}

// GetCompleteCalls returns all recorded Complete() calls.
// Useful for asserting the LLM was called correctly.
func (m *MockLLM) GetCompleteCalls() []CompleteCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent race conditions
	calls := make([]CompleteCall, len(m.completeCalls))
	copy(calls, m.completeCalls)
	return calls
}

// ChatCallCount returns the number of times Chat() was called.
func (m *MockLLM) ChatCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.chatCalls)
}

// CompleteCallCount returns the number of times Complete() was called.
func (m *MockLLM) CompleteCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.completeCalls)
}

// Reset clears all recorded calls.
// Useful when reusing a mock across multiple test cases.
func (m *MockLLM) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.chatCalls = nil
	m.completeCalls = nil
}

// WithChatResponse configures the mock to return a specific response.
// Useful for simple tests that need predictable responses.
func (m *MockLLM) WithChatResponse(content string, toolCalls []core.ToolCall) *MockLLM {
	m.ChatFunc = func(ctx context.Context, messages []core.Message) (*core.Response, error) {
		return &core.Response{
			Content:   content,
			ToolCalls: toolCalls,
		}, nil
	}
	return m
}

// WithChatError configures the mock to return an error.
// Useful for testing error handling.
func (m *MockLLM) WithChatError(err error) *MockLLM {
	m.ChatFunc = func(ctx context.Context, messages []core.Message) (*core.Response, error) {
		return nil, err
	}
	return m
}

// WithCompleteResponse configures the mock to return a specific completion.
func (m *MockLLM) WithCompleteResponse(response string) *MockLLM {
	m.CompleteFunc = func(ctx context.Context, prompt string) (string, error) {
		return response, nil
	}
	return m
}

// WithCompleteError configures the mock to return an error.
func (m *MockLLM) WithCompleteError(err error) *MockLLM {
	m.CompleteFunc = func(ctx context.Context, prompt string) (string, error) {
		return "", err
	}
	return m
}

// WithSequentialChatResponses configures the mock to return different responses
// for each call. Useful for testing multi-turn conversations or retry logic.
//
// WHY THIS WAY:
// Agents often make multiple LLM calls (e.g., initial query, then tool results).
// Tests need to simulate different responses for each call.
func (m *MockLLM) WithSequentialChatResponses(responses []*core.Response, errors []error) *MockLLM {
	callCount := 0
	m.ChatFunc = func(ctx context.Context, messages []core.Message) (*core.Response, error) {
		if callCount >= len(responses) {
			return nil, fmt.Errorf("mock: no more responses configured (call %d)", callCount+1)
		}

		resp := responses[callCount]
		var err error
		if callCount < len(errors) {
			err = errors[callCount]
		}
		callCount++

		return resp, err
	}
	return m
}

// Example usage in tests:
//
// Basic usage:
//   mock := mocks.NewMockLLM()
//   resp, err := mock.Chat(ctx, messages)
//   // resp.Content == "Mock response"
//
// Custom response:
//   mock := mocks.NewMockLLM().WithChatResponse("Custom answer", nil)
//   resp, err := mock.Chat(ctx, messages)
//   // resp.Content == "Custom answer"
//
// Error simulation:
//   mock := mocks.NewMockLLM().WithChatError(errors.New("API error"))
//   resp, err := mock.Chat(ctx, messages)
//   // err != nil
//
// Call tracking:
//   mock := mocks.NewMockLLM()
//   mock.Chat(ctx, messages)
//   calls := mock.GetChatCalls()
//   // calls[0].Messages contains the messages passed
//
// Sequential responses (multi-turn):
//   mock := mocks.NewMockLLM().WithSequentialChatResponses(
//       []*core.Response{
//           {Content: "First call", ToolCalls: []core.ToolCall{{...}}},
//           {Content: "Second call", ToolCalls: nil},
//       },
//       []error{nil, nil},
//   )
//   resp1, _ := mock.Chat(ctx, messages)  // Returns "First call" with tool calls
//   resp2, _ := mock.Chat(ctx, messages)  // Returns "Second call" without tool calls
