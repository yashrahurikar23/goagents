package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPTool_Name(t *testing.T) {
	tool := NewHTTPTool()
	if tool.Name() != "http" {
		t.Errorf("expected name 'http', got %s", tool.Name())
	}
}

func TestHTTPTool_Description(t *testing.T) {
	tool := NewHTTPTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("description should not be empty")
	}
	if !strings.Contains(desc, "HTTP") {
		t.Error("description should mention HTTP")
	}
}

func TestHTTPTool_Schema(t *testing.T) {
	tool := NewHTTPTool()
	schema := tool.Schema()

	if schema.Name != "http" {
		t.Errorf("expected schema name 'http', got %s", schema.Name)
	}

	// Check required parameters
	requiredParams := []string{"method", "url"}
	for _, param := range requiredParams {
		found := false
		for _, p := range schema.Parameters {
			if p.Name == param && p.Required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("required parameter '%s' not found in schema", param)
		}
	}

	// Check optional parameters
	optionalParams := []string{"headers", "query_params", "body"}
	for _, param := range optionalParams {
		found := false
		for _, p := range schema.Parameters {
			if p.Name == param && !p.Required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("optional parameter '%s' not found in schema", param)
		}
	}
}

func TestHTTPTool_Execute_GET(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	}))
	defer server.Close()

	tool := NewHTTPTool()
	args := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
	}

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("result should be a map")
	}

	if response["status_code"] != 200 {
		t.Errorf("expected status code 200, got %v", response["status_code"])
	}

	if response["success"] != true {
		t.Error("expected success to be true")
	}

	body, ok := response["body"].(map[string]interface{})
	if !ok {
		t.Fatal("body should be a map")
	}

	if body["message"] != "success" {
		t.Errorf("expected message 'success', got %v", body["message"])
	}
}

func TestHTTPTool_Execute_POST(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// Read and validate body
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if reqBody["name"] != "test" {
			t.Errorf("expected name 'test', got %v", reqBody["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      123,
			"message": "created",
		})
	}))
	defer server.Close()

	tool := NewHTTPTool()
	args := map[string]interface{}{
		"method": "POST",
		"url":    server.URL,
		"body": map[string]interface{}{
			"name": "test",
		},
	}

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("result should be a map")
	}

	if response["status_code"] != 201 {
		t.Errorf("expected status code 201, got %v", response["status_code"])
	}
}

func TestHTTPTool_Execute_WithHeaders(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization header 'Bearer test-token', got %s", auth)
		}

		customHeader := r.Header.Get("X-Custom-Header")
		if customHeader != "custom-value" {
			t.Errorf("expected X-Custom-Header 'custom-value', got %s", customHeader)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "ok"}`))
	}))
	defer server.Close()

	tool := NewHTTPTool()
	args := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
		"headers": map[string]interface{}{
			"Authorization":   "Bearer test-token",
			"X-Custom-Header": "custom-value",
		},
	}

	_, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestHTTPTool_Execute_WithQueryParams(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("q") != "search term" {
			t.Errorf("expected query param q='search term', got %s", query.Get("q"))
		}
		if query.Get("limit") != "10" {
			t.Errorf("expected query param limit='10', got %s", query.Get("limit"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results": []}`))
	}))
	defer server.Close()

	tool := NewHTTPTool()
	args := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
		"query_params": map[string]interface{}{
			"q":     "search term",
			"limit": 10,
		},
	}

	_, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestHTTPTool_Execute_InvalidMethod(t *testing.T) {
	tool := NewHTTPTool()
	args := map[string]interface{}{
		"method": "INVALID",
		"url":    "http://example.com",
	}

	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("expected error for invalid method")
	}
	if !strings.Contains(err.Error(), "invalid HTTP method") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestHTTPTool_Execute_MissingMethod(t *testing.T) {
	tool := NewHTTPTool()
	args := map[string]interface{}{
		"url": "http://example.com",
	}

	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("expected error for missing method")
	}
	if !strings.Contains(err.Error(), "missing required parameter: method") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestHTTPTool_Execute_MissingURL(t *testing.T) {
	tool := NewHTTPTool()
	args := map[string]interface{}{
		"method": "GET",
	}

	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("expected error for missing URL")
	}
	if !strings.Contains(err.Error(), "missing required parameter: url") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestHTTPTool_Execute_InvalidURL(t *testing.T) {
	tool := NewHTTPTool()
	args := map[string]interface{}{
		"method": "GET",
		"url":    "not-a-valid-url",
	}

	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "must start with http://") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestHTTPTool_Execute_Timeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create tool with short timeout
	tool := NewHTTPTool(WithTimeout(100 * time.Millisecond))
	args := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
	}

	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestHTTPTool_Execute_Retry(t *testing.T) {
	// Test that retry logic exists and works with max retries exceeded
	// We'll create a server that always fails to verify retries are attempted

	attempts := 0

	// Create server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Simulate connection error by closing connection
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	tool := NewHTTPTool(
		WithMaxRetries(2),
		WithRetryDelay(10*time.Millisecond),
	)
	args := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
	}

	_, err := tool.Execute(context.Background(), args)
	if err == nil {
		t.Error("expected error after max retries exceeded")
	}

	// Should have tried 3 times total (initial + 2 retries)
	if attempts < 3 {
		t.Errorf("expected at least 3 attempts, got %d", attempts)
	}

	if !strings.Contains(err.Error(), "failed after") {
		t.Errorf("expected 'failed after' in error message, got: %v", err)
	}
}

func TestHTTPTool_Execute_AllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("expected %s request, got %s", method, r.Method)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"method": "` + method + `"}`))
			}))
			defer server.Close()

			tool := NewHTTPTool()
			args := map[string]interface{}{
				"method": method,
				"url":    server.URL,
			}

			result, err := tool.Execute(context.Background(), args)
			if err != nil {
				t.Fatalf("Execute failed for %s: %v", method, err)
			}

			response := result.(map[string]interface{})
			if response["status_code"] != 200 {
				t.Errorf("expected status 200, got %v", response["status_code"])
			}
		})
	}
}

func TestHTTPTool_Execute_NonJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("plain text response"))
	}))
	defer server.Close()

	tool := NewHTTPTool()
	args := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
	}

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	response := result.(map[string]interface{})
	body, ok := response["body"].(string)
	if !ok {
		t.Fatal("body should be a string for non-JSON response")
	}

	if body != "plain text response" {
		t.Errorf("expected 'plain text response', got %s", body)
	}

	if response["content_type"] != "text/plain" {
		t.Errorf("expected content_type 'text/plain', got %v", response["content_type"])
	}
}

func TestHTTPTool_Execute_StatusCodes(t *testing.T) {
	testCases := []struct {
		statusCode int
		success    bool
	}{
		{200, true},
		{201, true},
		{204, true},
		{400, false},
		{401, false},
		{404, false},
		{500, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Status%d", tc.statusCode), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(`{}`))
			}))
			defer server.Close()

			tool := NewHTTPTool()
			args := map[string]interface{}{
				"method": "GET",
				"url":    server.URL,
			}

			result, err := tool.Execute(context.Background(), args)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			response := result.(map[string]interface{})
			if response["status_code"] != tc.statusCode {
				t.Errorf("expected status %d, got %v", tc.statusCode, response["status_code"])
			}

			if response["success"] != tc.success {
				t.Errorf("expected success %v for status %d, got %v", tc.success, tc.statusCode, response["success"])
			}
		})
	}
}

func TestHTTPTool_Options(t *testing.T) {
	t.Run("WithTimeout", func(t *testing.T) {
		tool := NewHTTPTool(WithTimeout(5 * time.Second))
		if tool.timeout != 5*time.Second {
			t.Errorf("expected timeout 5s, got %v", tool.timeout)
		}
	})

	t.Run("WithMaxRetries", func(t *testing.T) {
		tool := NewHTTPTool(WithMaxRetries(5))
		if tool.maxRetries != 5 {
			t.Errorf("expected maxRetries 5, got %d", tool.maxRetries)
		}
	})

	t.Run("WithRetryDelay", func(t *testing.T) {
		tool := NewHTTPTool(WithRetryDelay(2 * time.Second))
		if tool.retryDelay != 2*time.Second {
			t.Errorf("expected retryDelay 2s, got %v", tool.retryDelay)
		}
	})

	t.Run("WithUserAgent", func(t *testing.T) {
		tool := NewHTTPTool(WithUserAgent("CustomAgent/1.0"))
		if tool.userAgent != "CustomAgent/1.0" {
			t.Errorf("expected userAgent 'CustomAgent/1.0', got %s", tool.userAgent)
		}
	})

	t.Run("WithMaxBodySize", func(t *testing.T) {
		tool := NewHTTPTool(WithMaxBodySize(1024))
		if tool.maxBodySize != 1024 {
			t.Errorf("expected maxBodySize 1024, got %d", tool.maxBodySize)
		}
	})
}

func TestHTTPTool_Execute_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tool := NewHTTPTool()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	args := map[string]interface{}{
		"method": "GET",
		"url":    server.URL,
	}

	_, err := tool.Execute(ctx, args)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}
