package ollama

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

func TestChatStream(t *testing.T) {
	// Mock Ollama streaming response (newline-delimited JSON)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected path /api/chat, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")

		// Send multiple chunks
		chunks := []string{"Hello", " ", "world", "!"}
		for i, text := range chunks {
			resp := ChatResponse{
				Model:     "llama3.2:1b",
				CreatedAt: time.Now(),
				Message: ChatMessage{
					Role:    "assistant",
					Content: text,
				},
				Done: i == len(chunks)-1,
			}

			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "%s\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	client := New(
		WithBaseURL(server.URL),
		WithModel("llama3.2:1b"),
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
		w.Header().Set("Content-Type", "application/json")

		// Send first chunk
		resp := ChatResponse{
			Model: "llama3.2:1b",
			Message: ChatMessage{
				Role:    "assistant",
				Content: "Hello",
			},
			Done: false,
		}
		data, _ := json.Marshal(resp)
		fmt.Fprintf(w, "%s\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Simulate delay before next chunk
		time.Sleep(200 * time.Millisecond)

		// Try to send another chunk (should be cancelled)
		resp.Message.Content = " world"
		data, _ = json.Marshal(resp)
		fmt.Fprintf(w, "%s\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}))
	defer server.Close()

	client := New(
		WithBaseURL(server.URL),
		WithModel("llama3.2:1b"),
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
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	client := New(
		WithBaseURL(server.URL),
		WithModel("llama3.2:1b"),
	)

	messages := []core.Message{
		{Role: "user", Content: "Test"},
	}

	ctx := context.Background()
	_, err := client.ChatStream(ctx, messages)
	if err == nil {
		t.Fatal("Expected error from ChatStream")
	}

	if !strings.Contains(err.Error(), "status code: 500") {
		t.Errorf("Expected status code error, got: %v", err)
	}
}

func TestCompleteStream(t *testing.T) {
	// Mock Ollama streaming response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Send chunks
		chunks := []string{"The", " answer", " is", " 42"}
		for i, text := range chunks {
			resp := ChatResponse{
				Model: "llama3.2:1b",
				Message: ChatMessage{
					Role:    "assistant",
					Content: text,
				},
				Done: i == len(chunks)-1,
			}
			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "%s\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	client := New(
		WithBaseURL(server.URL),
		WithModel("llama3.2:1b"),
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
		w.Header().Set("Content-Type", "application/json")

		tokens := []string{"Hello", " ", "streaming", " ", "world"}
		for i, token := range tokens {
			resp := ChatResponse{
				Model: "llama3.2:1b",
				Message: ChatMessage{
					Role:    "assistant",
					Content: token,
				},
				Done: i == len(tokens)-1,
			}
			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "%s\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	client := New(
		WithBaseURL(server.URL),
		WithModel("llama3.2:1b"),
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

func TestStreamChunkMetadata(t *testing.T) {
	// Test that metadata is properly set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		resp := ChatResponse{
			Model: "llama3.2:1b",
			Message: ChatMessage{
				Role:    "assistant",
				Content: "Test",
			},
			Done: true,
		}
		data, _ := json.Marshal(resp)
		fmt.Fprintf(w, "%s\n", data)
	}))
	defer server.Close()

	client := New(
		WithBaseURL(server.URL),
		WithModel("llama3.2:1b"),
	)

	messages := []core.Message{
		{Role: "user", Content: "Test"},
	}

	ctx := context.Background()
	stream, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	var lastChunk core.StreamChunk
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Chunk error: %v", chunk.Error)
		}
		lastChunk = chunk
	}

	// Check metadata
	if lastChunk.Metadata == nil {
		t.Fatal("Expected metadata to be set")
	}

	if model, ok := lastChunk.Metadata["model"].(string); !ok || model != "llama3.2:1b" {
		t.Errorf("Expected model metadata 'llama3.2:1b', got: %v", lastChunk.Metadata["model"])
	}

	if done, ok := lastChunk.Metadata["done"].(bool); !ok || !done {
		t.Errorf("Expected done metadata to be true, got: %v", lastChunk.Metadata["done"])
	}
}
