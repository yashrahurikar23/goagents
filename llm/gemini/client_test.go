package gemini

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
	if client.model != ModelGemini15Flash {
		t.Errorf("Expected model %s, got %s", ModelGemini15Flash, client.model)
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
	model := ModelGemini15Pro
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
	maxTokens := 1024
	client := New(WithMaxTokens(maxTokens))
	if client.maxTokens == nil {
		t.Fatal("Expected non-nil maxTokens")
	}
	if *client.maxTokens != maxTokens {
		t.Errorf("Expected maxTokens %d, got %d", maxTokens, *client.maxTokens)
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
	topK := 40
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
	model := ModelGemini15Flash8B
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
	if err.Error() != "gemini: API key is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestChatSuccess(t *testing.T) {
	// Create mock response
	mockResponse := GenerateContentResponse{
		Candidates: []Candidate{
			{
				Content: Content{
					Role: "model",
					Parts: []Part{
						{Text: "Hello! How can I help you today?"},
					},
				},
				FinishReason: "STOP",
			},
		},
		UsageMetadata: UsageMetadata{
			PromptTokenCount:     10,
			CandidatesTokenCount: 20,
			TotalTokenCount:      30,
		},
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.String(), "key=test-key") {
			t.Error("Expected API key in URL")
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
	if resp.Meta["prompt_tokens"] != 10 {
		t.Errorf("Expected prompt_tokens in metadata")
	}
	if resp.Meta["completion_tokens"] != 20 {
		t.Errorf("Expected completion_tokens in metadata")
	}
}

func TestChatWithSystemMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify request
		var req GenerateContentRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.SystemInstruction == nil {
			t.Error("Expected systemInstruction in request")
		} else if len(req.SystemInstruction.Parts) == 0 ||
			req.SystemInstruction.Parts[0].Text != "You are a helpful assistant." {
			t.Error("System instruction not properly set")
		}

		// Check that system message is not in contents
		for _, content := range req.Contents {
			if content.Role == "system" {
				t.Error("System message should not be in contents array")
			}
		}

		// Send response
		mockResponse := GenerateContentResponse{
			Candidates: []Candidate{
				{
					Content: Content{
						Role:  "model",
						Parts: []Part{{Text: "Hello!"}},
					},
					FinishReason: "STOP",
				},
			},
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
			Error: APIError{
				Code:    400,
				Message: "Invalid request",
				Status:  "INVALID_ARGUMENT",
			},
		}
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

func TestChatBlockedContent(t *testing.T) {
	mockResponse := GenerateContentResponse{
		PromptFeedback: &PromptFeedback{
			BlockReason: "SAFETY",
		},
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
	messages := []core.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := client.Chat(ctx, messages)
	if err == nil {
		t.Fatal("Expected error for blocked content")
	}
	if !strings.Contains(err.Error(), "prompt blocked") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestCompleteSuccess(t *testing.T) {
	mockResponse := GenerateContentResponse{
		Candidates: []Candidate{
			{
				Content: Content{
					Role:  "model",
					Parts: []Part{{Text: "This is a completion"}},
				},
				FinishReason: "STOP",
			},
		},
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

	contents, systemInstruction := client.convertMessages(messages)

	// Should have 3 non-system messages
	if len(contents) != 3 {
		t.Errorf("Expected 3 contents, got %d", len(contents))
	}

	// System prompts should be combined
	if systemInstruction == nil {
		t.Fatal("Expected non-nil systemInstruction")
	}
	if len(systemInstruction.Parts) != 1 {
		t.Fatal("Expected 1 part in systemInstruction")
	}
	expectedSystem := "System prompt 1\n\nSystem prompt 2"
	if systemInstruction.Parts[0].Text != expectedSystem {
		t.Errorf("Expected system '%s', got '%s'",
			expectedSystem, systemInstruction.Parts[0].Text)
	}

	// Check message roles - Gemini uses "model" instead of "assistant"
	if contents[0].Role != "user" {
		t.Errorf("Expected first content role to be user, got %s", contents[0].Role)
	}
	if contents[1].Role != "model" {
		t.Errorf("Expected second content role to be model, got %s", contents[1].Role)
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
		ModelGemini20Flash,
		ModelGemini15Flash,
		ModelGemini15Flash8B,
		ModelGemini15Pro,
		ModelGeminiPro,
	}

	for _, model := range models {
		client := New(WithModel(model))
		if client.Model() != model {
			t.Errorf("Model mismatch for %s", model)
		}
	}
}

func TestGenerationConfigIncluded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GenerateContentRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.GenerationConfig == nil {
			t.Error("Expected generationConfig in request")
		} else {
			if req.GenerationConfig.Temperature == nil || *req.GenerationConfig.Temperature != 0.7 {
				t.Error("Temperature not set correctly")
			}
			if req.GenerationConfig.MaxOutputTokens == nil || *req.GenerationConfig.MaxOutputTokens != 1024 {
				t.Error("MaxOutputTokens not set correctly")
			}
		}

		mockResponse := GenerateContentResponse{
			Candidates: []Candidate{
				{
					Content: Content{
						Role:  "model",
						Parts: []Part{{Text: "Response"}},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
		WithTemperature(0.7),
		WithMaxTokens(1024),
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestMultiPartResponse(t *testing.T) {
	mockResponse := GenerateContentResponse{
		Candidates: []Candidate{
			{
				Content: Content{
					Role: "model",
					Parts: []Part{
						{Text: "Part 1. "},
						{Text: "Part 2. "},
						{Text: "Part 3."},
					},
				},
				FinishReason: "STOP",
			},
		},
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
	messages := []core.Message{
		{Role: "user", Content: "Hello"},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Part 1. Part 2. Part 3."
	if resp.Content != expected {
		t.Errorf("Expected content '%s', got '%s'", expected, resp.Content)
	}
}
