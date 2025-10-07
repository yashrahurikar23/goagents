package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yashrahurikar/goagents/core"
)

// Client represents an Ollama LLM client
type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
	options    *RequestOptions
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// New creates a new Ollama client
func New(opts ...ClientOption) *Client {
	c := &Client{
		baseURL: "http://localhost:11434",
		model:   "llama3.2:1b",
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		options: &RequestOptions{},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithBaseURL sets the base URL for the Ollama API
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithModel sets the default model to use
func WithModel(model string) ClientOption {
	return func(c *Client) {
		c.model = model
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTemperature sets the temperature parameter
func WithTemperature(temperature float64) ClientOption {
	return func(c *Client) {
		c.options.Temperature = &temperature
	}
}

// WithTopP sets the top_p parameter
func WithTopP(topP float64) ClientOption {
	return func(c *Client) {
		c.options.TopP = &topP
	}
}

// WithTopK sets the top_k parameter
func WithTopK(topK int) ClientOption {
	return func(c *Client) {
		c.options.TopK = &topK
	}
}

// WithMaxTokens sets the maximum number of tokens to generate
func WithMaxTokens(maxTokens int) ClientOption {
	return func(c *Client) {
		c.options.NumPredict = &maxTokens
	}
}

// WithStop sets the stop sequences
func WithStop(stop []string) ClientOption {
	return func(c *Client) {
		c.options.Stop = stop
	}
}

// Chat sends a chat request to Ollama
func (c *Client) Chat(ctx context.Context, messages []core.Message) (*core.Response, error) {
	// Convert messages to Ollama format
	ollamaMessages := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Create request
	req := ChatRequest{
		Model:    c.model,
		Messages: ollamaMessages,
		Stream:   false,
		Options:  c.options,
	}

	// Send request
	var resp ChatResponse
	if err := c.doRequest(ctx, "/api/chat", req, &resp); err != nil {
		return nil, fmt.Errorf("chat request failed: %w", err)
	}

	// Convert response
	response := &core.Response{
		Content: resp.Message.Content,
		Meta: map[string]interface{}{
			"model":                resp.Model,
			"created_at":           resp.CreatedAt,
			"done_reason":          resp.DoneReason,
			"total_duration":       resp.TotalDuration,
			"prompt_eval_count":    resp.PromptEvalCount,
			"prompt_eval_duration": resp.PromptEvalDuration,
			"eval_count":           resp.EvalCount,
			"eval_duration":        resp.EvalDuration,
		},
	}

	// Convert tool calls if any
	if len(resp.Message.ToolCalls) > 0 {
		response.ToolCalls = make([]core.ToolCall, len(resp.Message.ToolCalls))
		for i, tc := range resp.Message.ToolCalls {
			response.ToolCalls[i] = core.ToolCall{
				ID:   tc.ID,
				Name: tc.Function.Name,
				Args: tc.Function.Arguments,
			}
		}
	}

	return response, nil
}

// Complete sends a completion request to Ollama
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	// Create request
	req := GenerateRequest{
		Model:   c.model,
		Prompt:  prompt,
		Stream:  false,
		Options: c.options,
	}

	// Send request
	var resp GenerateResponse
	if err := c.doRequest(ctx, "/api/generate", req, &resp); err != nil {
		return "", fmt.Errorf("completion request failed: %w", err)
	}

	return resp.Response, nil
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// Stream sends a streaming chat request to Ollama
func (c *Client) Stream(ctx context.Context, messages []core.Message) (<-chan StreamChunk, error) {
	// Convert messages to Ollama format
	ollamaMessages := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Create request
	req := ChatRequest{
		Model:    c.model,
		Messages: ollamaMessages,
		Stream:   true,
		Options:  c.options,
	}

	// Create HTTP request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		httpResp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	// Create channel for streaming chunks
	chunks := make(chan StreamChunk)

	// Start goroutine to read streaming response
	go func() {
		defer close(chunks)
		defer httpResp.Body.Close()

		scanner := bufio.NewScanner(httpResp.Body)
		for scanner.Scan() {
			var resp ChatResponse
			if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
				chunks <- StreamChunk{Error: fmt.Errorf("failed to unmarshal response: %w", err)}
				return
			}

			chunks <- StreamChunk{
				Content: resp.Message.Content,
				Done:    resp.Done,
			}

			if resp.Done {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			chunks <- StreamChunk{Error: fmt.Errorf("error reading stream: %w", err)}
		}
	}()

	return chunks, nil
}

// doRequest sends an HTTP request and unmarshals the response
func (c *Client) doRequest(ctx context.Context, endpoint string, reqBody interface{}, respBody interface{}) error {
	// Marshal request body
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(responseBody, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("ollama error: %s", errResp.Error)
		}
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(responseBody))
	}

	// Debug: print response body length
	// fmt.Printf("DEBUG: Response body length: %d\n", len(responseBody))
	// fmt.Printf("DEBUG: Response body: %s\n", string(responseBody))

	// Unmarshal response
	if err := json.Unmarshal(responseBody, respBody); err != nil {
		return fmt.Errorf("failed to unmarshal response (body len=%d): %w\nBody: %s", len(responseBody), err, string(responseBody)[:min(500, len(responseBody))])
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ListModels lists all available models
func (c *Client) ListModels(ctx context.Context) (*ListModelsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Embedding generates embeddings for a prompt
func (c *Client) Embedding(ctx context.Context, prompt string) ([]float64, error) {
	req := EmbeddingRequest{
		Model:   c.model,
		Prompt:  prompt,
		Options: c.options,
	}

	var resp EmbeddingResponse
	if err := c.doRequest(ctx, "/api/embeddings", req, &resp); err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}

	return resp.Embedding, nil
}
