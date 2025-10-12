package anthropic

import (
"context"
"fmt"
"net/http"
"net/http/httptest"
"strings"
"testing"
"time"

"github.com/yashrahurikar23/goagents/core"
)

func TestChatStream(t *testing.T) {
	// Mock Anthropic streaming response (SSE format)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if r.Method != "POST" {
t.Errorf("Expected POST request, got %s", r.Method)
}

if r.Header.Get("x-api-key") != "test-key" {
t.Errorf("Expected API key header")
}

w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		// Send streaming events
		events := []string{
			"event: message_start\ndata: {\"type\": \"message_start\", \"message\": {\"id\": \"msg_test123\", \"type\": \"message\", \"role\": \"assistant\", \"model\": \"claude-3-5-sonnet-20241022\"}}\n\n",
			"event: content_block_start\ndata: {\"type\": \"content_block_start\", \"index\": 0, \"content_block\": {\"type\": \"text\", \"text\": \"\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"Hello\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \" \"}}\n\n",
			"event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"world!\"}}\n\n",
			"event: content_block_stop\ndata: {\"type\": \"content_block_stop\", \"index\": 0}\n\n",
			"event: message_delta\ndata: {\"type\": \"message_delta\", \"delta\": {\"stop_reason\": \"end_turn\", \"stop_sequence\": null}, \"usage\": {\"output_tokens\": 5}}\n\n",
			"event: message_stop\ndata: {\"type\": \"message_stop\"}\n\n",
		}

		for _, event := range events {
			fmt.Fprint(w, event)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(1 * time.Millisecond) // Small delay to simulate streaming
		}
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

		// Send first chunk
		event := "event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"Hello\"}}\n\n"
		fmt.Fprint(w, event)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Simulate delay before next chunk
		time.Sleep(200 * time.Millisecond)

		// Try to send another chunk (should be cancelled)
		event = "event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \" world\"}}\n\n"
		fmt.Fprint(w, event)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
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
w.WriteHeader(http.StatusBadRequest)
w.Write([]byte(`{"type": "error", "error": {"type": "invalid_request_error", "message": "Invalid request"}}`))
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
	_, err := client.ChatStream(ctx, messages)
	if err == nil {
		t.Fatal("Expected error from ChatStream")
	}

	if !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected status code error, got: %v", err)
	}
}

func TestCompleteStream(t *testing.T) {
	// Mock Anthropic streaming response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "text/event-stream")

		// Send chunks
		events := []string{
			"event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"The\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \" answer\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \" is\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \" 42\"}}\n\n",
			"event: message_stop\ndata: {\"type\": \"message_stop\"}\n\n",
		}

		for _, event := range events {
			fmt.Fprint(w, event)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(1 * time.Millisecond)
		}
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

		tokens := []string{"Hello", " ", "streaming", " ", "world"}
		for _, token := range tokens {
			event := fmt.Sprintf("event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"%s\"}}\n\n", token)
			fmt.Fprint(w, event)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(1 * time.Millisecond)
		}

		// End stream
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\": \"message_stop\"}\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
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
w.Header().Set("Content-Type", "text/event-stream")

		event := "event: content_block_delta\ndata: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"Test\"}}\n\n"
		fmt.Fprint(w, event)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// End stream
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\": \"message_stop\"}\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}))
	defer server.Close()

	client := New(
WithAPIKey("test-key"),
WithBaseURL(server.URL),
WithModel("claude-3-5-sonnet-20241022"),
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

	if model, ok := lastChunk.Metadata["model"].(string); !ok || model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model metadata 'claude-3-5-sonnet-20241022', got: %v", lastChunk.Metadata["model"])
	}
}
