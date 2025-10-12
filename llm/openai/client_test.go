package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yashrahurikar23/goagents/core"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want *Client
	}{
		{
			name: "default config",
			opts: nil,
			want: &Client{
				baseURL:    DefaultBaseURL,
				model:      DefaultModel,
				timeout:    DefaultTimeout,
				maxRetries: DefaultMaxRetries,
			},
		},
		{
			name: "with custom options",
			opts: []Option{
				WithAPIKey("test-key"),
				WithModel("gpt-3.5-turbo"),
				WithBaseURL("https://custom.api"),
				WithMaxRetries(5),
			},
			want: &Client{
				apiKey:     "test-key",
				model:      "gpt-3.5-turbo",
				baseURL:    "https://custom.api",
				maxRetries: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.opts...)

			if got.apiKey != tt.want.apiKey {
				t.Errorf("apiKey = %v, want %v", got.apiKey, tt.want.apiKey)
			}
			if got.model != tt.want.model {
				t.Errorf("model = %v, want %v", got.model, tt.want.model)
			}
			if got.baseURL != tt.want.baseURL {
				t.Errorf("baseURL = %v, want %v", got.baseURL, tt.want.baseURL)
			}
			if got.maxRetries != tt.want.maxRetries {
				t.Errorf("maxRetries = %v, want %v", got.maxRetries, tt.want.maxRetries)
			}
		})
	}
}

func TestCreateChatCompletion(t *testing.T) {
	mockResponse := ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []Choice{
			{
				Index: 0,
				Message: &ChatMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you today?",
				},
				FinishReason: "stop",
			},
		},
		Usage: &Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("Expected Authorization: Bearer test-key, got %s", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	req := ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateChatCompletion failed: %v", err)
	}

	if resp.ID != mockResponse.ID {
		t.Errorf("ID = %v, want %v", resp.ID, mockResponse.ID)
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Content != mockResponse.Choices[0].Message.Content {
		t.Errorf("Content = %v, want %v",
			resp.Choices[0].Message.Content,
			mockResponse.Choices[0].Message.Content)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		errorResponse  ErrorResponse
		wantErrType    string
		wantStatusCode int
	}{
		{
			name:       "rate limit error",
			statusCode: http.StatusTooManyRequests,
			errorResponse: ErrorResponse{
				Error: &APIError{
					Type:    "rate_limit_exceeded",
					Message: "Rate limit exceeded",
				},
			},
			wantErrType:    "rate_limit_exceeded",
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name:       "invalid request error",
			statusCode: http.StatusBadRequest,
			errorResponse: ErrorResponse{
				Error: &APIError{
					Type:    "invalid_request_error",
					Message: "Invalid request",
				},
			},
			wantErrType:    "invalid_request_error",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.errorResponse)
			}))
			defer server.Close()

			client := New(
				WithAPIKey("test-key"),
				WithBaseURL(server.URL),
				WithMaxRetries(0), // Disable retries for this test
			)

			req := ChatCompletionRequest{
				Messages: []ChatMessage{{Role: "user", Content: "test"}},
			}

			_, err := client.CreateChatCompletion(context.Background(), req)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			oaiErr, ok := err.(*OpenAIError)
			if !ok {
				t.Fatalf("Expected *OpenAIError, got %T", err)
			}

			if oaiErr.StatusCode != tt.wantStatusCode {
				t.Errorf("StatusCode = %v, want %v", oaiErr.StatusCode, tt.wantStatusCode)
			}

			if oaiErr.Type != tt.wantErrType {
				t.Errorf("Type = %v, want %v", oaiErr.Type, tt.wantErrType)
			}
		})
	}
}

func TestRetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Fail the first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: &APIError{
					Type:    "server_error",
					Message: "Internal server error",
				},
			})
			return
		}

		// Succeed on 3rd attempt
		json.NewEncoder(w).Encode(ChatCompletionResponse{
			ID: "chatcmpl-123",
			Choices: []Choice{
				{Message: &ChatMessage{Role: "assistant", Content: "Success"}},
			},
		})
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
		WithMaxRetries(3),
	)

	req := ChatCompletionRequest{
		Messages: []ChatMessage{{Role: "user", Content: "test"}},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if resp.ID != "chatcmpl-123" {
		t.Errorf("ID = %v, want chatcmpl-123", resp.ID)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("SystemMessage", func(t *testing.T) {
		msg := SystemMessage("You are helpful")
		if msg.Role != "system" {
			t.Errorf("Role = %v, want system", msg.Role)
		}
		if msg.Content != "You are helpful" {
			t.Errorf("Content = %v, want 'You are helpful'", msg.Content)
		}
	})

	t.Run("UserMessage", func(t *testing.T) {
		msg := UserMessage("Hello")
		if msg.Role != "user" {
			t.Errorf("Role = %v, want user", msg.Role)
		}
	})

	t.Run("AssistantMessage", func(t *testing.T) {
		msg := AssistantMessage("Hi there")
		if msg.Role != "assistant" {
			t.Errorf("Role = %v, want assistant", msg.Role)
		}
	})

	t.Run("UserMessageWithImage", func(t *testing.T) {
		msg := UserMessageWithImage("Describe", "https://example.com/img.jpg", "high")
		if msg.Role != "user" {
			t.Errorf("Role = %v, want user", msg.Role)
		}

		parts, ok := msg.Content.([]ContentPart)
		if !ok {
			t.Fatalf("Expected Content to be []ContentPart")
		}

		if len(parts) != 2 {
			t.Fatalf("Expected 2 content parts, got %d", len(parts))
		}

		if parts[0].Type != "text" || parts[0].Text != "Describe" {
			t.Errorf("First part should be text 'Describe'")
		}

		if parts[1].Type != "image_url" {
			t.Errorf("Second part should be image_url")
		}

		if parts[1].ImageURL.Detail != "high" {
			t.Errorf("Detail = %v, want high", parts[1].ImageURL.Detail)
		}
	})
}

func TestJSONSchema(t *testing.T) {
	schema := JSONSchema(
		map[string]interface{}{
			"name": PropertyString("User name"),
			"age":  PropertyNumber("User age"),
		},
		[]string{"name"},
	)

	if schema["type"] != "object" {
		t.Errorf("type = %v, want object", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties should be map[string]interface{}")
	}

	if len(props) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(props))
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required should be []string")
	}

	if len(required) != 1 || required[0] != "name" {
		t.Errorf("required = %v, want [name]", required)
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "rate limit error",
			err: &OpenAIError{
				StatusCode: http.StatusTooManyRequests,
				Type:       "rate_limit_exceeded",
			},
			want: true,
		},
		{
			name: "other error",
			err: &OpenAIError{
				StatusCode: http.StatusBadRequest,
				Type:       "invalid_request_error",
			},
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRateLimitError(tt.err)
			if got != tt.want {
				t.Errorf("IsRateLimitError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChatStream(t *testing.T) {
	// Mock SSE streaming response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("Expected http.ResponseWriter to support flushing")
		}

		// Send multiple chunks
		chunks := []string{"Hello", " ", "world", "!"}
		for i, text := range chunks {
			chunk := ChatCompletionStreamResponse{
				ID:      "chatcmpl-123",
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "gpt-4",
				Choices: []Choice{
					{
						Index: 0,
						Delta: &ChatMessage{
							Role:    "assistant",
							Content: text,
						},
						FinishReason: "",
					},
				},
			}

			// Last chunk has finish_reason
			if i == len(chunks)-1 {
				chunk.Choices[0].FinishReason = "stop"
			}

			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		// Send [DONE] message
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	messages := []core.Message{
		{Role: "user", Content: "Hello"},
	}

	ctx := context.Background()
	stream, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	var chunks []core.StreamChunk
	var fullContent string

	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Chunk error: %v", chunk.Error)
		}
		chunks = append(chunks, chunk)
		fullContent += chunk.Delta
	}

	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}

	if fullContent != "Hello world!" {
		t.Errorf("Expected content 'Hello world!', got '%s'", fullContent)
	}

	// Check last chunk has finish reason
	lastChunk := chunks[len(chunks)-1]
	if lastChunk.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got '%s'", lastChunk.FinishReason)
	}

	// Check accumulated content
	if lastChunk.Content != "Hello world!" {
		t.Errorf("Expected accumulated content 'Hello world!', got '%s'", lastChunk.Content)
	}
}

func TestChatStreamContextCancellation(t *testing.T) {
	// Mock server with slow streaming
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		// Send first chunk
		chunk := ChatCompletionStreamResponse{
			ID:    "chatcmpl-123",
			Model: "gpt-4",
			Choices: []Choice{
				{Delta: &ChatMessage{Content: "Hello"}},
			},
		}
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		// Simulate delay before next chunk
		time.Sleep(200 * time.Millisecond)

		// Try to send another chunk (should be cancelled)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	messages := []core.Message{
		{Role: "user", Content: "Test"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	stream, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	var chunks []core.StreamChunk
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	// Should receive at least one chunk before cancellation
	if len(chunks) == 0 {
		t.Error("Expected at least one chunk before cancellation")
	}
}

func TestChatStreamError(t *testing.T) {
	// Mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: &APIError{
				Type:    "server_error",
				Message: "Internal server error",
			},
		})
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	messages := []core.Message{
		{Role: "user", Content: "Test"},
	}

	ctx := context.Background()
	stream, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	var errorReceived bool
	for chunk := range stream {
		if chunk.Error != nil {
			errorReceived = true
			if !strings.Contains(chunk.Error.Error(), "server_error") {
				t.Errorf("Expected error to contain 'server_error', got: %v", chunk.Error)
			}
		}
	}

	if !errorReceived {
		t.Error("Expected to receive error chunk")
	}
}

func TestCompleteStream(t *testing.T) {
	// Mock SSE streaming response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		// Send chunks
		chunks := []string{"The", " answer", " is", " 42"}
		for _, text := range chunks {
			chunk := ChatCompletionStreamResponse{
				ID:    "chatcmpl-123",
				Model: "gpt-4",
				Choices: []Choice{
					{Delta: &ChatMessage{Content: text}},
				},
			}
			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	ctx := context.Background()
	stream, err := client.CompleteStream(ctx, "What is the answer?")
	if err != nil {
		t.Fatalf("CompleteStream failed: %v", err)
	}

	var fullContent string
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Chunk error: %v", chunk.Error)
		}
		fullContent += chunk.Delta
	}

	if fullContent != "The answer is 42" {
		t.Errorf("Expected content 'The answer is 42', got '%s'", fullContent)
	}
}

func TestStreamChunkAccumulation(t *testing.T) {
	// Test that Content field accumulates properly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		tokens := []string{"Hello", " ", "streaming", " ", "world"}
		for _, token := range tokens {
			chunk := ChatCompletionStreamResponse{
				ID:    "chatcmpl-123",
				Model: "gpt-4",
				Choices: []Choice{
					{Delta: &ChatMessage{Content: token}},
				},
			}
			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	messages := []core.Message{
		{Role: "user", Content: "Test"},
	}

	ctx := context.Background()
	stream, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	expectedAccumulations := []string{
		"Hello",
		"Hello ",
		"Hello streaming",
		"Hello streaming ",
		"Hello streaming world",
	}

	var i int
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Chunk error: %v", chunk.Error)
		}

		if i < len(expectedAccumulations) {
			if chunk.Content != expectedAccumulations[i] {
				t.Errorf("Chunk %d: expected Content '%s', got '%s'",
					i, expectedAccumulations[i], chunk.Content)
			}
		}
		i++
	}

	if i != len(expectedAccumulations) {
		t.Errorf("Expected %d chunks, got %d", len(expectedAccumulations), i)
	}
}
