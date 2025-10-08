// Package anthropic provides an implementation of the core.LLM interface for Anthropic's Claude models.
//
// PURPOSE:
// This package implements a production-ready HTTP client for Anthropic's Claude API,
// translating between GoAgents' universal core.LLM interface and Anthropic's specific
// API requirements.
//
// WHY THIS EXISTS:
// Anthropic's Claude models have unique requirements that differ from other LLM providers:
// 1. System prompts are sent as a separate "system" parameter, not in the messages array
// 2. Content is structured as arrays of ContentItems to support multimodal features
// 3. Authentication uses x-api-key header instead of Bearer tokens
// 4. API version is specified via anthropic-version header for stability
//
// This client abstracts these differences, allowing users to work with Claude models
// using the same interface as other LLM providers in GoAgents.
//
// KEY DESIGN DECISIONS:
//   - System message extraction: Anthropic requires system prompts separate from conversation,
//     so we extract and combine all "system" role messages into a single system parameter
//   - Metadata enrichment: We include token usage, stop_reason, and model info in response
//     metadata to enable usage tracking, cost calculation, and debugging
//   - Error handling: We parse Anthropic's error format and provide clear error messages
//     including error type and message for easier troubleshooting
//
// METHODS:
// - New(opts): Creates a client with functional options for flexible configuration
// - Chat(ctx, messages): Send conversation history, get response (implements core.LLM)
// - Complete(ctx, prompt): Convenience method for single-turn completions
// - Model(): Returns the model name being used
//
// INTERNAL METHODS:
// - doRequest: Handles HTTP request/response cycle with proper headers and error handling
// - convertMessages: Transforms core.Message to Anthropic's Message format, extracts system prompts
// - convertResponse: Transforms Anthropic's Response to core.Response with enriched metadata
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/yashrahurikar23/goagents/core"
)

// Client is the Anthropic API client that implements core.LLM interface.
// WHY: Encapsulates all configuration and state needed to interact with Claude API,
// including authentication, model selection, sampling parameters, and HTTP transport.
type Client struct {
	apiKey      string       // WHY: Required for authentication with Anthropic API
	baseURL     string       // WHY: Allows pointing to different endpoints (proxy, staging, regional)
	model       string       // WHY: Specifies which Claude model to use (3.5 Sonnet, 3 Opus, etc.)
	maxTokens   int          // WHY: Controls response length and API costs
	temperature *float64     // WHY: Pointer to distinguish "not set" from "set to 0.0"; controls randomness
	topP        *float64     // WHY: Pointer for optional nucleus sampling parameter
	topK        *int         // WHY: Pointer for optional top-k sampling parameter
	apiVersion  string       // WHY: Date-based API version for stability and feature control
	httpClient  *http.Client // WHY: Allows custom HTTP configuration (timeouts, proxies, TLS)
}

// New creates a new Anthropic client with the given options.
// WHY: Uses functional options pattern to provide flexible configuration with sensible
// defaults. This allows adding new options in the future without breaking existing code.
// Defaults to Claude 3.5 Sonnet, 4096 max tokens, and 60-second timeout.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		model:      ModelClaude35Sonnet,
		maxTokens:  DefaultMaxTokens,
		apiVersion: DefaultAPIVersion,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	// Apply all provided options in order
	// WHY: Options are applied sequentially, allowing later options to override earlier ones
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Chat sends a conversation and receives a response (implements core.LLM).
// WHY: This is the primary method for interacting with Claude, supporting full conversation
// history for multi-turn dialogues. It handles:
// 1. System message extraction (Claude requires system prompts separate from messages)
// 2. Message format conversion (core.Message -> Anthropic Message format)
// 3. API communication with proper authentication and headers
// 4. Response conversion back to core format with enriched metadata
//
// BUSINESS LOGIC:
// - Validates API key is configured (fails fast if missing)
// - Extracts all system messages and combines them into Claude's system parameter
// - Sends user/assistant messages as conversation history
// - Includes token usage in response metadata for cost tracking and optimization
func (c *Client) Chat(ctx context.Context, messages []core.Message) (*core.Response, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("anthropic: API key is required")
	}

	// Convert core.Message to Anthropic format and extract system prompts
	// WHY: Anthropic treats system prompts differently from other providers,
	// requiring them in a separate "system" field rather than in the messages array
	anthropicMessages, systemPrompt := c.convertMessages(messages)

	// Build request with all configured parameters
	req := Request{
		Model:       c.model,
		Messages:    anthropicMessages,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
		TopP:        c.topP,
		TopK:        c.topK,
		System:      systemPrompt,
	}

	// Make API call with context for cancellation and timeouts
	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("anthropic: request failed: %w", err)
	}

	// Convert response to core format with metadata
	return c.convertResponse(resp), nil
}

// Complete sends a single prompt and receives a completion (implements core.LLM).
// WHY: This is a convenience method for simple, single-turn interactions where you
// don't need conversation history. It wraps Chat() but provides a simpler interface
// that just takes a string and returns a string.
//
// WHEN TO USE:
// - One-off completions (e.g., "Summarize this text")
// - Simple Q&A without conversation context
// - Quick prototyping and testing
//
// For multi-turn conversations or when you need metadata, use Chat() directly.
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

// doRequest makes the HTTP request to the Anthropic API.
// WHY: Centralizes all HTTP communication logic for consistency and maintainability.
// Handles the complete request/response cycle including:
// - Request marshaling and HTTP setup
// - Authentication and API versioning headers
// - Context-aware cancellation and timeouts
// - Error parsing and proper error messages
// - Response parsing
//
// KEY DESIGN DECISIONS:
// - Uses context for cancellation: Allows callers to cancel requests and enforces timeouts
// - x-api-key header authentication: Required by Anthropic instead of Bearer token
// - anthropic-version header: Ensures consistent API behavior across version updates
// - Error type extraction: Provides detailed error info (type + message) for debugging
func (c *Client) doRequest(ctx context.Context, req Request) (*Response, error) {
	// Marshal request body to JSON
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context for cancellation
	// WHY: Context allows timeout enforcement and request cancellation
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	// WHY: Content-Type tells API we're sending JSON
	// WHY: x-api-key is Anthropic's authentication method (not Bearer token)
	// WHY: anthropic-version ensures stable API behavior even as API evolves
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", c.apiVersion)

	// Execute HTTP request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read full response body
	// WHY: Need full body for both success and error cases
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for API errors
	// WHY: Non-200 status indicates an error; parse error details for helpful message
	if httpResp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			// If we can't parse error response, return raw body
			return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(respBody))
		}
		// Return structured error with type and message for easier debugging
		return nil, fmt.Errorf("API error: %s (type: %s)", errResp.Error.Message, errResp.Error.Type)
	}

	// Parse successful response
	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// convertMessages converts core messages to Anthropic format and extracts system prompts.
// WHY: Anthropic has a unique requirement where system prompts must be sent in a separate
// "system" parameter rather than as messages with role="system". This differs from OpenAI
// and other providers who include system messages in the messages array.
//
// BUSINESS LOGIC:
// 1. Extract all messages with role="system" and combine them
// 2. Convert user/assistant messages to Anthropic's ContentItem array format
// 3. Return both the message array and the combined system prompt
//
// WHY CONTENTITEM ARRAY:
// Anthropic uses ContentItem arrays to support multimodal content (text, images, etc.)
// in the future without breaking the API. Even for text-only, we use this structure.
//
// WHY COMBINE SYSTEM PROMPTS:
// If multiple system messages are provided, we join them with double newlines to
// preserve structure while consolidating into Anthropic's single system parameter.
func (c *Client) convertMessages(messages []core.Message) ([]Message, string) {
	var anthropicMessages []Message
	var systemPrompts []string

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			// System messages are handled separately in Anthropic API
			// WHY: Anthropic requires system prompts in a separate "system" parameter
			systemPrompts = append(systemPrompts, msg.Content)
		case "user", "assistant":
			// Convert to Anthropic's message format with ContentItem array
			// WHY: ContentItem array enables future multimodal support (images, etc.)
			anthropicMessages = append(anthropicMessages, Message{
				Role: msg.Role,
				Content: []ContentItem{
					{Type: "text", Text: msg.Content},
				},
			})
		}
	}

	// Combine all system prompts into single string
	// WHY: Anthropic only accepts one system parameter, so we join multiple
	// system messages with double newlines to preserve structure
	systemPrompt := ""
	if len(systemPrompts) > 0 {
		systemPrompt = strings.Join(systemPrompts, "\n\n")
	}

	return anthropicMessages, systemPrompt
}

// convertResponse converts Anthropic response to core format with enriched metadata.
// WHY: Anthropic's response structure differs from core.Response, so we need to:
// 1. Extract text from ContentItem array
// 2. Enrich metadata with usage stats for cost tracking and optimization
// 3. Include stop reason for understanding why completion ended
//
// METADATA ENRICHMENT:
// We include token usage (input, output, total) because:
// - Enables cost calculation (tokens are billed)
// - Allows monitoring API usage patterns
// - Helps optimize max_tokens settings
// - Useful for debugging unexpected truncation (if stop_reason is "max_tokens")
//
// WHY EXTRACT FIRST CONTENT ITEM:
// Claude currently returns single text content item, but API supports multiple
// items for future multimodal responses. We extract first item for simplicity.
func (c *Client) convertResponse(resp *Response) *core.Response {
	// Extract text content from response
	// WHY: Response.Content is an array to support future multimodal content,
	// but for text responses we just need the first item's text
	var contentText string
	if len(resp.Content) > 0 {
		contentText = resp.Content[0].Text
	}

	// Build enriched metadata
	// WHY: Include comprehensive metadata for debugging, cost tracking, and optimization:
	// - model: Confirms which model actually responded (useful for debugging)
	// - stop_reason: Why completion ended (end_turn, max_tokens, stop_sequence)
	// - token counts: For cost calculation and usage monitoring
	meta := map[string]interface{}{
		"model":         resp.Model,
		"stop_reason":   resp.StopReason,
		"input_tokens":  resp.Usage.InputTokens,
		"output_tokens": resp.Usage.OutputTokens,
		"total_tokens":  resp.Usage.InputTokens + resp.Usage.OutputTokens,
	}

	// Include stop_sequence if present
	// WHY: If completion stopped due to custom stop sequence, include it for debugging
	if resp.StopSequence != nil {
		meta["stop_sequence"] = *resp.StopSequence
	}

	return &core.Response{
		Content: contentText,
		Meta:    meta,
	}
}

// Model returns the model name being used.
// WHY: Implements core.LLM interface requirement and allows users to verify
// which model they're using without accessing private fields. Useful for
// logging, debugging, and dynamic model selection logic.
func (c *Client) Model() string {
	return c.model
}
