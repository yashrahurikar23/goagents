package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yashrahurikar23/goagents/core"
)

const (
	// DefaultBaseURL is the default OpenAI API base URL.
	// WHY: v1 endpoint is stable and widely supported
	DefaultBaseURL = "https://api.openai.com/v1"

	// DefaultModel is the default model to use.
	// WHY: GPT-4 provides best quality-to-cost ratio for agent use cases
	// Can be overridden per-request or via WithModel option
	DefaultModel = "gpt-4"

	// DefaultTimeout is the default HTTP timeout.
	// WHY: 60s accommodates complex prompts while preventing indefinite hangs
	// Streaming requests may take longer due to incremental delivery
	DefaultTimeout = 60 * time.Second

	// DefaultMaxRetries is the default number of retries.
	// WHY: 3 retries handles most transient failures (rate limits, server errors)
	// Uses exponential backoff: 1s, 2s, 4s to respect rate limits
	DefaultMaxRetries = 3
)

// Client is an OpenAI API client.
//
// WHY THIS DESIGN:
// - Immutable after creation: Thread-safe for concurrent use
// - Functional options: Extensible configuration without breaking API
// - Custom HTTP client support: Enables middleware, custom transports, testing
// - Per-client settings: Multiple clients can coexist with different configs
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	maxRetries int
	timeout    time.Duration
}

// Option is a functional option for configuring the Client.
//
// WHY FUNCTIONAL OPTIONS:
// - Backward compatible: New options don't break existing code
// - Self-documenting: Option names clearly indicate what they configure
// - Type-safe: Compiler catches configuration errors
// - Composable: Options can be stored and reused
type Option func(*Client)

// WithAPIKey sets the API key.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithBaseURL sets the base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithModel sets the default model.
func WithModel(model string) Option {
	return func(c *Client) {
		c.model = model
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the HTTP timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// New creates a new OpenAI client with the given options.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		model:      DefaultModel,
		timeout:    DefaultTimeout,
		maxRetries: DefaultMaxRetries,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: c.timeout,
		}
	}

	return c
}

// Chat implements the core.LLM interface for chat completions.
//
// WHY THIS WAY:
// - Converts between core types and OpenAI types for framework compatibility
// - Preserves tool calls in response for agent function calling
// - Returns structured error types (ErrLLMFailure) for proper error handling
// - Stores usage metadata for token tracking and cost monitoring
//
// BUSINESS LOGIC:
// - Must have at least one choice in response (API contract)
// - Tool calls parsed as JSON arguments for type safety
// - Content may be empty when tool calls are present
//
// WHEN TO USE:
// - Multi-turn conversations with chat history
// - Function calling / tool use scenarios
// - When you need structured responses with metadata
func (c *Client) Chat(ctx context.Context, messages []core.Message) (*core.Response, error) {
	// WHY: Convert between framework types and OpenAI API types
	// Maintains clean separation between core interfaces and provider implementations
	chatMessages := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	req := ChatCompletionRequest{
		Model:    c.model,
		Messages: chatMessages,
	}

	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, &core.ErrLLMFailure{
			Provider: "openai",
			Err:      err,
		}
	}

	if len(resp.Choices) == 0 {
		return nil, &core.ErrLLMFailure{
			Provider: "openai",
			Err:      fmt.Errorf("no choices in response"),
		}
	}

	choice := resp.Choices[0]

	// Convert tool calls if present
	var toolCalls []core.ToolCall
	if choice.Message != nil && len(choice.Message.ToolCalls) > 0 {
		toolCalls = make([]core.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			var args map[string]interface{}
			if tc.Function != nil && tc.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					return nil, fmt.Errorf("failed to parse tool call arguments: %w", err)
				}
			}

			toolCalls[i] = core.ToolCall{
				ID:   tc.ID,
				Name: tc.Function.Name,
				Args: args,
			}
		}
	}

	content := ""
	if choice.Message != nil {
		if str, ok := choice.Message.Content.(string); ok {
			content = str
		}
	}

	return &core.Response{
		Content:   content,
		ToolCalls: toolCalls,
		Meta: map[string]interface{}{
			"model":         resp.Model,
			"finish_reason": choice.FinishReason,
			"usage":         resp.Usage,
		},
	}, nil
}

// Complete implements the core.LLM interface for simple completions.
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	messages := []core.Message{
		{Role: "user", Content: prompt},
	}

	resp, err := c.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// CreateChatCompletion creates a chat completion.
func (c *Client) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	var resp ChatCompletionResponse
	if err := c.doRequest(ctx, "POST", "/chat/completions", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ChatStream implements the core.StreamingLLM interface for streaming chat completions.
//
// WHY THIS METHOD:
// - Implements core.StreamingLLM interface for framework compatibility
// - Returns a channel of core.StreamChunk for real-time token delivery
// - Handles context cancellation gracefully
// - Accumulates content across chunks for convenience
//
// IMPLEMENTATION:
// - Uses CreateChatCompletionStream internally
// - Converts OpenAI SSE chunks to core.StreamChunk format
// - Runs stream processing in a goroutine
// - Closes channel when stream completes or errors occur
func (c *Client) ChatStream(ctx context.Context, messages []core.Message, opts ...interface{}) (<-chan core.StreamChunk, error) {
	// Convert core.Message to ChatMessage
	chatMessages := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	req := ChatCompletionRequest{
		Model:    c.model,
		Messages: chatMessages,
	}

	// Create buffered channel for chunks
	chunkChan := make(chan core.StreamChunk, 10)

	// Track accumulated content and index
	var content string
	index := 0

	// Start streaming in a goroutine
	go func() {
		defer close(chunkChan)

		streamOpts := StreamOptions{
			OnChunk: func(chunk *ChatCompletionStreamResponse) error {
				// Check for context cancellation
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				// Extract delta from chunk
				if len(chunk.Choices) == 0 {
					return nil
				}

				choice := chunk.Choices[0]
				var delta string

				if choice.Delta != nil {
					if deltaContent, ok := choice.Delta.Content.(string); ok {
						delta = deltaContent
						content += delta
					}
				}

				// Create StreamChunk
				streamChunk := core.StreamChunk{
					Content:      content,
					Delta:        delta,
					Index:        index,
					FinishReason: choice.FinishReason,
					Metadata: map[string]interface{}{
						"model": chunk.Model,
						"id":    chunk.ID,
					},
					Timestamp: time.Now(),
				}

				index++

				// Send chunk on channel
				select {
				case <-ctx.Done():
					return ctx.Err()
				case chunkChan <- streamChunk:
				}

				return nil
			},
			OnComplete: func() error {
				return nil
			},
			OnError: func(err error) {
				// Send error chunk
				errorChunk := core.StreamChunk{
					Content:   content,
					Index:     index,
					Error:     err,
					Timestamp: time.Now(),
				}
				select {
				case <-ctx.Done():
				case chunkChan <- errorChunk:
				}
			},
		}

		if err := c.CreateChatCompletionStream(ctx, req, streamOpts); err != nil {
			// Send final error chunk
			errorChunk := core.StreamChunk{
				Content:   content,
				Index:     index,
				Error:     err,
				Timestamp: time.Now(),
			}
			select {
			case <-ctx.Done():
			case chunkChan <- errorChunk:
			}
		}
	}()

	return chunkChan, nil
}

// CompleteStream implements the core.StreamingLLM interface for streaming completions.
//
// WHY THIS METHOD:
// - Provides simple streaming interface for single-prompt use cases
// - Wraps prompt as user message and delegates to ChatStream
// - Maintains consistency with Complete() method signature
func (c *Client) CompleteStream(ctx context.Context, prompt string, opts ...interface{}) (<-chan core.StreamChunk, error) {
	messages := []core.Message{
		{Role: "user", Content: prompt},
	}
	return c.ChatStream(ctx, messages, opts...)
}

// CreateChatCompletionStream creates a streaming chat completion.
//
// WHY STREAMING:
// - Enables real-time UI updates as tokens arrive (better UX)
// - Reduces time-to-first-token latency perception
// - Allows early termination based on content
// - Memory efficient: processes chunks without buffering entire response
//
// WHY CALLBACK PATTERN:
// - Provides control over chunk processing to caller
// - Enables error handling at chunk level
// - Supports different consumption patterns (print, accumulate, filter)
//
// IMPLEMENTATION NOTES:
// - Uses Server-Sent Events (SSE) format from OpenAI
// - "[DONE]" message signals stream completion
// - Errors during streaming call OnError handler but don't stop stream
// - Context cancellation stops streaming immediately
func (c *Client) CreateChatCompletionStream(ctx context.Context, req ChatCompletionRequest, opts StreamOptions) error {
	if req.Model == "" {
		req.Model = c.model
	}
	req.Stream = true

	httpReq, err := c.newRequest(ctx, "POST", "/chat/completions", req)
	if err != nil {
		return err
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(httpResp)
	}

	// Process SSE stream
	reader := bufio.NewReader(httpResp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			if opts.OnError != nil {
				opts.OnError(err)
			}
			return fmt.Errorf("stream read error: %w", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// SSE format: "data: {...}"
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		data := bytes.TrimPrefix(line, []byte("data: "))

		// Check for stream end
		if bytes.Equal(data, []byte("[DONE]")) {
			if opts.OnComplete != nil {
				if err := opts.OnComplete(); err != nil {
					return err
				}
			}
			break
		}

		// Parse chunk
		var chunk ChatCompletionStreamResponse
		if err := json.Unmarshal(data, &chunk); err != nil {
			if opts.OnError != nil {
				opts.OnError(err)
			}
			continue
		}

		// Call chunk handler
		if opts.OnChunk != nil {
			if err := opts.OnChunk(&chunk); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateEmbedding creates embeddings for the given input.
func (c *Client) CreateEmbedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "text-embedding-ada-002"
	}

	var resp EmbeddingResponse
	if err := c.doRequest(ctx, "POST", "/embeddings", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateModeration checks content for policy violations.
func (c *Client) CreateModeration(ctx context.Context, req ModerationRequest) (*ModerationResponse, error) {
	if req.Model == "" {
		req.Model = "omni-moderation-latest"
	}

	var resp ModerationResponse
	if err := c.doRequest(ctx, "POST", "/moderations", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ListModels lists available models.
func (c *Client) ListModels(ctx context.Context) (*ModelListResponse, error) {
	var resp ModelListResponse
	if err := c.doRequest(ctx, "GET", "/models", nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// doRequest performs an HTTP request with retry logic.
func (c *Client) doRequest(ctx context.Context, method, path string, reqBody, respBody interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s, ...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := c.newRequest(ctx, method, path, reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		defer resp.Body.Close()

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if respBody != nil {
				if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}
			}
			return nil
		}

		// Handle errors
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			// Retry on rate limit or server errors
			lastErr = c.handleErrorResponse(resp)
			continue
		}

		// Don't retry on client errors
		return c.handleErrorResponse(resp)
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// newRequest creates a new HTTP request.
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return req, nil
}

// handleErrorResponse handles an error response.
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read error response: %w", resp.StatusCode, err)
	}

	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if errorResp.Error != nil {
		return &OpenAIError{
			StatusCode: resp.StatusCode,
			Type:       errorResp.Error.Type,
			Message:    errorResp.Error.Message,
			Code:       errorResp.Error.Code,
		}
	}

	return fmt.Errorf("HTTP %d: unknown error", resp.StatusCode)
}

// OpenAIError represents an OpenAI API error.
type OpenAIError struct {
	StatusCode int
	Type       string
	Message    string
	Code       interface{}
}

func (e *OpenAIError) Error() string {
	return fmt.Sprintf("openai: %s (HTTP %d): %s", e.Type, e.StatusCode, e.Message)
}

// IsRateLimitError checks if the error is a rate limit error.
func IsRateLimitError(err error) bool {
	if oaiErr, ok := err.(*OpenAIError); ok {
		return oaiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

// IsTimeoutError checks if the error is a timeout error.
func IsTimeoutError(err error) bool {
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded")
}

// Helper functions for creating messages

// SystemMessage creates a system message.
func SystemMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    "system",
		Content: content,
	}
}

// UserMessage creates a user message.
func UserMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    "user",
		Content: content,
	}
}

// UserMessageWithImage creates a user message with text and image.
func UserMessageWithImage(text, imageURL string, detail ...string) ChatMessage {
	detailLevel := "auto"
	if len(detail) > 0 {
		detailLevel = detail[0]
	}

	return ChatMessage{
		Role: "user",
		Content: []ContentPart{
			{Type: "text", Text: text},
			{
				Type: "image_url",
				ImageURL: &ImageURL{
					URL:    imageURL,
					Detail: detailLevel,
				},
			},
		},
	}
}

// AssistantMessage creates an assistant message.
func AssistantMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    "assistant",
		Content: content,
	}
}

// ToolMessage creates a tool message.
func ToolMessage(toolCallID, content string) ChatMessage {
	return ChatMessage{
		Role:       "tool",
		Content:    content,
		ToolCallID: toolCallID,
	}
}

// Helper functions for creating tools

// NewFunction creates a new function definition.
func NewFunction(name, description string, parameters map[string]interface{}) *Function {
	return &Function{
		Name:        name,
		Description: description,
		Parameters:  parameters,
	}
}

// NewTool creates a new tool from a function.
func NewTool(fn *Function) Tool {
	return Tool{
		Type:     "function",
		Function: fn,
	}
}

// JSONSchema creates a JSON schema for function parameters.
func JSONSchema(properties map[string]interface{}, required []string) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// PropertyString creates a string property for JSON schema.
func PropertyString(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
	}
}

// PropertyNumber creates a number property for JSON schema.
func PropertyNumber(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "number",
		"description": description,
	}
}

// PropertyBoolean creates a boolean property for JSON schema.
func PropertyBoolean(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
}

// PropertyArray creates an array property for JSON schema.
func PropertyArray(description string, items map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items":       items,
	}
}

// PropertyEnum creates an enum property for JSON schema.
func PropertyEnum(description string, values []string) map[string]interface{} {
	enumValues := make([]interface{}, len(values))
	for i, v := range values {
		enumValues[i] = v
	}
	return map[string]interface{}{
		"type":        "string",
		"description": description,
		"enum":        enumValues,
	}
}
