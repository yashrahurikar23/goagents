/*
Package anthropic provides an Anthropic Claude LLM client for GoAgents.

PURPOSE:
This package implements the core.LLM interface for Anthropic's Claude models,
enabling AI agents to use Claude's advanced capabilities including 200K context
windows, system prompts, and sophisticated reasoning.

WHY THIS EXISTS:
Anthropic Claude offers unique advantages for agent applications:
- Exceptional reasoning and code understanding
- Large context windows (200K tokens) for complex tasks
- Strong safety and reliability characteristics
- System prompts for precise behavior control

KEY DESIGN DECISIONS:
- System prompts handled separately from messages (Anthropic API requirement)
- Pointer types for optional parameters to distinguish zero values from unset
- Model constants with dates for version tracking and compatibility
- Streaming support prepared but not yet exposed (infrastructure ready)

API STRUCTURE:
The Anthropic API uses a messages-based format where system prompts are a
separate field rather than part of the messages array. This differs from
OpenAI and requires special handling in message conversion.

AVAILABLE MODELS:
- Claude 3.5 Sonnet: Latest, best balance of performance and speed
- Claude 3.5 Haiku: Fastest, most cost-effective for simple tasks
- Claude 3 Opus: Most capable, best for complex reasoning
- Claude 3 Sonnet: Balanced performance and cost
- Claude 3 Haiku: Fast and affordable baseline
*/
package anthropic

import "time"

// Message represents a single message in a conversation for Anthropic's API.
//
// WHY THIS STRUCTURE:
// - Content as array supports future multimodal capabilities (images, etc.)
// - Matches Anthropic's API format exactly for reliable serialization
type Message struct {
	Role    string        `json:"role"`    // "user" or "assistant"
	Content []ContentItem `json:"content"` // Array to support multiple content types
}

// ContentItem represents a single content item (text, image, etc.).
//
// WHY ARRAY OF ITEMS:
// Anthropic's API supports rich content including text, images, and tools.
// The array structure allows mixing content types in a single message.
type ContentItem struct {
	Type string `json:"type"`           // "text", "image", etc.
	Text string `json:"text,omitempty"` // Text content when type is "text"
}

// Request represents the Anthropic API request structure.
//
// WHY POINTERS FOR OPTIONALS:
// Using pointers (*float64, *int) allows distinguishing between:
// - Not set (nil) - API uses its defaults
// - Explicitly set to zero (&0) - API uses zero value
// This matters for parameters like Temperature where 0 is valid but different from unset.
//
// BUSINESS LOGIC:
// - MaxTokens is required by Anthropic API (no default)
// - System prompts are separate from messages (API design)
// - Temperature/TopP/TopK control randomness for different use cases
type Request struct {
	Model         string    `json:"model"`                    // Model identifier (e.g., "claude-3-5-sonnet-20241022")
	Messages      []Message `json:"messages"`                 // Conversation history
	MaxTokens     int       `json:"max_tokens"`               // Maximum tokens to generate (required)
	Temperature   *float64  `json:"temperature,omitempty"`    // 0.0-1.0, controls randomness
	TopP          *float64  `json:"top_p,omitempty"`          // 0.0-1.0, nucleus sampling
	TopK          *int      `json:"top_k,omitempty"`          // Top-k sampling
	StopSequences []string  `json:"stop_sequences,omitempty"` // Sequences that stop generation
	Stream        bool      `json:"stream,omitempty"`         // Enable streaming (not yet exposed)
	System        string    `json:"system,omitempty"`         // System prompt (separate from messages)
}

// Response represents the Anthropic API response structure.
//
// WHY THIS STRUCTURE:
// Matches Anthropic's response format exactly, including metadata needed for:
// - Token usage tracking (billing and rate limits)
// - Stop reason analysis (completion vs truncation)
// - Response identification and logging
type Response struct {
	ID           string            `json:"id"`            // Unique response ID
	Type         string            `json:"type"`          // Response type ("message")
	Role         string            `json:"role"`          // Always "assistant" for responses
	Content      []ResponseContent `json:"content"`       // Generated content
	Model        string            `json:"model"`         // Model used for generation
	StopReason   string            `json:"stop_reason"`   // Why generation stopped
	StopSequence *string           `json:"stop_sequence"` // Which stop sequence triggered (if any)
	Usage        Usage             `json:"usage"`         // Token usage statistics
}

// ResponseContent represents content in the response.
type ResponseContent struct {
	Type string `json:"type"` // Content type ("text")
	Text string `json:"text"` // Generated text
}

// Usage represents token usage information.
//
// WHY TRACK USAGE:
// Essential for:
// - Cost tracking and billing
// - Rate limit monitoring
// - Performance optimization
// - Context window management
type Usage struct {
	InputTokens  int `json:"input_tokens"`  // Tokens in the prompt
	OutputTokens int `json:"output_tokens"` // Tokens in the response
}

// ErrorResponse represents an error from the Anthropic API.
//
// WHY NESTED STRUCTURE:
// Matches Anthropic's error format which wraps the actual error details
// in a nested "error" object. This structure enables proper error parsing
// and type-specific error handling.
type ErrorResponse struct {
	Type  string `json:"type"` // Error category
	Error struct {
		Type    string `json:"type"`    // Specific error type
		Message string `json:"message"` // Human-readable error message
	} `json:"error"`
}

// StreamEvent represents a streaming event from Anthropic.
//
// WHY PREPARED BUT NOT EXPOSED:
// Streaming infrastructure is ready for future v0.3.0 feature.
// Types defined now ensure API compatibility when streaming is enabled.
type StreamEvent struct {
	Type    string      `json:"type"`              // Event type
	Message Response    `json:"message,omitempty"` // Complete message (start/end events)
	Index   int         `json:"index,omitempty"`   // Content index
	Delta   StreamDelta `json:"delta,omitempty"`   // Incremental update
}

// StreamDelta represents incremental updates in a stream.
type StreamDelta struct {
	Type       string `json:"type,omitempty"`        // Delta type
	Text       string `json:"text,omitempty"`        // Incremental text
	StopReason string `json:"stop_reason,omitempty"` // Final stop reason
}

// Model constants for Anthropic Claude models.
//
// WHY DATE-VERSIONED CONSTANTS:
// - Anthropic versions models by date (YYYYMMDD format)
// - Explicit versions ensure reproducible behavior across updates
// - Constants prevent typos and enable IDE auto-completion
//
// MODEL SELECTION GUIDE:
// - Use Claude 3.5 Sonnet for best balance (recommended default)
// - Use Claude 3.5 Haiku for speed and cost optimization
// - Use Claude 3 Opus for maximum reasoning capability
const (
	ModelClaude35Sonnet = "claude-3-5-sonnet-20241022" // Latest Sonnet, best general purpose
	ModelClaude3Opus    = "claude-3-opus-20240229"     // Most capable, expensive
	ModelClaude3Sonnet  = "claude-3-sonnet-20240229"   // Balanced, good for most tasks
	ModelClaude3Haiku   = "claude-3-haiku-20240307"    // Fast and economical
	ModelClaude35Haiku  = "claude-3-5-haiku-20241022"  // Latest Haiku, fastest
)

// DefaultMaxTokens is the default maximum tokens for completion.
//
// WHY 4096:
// Balances between:
// - Allowing substantial responses (full page of text)
// - Preventing excessive costs for unintentional long generations
// - Compatibility with most use cases
const DefaultMaxTokens = 4096

// DefaultTimeout is the default HTTP timeout.
//
// WHY 60 SECONDS:
// - Claude can take 30-45s for complex requests with large context
// - Allows time for rate limiting retries
// - Prevents indefinite hangs while being generous for slow requests
const DefaultTimeout = 60 * time.Second

// DefaultBaseURL is the default Anthropic API base URL.
const DefaultBaseURL = "https://api.anthropic.com/v1"

// DefaultAPIVersion is the default API version header.
//
// WHY VERSION HEADER:
// Anthropic uses a version header instead of URL versioning.
// This allows API evolution while maintaining backward compatibility.
const DefaultAPIVersion = "2023-06-01"
