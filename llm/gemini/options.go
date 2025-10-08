// Package gemini provides functional options for configuring the Google Gemini API client.
//
// PURPOSE:
// This package implements the functional options pattern to provide flexible, backward-compatible
// client configuration without breaking existing code as new options are added.
//
// WHY THIS EXISTS:
// The functional options pattern is chosen over builder pattern or config structs because:
// 1. Clean, readable API: NewClient(WithAPIKey("key"), WithModel("gemini-1.5-pro"))
// 2. Backward compatible: New options can be added without breaking existing code
// 3. Optional parameters: Sensible defaults are used when options aren't provided
// 4. Idiomatic Go: This is the standard pattern for complex configuration in Go
//
// KEY DESIGN DECISIONS:
//   - Each option returns a closure: Enables deferred application during NewClient()
//   - Options applied in order: Later options can override earlier ones if needed
//   - Pointer types for sampling: Distinguishes "not set" (nil) from "set to zero" (0)
//   - Custom HTTP client support: Allows users to inject their own client with middleware,
//     proxies, custom TLS, or connection pooling
//
// AVAILABLE OPTIONS:
// - Authentication: WithAPIKey (required)
// - Model Selection: WithModel (defaults to Gemini 1.5 Flash)
// - Network: WithBaseURL, WithHTTPClient, WithTimeout
// - Generation: WithMaxTokens, WithTemperature, WithTopP, WithTopK
package gemini

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the Gemini client.
// WHY: Using function closures allows options to be applied during client construction,
// enabling flexible configuration without exposing internal client fields.
type Option func(*Client)

// WithAPIKey sets the Google API key for authentication.
// WHY: API key authentication is required by Google's Gemini API. This is kept as a
// separate option rather than a constructor parameter to maintain consistent functional
// options pattern and allow future addition of other auth methods without breaking changes.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithModel sets the Gemini model to use for completions.
// WHY: Different Gemini models have different capabilities, speeds, and costs:
// - Gemini 1.5 Flash: Fast and efficient (default)
// - Gemini 1.5 Pro: Most capable for complex reasoning
// - Gemini 2.0 Flash: Latest experimental features
// Exposing this as an option allows users to choose the right model for their use case
// without creating separate clients.
func WithModel(model string) Option {
	return func(c *Client) {
		c.model = model
	}
}

// WithBaseURL sets a custom base URL for the API.
// WHY: Allows users to:
// 1. Use proxy servers or custom API gateways
// 2. Point to test/staging environments (if Google provides them)
// 3. Use regional endpoints when available
// This is essential for enterprise deployments with custom networking requirements.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
// WHY: Gives users full control over HTTP behavior without us having to expose every
// possible configuration. Users can provide clients with:
// - Custom TLS certificates for corporate proxies
// - Proxy configuration for network security
// - Request/response middleware for logging/monitoring
// - Connection pooling settings for performance
// - Custom transport layer for advanced networking
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the HTTP timeout for API requests.
// WHY: Creates a new HTTP client with the specified timeout if one doesn't exist.
// Gemini requests can take time (especially for long contexts or complex reasoning),
// so timeout needs to be configurable. Default is 60 seconds, but users may need:
// - Longer timeouts for complex reasoning or long context windows
// - Shorter timeouts for latency-sensitive applications
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
	}
}

// WithMaxTokens sets the maximum output tokens for completion responses.
// WHY: Token limits control:
// 1. Response length (prevents overly long responses)
// 2. API costs (tokens are billed by Google)
// 3. Processing time (fewer tokens = faster response)
// Different use cases need different limits: chatbots might use 500-1000, while
// document generation might use 2000-4000. Gemini defaults to 2048.
func WithMaxTokens(maxTokens int) Option {
	return func(c *Client) {
		c.maxTokens = &maxTokens
	}
}

// WithTemperature sets the temperature for sampling.
// WHY: Temperature controls randomness and creativity in responses:
// - 0.0: Deterministic, focused responses (good for factual Q&A, code generation)
// - 0.7: Balanced creativity and coherence (good for general conversation)
// - 1.0+: More creative and diverse outputs (good for brainstorming, creative writing)
// Using pointer allows distinguishing "not set" (nil, uses Gemini's default)
// from "set to 0.0" (explicitly requesting deterministic output).
func WithTemperature(temperature float64) Option {
	return func(c *Client) {
		c.temperature = &temperature
	}
}

// WithTopP sets the top-p (nucleus sampling) parameter.
// WHY: Top-p controls diversity by only sampling from the top probability mass:
// - 0.9: Sample from top 90% probability mass (good balance)
// - 0.5: More focused, conservative responses
// - 1.0: Consider all possible tokens
// This is an alternative to temperature. Using both simultaneously is generally not
// recommended - choose one or the other. Pointer type allows "not set" vs "set to 0.0".
func WithTopP(topP float64) Option {
	return func(c *Client) {
		c.topP = &topP
	}
}

// WithTopK sets the top-k sampling parameter.
// WHY: Top-k limits sampling to the k most likely next tokens:
// - Lower values (1-10): More focused, deterministic responses
// - Higher values (40-100): More diverse responses
// This is another alternative to temperature for controlling randomness.
// Google's models often work well with TopK. Pointer type allows "not set" vs "set to 0".
func WithTopK(topK int) Option {
	return func(c *Client) {
		c.topK = &topK
	}
}
