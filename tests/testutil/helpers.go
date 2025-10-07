// Package testutil provides common test utilities and helpers.
// This package is NOT part of the production goagent package.
package testutil

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// AssertNoError fails the test if err is not nil.
// Using t.Helper() ensures the failure is reported at the caller's line.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

// AssertEqual fails the test if got != want.
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// AssertNotEqual fails the test if got == want.
func AssertNotEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if got == want {
		t.Fatalf("got %v, want different value", got)
	}
}

// AssertTrue fails the test if condition is false.
func AssertTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Fatalf("assertion failed: %s", message)
	}
}

// AssertFalse fails the test if condition is true.
func AssertFalse(t *testing.T, condition bool, message string) {
	t.Helper()
	if condition {
		t.Fatalf("assertion failed: %s", message)
	}
}

// AssertContains fails the test if haystack doesn't contain needle.
func AssertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !contains(haystack, needle) {
		t.Fatalf("expected %q to contain %q", haystack, needle)
	}
}

// AssertNotContains fails the test if haystack contains needle.
func AssertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if contains(haystack, needle) {
		t.Fatalf("expected %q to not contain %q", haystack, needle)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// AssertNil fails the test if v is not nil.
func AssertNil(t *testing.T, v interface{}) {
	t.Helper()
	if v != nil {
		t.Fatalf("expected nil, got %v", v)
	}
}

// AssertNotNil fails the test if v is nil.
func AssertNotNil(t *testing.T, v interface{}) {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil value")
	}
}

// AssertLen fails the test if the slice doesn't have the expected length.
func AssertLen(t *testing.T, slice interface{}, expectedLen int) {
	t.Helper()
	// This is a simplified version - a real implementation would use reflection
	// to handle any slice type
	switch v := slice.(type) {
	case []interface{}:
		if len(v) != expectedLen {
			t.Fatalf("expected length %d, got %d", expectedLen, len(v))
		}
	case []string:
		if len(v) != expectedLen {
			t.Fatalf("expected length %d, got %d", expectedLen, len(v))
		}
	case []int:
		if len(v) != expectedLen {
			t.Fatalf("expected length %d, got %d", expectedLen, len(v))
		}
	default:
		t.Fatal("unsupported type for AssertLen")
	}
}

// Timeout creates a context with timeout for tests.
// Default timeout is 5 seconds if duration is 0.
func Timeout(duration time.Duration) (context.Context, context.CancelFunc) {
	if duration == 0 {
		duration = 5 * time.Second
	}
	return context.WithTimeout(context.Background(), duration)
}

// LoadFixture loads a JSON fixture file from tests/fixtures/
//
// WHY THIS EXISTS:
// Tests often need sample API responses or test data.
// Loading from files keeps test code clean and data reusable.
func LoadFixture(t *testing.T, filename string, v interface{}) {
	t.Helper()

	// Build path relative to tests/fixtures/
	path := filepath.Join("tests", "fixtures", filename)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", filename, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("failed to parse fixture %s: %v", filename, err)
	}
}

// SaveFixture saves data as a JSON fixture file.
// Useful for capturing real API responses for later testing.
func SaveFixture(t *testing.T, filename string, v interface{}) {
	t.Helper()

	path := filepath.Join("tests", "fixtures", filename)

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create fixture directory: %v", err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal fixture: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
}

// SkipIfShort skips the test if -short flag is used.
// Useful for slow tests (integration, E2E).
func SkipIfShort(t *testing.T, reason string) {
	t.Helper()
	if testing.Short() {
		t.Skipf("skipping in short mode: %s", reason)
	}
}

// RequireEnv skips the test if the environment variable is not set.
// Useful for E2E tests that need API keys.
func RequireEnv(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("skipping: %s environment variable not set", key)
	}
	return value
}

// WithCleanup runs cleanup function even if test panics.
// Similar to t.Cleanup() but with explicit control.
func WithCleanup(t *testing.T, setup func() interface{}, cleanup func(interface{})) interface{} {
	t.Helper()
	resource := setup()
	t.Cleanup(func() {
		cleanup(resource)
	})
	return resource
}

// Benchmark helpers

// BenchmarkMemory runs a benchmark and reports memory allocations.
func BenchmarkMemory(b *testing.B, fn func()) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

// Example usage in tests:
//
// Basic assertions:
//   testutil.AssertNoError(t, err)
//   testutil.AssertEqual(t, got, want)
//   testutil.AssertContains(t, response, "success")
//
// Context with timeout:
//   ctx, cancel := testutil.Timeout(2 * time.Second)
//   defer cancel()
//   result, err := client.Chat(ctx, messages)
//
// Load fixture:
//   var response ChatCompletionResponse
//   testutil.LoadFixture(t, "openai/chat_response.json", &response)
//
// Skip tests conditionally:
//   testutil.SkipIfShort(t, "requires real API call")
//   apiKey := testutil.RequireEnv(t, "OPENAI_API_KEY")
//
// Cleanup:
//   server := testutil.WithCleanup(t,
//       func() interface{} { return mocks.NewMockHTTPServer(...) },
//       func(s interface{}) { s.(*mocks.MockHTTPServer).Close() },
//   ).(*mocks.MockHTTPServer)
