package tools

import (
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

// HTTPTool provides HTTP request capabilities for agents.
// It supports GET, POST, PUT, DELETE, and PATCH methods with configurable
// timeout, retries, and error handling.
type HTTPTool struct {
	client      *http.Client
	maxRetries  int
	retryDelay  time.Duration
	timeout     time.Duration
	userAgent   string
	maxBodySize int64 // Maximum response body size in bytes (default: 10MB)
}

// HTTPOption is a function that configures an HTTPTool.
type HTTPOption func(*HTTPTool)

// WithTimeout sets the HTTP client timeout (default: 30 seconds).
func WithTimeout(timeout time.Duration) HTTPOption {
	return func(h *HTTPTool) {
		h.timeout = timeout
		h.client.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retry attempts (default: 3).
func WithMaxRetries(maxRetries int) HTTPOption {
	return func(h *HTTPTool) {
		h.maxRetries = maxRetries
	}
}

// WithRetryDelay sets the base delay between retries (default: 1 second).
// Uses exponential backoff: delay * (2 ^ attempt).
func WithRetryDelay(delay time.Duration) HTTPOption {
	return func(h *HTTPTool) {
		h.retryDelay = delay
	}
}

// WithUserAgent sets a custom User-Agent header (default: "GoAgents-HTTPTool/1.0").
func WithUserAgent(userAgent string) HTTPOption {
	return func(h *HTTPTool) {
		h.userAgent = userAgent
	}
}

// WithMaxBodySize sets the maximum response body size in bytes (default: 10MB).
func WithMaxBodySize(size int64) HTTPOption {
	return func(h *HTTPTool) {
		h.maxBodySize = size
	}
}

// NewHTTPTool creates a new HTTP tool with the given options.
func NewHTTPTool(opts ...HTTPOption) *HTTPTool {
	h := &HTTPTool{
		client:      &http.Client{},
		maxRetries:  3,
		retryDelay:  1 * time.Second,
		timeout:     30 * time.Second,
		userAgent:   "GoAgents-HTTPTool/1.0",
		maxBodySize: 10 * 1024 * 1024, // 10MB
	}

	// Apply options
	for _, opt := range opts {
		opt(h)
	}

	// Set timeout on client
	h.client.Timeout = h.timeout

	return h
}

// Name returns the tool name.
func (h *HTTPTool) Name() string {
	return "http"
}

// Description returns a human-readable description of the tool.
func (h *HTTPTool) Description() string {
	return "Make HTTP requests to external APIs and web services. Supports GET, POST, PUT, DELETE, and PATCH methods with headers, query parameters, and JSON request/response bodies."
}

// Schema returns the tool's parameter schema for LLM function calling.
func (h *HTTPTool) Schema() *core.ToolSchema {
	return &core.ToolSchema{
		Name:        h.Name(),
		Description: h.Description(),
		Parameters: []core.Parameter{
			{
				Name:        "method",
				Type:        "string",
				Required:    true,
				Description: "HTTP method: GET, POST, PUT, DELETE, or PATCH",
			},
			{
				Name:        "url",
				Type:        "string",
				Required:    true,
				Description: "The URL to request (must include protocol: http:// or https://)",
			},
			{
				Name:        "headers",
				Type:        "object",
				Required:    false,
				Description: "Optional HTTP headers as key-value pairs (e.g., {\"Authorization\": \"Bearer token\"})",
			},
			{
				Name:        "query_params",
				Type:        "object",
				Required:    false,
				Description: "Optional query parameters as key-value pairs (e.g., {\"q\": \"search term\", \"limit\": \"10\"})",
			},
			{
				Name:        "body",
				Type:        "object",
				Required:    false,
				Description: "Optional request body for POST/PUT/PATCH (will be sent as JSON)",
			},
		},
	}
}

// Execute performs the HTTP request with the given parameters.
func (h *HTTPTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Extract and validate method
	methodRaw, ok := args["method"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: method")
	}
	method, ok := methodRaw.(string)
	if !ok {
		return nil, fmt.Errorf("method must be a string")
	}
	method = strings.ToUpper(method)

	// Validate method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true,
	}
	if !validMethods[method] {
		return nil, fmt.Errorf("invalid HTTP method: %s (must be GET, POST, PUT, DELETE, or PATCH)", method)
	}

	// Extract and validate URL
	urlRaw, ok := args["url"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: url")
	}
	url, ok := urlRaw.(string)
	if !ok {
		return nil, fmt.Errorf("url must be a string")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("url must start with http:// or https://")
	}

	// Build request
	var reqBody io.Reader
	if bodyRaw, ok := args["body"]; ok && bodyRaw != nil {
		// Marshal body to JSON
		bodyJSON, err := json.Marshal(bodyRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(bodyJSON)
	}

	// Add query parameters to URL
	if queryParamsRaw, ok := args["query_params"]; ok && queryParamsRaw != nil {
		queryParams, ok := queryParamsRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("query_params must be an object")
		}
		url = h.addQueryParams(url, queryParams)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", h.userAgent)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add custom headers
	if headersRaw, ok := args["headers"]; ok && headersRaw != nil {
		headers, ok := headersRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("headers must be an object")
		}
		for key, value := range headers {
			valueStr, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("header value for '%s' must be a string", key)
			}
			req.Header.Set(key, valueStr)
		}
	}

	// Execute with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= h.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := h.retryDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, lastErr = h.client.Do(req)
		if lastErr == nil {
			// Success
			break
		}

		// Check if error is retryable
		if !h.isRetryable(lastErr) {
			return nil, fmt.Errorf("request failed: %w", lastErr)
		}

		// Close response body if any
		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", h.maxRetries, lastErr)
	}
	defer resp.Body.Close()

	// Read response body with size limit
	limitReader := io.LimitReader(resp.Body, h.maxBodySize)
	bodyBytes, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Build response object
	response := map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     h.formatHeaders(resp.Header),
	}

	// Try to parse as JSON, fallback to string
	var bodyJSON interface{}
	if err := json.Unmarshal(bodyBytes, &bodyJSON); err == nil {
		response["body"] = bodyJSON
		response["content_type"] = "application/json"
	} else {
		response["body"] = string(bodyBytes)
		response["content_type"] = resp.Header.Get("Content-Type")
	}

	// Add success flag
	response["success"] = resp.StatusCode >= 200 && resp.StatusCode < 300

	return response, nil
}

// addQueryParams adds query parameters to a URL.
func (h *HTTPTool) addQueryParams(url string, params map[string]interface{}) string {
	if len(params) == 0 {
		return url
	}

	separator := "?"
	if strings.Contains(url, "?") {
		separator = "&"
	}

	var pairs []string
	for key, value := range params {
		valueStr := fmt.Sprintf("%v", value)
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, valueStr))
	}

	return url + separator + strings.Join(pairs, "&")
}

// formatHeaders converts http.Header to a simple map.
func (h *HTTPTool) formatHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// isRetryable checks if an error is retryable.
func (h *HTTPTool) isRetryable(err error) bool {
	// Network errors are generally retryable
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Common retryable errors
	retryableErrors := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
		"too many redirects",
		"eof", // Connection closed unexpectedly
		"broken pipe",
		"no such host",
	}

	errStrLower := strings.ToLower(errStr)
	for _, retryable := range retryableErrors {
		if strings.Contains(errStrLower, retryable) {
			return true
		}
	}

	return false
}
