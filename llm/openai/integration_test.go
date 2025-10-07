// Package openai_test contains integration tests for the OpenAI client.
// These tests make real API calls to OpenAI and require OPENAI_API_KEY.
//
// Run with: OPENAI_API_KEY=sk-xxx go test -v ./llm/openai/...
// Skip with: go test -short ./llm/openai/...
package openai_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yashrahurikar23/goagents/core"
	"github.com/yashrahurikar23/goagents/llm/openai"
)

// getAPIKey returns the OpenAI API key from environment or skips the test
func getAPIKey(t *testing.T) string {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	return apiKey
}

// TestOpenAI_Chat tests the Chat() method with real API
func TestOpenAI_Chat(t *testing.T) {
	apiKey := getAPIKey(t)

	client := openai.New(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-3.5-turbo"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []core.Message{
		core.SystemMessage("You are a helpful assistant."),
		core.UserMessage("Say 'Hello, integration test!' and nothing else."),
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if resp.Content == "" {
		t.Error("expected non-empty content")
	}

	// Response should contain our expected text
	if !strings.Contains(strings.ToLower(resp.Content), "hello") {
		t.Logf("Response: %q", resp.Content)
		t.Error("expected response to contain 'hello'")
	}

	t.Logf("✅ Chat response: %q", resp.Content)
}

// TestOpenAI_Complete tests the Complete() method with real API
func TestOpenAI_Complete(t *testing.T) {
	apiKey := getAPIKey(t)

	client := openai.New(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-3.5-turbo"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := "Complete this sentence: The capital of France is"
	resp, err := client.Complete(ctx, prompt)
	if err != nil {
		t.Fatalf("Complete() failed: %v", err)
	}

	if resp == "" {
		t.Fatal("expected non-empty response")
	}

	// Response should mention Paris
	if !strings.Contains(strings.ToLower(resp), "paris") {
		t.Logf("Response: %q", resp)
		t.Error("expected response to mention Paris")
	}

	t.Logf("✅ Complete response: %q", resp)
}

// TestOpenAI_CreateChatCompletion tests the full CreateChatCompletion with real API
func TestOpenAI_CreateChatCompletion(t *testing.T) {
	apiKey := getAPIKey(t)

	client := openai.New(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-3.5-turbo"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	maxTokens := 10
	temperature := 0.7
	req := openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: "What is 2+2? Answer with just the number.",
			},
		},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	content := resp.Choices[0].Message.Content
	contentStr, ok := content.(string)
	if !ok || contentStr == "" {
		t.Error("expected non-empty string content")
	}

	// Should contain "4"
	if !strings.Contains(contentStr, "4") {
		t.Logf("Response: %q", contentStr)
		t.Error("expected response to contain '4'")
	}

	t.Logf("✅ CreateChatCompletion response: %q", contentStr)
}

// TestOpenAI_Streaming tests streaming with real API
func TestOpenAI_Streaming(t *testing.T) {
	apiKey := getAPIKey(t)

	client := openai.New(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-3.5-turbo"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatMessage{
			{
				Role:    "user",
				Content: "Count from 1 to 5, one number per line.",
			},
		},
		Stream: true,
	}

	chunks := []string{}
	completed := false

	opts := openai.StreamOptions{
		OnChunk: func(chunk *openai.ChatCompletionStreamResponse) error {
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil && chunk.Choices[0].Delta.Content != nil {
				if content, ok := chunk.Choices[0].Delta.Content.(string); ok {
					t.Logf("📦 Chunk: %q", content)
					chunks = append(chunks, content)
				}
			}
			return nil
		},
		OnComplete: func() error {
			t.Log("✅ Stream completed")
			completed = true
			return nil
		},
		OnError: func(err error) {
			t.Errorf("❌ Stream error: %v", err)
		},
	}

	err := client.CreateChatCompletionStream(ctx, req, opts)
	if err != nil {
		t.Fatalf("CreateChatCompletionStream() failed: %v", err)
	}

	if !completed {
		t.Error("OnComplete callback was not called")
	}

	if len(chunks) == 0 {
		t.Error("expected to receive chunks")
	}

	// Combine all chunks
	fullResponse := strings.Join(chunks, "")
	t.Logf("✅ Full streaming response: %q", fullResponse)

	// Should contain numbers
	hasNumbers := false
	for i := 1; i <= 5; i++ {
		if strings.Contains(fullResponse, string(rune('0'+i))) {
			hasNumbers = true
			break
		}
	}

	if !hasNumbers {
		t.Error("expected response to contain numbers 1-5")
	}
}

// TestOpenAI_FunctionCalling tests function calling with real API
func TestOpenAI_FunctionCalling(t *testing.T) {
	apiKey := getAPIKey(t)

	client := openai.New(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-3.5-turbo"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Define a calculator tool
	req := openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatMessage{
			{
				Role:    "user",
				Content: "What is 25 multiplied by 4?",
			},
		},
		Tools: []openai.Tool{
			{
				Type: "function",
				Function: &openai.Function{
					Name:        "calculator",
					Description: "Performs basic arithmetic operations",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{
								"type":        "string",
								"description": "The operation to perform: add, subtract, multiply, divide",
								"enum":        []string{"add", "subtract", "multiply", "divide"},
							},
							"a": map[string]interface{}{
								"type":        "number",
								"description": "First number",
							},
							"b": map[string]interface{}{
								"type":        "number",
								"description": "Second number",
							},
						},
						"required": []string{"operation", "a", "b"},
					},
				},
			},
		},
		ToolChoice: "auto",
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion() with tools failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	choice := resp.Choices[0]

	// Check if function was called
	if len(choice.Message.ToolCalls) == 0 {
		t.Log("No tool calls in response (LLM might have answered directly)")
		t.Logf("Response: %q", choice.Message.Content)
		// Not necessarily an error - LLM might answer directly
		return
	}

	toolCall := choice.Message.ToolCalls[0]
	t.Logf("✅ Function called: %s", toolCall.Function.Name)
	t.Logf("✅ Arguments: %s", toolCall.Function.Arguments)

	// Verify it called calculator
	if toolCall.Function.Name != "calculator" {
		t.Errorf("expected function 'calculator', got %q", toolCall.Function.Name)
	}

	// Arguments should contain multiply operation
	if !strings.Contains(toolCall.Function.Arguments, "multiply") &&
		!strings.Contains(toolCall.Function.Arguments, "mul") {
		t.Logf("Arguments: %s", toolCall.Function.Arguments)
		t.Error("expected arguments to contain 'multiply' operation")
	}
}

// TestOpenAI_ErrorHandling tests error scenarios with real API
func TestOpenAI_ErrorHandling_InvalidModel(t *testing.T) {
	apiKey := getAPIKey(t)

	client := openai.New(
		openai.WithAPIKey(apiKey),
		openai.WithModel("invalid-model-that-does-not-exist"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messages := []core.Message{
		core.UserMessage("Test"),
	}

	_, err := client.Chat(ctx, messages)
	if err == nil {
		t.Error("expected error for invalid model, got nil")
	}

	t.Logf("✅ Got expected error: %v", err)
}

// TestOpenAI_ContextCancellation tests context cancellation
func TestOpenAI_ContextCancellation(t *testing.T) {
	apiKey := getAPIKey(t)

	client := openai.New(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-3.5-turbo"),
	)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	messages := []core.Message{
		core.UserMessage("Test"),
	}

	_, err := client.Chat(ctx, messages)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}

	// Should be a context cancellation error
	if !strings.Contains(err.Error(), "context") {
		t.Logf("Error: %v", err)
		t.Error("expected context cancellation error")
	}

	t.Logf("✅ Got expected context error: %v", err)
}

// TestOpenAI_MultipleModels tests different GPT models
func TestOpenAI_MultipleModels(t *testing.T) {
	apiKey := getAPIKey(t)

	models := []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4-turbo-preview",
	}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			client := openai.New(
				openai.WithAPIKey(apiKey),
				openai.WithModel(model),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			messages := []core.Message{
				core.UserMessage("Say 'OK' and nothing else."),
			}

			resp, err := client.Chat(ctx, messages)
			if err != nil {
				// GPT-4 might not be available for all accounts
				if strings.Contains(err.Error(), "does not exist") ||
					strings.Contains(err.Error(), "not found") {
					t.Skipf("Model %s not available: %v", model, err)
				}
				t.Fatalf("Chat() with %s failed: %v", model, err)
			}

			if resp.Content == "" {
				t.Errorf("expected non-empty response from %s", model)
			}

			t.Logf("✅ %s response: %q", model, resp.Content)
		})
	}
}

// Benchmark tests (will be skipped without API key)

func BenchmarkOpenAI_Chat(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		b.Skip("OPENAI_API_KEY not set")
	}

	client := openai.New(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-3.5-turbo"),
	)

	messages := []core.Message{
		core.UserMessage("Say 'benchmark' and nothing else."),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := client.Chat(ctx, messages)
		cancel()

		if err != nil {
			b.Fatalf("Chat() failed: %v", err)
		}
	}
}
