// Package mocks provides mock implementations for testing.
package mocks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockHTTPServer creates a mock HTTP server for testing HTTP clients.
//
// WHY THIS EXISTS:
// Testing HTTP clients requires a controlled server that:
// - Returns predictable responses
// - Simulates error conditions (timeouts, 429s, 500s)
// - Validates request structure
// - Works offline without real API calls
//
// DESIGN DECISIONS:
// - Uses httptest.Server from standard library (no external deps)
// - Handler function approach allows flexible per-test configuration
// - Request tracking enables assertion of API usage
// - Response builders make common scenarios easy
type MockHTTPServer struct {
	Server *httptest.Server

	// Handler is called for each request
	Handler http.HandlerFunc

	// RequestHistory tracks all requests
	Requests []*http.Request
}

// NewMockHTTPServer creates a new mock HTTP server.
//
// WHY THIS WAY:
// The handler function approach allows tests to customize behavior
// without creating a new server type for each scenario.
func NewMockHTTPServer(handler http.HandlerFunc) *MockHTTPServer {
	mock := &MockHTTPServer{
		Handler:  handler,
		Requests: []*http.Request{},
	}

	// Wrap handler to track requests
	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.Requests = append(mock.Requests, r)
		if mock.Handler != nil {
			mock.Handler(w, r)
		} else {
			// Default: return 200 OK
			w.WriteHeader(http.StatusOK)
		}
	}))

	return mock
}

// Close shuts down the mock server.
// Should be called with defer in tests.
func (m *MockHTTPServer) Close() {
	m.Server.Close()
}

// URL returns the mock server's base URL.
func (m *MockHTTPServer) URL() string {
	return m.Server.URL
}

// RequestCount returns the number of requests received.
func (m *MockHTTPServer) RequestCount() int {
	return len(m.Requests)
}

// LastRequest returns the most recent request, or nil if none.
func (m *MockHTTPServer) LastRequest() *http.Request {
	if len(m.Requests) == 0 {
		return nil
	}
	return m.Requests[len(m.Requests)-1]
}

// Reset clears request history.
func (m *MockHTTPServer) Reset() {
	m.Requests = []*http.Request{}
}

// Common Response Builders
// These make it easy to simulate common API scenarios

// ChatCompletionResponse creates a mock OpenAI chat completion response.
func ChatCompletionResponse(content string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"id":      "chatcmpl-mock123",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "gpt-4",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": content,
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}

		json.NewEncoder(w).Encode(response)
	}
}

// ChatCompletionWithToolCallResponse creates a response with function calling.
func ChatCompletionWithToolCallResponse(toolName, toolArgs string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"id":      "chatcmpl-mock123",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "gpt-4",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": nil,
						"tool_calls": []map[string]interface{}{
							{
								"id":   "call_mock123",
								"type": "function",
								"function": map[string]interface{}{
									"name":      toolName,
									"arguments": toolArgs,
								},
							},
						},
					},
					"finish_reason": "tool_calls",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}

		json.NewEncoder(w).Encode(response)
	}
}

// StreamingResponse creates a mock SSE streaming response.
func StreamingResponse(chunks []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		for _, chunk := range chunks {
			data := map[string]interface{}{
				"id":      "chatcmpl-mock123",
				"object":  "chat.completion.chunk",
				"created": 1234567890,
				"model":   "gpt-4",
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"delta": map[string]interface{}{
							"content": chunk,
						},
						"finish_reason": nil,
					},
				},
			}

			jsonData, _ := json.Marshal(data)
			w.Write([]byte("data: "))
			w.Write(jsonData)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}

		// Send [DONE] marker
		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}
}

// ErrorResponse creates an error response.
func ErrorResponse(statusCode int, errorType, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		response := map[string]interface{}{
			"error": map[string]interface{}{
				"type":    errorType,
				"message": message,
			},
		}

		json.NewEncoder(w).Encode(response)
	}
}

// RateLimitResponse creates a 429 rate limit response.
func RateLimitResponse() http.HandlerFunc {
	return ErrorResponse(http.StatusTooManyRequests, "rate_limit_exceeded", "Rate limit exceeded")
}

// ServerErrorResponse creates a 500 server error response.
func ServerErrorResponse() http.HandlerFunc {
	return ErrorResponse(http.StatusInternalServerError, "server_error", "Internal server error")
}

// InvalidRequestResponse creates a 400 bad request response.
func InvalidRequestResponse(message string) http.HandlerFunc {
	return ErrorResponse(http.StatusBadRequest, "invalid_request_error", message)
}

// SequentialResponses returns different responses for each call.
// Useful for testing retry logic.
//
// WHY THIS WAY:
// Many tests need to simulate: fail, fail, success
// or: rate limit, then success
func SequentialResponses(handlers ...http.HandlerFunc) http.HandlerFunc {
	callCount := 0
	return func(w http.ResponseWriter, r *http.Request) {
		if callCount < len(handlers) {
			handler := handlers[callCount]
			callCount++
			handler(w, r)
		} else {
			// Out of configured responses
			http.Error(w, "No more responses configured", http.StatusInternalServerError)
		}
	}
}

// ConditionalResponse returns different responses based on request content.
// Useful for simulating stateful interactions.
func ConditionalResponse(conditions map[string]http.HandlerFunc, defaultHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check request path or body for conditions
		for key, handler := range conditions {
			if strings.Contains(r.URL.Path, key) || strings.Contains(r.Header.Get("Content-Type"), key) {
				handler(w, r)
				return
			}
		}

		if defaultHandler != nil {
			defaultHandler(w, r)
		} else {
			http.Error(w, "No matching condition", http.StatusNotFound)
		}
	}
}

// Example usage in tests:
//
// Basic mock server:
//   server := mocks.NewMockHTTPServer(mocks.ChatCompletionResponse("Hello!"))
//   defer server.Close()
//   // Use server.URL() as base URL for OpenAI client
//
// Test retry logic (fail twice, then succeed):
//   server := mocks.NewMockHTTPServer(mocks.SequentialResponses(
//       mocks.ServerErrorResponse(),
//       mocks.RateLimitResponse(),
//       mocks.ChatCompletionResponse("Success!"),
//   ))
//   defer server.Close()
//
// Test streaming:
//   server := mocks.NewMockHTTPServer(mocks.StreamingResponse([]string{
//       "Hello", " ", "world", "!",
//   }))
//   defer server.Close()
//
// Verify requests:
//   server := mocks.NewMockHTTPServer(mocks.ChatCompletionResponse("OK"))
//   defer server.Close()
//   // ... make requests ...
//   if server.RequestCount() != 1 {
//       t.Errorf("expected 1 request, got %d", server.RequestCount())
//   }
//   req := server.LastRequest()
//   if req.Header.Get("Authorization") != "Bearer sk-test" {
//       t.Error("missing or wrong auth header")
//   }
