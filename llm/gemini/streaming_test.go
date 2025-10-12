package gemini

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

// TestChatStream verifies basic streaming functionality with real-time token delivery
func TestChatStream(t *testing.T) {
	// Create mock server that streams responses as newline-delimited JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if !strings.Contains(r.URL.Path, "streamGenerateContent") {
			t.Errorf("Expected streamGenerateContent endpoint, got %s", r.URL.Path)
		}

		// Stream multiple chunks as newline-delimited JSON
		responses := []GenerateContentResponse{
			{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: "Hello"},
							},
							Role: "model",
						},
					},
				},
			},
			{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: " world"},
							},
							Role: "model",
						},
					},
				},
			},
			{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: "!"},
							},
							Role: "model",
						},
						FinishReason: "STOP",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter doesn't support flushing")
		}

		for _, resp := range responses {
			data, err := json.Marshal(resp)
			if err != nil {
				t.Fatalf("Failed to marshal response: %v", err)
			}
			fmt.Fprintf(w, "%s\n", data)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	// Create client pointing to mock server
	client := New(WithAPIKey("test-key"), WithBaseURL(server.URL), WithModel("gemini-pro"))

	// Call ChatStream
	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "Say hello"},
	}

	chunkChan, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	// Collect chunks
	var chunks []core.StreamChunk
	for chunk := range chunkChan {
		chunks = append(chunks, chunk)
		if chunk.Error != nil {
			t.Fatalf("Received error chunk: %v", chunk.Error)
		}
	}

	// Verify we received all chunks
	if len(chunks) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(chunks))
	}

	// Verify content accumulation
	expectedDeltas := []string{"Hello", " world", "!"}
	for i, chunk := range chunks {
		if chunk.Delta != expectedDeltas[i] {
			t.Errorf("Chunk %d: expected delta %q, got %q", i, expectedDeltas[i], chunk.Delta)
		}
	}

	// Verify final accumulated content
	finalChunk := chunks[len(chunks)-1]
	if finalChunk.Content != "Hello world!" {
		t.Errorf("Expected final content 'Hello world!', got %q", finalChunk.Content)
	}

	// Verify finish reason
	if finalChunk.FinishReason != "stop" {
		t.Errorf("Expected finish reason 'stop', got %q", finalChunk.FinishReason)
	}

	// Verify metadata
	if finalChunk.Metadata == nil {
		t.Error("Expected metadata to be set")
	} else if model, ok := finalChunk.Metadata["model"].(string); !ok || model != "gemini-pro" {
		t.Errorf("Expected metadata model 'gemini-pro', got %v", finalChunk.Metadata["model"])
	}
}

// TestChatStreamContextCancellation verifies that cancelling the context stops streaming
func TestChatStreamContextCancellation(t *testing.T) {
	// Create mock server that streams many chunks
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter doesn't support flushing")
		}

		// Stream many chunks
		for i := 0; i < 100; i++ {
			resp := GenerateContentResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: fmt.Sprintf("chunk%d ", i)},
							},
							Role: "model",
						},
					},
				},
			}
			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "%s\n", data)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	client := New(WithAPIKey("test-key"), WithBaseURL(server.URL), WithModel("gemini-pro"))

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	messages := []core.Message{
		{Role: "user", Content: "Test"},
	}

	chunkChan, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	// Read a few chunks then cancel
	chunkCount := 0
	for chunk := range chunkChan {
		if chunk.Error != nil {
			// Context cancellation error is expected
			if chunk.Error == context.Canceled || strings.Contains(chunk.Error.Error(), "context canceled") {
				break
			}
			t.Fatalf("Unexpected error: %v", chunk.Error)
		}
		chunkCount++
		if chunkCount == 3 {
			cancel()
		}
	}

	// Verify we received some chunks but not all 100
	if chunkCount >= 100 {
		t.Errorf("Expected cancellation to stop streaming, but received %d chunks", chunkCount)
	}
}

// TestChatStreamError verifies error handling during streaming
func TestChatStreamError(t *testing.T) {
	// Create mock server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		// Send error response as body
		errResp := ErrorResponse{
			Error: APIError{
				Code:    400,
				Message: "Content generation failed",
				Status:  "INVALID_ARGUMENT",
			},
		}
		json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	client := New(WithAPIKey("test-key"), WithBaseURL(server.URL), WithModel("gemini-pro"))

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "Test"},
	}

	_, err := client.ChatStream(ctx, messages)
	if err == nil {
		t.Fatal("Expected ChatStream to return an error")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "API error") {
		t.Errorf("Expected API error message, got: %v", err)
	}
}

// TestCompleteStream verifies the convenience method for single-prompt streaming
func TestCompleteStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter doesn't support flushing")
		}

		responses := []GenerateContentResponse{
			{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: "Complete"},
							},
							Role: "model",
						},
					},
				},
			},
			{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: " response"},
							},
							Role: "model",
						},
						FinishReason: "STOP",
					},
				},
			},
		}

		for _, resp := range responses {
			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "%s\n", data)
			flusher.Flush()
		}
	}))
	defer server.Close()

	client := New(WithAPIKey("test-key"), WithBaseURL(server.URL), WithModel("gemini-pro"))

	ctx := context.Background()
	chunkChan, err := client.CompleteStream(ctx, "Test prompt")
	if err != nil {
		t.Fatalf("CompleteStream failed: %v", err)
	}

	var chunks []core.StreamChunk
	for chunk := range chunkChan {
		if chunk.Error != nil {
			t.Fatalf("Received error chunk: %v", chunk.Error)
		}
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}

	finalChunk := chunks[len(chunks)-1]
	if finalChunk.Content != "Complete response" {
		t.Errorf("Expected 'Complete response', got %q", finalChunk.Content)
	}
}

// TestStreamChunkAccumulation verifies that content is properly accumulated across chunks
func TestStreamChunkAccumulation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter doesn't support flushing")
		}

		// Stream word by word
		words := []string{"The", " quick", " brown", " fox", " jumps"}
		for _, word := range words {
			resp := GenerateContentResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: word},
							},
							Role: "model",
						},
					},
				},
			}
			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "%s\n", data)
			flusher.Flush()
		}

		// Final chunk with finish reason
		resp := GenerateContentResponse{
			Candidates: []Candidate{
				{
					Content: Content{
						Parts: []Part{
							{Text: ""},
						},
						Role: "model",
					},
					FinishReason: "STOP",
				},
			},
		}
		data, _ := json.Marshal(resp)
		fmt.Fprintf(w, "%s\n", data)
		flusher.Flush()
	}))
	defer server.Close()

	client := New(WithAPIKey("test-key"), WithBaseURL(server.URL), WithModel("gemini-pro"))

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "Test"},
	}

	chunkChan, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	// Verify accumulation
	expectedAccumulations := []string{
		"The",
		"The quick",
		"The quick brown",
		"The quick brown fox",
		"The quick brown fox jumps",
		"The quick brown fox jumps",
	}

	var chunks []core.StreamChunk
	for chunk := range chunkChan {
		if chunk.Error != nil {
			t.Fatalf("Received error chunk: %v", chunk.Error)
		}
		chunks = append(chunks, chunk)
	}

	if len(chunks) != len(expectedAccumulations) {
		t.Errorf("Expected %d chunks, got %d", len(expectedAccumulations), len(chunks))
	}

	for i, chunk := range chunks {
		if chunk.Content != expectedAccumulations[i] {
			t.Errorf("Chunk %d: expected accumulated content %q, got %q",
				i, expectedAccumulations[i], chunk.Content)
		}
	}
}

// TestStreamChunkMetadata verifies that metadata is properly populated in stream chunks
func TestStreamChunkMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter doesn't support flushing")
		}

		resp := GenerateContentResponse{
			Candidates: []Candidate{
				{
					Content: Content{
						Parts: []Part{
							{Text: "Test response"},
						},
						Role: "model",
					},
					FinishReason: "STOP",
				},
			},
		}
		data, _ := json.Marshal(resp)
		fmt.Fprintf(w, "%s\n", data)
		flusher.Flush()
	}))
	defer server.Close()

	client := New(WithAPIKey("test-key"), WithBaseURL(server.URL), WithModel("gemini-1.5-pro"))

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "Test"},
	}

	chunkChan, err := client.ChatStream(ctx, messages)
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	var lastChunk core.StreamChunk
	for chunk := range chunkChan {
		if chunk.Error != nil {
			t.Fatalf("Received error chunk: %v", chunk.Error)
		}
		lastChunk = chunk

		// Verify metadata exists
		if chunk.Metadata == nil {
			t.Error("Expected metadata to be set")
		}

		// Verify model is in metadata
		if model, ok := chunk.Metadata["model"].(string); !ok {
			t.Error("Expected model in metadata")
		} else if model != "gemini-1.5-pro" {
			t.Errorf("Expected model 'gemini-1.5-pro', got %q", model)
		}

		// Verify index increments
		if chunk.Index < 0 {
			t.Errorf("Expected positive index, got %d", chunk.Index)
		}

		// Verify timestamp is set
		if chunk.Timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}
	}

	// Verify finish reason on last chunk
	if lastChunk.FinishReason != "stop" {
		t.Errorf("Expected finish reason 'stop', got %q", lastChunk.FinishReason)
	}
}
