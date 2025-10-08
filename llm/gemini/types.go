// Package gemini provides type definitions for Google's Gemini API integration.
//
// PURPOSE:
// This package defines the complete type system for interacting with Google's Gemini models,
// mapping between GoAgents' universal interface and Gemini's specific API requirements.
//
// WHY THIS EXISTS:
// Gemini has unique characteristics that differ from other LLM providers:
// 1. Role Mapping: Uses "model" instead of "assistant" for AI responses
// 2. Safety Features: Comprehensive safety ratings for content filtering
// 3. Multi-Part Structure: Content is split into Parts arrays for multimodal support
// 4. System Instructions: Separate systemInstruction field (similar to Anthropic)
// 5. Candidate Responses: API can return multiple response candidates (though we use first)
//
// These types ensure proper serialization/deserialization when communicating with Gemini API.
//
// KEY DESIGN DECISIONS:
//   - Parts Array: Even text-only content uses Part arrays to support future multimodal
//     content (images, video, audio) without breaking changes
//   - Pointer Types: Optional fields use pointers to distinguish between "not set" (nil)
//     and "set to zero value" (0, 0.0)
//   - Safety Ratings: Included throughout to enable content filtering and policy compliance
//   - PromptFeedback: Separate from response to handle blocked prompts before generation
//
// API STRUCTURE:
// Request: Contents (array of messages) + GenerationConfig + SystemInstruction
// Response: Candidates (array) + UsageMetadata + PromptFeedback
// Each Candidate: Content + FinishReason + SafetyRatings
//
// AVAILABLE MODELS:
// - Gemini 2.0 Flash (experimental): Latest experimental model
// - Gemini 1.5 Flash: Fast, efficient model for general use
// - Gemini 1.5 Flash 8B: Smaller, faster variant
// - Gemini 1.5 Pro: Most capable model for complex tasks
// - Gemini Pro: Legacy stable model
// - Gemini Pro Vision: Legacy vision model (deprecated in favor of 1.5+)
package gemini

import "time"

// Content represents content in a message (request or response).
// WHY: Gemini uses a multi-part structure to support multimodal content.
// Even for text-only conversations, we use this structure for consistency
// and forward compatibility with images, video, and audio.
type Content struct {
	Role  string `json:"role,omitempty"` // WHY: "user" or "model" (not "assistant" like other APIs)
	Parts []Part `json:"parts"`          // WHY: Array enables multimodal content mixing
}

// Part represents a single part of content (text, image, etc.).
// WHY: Gemini's multimodal architecture requires content to be split into parts.
// Currently we only use Text, but future versions will support ImagePart, VideoPart, etc.
type Part struct {
	Text string `json:"text"` // WHY: Text content for this part
}

// GenerateContentRequest represents a request to generate content.
// WHY: This is the main request structure for Gemini's generateContent endpoint.
// Includes conversation history (Contents), generation parameters (GenerationConfig),
// and optional system instructions for behavior control.
type GenerateContentRequest struct {
	Contents          []Content         `json:"contents"`                    // WHY: Full conversation history for context
	GenerationConfig  *GenerationConfig `json:"generationConfig,omitempty"`  // WHY: Pointer allows omitting entire config if using defaults
	SystemInstruction *Content          `json:"systemInstruction,omitempty"` // WHY: Separate system instruction like Anthropic (not in Contents array)
}

// GenerationConfig contains parameters that control text generation behavior.
// WHY: Provides fine-grained control over randomness, length, and stopping conditions.
// All fields are pointers to distinguish "not set" from "set to zero".
type GenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`     // WHY: Controls randomness (0.0=deterministic, 1.0+=creative)
	TopP            *float64 `json:"topP,omitempty"`            // WHY: Nucleus sampling parameter
	TopK            *int     `json:"topK,omitempty"`            // WHY: Top-k sampling parameter
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"` // WHY: Controls response length and costs
	StopSequences   []string `json:"stopSequences,omitempty"`   // WHY: Custom sequences that stop generation
}

// GenerateContentResponse represents the API response from Gemini.
// WHY: Gemini returns multiple candidates (for future features), usage metadata (for billing),
// and prompt feedback (for safety/blocking info).
type GenerateContentResponse struct {
	Candidates     []Candidate     `json:"candidates"`               // WHY: Array allows multiple responses (we use first)
	UsageMetadata  UsageMetadata   `json:"usageMetadata,omitempty"`  // WHY: Token counts for cost tracking
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"` // WHY: Safety info about the prompt itself
}

// Candidate represents a single generated response candidate.
// WHY: Gemini can generate multiple candidates for ranking/selection.
// Each candidate has content, finish reason, and safety ratings.
type Candidate struct {
	Content       Content        `json:"content"`                 // WHY: The actual generated response
	FinishReason  string         `json:"finishReason,omitempty"`  // WHY: Why generation stopped (STOP, MAX_TOKENS, SAFETY, etc.)
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"` // WHY: Content safety analysis for filtering
}

// SafetyRating represents content safety analysis.
// WHY: Google provides comprehensive safety ratings to help filter harmful content.
// Categories include harassment, hate speech, sexually explicit, and dangerous content.
// This enables policy-compliant applications.
type SafetyRating struct {
	Category    string `json:"category"`    // WHY: Type of safety concern (HARM_CATEGORY_HARASSMENT, etc.)
	Probability string `json:"probability"` // WHY: Likelihood (NEGLIGIBLE, LOW, MEDIUM, HIGH)
}

// UsageMetadata contains token usage information for cost tracking and optimization.
// WHY: Gemini bills by tokens, so we track:
// - PromptTokenCount: Input tokens (conversation history + system instruction)
// - CandidatesTokenCount: Output tokens (generated response)
// - TotalTokenCount: Sum for easy cost calculation
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`     // WHY: Input cost tracking
	CandidatesTokenCount int `json:"candidatesTokenCount"` // WHY: Output cost tracking
	TotalTokenCount      int `json:"totalTokenCount"`      // WHY: Total cost calculation
}

// PromptFeedback contains feedback about the prompt before generation.
// WHY: Gemini analyzes prompts for safety violations before generating responses.
// If prompt is blocked, BlockReason explains why. This is separate from response
// safety ratings because prompt can be blocked before any content is generated.
type PromptFeedback struct {
	BlockReason   string         `json:"blockReason,omitempty"`   // WHY: Why prompt was blocked (if it was)
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"` // WHY: Safety analysis of the prompt
}

// ErrorResponse represents an error from the Gemini API.
// WHY: Gemini uses standard Google API error format with nested error object.
type ErrorResponse struct {
	Error APIError `json:"error"` // WHY: Nested structure matches Google's standard error format
}

// APIError represents error details from Gemini API.
// WHY: Provides comprehensive error information including HTTP code, message, and status.
type APIError struct {
	Code    int    `json:"code"`    // WHY: HTTP status code
	Message string `json:"message"` // WHY: Human-readable error message
	Status  string `json:"status"`  // WHY: Error status string (INVALID_ARGUMENT, PERMISSION_DENIED, etc.)
}

// Model constants for Google Gemini models.
// WHY: Gemini offers multiple models with different capabilities and costs:
const (
	ModelGemini20Flash   = "gemini-2.0-flash-exp" // WHY: Latest experimental - fastest, most capable (experimental)
	ModelGemini15Flash   = "gemini-1.5-flash"     // WHY: Production stable - fast, efficient, good for most use cases
	ModelGemini15Flash8B = "gemini-1.5-flash-8b"  // WHY: Smaller variant - faster, lower cost, good for simple tasks
	ModelGemini15Pro     = "gemini-1.5-pro"       // WHY: Most capable - best for complex reasoning and long contexts
	ModelGeminiPro       = "gemini-pro"           // WHY: Legacy stable model (consider migrating to 1.5)
	ModelGeminiProVision = "gemini-pro-vision"    // WHY: Legacy vision model (deprecated, use 1.5+ instead)
)

// DefaultMaxTokens is the default maximum output tokens.
// WHY: 2048 tokens (~1500 words) provides a good balance:
// - Long enough for detailed responses
// - Short enough to control costs
// - Typical for conversational AI
const DefaultMaxTokens = 2048

// DefaultTimeout is the default HTTP timeout.
// WHY: 60 seconds accounts for:
// - Network latency
// - API processing time (especially for long contexts)
// - Generation time for long responses
// Gemini can be slower for complex reasoning tasks.
const DefaultTimeout = 60 * time.Second

// DefaultBaseURL is the default Gemini API base URL.
// WHY: Points to v1beta API which includes latest features while maintaining stability.
// Production users may want to use v1 once it's available.
const DefaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
