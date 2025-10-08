package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yashrahurikar23/goagents/core"
)

func TestNew(t *testing.T) {
	client := New()
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.baseURL != DefaultBaseURL {
		t.Errorf("Expected baseURL %s, got %s", DefaultBaseURL, client.baseURL)
	}
	if client.model != ModelClaude35Sonnet {
		t.Errorf("Expected model %s, got %s", ModelClaude35Sonnet, client.model)
	}
	if client.maxTokens != DefaultMaxTokens {
		t.Errorf("Expected maxTokens %d, got %d", DefaultMaxTokens, client.maxTokens)
	}
}

func TestWithAPIKey(t *testing.T) {
	apiKey := "test-api-key"
	client := New(WithAPIKey(apiKey))
	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey %s, got %s", apiKey, client.apiKey)
	}
}

func TestWithModel(t *testing.T) {
	model := ModelClaude3Opus
	client := New(WithModel(model))
	if client.model != model {
		t.Errorf("Expected model %s, got %s", model, client.model)
	}
}

func TestWithBaseURL(t *testing.T) {
	baseURL := "https://custom.api.com"
	client := New(WithBaseURL(baseURL))
	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}
}

func TestWithMaxTokens(t *testing.T) {
	maxTokens := 2048
	client := New(WithMaxTokens(maxTokens))
	if client.maxTokens != maxTokens {
		t.Errorf("Expected maxTokens %d, got %d", maxTokens, client.maxTokens)
	}
}

func TestWithTemperature(t *testing.T) {
	temp := 0.7
	client := New(WithTemperature(temp))
	if client.temperature == nil {
		t.Fatal("Expected non-nil temperature")
	}
	if *client.temperature != temp {
		t.Errorf("Expected temperature %f, got %f", temp, *client.temperature)
	}
}

func TestWithTopP(t *testing.T) {
	topP := 0.9
	client := New(WithTopP(topP))
	if client.topP == nil {
		t.Fatal("Expected non-nil topP")
	}
	if *client.topP != topP {
		t.Errorf("Expected topP %f, got %f", topP, *client.topP)
	}
}

func TestWithTopK(t *testing.T) {
	topK := 50
	client := New(WithTopK(topK))
	if client.topK == nil {
		t.Fatal("Expected non-nil topK")
	}
	if *client.topK != topK {
		t.Errorf("Expected topK %d, got %d", topK, *client.topK)
	}
}

func TestWithTimeout(t *testing.T) {
	timeout := 30 * time.Second
	client := New(WithTimeout(timeout))
	if client.httpClient.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, client.httpClient.Timeout)
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	client := New(WithHTTPClient(customClient))
	if client.httpClient != customClient {
		t.Error("Expected custom HTTP client")
	}
}

func TestModel(t *testing.T) {
	model := ModelClaude3Haiku
	client := New(WithModel(model))
	if client.Model() != model {
		t.Errorf("Expected Model() to return %s, got %s", model, client.Model())
	}
}

func TestChatWithoutAPIKey(t *testing.T) {
	client := New()
	ctx := context.Background()

	messages := []core.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := client.Chat(ctx, messages)
	if err == nil {
		t.Error("Expected error when API key is not set")
	}
	if err.Error() != "anthropic: API key is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestChatSuccess(t *testing.T) {
	// Create mock server
	mockResponse := Response{
		ID:         "msg_test123",
		Type:       "message",
		Role:       "assistant",
		Model:      ModelClaude35Sonnet,
		StopReason: "end_turn",
		Content: []ResponseContent{
			{Type: "text", Text: "Hello! How can I help you today?"},
		},
		Usage: Usage{
			InputTokens:  10,
			OutputTokens: 20,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("Expected API key header")
		}
		if r.Header.Get("anthropic-version") != DefaultAPIVersion {
			t.Errorf("Expected API version header")
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "Hello"},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Content != "Hello! How can I help you today?" {
		t.Errorf("Unexpected response content: %s", resp.Content)
	}

	// Check metadata
	if resp.Meta["model"] != ModelClaude35Sonnet {
		t.Errorf("Expected model in metadata")
	}
	if resp.Meta["input_tokens"] != 10 {
		t.Errorf("Expected input_tokens in metadata")
	}
	if resp.Meta["output_tokens"] != 20 {
		t.Errorf("Expected output_tokens in metadata")
	}
}

func TestChatWithSystemMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify request
		var req Request
		json.NewDecoder(r.Body).Decode(&req)

		if req.System != "You are a helpful assistant." {
			t.Errorf("Expected system prompt in request, got: %s", req.System)
		}

		// Check that system message is not in messages array
		for _, msg := range req.Messages {
			if msg.Role == "system" {
				t.Error("System message should not be in messages array")
			}
		}

		// Send response
		mockResponse := Response{
			ID:         "msg_test123",
			Type:       "message",
			Role:       "assistant",
			Model:      ModelClaude35Sonnet,
			StopReason: "end_turn",
			Content:    []ResponseContent{{Type: "text", Text: "Hello!"}},
			Usage:      Usage{InputTokens: 10, OutputTokens: 5},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
	}

	_, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestChatAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errResp := ErrorResponse{
			Type: "error",
		}
		errResp.Error.Type = "invalid_request_error"
		errResp.Error.Message = "Invalid request"
		json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := client.Chat(ctx, messages)
	if err == nil {
		t.Fatal("Expected error for API error response")
	}
	if !strings.Contains(err.Error(), "Invalid request") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestCompleteSuccess(t *testing.T) {
	mockResponse := Response{
		ID:         "msg_test123",
		Type:       "message",
		Role:       "assistant",
		Model:      ModelClaude35Sonnet,
		StopReason: "end_turn",
		Content:    []ResponseContent{{Type: "text", Text: "This is a completion"}},
		Usage:      Usage{InputTokens: 5, OutputTokens: 10},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	ctx := context.Background()
	result, err := client.Complete(ctx, "Test prompt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "This is a completion" {
		t.Errorf("Unexpected completion result: %s", result)
	}
}

func TestConvertMessages(t *testing.T) {
	client := New()

	messages := []core.Message{
		{Role: "system", Content: "System prompt 1"},
		{Role: "system", Content: "System prompt 2"},
		{Role: "user", Content: "User message"},
		{Role: "assistant", Content: "Assistant response"},
		{Role: "user", Content: "Another user message"},
	}

	anthropicMsgs, systemPrompt := client.convertMessages(messages)

	// Should have 3 non-system messages
	if len(anthropicMsgs) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(anthropicMsgs))
	}

	// System prompts should be combined
	expectedSystem := "System prompt 1\n\nSystem prompt 2"
	if systemPrompt != expectedSystem {
		t.Errorf("Expected system prompt '%s', got '%s'", expectedSystem, systemPrompt)
	}

	// Check message roles
	if anthropicMsgs[0].Role != "user" {
		t.Errorf("Expected first message role to be user, got %s", anthropicMsgs[0].Role)
	}
	if anthropicMsgs[1].Role != "assistant" {
		t.Errorf("Expected second message role to be assistant, got %s", anthropicMsgs[1].Role)
	}
}

func TestContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	messages := []core.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := client.Chat(ctx, messages)
	if err == nil {
		t.Fatal("Expected error due to context cancellation")
	}
}

func TestMultipleModels(t *testing.T) {
	models := []string{
		ModelClaude35Sonnet,
		ModelClaude3Opus,
		ModelClaude3Sonnet,
		ModelClaude3Haiku,
		ModelClaude35Haiku,
	}

	for _, model := range models {
		client := New(WithModel(model))
		if client.Model() != model {
			t.Errorf("Model mismatch for %s", model)
		}
	}
}

func TestRequestMarshaling(t *testing.T) {
	temp := 0.7
	topP := 0.9
	topK := 50

	req := Request{
		Model:         ModelClaude35Sonnet,
		MaxTokens:     1024,
		Temperature:   &temp,
		TopP:          &topP,
		TopK:          &topK,
		StopSequences: []string{"STOP"},
		System:        "You are helpful",
		Messages: []Message{
			{
				Role: "user",
				Content: []ContentItem{
					{Type: "text", Text: "Hello"},
				},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify JSON contains expected fields
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if parsed["model"] != ModelClaude35Sonnet {
		t.Error("Model not in JSON")
	}
	if parsed["max_tokens"] != float64(1024) {
		t.Error("MaxTokens not in JSON")
	}
	if parsed["temperature"] != 0.7 {
		t.Error("Temperature not in JSON")
	}
}
