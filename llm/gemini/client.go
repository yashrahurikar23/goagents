// Package gemini provides an implementation of the core.LLM interface for Google's Gemini models.
//
// PURPOSE:
// This package implements a production-ready HTTP client for Google's Gemini API,
// translating between GoAgents' universal core.LLM interface and Gemini's specific
// API requirements.
//
// WHY THIS EXISTS:
// Google's Gemini models have unique requirements that differ from other LLM providers:
// 1. Role Mapping: Gemini uses "model" instead of "assistant" for AI responses
// 2. Safety Features: Comprehensive safety ratings and prompt blocking for content policy
// 3. System Instructions: Separate systemInstruction field (similar to Anthropic)
// 4. Multi-Part Content: Parts arrays for future multimodal support (images, video, audio)
// 5. Candidate Array: API returns array of candidates (though we use first one)
// 6. URL-based API Key: API key is passed as query parameter, not in headers
//
// This client abstracts these differences, allowing users to work with Gemini models
// using the same interface as other LLM providers in GoAgents.
//
// KEY DESIGN DECISIONS:
// - Role name translation: Automatically converts "assistant" → "model" for Gemini
// - Safety checking: Validates PromptFeedback to detect blocked prompts before generation
// - System message extraction: Combines all system messages into systemInstruction
// - Metadata enrichment: Includes token usage and finish reason for tracking and debugging
// - Error handling: Parses Google's standard error format with code, message, and status
//
// METHODS:
// - New(opts): Creates a client with functional options for flexible configuration
// - Chat(ctx, messages): Send conversation history, get response (implements core.LLM)
// - Complete(ctx, prompt): Convenience method for single-turn completions
// - Model(): Returns the model name being used
//
// INTERNAL METHODS:
// - doRequest: Handles HTTP request/response with safety checking
// - convertMessages: Transforms core.Message to Gemini format, maps roles, extracts system
// - convertResponse: Transforms Gemini response to core.Response with token metadata
package gemini

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

// Client is the Google Gemini API client that implements core.LLM interface.
// WHY: Encapsulates all configuration and state needed to interact with Gemini API,
// including authentication, model selection, sampling parameters, and HTTP transport.
type Client struct {
	apiKey      string       // WHY: Required for authentication (passed as URL query param)
	baseURL     string       // WHY: Allows pointing to different endpoints (proxy, regional)
	model       string       // WHY: Specifies which Gemini model to use (1.5 Flash, 1.5 Pro, etc.)
	maxTokens   *int         // WHY: Pointer to distinguish "not set" from "set to 0"
	temperature *float64     // WHY: Pointer for optional temperature control
	topP        *float64     // WHY: Pointer for optional nucleus sampling
	topK        *int         // WHY: Pointer for optional top-k sampling
	httpClient  *http.Client // WHY: Allows custom HTTP configuration (timeouts, proxies, TLS)
}

// New creates a new Gemini client with the given options.
// WHY: Uses functional options pattern to provide flexible configuration with sensible
// defaults. This allows adding new options in the future without breaking existing code.
// Defaults to Gemini 1.5 Flash and 60-second timeout.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		model:   ModelGemini15Flash,
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
// WHY: This is the primary method for interacting with Gemini, supporting full conversation
// history for multi-turn dialogues. It handles:
// 1. System message extraction (Gemini requires separate systemInstruction field)
// 2. Role name translation ("assistant" → "model" for Gemini compatibility)
// 3. Generation config building (only if any params are set)
// 4. API communication with proper authentication
// 5. Safety checking (prompt blocking, content filtering)
// 6. Response conversion back to core format with enriched metadata
//
// BUSINESS LOGIC:
// - Validates API key is configured (fails fast if missing)
// - Extracts all system messages and combines them into systemInstruction
// - Only includes GenerationConfig if at least one parameter is set (cleaner API calls)
// - Converts "assistant" role to "model" (Gemini requirement)
// - Checks for blocked prompts before returning response
func (c *Client) Chat(ctx context.Context, messages []core.Message) (*core.Response, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("gemini: API key is required")
	}

	// Convert core.Message to Gemini format and extract system instructions
	// WHY: Gemini requires "model" role instead of "assistant", and system messages
	// must be in separate systemInstruction field
	contents, systemInstruction := c.convertMessages(messages)

	// Build generation config only if at least one parameter is set
	// WHY: Omitting the entire config object when not needed results in cleaner API calls
	// and lets Gemini use its default values
	var genConfig *GenerationConfig
	if c.temperature != nil || c.topP != nil || c.topK != nil || c.maxTokens != nil {
		genConfig = &GenerationConfig{
			Temperature:     c.temperature,
			TopP:            c.topP,
			TopK:            c.topK,
			MaxOutputTokens: c.maxTokens,
		}
	}

	// Build request with conversation and optional config
	req := GenerateContentRequest{
		Contents:          contents,
		GenerationConfig:  genConfig,
		SystemInstruction: systemInstruction,
	}

	// Make API call with context for cancellation and timeouts
	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gemini: request failed: %w", err)
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
// - One-off completions (e.g., "Translate this text")
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

// ChatStream implements the core.StreamingLLM interface for streaming chat completions.
//
// WHY THIS METHOD:
// - Implements core.StreamingLLM interface for framework compatibility
// - Returns a channel of core.StreamChunk for real-time token delivery
// - Handles context cancellation gracefully
// - Accumulates content across chunks for convenience
//
// IMPLEMENTATION:
// - Uses Gemini's streaming format (newline-delimited JSON)
// - Converts Gemini responses to core.StreamChunk format
// - Runs stream processing in a goroutine
// - Closes channel when stream completes or errors occur
func (c *Client) ChatStream(ctx context.Context, messages []core.Message, opts ...interface{}) (<-chan core.StreamChunk, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Convert messages to Gemini format
	geminiContents, systemInstruction := c.convertMessages(messages)

	// Build generation config
	var genConfig *GenerationConfig
	if c.maxTokens != nil || c.temperature != nil || c.topP != nil || c.topK != nil {
		genConfig = &GenerationConfig{
			MaxOutputTokens: c.maxTokens,
			Temperature:     c.temperature,
			TopP:            c.topP,
			TopK:            c.topK,
		}
	}

	req := GenerateContentRequest{
		Contents:          geminiContents,
		SystemInstruction: systemInstruction,
		GenerationConfig:  genConfig,
	}

	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL with streaming endpoint
	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s", c.baseURL, c.model, c.apiKey)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		httpResp.Body.Close()
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(bodyBytes))
	}

	// Create buffered channel for chunks
	chunkChan := make(chan core.StreamChunk, 10)

	// Track accumulated content and index
	var content string
	index := 0

	// Start goroutine to read streaming response
	go func() {
		defer close(chunkChan)
		defer httpResp.Body.Close()

		scanner := bufio.NewScanner(httpResp.Body)
		for scanner.Scan() {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				errorChunk := core.StreamChunk{
					Content:   content,
					Index:     index,
					Error:     ctx.Err(),
					Timestamp: time.Now(),
				}
				select {
				case chunkChan <- errorChunk:
				case <-ctx.Done():
				}
				return
			default:
			}

			line := scanner.Text()
			if line == "" {
				continue
			}

			// Parse the response chunk
			var resp GenerateContentResponse
			if err := json.Unmarshal([]byte(line), &resp); err != nil {
				// Try to parse as error response
				var errResp ErrorResponse
				if err := json.Unmarshal([]byte(line), &errResp); err == nil && errResp.Error.Code != 0 {
					errorChunk := core.StreamChunk{
						Content:   content,
						Index:     index,
						Error:     fmt.Errorf("API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code),
						Timestamp: time.Now(),
					}
					select {
					case chunkChan <- errorChunk:
					case <-ctx.Done():
					}
					return
				}
				// Ignore other parse errors
				continue
			}

			// Extract content from candidates
			if len(resp.Candidates) == 0 {
				continue
			}

			candidate := resp.Candidates[0]

			// Extract text from content parts
			var delta string
			if len(candidate.Content.Parts) > 0 {
				for _, part := range candidate.Content.Parts {
					delta += part.Text
				}
			}

			if delta != "" {
				content += delta
			}

			var finishReason string
			if candidate.FinishReason != "" && candidate.FinishReason != "FINISH_REASON_UNSPECIFIED" {
				finishReason = strings.ToLower(candidate.FinishReason)
			}

			// Create StreamChunk
			streamChunk := core.StreamChunk{
				Content:      content,
				Delta:        delta,
				Index:        index,
				FinishReason: finishReason,
				Metadata: map[string]interface{}{
					"model": c.model,
				},
				Timestamp: time.Now(),
			}

			index++

			// Send chunk on channel
			select {
			case <-ctx.Done():
				return
			case chunkChan <- streamChunk:
			}

			if finishReason != "" {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errorChunk := core.StreamChunk{
				Content:   content,
				Index:     index,
				Error:     fmt.Errorf("error reading stream: %w", err),
				Timestamp: time.Now(),
			}
			select {
			case chunkChan <- errorChunk:
			case <-ctx.Done():
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

// doRequest makes the HTTP request to the Gemini API.
// WHY: Centralizes all HTTP communication logic for consistency and maintainability.
// Handles the complete request/response cycle including:
// - Request marshaling and URL construction
// - API key authentication via query parameter
// - Context-aware cancellation and timeouts
// - Error parsing with Google's error format
// - Safety checking for blocked prompts
// - Response parsing
//
// KEY DESIGN DECISIONS:
// - URL-based API key: Gemini passes API key as query parameter (unlike header-based auth)
// - generateContent endpoint: Uses Gemini's generateContent RPC-style endpoint
// - PromptFeedback checking: Validates prompt wasn't blocked before returning response
// - Error detail extraction: Provides code, message, and status for debugging
func (c *Client) doRequest(ctx context.Context, req GenerateContentRequest) (*GenerateContentResponse, error) {
	// Marshal request body to JSON
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL with model and API key as query parameter
	// WHY: Gemini uses URL query parameter for API key (not Authorization header)
	// WHY: Uses RPC-style endpoint: models/{model}:generateContent
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)

	// Create HTTP request with context for cancellation
	// WHY: Context allows timeout enforcement and request cancellation
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	// WHY: Content-Type tells API we're sending JSON
	httpReq.Header.Set("Content-Type", "application/json")

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
	// WHY: Non-200 status indicates an error; parse Google's standard error format
	if httpResp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			// If we can't parse error response, return raw body
			return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(respBody))
		}
		// Return structured error with code, message, and status for debugging
		return nil, fmt.Errorf("API error: %s (code: %d, status: %s)",
			errResp.Error.Message, errResp.Error.Code, errResp.Error.Status)
	}

	// Parse successful response
	var resp GenerateContentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for blocked content due to safety filters
	// WHY: Gemini analyzes prompts for safety violations before generating responses.
	// If blocked, we need to return an error explaining why rather than an empty response.
	// BlockReason examples: SAFETY, PROHIBITED_CONTENT, etc.
	if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != "" {
		return nil, fmt.Errorf("prompt blocked: %s", resp.PromptFeedback.BlockReason)
	}

	return &resp, nil
}

// convertMessages converts core messages to Gemini format and extracts system instructions.
// WHY: Gemini has two unique requirements:
// 1. Role Mapping: Uses "model" instead of "assistant" for AI responses
// 2. System Instructions: System messages must be in separate systemInstruction field
//
// This differs from OpenAI (uses "assistant") and requires special handling.
//
// BUSINESS LOGIC:
// 1. Extract all messages with role="system" into systemInstruction
// 2. Convert user messages as-is with role="user"
// 3. Convert assistant messages to role="model" (Gemini requirement)
// 4. Wrap text in Parts arrays for future multimodal support
//
// WHY COMBINE SYSTEM MESSAGES:
// If multiple system messages exist, we combine them with double newlines to
// preserve structure while consolidating into Gemini's single systemInstruction field.
//
// WHY "MODEL" ROLE:
// Gemini's API design uses "model" to represent the AI's role in conversation,
// distinguishing it from "user". Other providers use "assistant" or "ai".
func (c *Client) convertMessages(messages []core.Message) ([]Content, *Content) {
	var contents []Content
	var systemParts []Part

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			// System messages go into separate SystemInstruction field
			// WHY: Gemini treats system instructions separately from conversation history
			systemParts = append(systemParts, Part{Text: msg.Content})
		case "user":
			// User messages map directly
			contents = append(contents, Content{
				Role:  "user",
				Parts: []Part{{Text: msg.Content}},
			})
		case "assistant":
			// Assistant role must be converted to "model" for Gemini
			// WHY: Gemini's API uses "model" instead of "assistant"
			contents = append(contents, Content{
				Role:  "model",
				Parts: []Part{{Text: msg.Content}},
			})
		}
	}

	// Build systemInstruction if any system messages exist
	var systemInstruction *Content
	if len(systemParts) > 0 {
		// Combine multiple system messages into one
		// WHY: Gemini accepts single systemInstruction, so we join multiple
		// system messages with double newlines to preserve structure
		if len(systemParts) > 1 {
			var combined strings.Builder
			for i, part := range systemParts {
				if i > 0 {
					combined.WriteString("\n\n")
				}
				combined.WriteString(part.Text)
			}
			systemParts = []Part{{Text: combined.String()}}
		}
		systemInstruction = &Content{
			Parts: systemParts,
		}
	}

	return contents, systemInstruction
}

// convertResponse converts Gemini response to core format with enriched metadata.
// WHY: Gemini's response structure differs from core.Response, so we need to:
// 1. Extract text from first candidate (Gemini returns array, we use first)
// 2. Combine multiple Parts into single string (for future multimodal support)
// 3. Enrich metadata with usage stats for cost tracking and optimization
// 4. Include finish reason for understanding why generation stopped
//
// METADATA ENRICHMENT:
// We include token usage (prompt, completion, total) because:
// - Enables cost calculation (Google bills by tokens)
// - Allows monitoring API usage patterns
// - Helps optimize generation parameters
// - Useful for debugging unexpected truncation
//
// WHY FIRST CANDIDATE:
// Gemini API can return multiple candidates for ranking/selection in the future,
// but currently returns one. We extract the first for simplicity.
//
// WHY COMBINE PARTS:
// Response Parts is an array to support future multimodal content (text + images),
// but for text-only responses we combine all parts into single string.
func (c *Client) convertResponse(resp *GenerateContentResponse) *core.Response {
	// Extract text from first candidate
	var contentText string
	var finishReason string

	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		finishReason = candidate.FinishReason

		// Combine all parts from response
		// WHY: Parts array supports multimodal content, but for text-only
		// we join all text parts into single string
		var parts []string
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				parts = append(parts, part.Text)
			}
		}
		contentText = strings.Join(parts, "")
	}

	// Build enriched metadata
	// WHY: Include comprehensive metadata for debugging, cost tracking, and optimization:
	// - model: Confirms which model responded
	// - finish_reason: Why generation stopped (STOP, MAX_TOKENS, SAFETY, etc.)
	// - token counts: For cost calculation and usage monitoring
	meta := map[string]interface{}{
		"model":         c.model,
		"finish_reason": finishReason,
	}

	// Include token usage if available
	// WHY: Token counts enable cost tracking and help users optimize their usage
	if resp.UsageMetadata.TotalTokenCount > 0 {
		meta["prompt_tokens"] = resp.UsageMetadata.PromptTokenCount
		meta["completion_tokens"] = resp.UsageMetadata.CandidatesTokenCount
		meta["total_tokens"] = resp.UsageMetadata.TotalTokenCount
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
