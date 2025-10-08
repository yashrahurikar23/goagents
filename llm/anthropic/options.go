// Package anthropic provides functional options for configuring the Anthropic Claude API client.
//
// PURPOSE:
// This package implements the functional options pattern to provide flexible, backward-compatible
// client configuration without breaking existing code as new options are added.
//
// WHY THIS EXISTS:
// The functional options pattern is chosen over builder pattern or config structs because:
// 1. It provides clean, readable API: NewClient(WithAPIKey("key"), WithModel("claude-3"))
// 2. It allows adding new options without breaking existing code (backward compatible)
// 3. It enables optional parameters with sensible defaults
// 4. It's the idiomatic Go pattern for complex configuration
//
// KEY DESIGN DECISIONS:
// - Each option returns a closure that modifies the client: Enables deferred application during NewClient()
// - Options are applied in order: Later options can override earlier ones if needed
// - Pointer types for sampling params: Distinguishes "not set" (nil) from "set to zero" (0.0)
// - Custom HTTP client support: Allows users to provide their own client with custom middleware, proxies, TLS config
//
// AVAILABLE OPTIONS:
// - Authentication: WithAPIKey (required)
// - Model Selection: WithModel (defaults to Claude 3.5 Sonnet)
// - Network: WithBaseURL, WithHTTPClient, WithTimeout
// - Generation: WithMaxTokens, WithTemperature, WithTopP, WithTopK
// - API: WithAPIVersion
package anthropic

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the Anthropic client.
// WHY: Using function closures allows options to be applied during client construction,
// enabling flexible configuration without exposing internal client fields.
type Option func(*Client)

// WithAPIKey sets the API key for authentication with Anthropic's API.
// WHY: API key authentication is required by Anthropic. This is kept as a separate option
// rather than a constructor parameter to maintain consistent functional options pattern
// and allow future addition of other auth methods (e.g., OAuth) without breaking changes.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithModel sets the Claude model to use for completions.
// WHY: Different models have different capabilities, speeds, and costs. Exposing this
// as an option allows users to choose the right model for their use case without
// creating separate clients. Defaults to Claude 3.5 Sonnet if not specified.
func WithModel(model string) Option {
	return func(c *Client) {
		c.model = model
	}
}

// WithBaseURL sets a custom base URL for the API.
// WHY: Allows users to:
// 1. Use proxy servers or custom API gateways
// 2. Point to test/staging environments
// 3. Use regional endpoints if available
// This is essential for enterprise deployments with custom networking requirements.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
// WHY: Gives users full control over HTTP behavior without us having to expose every possible
// configuration. Users can provide clients with:
// - Custom TLS certificates
// - Proxy configuration
// - Request/response middleware
// - Connection pooling settings
// - Custom transport layer
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the HTTP timeout for API requests.
// WHY: Creates a new HTTP client with the specified timeout if one doesn't exist.
// Claude requests can take time (especially for long contexts), so timeout needs to be
// configurable. Default is 60 seconds, but users may need longer for complex reasoning
// tasks or shorter for latency-sensitive applications.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
	}
}

// WithMaxTokens sets the maximum tokens for completion responses.
// WHY: Token limits control:
// 1. Response length (prevents overly long responses)
// 2. API costs (tokens are billed)
// 3. Processing time (fewer tokens = faster response)
// Different use cases need different limits: chatbots might use 500-1000, while
// document generation might use 4000-8000.
func WithMaxTokens(maxTokens int) Option {
	return func(c *Client) {
		c.maxTokens = maxTokens
	}
}

// WithTemperature sets the temperature for sampling.
// WHY: Temperature controls randomness/creativity in responses:
// - 0.0: Deterministic, focused responses (good for factual Q&A, code generation)
// - 0.7: Balanced creativity and coherence (good for general conversation)
// - 1.0+: More creative and diverse outputs (good for brainstorming, creative writing)
// Using pointer allows distinguishing "not set" (nil) from "set to 0.0" (deterministic).
func WithTemperature(temperature float64) Option {
	return func(c *Client) {
		c.temperature = &temperature
	}
}

// WithTopP sets the top-p (nucleus sampling) parameter.
// WHY: Top-p controls diversity by only sampling from the top probability mass:
// - 0.9: Sample from top 90% probability mass (good balance)
// - 0.5: More focused responses
// - 1.0: Consider all tokens
// This is an alternative to temperature. Using both simultaneously is generally not recommended.
// Pointer type allows "not set" vs "set to 0.0".
func WithTopP(topP float64) Option {
	return func(c *Client) {
		c.topP = &topP
	}
}

// WithTopK sets the top-k sampling parameter.
// WHY: Top-k limits sampling to the k most likely next tokens:
// - Lower values (1-10): More focused, deterministic
// - Higher values (50-100): More diverse
// This is another alternative to temperature for controlling randomness.
// Pointer type allows "not set" vs "set to 0".
func WithTopK(topK int) Option {
	return func(c *Client) {
		c.topK = &topK
	}
}

// WithAPIVersion sets the Anthropic API version header.
// WHY: Anthropic uses date-based API versions (e.g., "2023-06-01") to allow gradual
// API evolution without breaking existing integrations. Users can:
// 1. Pin to a specific version for stability
// 2. Upgrade to newer versions to access new features
// 3. Test new versions before switching production
// Defaults to "2023-06-01" which is stable and widely supported.
func WithAPIVersion(version string) Option {
	return func(c *Client) {
		c.apiVersion = version
	}
}
