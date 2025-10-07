/*
Package openai provides a comprehensive OpenAI API client for Go.

PURPOSE:
This package implements a complete client for interacting with OpenAI's API endpoints,
including chat completions, embeddings, moderation, and streaming responses. It serves
as the primary LLM provider for the GoAgent framework.

WHY THIS EXISTS:
- Provides type-safe, idiomatic Go interface to OpenAI's REST API
- Implements automatic retry logic with exponential backoff for reliability
- Supports all OpenAI features: function calling, vision, JSON mode, streaming
- Integrates seamlessly with core.LLM interface for framework compatibility

KEY DESIGN DECISIONS:
- Functional options pattern: Enables flexible client configuration without breaking changes
- Separate types file: Keeps API contracts clear and separate from implementation
- Streaming via callbacks: Allows real-time processing without buffering entire response
- Automatic retries: Handles transient failures (rate limits, server errors) transparently
- Context-aware: All operations respect context cancellation and timeouts

MAIN COMPONENTS:
- Client: Main client with configuration and HTTP handling
- ChatCompletionRequest/Response: Chat API request/response types
- EmbeddingRequest/Response: Embeddings API types
- ModerationRequest/Response: Content moderation types
- StreamOptions: Streaming configuration with callbacks

USAGE PATTERNS:
1. Basic chat: client.Complete(ctx, "prompt")
2. Advanced chat: client.Chat(ctx, messages) with tool calls
3. Streaming: client.CreateChatCompletionStream(ctx, req, streamOpts)
4. Embeddings: client.CreateEmbedding(ctx, embeddingReq)
5. Vision: UserMessageWithImage(text, imageURL)
*/
package openai

import "time"

// ChatCompletionRequest represents a request to the chat completions API.
//
// WHY THESE FIELDS:
// - Model: Allows per-request model override (e.g., gpt-4 vs gpt-3.5-turbo)
// - Temperature/TopP: Control randomness vs determinism in responses
// - MaxTokens: Prevent unexpectedly long/expensive responses
// - Tools/Functions: Enable function calling for agent capabilities
// - ResponseFormat: Support JSON mode for structured outputs
// - ReasoningEffort/Verbosity: GPT-5 specific controls for reasoning depth
type ChatCompletionRequest struct {
	Model            string             `json:"model"`
	Messages         []ChatMessage      `json:"messages"`
	MaxTokens        *int               `json:"max_tokens,omitempty"`
	Temperature      *float64           `json:"temperature,omitempty"`
	TopP             *float64           `json:"top_p,omitempty"`
	N                *int               `json:"n,omitempty"`
	Stream           bool               `json:"stream,omitempty"`
	Stop             []string           `json:"stop,omitempty"`
	PresencePenalty  *float64           `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64           `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float64 `json:"logit_bias,omitempty"`
	User             string             `json:"user,omitempty"`
	ResponseFormat   *ResponseFormat    `json:"response_format,omitempty"`
	Seed             *int               `json:"seed,omitempty"`
	Tools            []Tool             `json:"tools,omitempty"`
	ToolChoice       interface{}        `json:"tool_choice,omitempty"` // string or object
	Functions        []Function         `json:"functions,omitempty"`
	FunctionCall     interface{}        `json:"function_call,omitempty"` // string or object
	ReasoningEffort  string             `json:"reasoning_effort,omitempty"`
	Verbosity        string             `json:"verbosity,omitempty"`
}

// ChatMessage represents a message in a conversation.
type ChatMessage struct {
	Role         string        `json:"role"`              // system, user, assistant, tool, function
	Content      interface{}   `json:"content,omitempty"` // string or []ContentPart
	Name         string        `json:"name,omitempty"`
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`
	ToolCallID   string        `json:"tool_call_id,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

// ContentPart represents a part of message content (text or image).
type ContentPart struct {
	Type     string    `json:"type"` // text, image_url
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL in a message.
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // low, high, auto
}

// ResponseFormat specifies the format of the response.
type ResponseFormat struct {
	Type string `json:"type"` // text, json_object, json_schema
}

// Tool represents a tool that the model can call.
type Tool struct {
	Type     string    `json:"type"` // function
	Function *Function `json:"function,omitempty"`
}

// Function represents a function definition.
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Strict      bool                   `json:"strict,omitempty"`
}

// ToolCall represents a tool call made by the model.
type ToolCall struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"` // function
	Function *FunctionCall `json:"function,omitempty"`
}

// FunctionCall represents a function call.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionResponse represents a response from the chat completions API.
type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             *Usage   `json:"usage,omitempty"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
}

// Choice represents a completion choice.
type Choice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"` // For streaming
	FinishReason string       `json:"finish_reason,omitempty"`
	LogProbs     *LogProbs    `json:"logprobs,omitempty"`
}

// LogProbs represents log probabilities.
type LogProbs struct {
	Content []TokenLogProb `json:"content"`
}

// TokenLogProb represents log probability for a token.
type TokenLogProb struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
}

// Usage represents token usage statistics.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionStreamResponse represents a streaming response chunk.
type ChatCompletionStreamResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
}

// EmbeddingRequest represents a request to the embeddings API.
type EmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"`                     // string or []string
	EncodingFormat string      `json:"encoding_format,omitempty"` // float, base64
	Dimensions     *int        `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

// EmbeddingResponse represents a response from the embeddings API.
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  *Usage      `json:"usage"`
}

// Embedding represents a single embedding.
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// ModerationRequest represents a request to the moderation API.
type ModerationRequest struct {
	Model string      `json:"model,omitempty"`
	Input interface{} `json:"input"` // string, []string, or []ModerationInput
}

// ModerationInput represents multimodal moderation input.
type ModerationInput struct {
	Type     string    `json:"type"` // text, image_url
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ModerationResponse represents a response from the moderation API.
type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

// ModerationResult represents a moderation result.
type ModerationResult struct {
	Flagged                   bool                `json:"flagged"`
	Categories                map[string]bool     `json:"categories"`
	CategoryScores            map[string]float64  `json:"category_scores"`
	CategoryAppliedInputTypes map[string][]string `json:"category_applied_input_types,omitempty"`
}

// ModelListResponse represents a list of available models.
type ModelListResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// Model represents an OpenAI model.
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error *APIError `json:"error"`
}

// APIError represents an API error.
type APIError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   interface{} `json:"param"`
	Code    interface{} `json:"code"`
}

// StreamOptions configures streaming behavior.
type StreamOptions struct {
	OnChunk    func(*ChatCompletionStreamResponse) error
	OnComplete func() error
	OnError    func(error)
	BufferSize int
	Timeout    time.Duration
}
