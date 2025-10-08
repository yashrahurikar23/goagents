package gemini

import (
	"context"
	"os"
	"testing"

	"github.com/yashrahurikar23/goagents/core"
)

// TestIntegration_Chat tests actual API calls (requires GEMINI_API_KEY)
func TestIntegration_Chat(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelGemini15Flash), // Fast and free
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "Say 'Hello, World!' and nothing else."},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Expected non-empty response content")
	}

	t.Logf("Response: %s", resp.Content)
	t.Logf("Tokens: %d prompt, %d completion",
		resp.Meta["prompt_tokens"],
		resp.Meta["completion_tokens"])
}

// TestIntegration_Complete tests the Complete method
func TestIntegration_Complete(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelGemini15Flash),
	)

	ctx := context.Background()
	result, err := client.Complete(ctx, "What is 2+2? Answer with just the number.")
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	t.Logf("Result: %s", result)
}

// TestIntegration_WithSystemInstruction tests with system message
func TestIntegration_WithSystemInstruction(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelGemini15Flash),
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "system", Content: "You are a pirate. Respond in pirate speak."},
		{Role: "user", Content: "Hello! How are you?"},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	t.Logf("Pirate response: %s", resp.Content)
}

// TestIntegration_MultiTurn tests multi-turn conversation
func TestIntegration_MultiTurn(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelGemini15Flash),
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "My name is Alice."},
		{Role: "assistant", Content: "Nice to meet you, Alice!"},
		{Role: "user", Content: "What is my name?"},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	t.Logf("Response: %s", resp.Content)
}

// TestIntegration_Temperature tests with temperature setting
func TestIntegration_Temperature(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelGemini15Flash),
		WithTemperature(0.1), // Low temperature for more deterministic output
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "What is the capital of France?"},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	t.Logf("Response: %s", resp.Content)
}

// TestIntegration_MaxTokens tests with max tokens limit
func TestIntegration_MaxTokens(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelGemini15Flash),
		WithMaxTokens(20), // Very short response
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "Write a long story about a robot."},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	t.Logf("Response (truncated): %s", resp.Content)
	t.Logf("Finish reason: %v", resp.Meta["finish_reason"])
}
