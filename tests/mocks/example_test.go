// Package mocks_test demonstrates how to use the mock implementations.
package mocks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yashrahurikar/goagents/core"
	"github.com/yashrahurikar/goagents/tests/mocks"
	"github.com/yashrahurikar/goagents/tests/testutil"
)

// TestMockLLM_BasicUsage demonstrates basic mock LLM usage
func TestMockLLM_BasicUsage(t *testing.T) {
	// Create a mock LLM with default behavior
	mockLLM := mocks.NewMockLLM()

	ctx := context.Background()
	messages := []core.Message{
		core.UserMessage("Hello"),
	}

	// Call Chat - returns default "Mock response"
	resp, err := mockLLM.Chat(ctx, messages)

	testutil.AssertNoError(t, err)
	testutil.AssertNotNil(t, resp)
	testutil.AssertEqual(t, resp.Content, "Mock response")

	// Verify call was tracked
	calls := mockLLM.GetChatCalls()
	testutil.AssertEqual(t, len(calls), 1)
	testutil.AssertEqual(t, len(calls[0].Messages), 1)
	testutil.AssertEqual(t, calls[0].Messages[0].Content, "Hello")
}

// TestMockLLM_CustomResponse demonstrates custom response configuration
func TestMockLLM_CustomResponse(t *testing.T) {
	// Configure mock to return specific response
	mockLLM := mocks.NewMockLLM().WithChatResponse(
		"The answer is 42",
		nil, // No tool calls
	)

	ctx := context.Background()
	messages := []core.Message{
		core.UserMessage("What is the meaning of life?"),
	}

	resp, err := mockLLM.Chat(ctx, messages)

	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, resp.Content, "The answer is 42")
}

// TestMockLLM_ErrorHandling demonstrates error simulation
func TestMockLLM_ErrorHandling(t *testing.T) {
	// Configure mock to return an error
	expectedErr := errors.New("API rate limit exceeded")
	mockLLM := mocks.NewMockLLM().WithChatError(expectedErr)

	ctx := context.Background()
	messages := []core.Message{
		core.UserMessage("Hello"),
	}

	resp, err := mockLLM.Chat(ctx, messages)

	testutil.AssertError(t, err)
	if resp != nil {
		t.Errorf("expected nil response, got %v", resp)
	}
	testutil.AssertEqual(t, err.Error(), expectedErr.Error())
}

// TestMockLLM_SequentialResponses demonstrates multi-turn conversations
func TestMockLLM_SequentialResponses(t *testing.T) {
	// Configure mock to return different responses for each call
	mockLLM := mocks.NewMockLLM().WithSequentialChatResponses(
		[]*core.Response{
			{Content: "First response"},
			{Content: "Second response"},
		},
		[]error{nil, nil},
	)

	ctx := context.Background()
	messages := []core.Message{
		core.UserMessage("Hello"),
	}

	// First call
	resp1, err := mockLLM.Chat(ctx, messages)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, resp1.Content, "First response")

	// Second call
	resp2, err := mockLLM.Chat(ctx, messages)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, resp2.Content, "Second response")

	// Verify both calls were tracked
	testutil.AssertEqual(t, mockLLM.ChatCallCount(), 2)
}

// TestMockTool_BasicUsage demonstrates basic mock tool usage
func TestMockTool_BasicUsage(t *testing.T) {
	// Create a mock tool
	tool := mocks.NewMockTool("calculator", "Performs calculations")

	ctx := context.Background()
	args := map[string]interface{}{
		"operation": "add",
		"a":         5,
		"b":         3,
	}

	// Execute - returns default success result
	result, err := tool.Execute(ctx, args)

	testutil.AssertNoError(t, err)
	testutil.AssertNotNil(t, result)

	// Verify call was tracked
	calls := tool.GetCalls()
	testutil.AssertEqual(t, len(calls), 1)
	testutil.AssertEqual(t, calls[0].Args["a"], 5)
}

// TestMockTool_CustomResult demonstrates custom result configuration
func TestMockTool_CustomResult(t *testing.T) {
	// Configure mock to return specific result
	tool := mocks.NewMockTool("calculator", "Performs calculations").
		WithExecuteResult(8)

	ctx := context.Background()
	args := map[string]interface{}{
		"operation": "add",
		"a":         5,
		"b":         3,
	}

	result, err := tool.Execute(ctx, args)

	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, result, 8)
}

// TestMockHTTPServer_ChatCompletion demonstrates HTTP mock usage
func TestMockHTTPServer_ChatCompletion(t *testing.T) {
	// Create mock server that returns a chat completion
	server := mocks.NewMockHTTPServer(
		mocks.ChatCompletionResponse("Hello from mock!"),
	)
	defer server.Close()

	// Server is now available at server.URL()
	// You would configure your OpenAI client to use this URL
	testutil.AssertNotNil(t, server.URL())

	// After making requests, verify the server received them
	// (This example doesn't make actual requests)
	testutil.AssertEqual(t, server.RequestCount(), 0)
}

// TestMockHTTPServer_RetryLogic demonstrates testing retry behavior
func TestMockHTTPServer_RetryLogic(t *testing.T) {
	// Simulate: fail, fail, succeed
	server := mocks.NewMockHTTPServer(
		mocks.SequentialResponses(
			mocks.ServerErrorResponse(),        // First call: 500 error
			mocks.RateLimitResponse(),          // Second call: 429 rate limit
			mocks.ChatCompletionResponse("OK"), // Third call: success
		),
	)
	defer server.Close()

	// Your client would retry and eventually succeed
	// (Actual retry logic would be in the client being tested)
	testutil.AssertNotNil(t, server.URL())
}

// Example: Testing a hypothetical agent that uses LLM and Tool
func TestAgent_WithMocks(t *testing.T) {
	// This demonstrates how you'd test an agent using mocks

	// Set up mock LLM to return a tool call, then a final response
	mockLLM := mocks.NewMockLLM().WithSequentialChatResponses(
		[]*core.Response{
			{
				Content: "",
				ToolCalls: []core.ToolCall{
					{
						ID:   "call_123",
						Name: "calculator",
						Args: map[string]interface{}{
							"operation": "add",
							"a":         5,
							"b":         3,
						},
					},
				},
			},
			{
				Content: "The answer is 8",
			},
		},
		[]error{nil, nil},
	)

	// Set up mock tool to return calculation result
	mockTool := mocks.NewMockTool("calculator", "Performs calculations").
		WithExecuteResult(8)

	// In a real test, you'd create an agent with these mocks
	// agent := agent.NewFunctionAgent(mockLLM)
	// agent.AddTool(mockTool)
	// resp, err := agent.Run(ctx, "What is 5 + 3?")

	// Verify the mocks were used correctly
	testutil.AssertEqual(t, mockLLM.ChatCallCount(), 0) // Would be 2 after agent.Run()
	testutil.AssertEqual(t, mockTool.CallCount(), 0)    // Would be 1 after agent.Run()
}
