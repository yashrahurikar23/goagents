package ollama

import (
	"context"
	"testing"
	"time"

	"github.com/yashrahurikar23/goagents/core"
)

// TestOllamaIntegration tests the Ollama client with a real server
// Run with: go test -v ./llm/ollama/... -tags=integration
func TestOllamaIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create client with a small model
	client := New(
		WithModel("gemma3:270m"),
		WithTemperature(0.7),
	)

	ctx := context.Background()

	t.Run("Complete", func(t *testing.T) {
		response, err := client.Complete(ctx, "Say 'Hello' in one word")
		if err != nil {
			t.Fatalf("Complete failed: %v", err)
		}

		if response == "" {
			t.Error("Expected non-empty response")
		}

		t.Logf("Complete response: %s", response)
	})

	t.Run("Chat", func(t *testing.T) {
		messages := []core.Message{
			core.UserMessage("What is 2+2? Answer with just the number."),
		}

		response, err := client.Chat(ctx, messages)
		if err != nil {
			t.Fatalf("Chat failed: %v", err)
		}

		if response.Content == "" {
			t.Error("Expected non-empty content")
		}

		// Check metadata
		if response.Meta == nil {
			t.Error("Expected meta to be set")
		}
		if model, ok := response.Meta["model"].(string); !ok || model == "" {
			t.Error("Expected model in meta")
		}

		t.Logf("Chat response: %s", response.Content)
		t.Logf("Meta: %+v", response.Meta)
	})

	t.Run("ChatWithHistory", func(t *testing.T) {
		messages := []core.Message{
			core.UserMessage("My name is Alice"),
			core.AssistantMessage("Hello Alice! Nice to meet you."),
			core.UserMessage("What's my name?"),
		}

		response, err := client.Chat(ctx, messages)
		if err != nil {
			t.Fatalf("Chat failed: %v", err)
		}

		if response.Content == "" {
			t.Error("Expected non-empty content")
		}

		t.Logf("Chat with history response: %s", response.Content)
	})

	t.Run("Stream", func(t *testing.T) {
		messages := []core.Message{
			core.UserMessage("Count from 1 to 3"),
		}

		chunks, err := client.Stream(ctx, messages)
		if err != nil {
			t.Fatalf("Stream failed: %v", err)
		}

		var fullResponse string
		chunkCount := 0

		for chunk := range chunks {
			if chunk.Error != nil {
				t.Fatalf("Stream chunk error: %v", chunk.Error)
			}

			fullResponse += chunk.Content
			chunkCount++

			if chunk.Done {
				break
			}
		}

		if fullResponse == "" {
			t.Error("Expected non-empty streaming response")
		}

		if chunkCount == 0 {
			t.Error("Expected at least one chunk")
		}

		t.Logf("Stream response (%d chunks): %s", chunkCount, fullResponse)
	})

	t.Run("ListModels", func(t *testing.T) {
		models, err := client.ListModels(ctx)
		if err != nil {
			t.Fatalf("ListModels failed: %v", err)
		}

		if len(models.Models) == 0 {
			t.Error("Expected at least one model")
		}

		t.Logf("Found %d models:", len(models.Models))
		for _, model := range models.Models {
			t.Logf("  - %s (%s, %s)", model.Name, model.Details.Family, model.Details.ParameterSize)
		}
	})

	t.Run("WithSystemPrompt", func(t *testing.T) {
		messages := []core.Message{
			core.SystemMessage("You are a helpful assistant that only responds with 'Yes' or 'No'."),
			core.UserMessage("Is the sky blue?"),
		}

		response, err := client.Chat(ctx, messages)
		if err != nil {
			t.Fatalf("Chat failed: %v", err)
		}

		t.Logf("Response with system prompt: %s", response.Content)
	})

	t.Run("WithOptions", func(t *testing.T) {
		customClient := New(
			WithModel("gemma3:270m"),
			WithTemperature(0.1), // Low temperature for deterministic
			WithMaxTokens(10),
		)

		response, err := customClient.Complete(ctx, "Once upon a time")
		if err != nil {
			t.Fatalf("Complete failed: %v", err)
		}

		t.Logf("Response with custom options: %s", response)
	})
}

// TestOllamaMultipleModels tests with different models
func TestOllamaMultipleModels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	models := []string{"gemma3:270m", "llama3.2:1b"}

	for _, modelName := range models {
		t.Run(modelName, func(t *testing.T) {
			client := New(WithModel(modelName))

			response, err := client.Complete(ctx, "Say hello")
			if err != nil {
				t.Fatalf("Complete with %s failed: %v", modelName, err)
			}

			if response == "" {
				t.Error("Expected non-empty response")
			}

			t.Logf("%s response: %s", modelName, response)
		})
	}
}

// TestOllamaPerformance tests basic performance characteristics
func TestOllamaPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	client := New(WithModel("gemma3:270m"))
	ctx := context.Background()

	start := time.Now()
	_, err := client.Complete(ctx, "Say hi")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	t.Logf("Response time: %v", duration)

	// Should be reasonably fast (within 10 seconds for small model)
	if duration > 10*time.Second {
		t.Logf("Warning: Response took longer than expected: %v", duration)
	}
}

// TestOllamaErrorHandling tests error scenarios
func TestOllamaErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("InvalidModel", func(t *testing.T) {
		client := New(WithModel("invalid-model-that-does-not-exist"))
		ctx := context.Background()

		_, err := client.Complete(ctx, "test")
		if err == nil {
			t.Error("Expected error for invalid model")
		}

		t.Logf("Expected error: %v", err)
	})

	t.Run("EmptyMessage", func(t *testing.T) {
		client := New(WithModel("gemma3:270m"))
		ctx := context.Background()

		messages := []core.Message{
			core.UserMessage(""),
		}

		response, err := client.Chat(ctx, messages)
		// Some models may handle empty input gracefully
		if err != nil {
			t.Logf("Error with empty message: %v", err)
		} else {
			t.Logf("Response to empty message: %s", response.Content)
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		client := New(WithModel("gemma3:270m"))

		// Create a context that's immediately cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.Complete(ctx, "test")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}

		t.Logf("Expected cancellation error: %v", err)
	})
}
