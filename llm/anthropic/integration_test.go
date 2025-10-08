package anthropic

import (
	"context"
	"os"
	"testing"

	"github.com/yashrahurikar23/goagents/core"
)

// TestIntegration_Chat tests actual API calls (requires ANTHROPIC_API_KEY)
func TestIntegration_Chat(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelClaude35Haiku), // Use faster/cheaper model for tests
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
	t.Logf("Tokens: %d input, %d output",
		resp.Meta["input_tokens"],
		resp.Meta["output_tokens"])
}

// TestIntegration_Complete tests the Complete method
func TestIntegration_Complete(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelClaude35Haiku),
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

// TestIntegration_WithSystemPrompt tests with system message
func TestIntegration_WithSystemPrompt(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelClaude35Haiku),
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
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelClaude35Haiku),
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
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel(ModelClaude35Haiku),
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
