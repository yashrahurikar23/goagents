package recovery

import (
	"testing"
	"time"
)

// TestNewTracker tests tracker creation with default options
func TestNewTracker(t *testing.T) {
	tracker := NewTracker()

	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	if tracker.maxAttempts != 3 {
		t.Errorf("expected maxAttempts=3, got %d", tracker.maxAttempts)
	}

	if tracker.retryWindow != 5*time.Minute {
		t.Errorf("expected retryWindow=5m, got %v", tracker.retryWindow)
	}

	if tracker.errors == nil {
		t.Error("errors map not initialized")
	}
}

// TestNewTracker_WithOptions tests tracker creation with custom options
func TestNewTracker_WithOptions(t *testing.T) {
	tracker := NewTracker(
		WithMaxAttempts(5),
		WithRetryWindow(10*time.Minute),
	)

	if tracker.maxAttempts != 5 {
		t.Errorf("expected maxAttempts=5, got %d", tracker.maxAttempts)
	}

	if tracker.retryWindow != 10*time.Minute {
		t.Errorf("expected retryWindow=10m, got %v", tracker.retryWindow)
	}
}

// TestRecordError tests error recording
func TestRecordError(t *testing.T) {
	tracker := NewTracker()

	record := ErrorRecord{
		FilePath:   "test.go",
		Operation:  "fuzzy_replace",
		ErrorType:  ErrorTypeNoMatch,
		ErrorMsg:   "search text not found",
		SearchText: "func test() {}",
		Confidence: 0.0,
	}

	tracker.RecordError(record)

	history := tracker.GetHistory("test.go")
	if len(history) != 1 {
		t.Fatalf("expected 1 error, got %d", len(history))
	}

	if history[0].FilePath != "test.go" {
		t.Errorf("expected FilePath=test.go, got %s", history[0].FilePath)
	}

	if history[0].ErrorType != ErrorTypeNoMatch {
		t.Errorf("expected ErrorType=no_match, got %s", history[0].ErrorType)
	}

	if history[0].Attempt != 1 {
		t.Errorf("expected Attempt=1, got %d", history[0].Attempt)
	}
}

// TestRecordError_MultipleAttempts tests attempt counting
func TestRecordError_MultipleAttempts(t *testing.T) {
	tracker := NewTracker()

	// Record 3 consecutive errors
	for i := 0; i < 3; i++ {
		tracker.RecordError(ErrorRecord{
			FilePath:  "test.go",
			Operation: "fuzzy_replace",
			ErrorType: ErrorTypeNoMatch,
			ErrorMsg:  "search text not found",
		})
	}

	history := tracker.GetHistory("test.go")
	if len(history) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(history))
	}

	// Check attempt numbers
	for i, record := range history {
		expectedAttempt := i + 1
		if record.Attempt != expectedAttempt {
			t.Errorf("error %d: expected Attempt=%d, got %d", i, expectedAttempt, record.Attempt)
		}
	}
}

// TestGetAdvice_NoErrors tests advice when no errors exist
func TestGetAdvice_NoErrors(t *testing.T) {
	tracker := NewTracker()

	advice := tracker.GetAdvice("test.go", "fuzzy_replace")

	if !advice.ShouldRetry {
		t.Error("expected ShouldRetry=true when no errors")
	}

	if advice.AttemptsRemaining != 3 {
		t.Errorf("expected AttemptsRemaining=3, got %d", advice.AttemptsRemaining)
	}

	if len(advice.Suggestions) != 0 {
		t.Errorf("expected 0 suggestions, got %d", len(advice.Suggestions))
	}
}

// TestGetAdvice_AfterFirstError tests advice after one error
func TestGetAdvice_AfterFirstError(t *testing.T) {
	tracker := NewTracker()

	tracker.RecordError(ErrorRecord{
		FilePath:   "test.go",
		Operation:  "fuzzy_replace",
		ErrorType:  ErrorTypeNoMatch,
		SearchText: "func test() {}",
	})

	advice := tracker.GetAdvice("test.go", "fuzzy_replace")

	if !advice.ShouldRetry {
		t.Error("expected ShouldRetry=true after first error")
	}

	if advice.AttemptsRemaining != 2 {
		t.Errorf("expected AttemptsRemaining=2, got %d", advice.AttemptsRemaining)
	}

	if len(advice.Suggestions) == 0 {
		t.Error("expected suggestions after first error")
	}
}

// TestGetAdvice_MaxAttemptsReached tests advice when max attempts reached
func TestGetAdvice_MaxAttemptsReached(t *testing.T) {
	tracker := NewTracker(WithMaxAttempts(3))

	// Record 3 errors
	for i := 0; i < 3; i++ {
		tracker.RecordError(ErrorRecord{
			FilePath:  "test.go",
			Operation: "fuzzy_replace",
			ErrorType: ErrorTypeNoMatch,
		})
	}

	advice := tracker.GetAdvice("test.go", "fuzzy_replace")

	if advice.ShouldRetry {
		t.Error("expected ShouldRetry=false after max attempts")
	}

	if advice.AttemptsRemaining != 0 {
		t.Errorf("expected AttemptsRemaining=0, got %d", advice.AttemptsRemaining)
	}
}

// TestGetAdvice_NoMatchSuggestions tests suggestions for no-match errors
func TestGetAdvice_NoMatchSuggestions(t *testing.T) {
	tracker := NewTracker()

	tracker.RecordError(ErrorRecord{
		FilePath:   "test.go",
		Operation:  "fuzzy_replace",
		ErrorType:  ErrorTypeNoMatch,
		SearchText: "func test() {}",
	})

	advice := tracker.GetAdvice("test.go", "fuzzy_replace")

	if len(advice.Suggestions) == 0 {
		t.Fatal("expected suggestions for no-match error")
	}

	// Should suggest reading file first
	hasReadSuggestion := false
	for _, suggestion := range advice.Suggestions {
		if contains(suggestion, "read_file") || contains(suggestion, "Read the file") {
			hasReadSuggestion = true
			break
		}
	}

	if !hasReadSuggestion {
		t.Error("expected suggestion to read file first")
	}
}

// TestGetAdvice_LowConfidenceSuggestions tests suggestions for low confidence errors
func TestGetAdvice_LowConfidenceSuggestions(t *testing.T) {
	tracker := NewTracker()

	tracker.RecordError(ErrorRecord{
		FilePath:   "test.go",
		Operation:  "fuzzy_replace",
		ErrorType:  ErrorTypeLowConfidence,
		Confidence: 0.75,
	})

	advice := tracker.GetAdvice("test.go", "fuzzy_replace")

	if len(advice.Suggestions) == 0 {
		t.Fatal("expected suggestions for low confidence error")
	}

	// Should mention confidence threshold
	hasConfidenceMention := false
	for _, suggestion := range advice.Suggestions {
		if contains(suggestion, "confidence") || contains(suggestion, "85%") {
			hasConfidenceMention = true
			break
		}
	}

	if !hasConfidenceMention {
		t.Error("expected mention of confidence threshold in suggestions")
	}
}

// TestGetAdvice_AlternativeApproach tests alternative approach suggestions
func TestGetAdvice_AlternativeApproach(t *testing.T) {
	tracker := NewTracker()

	// Record multiple errors
	for i := 0; i < 2; i++ {
		tracker.RecordError(ErrorRecord{
			FilePath:  "test.go",
			Operation: "fuzzy_replace",
			ErrorType: ErrorTypeNoMatch,
		})
	}

	advice := tracker.GetAdvice("test.go", "fuzzy_replace")

	if advice.AlternativeApproach == "" {
		t.Error("expected alternative approach after multiple failures")
	}

	if !contains(advice.AlternativeApproach, "line_replace") {
		t.Error("expected suggestion to use line_replace as alternative")
	}
}

// TestGetHistory tests history retrieval
func TestGetHistory(t *testing.T) {
	tracker := NewTracker()

	// Record errors for multiple files
	tracker.RecordError(ErrorRecord{
		FilePath:  "file1.go",
		ErrorType: ErrorTypeNoMatch,
	})

	tracker.RecordError(ErrorRecord{
		FilePath:  "file2.go",
		ErrorType: ErrorTypeLowConfidence,
	})

	tracker.RecordError(ErrorRecord{
		FilePath:  "file1.go",
		ErrorType: ErrorTypeNoMatch,
	})

	// Check file1 history
	history1 := tracker.GetHistory("file1.go")
	if len(history1) != 2 {
		t.Errorf("expected 2 errors for file1.go, got %d", len(history1))
	}

	// Check file2 history
	history2 := tracker.GetHistory("file2.go")
	if len(history2) != 1 {
		t.Errorf("expected 1 error for file2.go, got %d", len(history2))
	}

	// Check non-existent file
	history3 := tracker.GetHistory("nonexistent.go")
	if len(history3) != 0 {
		t.Errorf("expected 0 errors for nonexistent.go, got %d", len(history3))
	}
}

// TestClearHistory tests clearing history for a file
func TestClearHistory(t *testing.T) {
	tracker := NewTracker()

	tracker.RecordError(ErrorRecord{
		FilePath:  "test.go",
		ErrorType: ErrorTypeNoMatch,
	})

	// Verify error was recorded
	if len(tracker.GetHistory("test.go")) != 1 {
		t.Fatal("error not recorded")
	}

	// Clear history
	tracker.ClearHistory("test.go")

	// Verify history cleared
	if len(tracker.GetHistory("test.go")) != 0 {
		t.Error("history not cleared")
	}
}

// TestClearAllHistory tests clearing all history
func TestClearAllHistory(t *testing.T) {
	tracker := NewTracker()

	// Record errors for multiple files
	tracker.RecordError(ErrorRecord{FilePath: "file1.go", ErrorType: ErrorTypeNoMatch})
	tracker.RecordError(ErrorRecord{FilePath: "file2.go", ErrorType: ErrorTypeNoMatch})
	tracker.RecordError(ErrorRecord{FilePath: "file3.go", ErrorType: ErrorTypeNoMatch})

	// Clear all
	tracker.ClearAllHistory()

	// Verify all cleared
	if len(tracker.GetHistory("file1.go")) != 0 {
		t.Error("file1.go history not cleared")
	}
	if len(tracker.GetHistory("file2.go")) != 0 {
		t.Error("file2.go history not cleared")
	}
	if len(tracker.GetHistory("file3.go")) != 0 {
		t.Error("file3.go history not cleared")
	}
}

// TestGetStats tests statistics retrieval
func TestGetStats(t *testing.T) {
	tracker := NewTracker()

	// Record various errors
	tracker.RecordError(ErrorRecord{FilePath: "file1.go", ErrorType: ErrorTypeNoMatch})
	tracker.RecordError(ErrorRecord{FilePath: "file1.go", ErrorType: ErrorTypeNoMatch})
	tracker.RecordError(ErrorRecord{FilePath: "file2.go", ErrorType: ErrorTypeLowConfidence})
	tracker.RecordError(ErrorRecord{FilePath: "file3.go", ErrorType: ErrorTypeFileAccess})

	stats := tracker.GetStats()

	totalErrors := stats["total_errors"].(int)
	if totalErrors != 4 {
		t.Errorf("expected total_errors=4, got %d", totalErrors)
	}

	filesTracked := stats["files_tracked"].(int)
	if filesTracked != 3 {
		t.Errorf("expected files_tracked=3, got %d", filesTracked)
	}

	errorsByType := stats["errors_by_type"].(map[ErrorType]int)
	if errorsByType[ErrorTypeNoMatch] != 2 {
		t.Errorf("expected 2 no_match errors, got %d", errorsByType[ErrorTypeNoMatch])
	}
	if errorsByType[ErrorTypeLowConfidence] != 1 {
		t.Errorf("expected 1 low_confidence error, got %d", errorsByType[ErrorTypeLowConfidence])
	}
	if errorsByType[ErrorTypeFileAccess] != 1 {
		t.Errorf("expected 1 file_access error, got %d", errorsByType[ErrorTypeFileAccess])
	}
}

// TestShouldRetry tests the convenience method
func TestShouldRetry(t *testing.T) {
	tracker := NewTracker(WithMaxAttempts(2))

	// Initially should retry
	if !tracker.ShouldRetry("test.go") {
		t.Error("expected ShouldRetry=true initially")
	}

	// After first error, still should retry
	tracker.RecordError(ErrorRecord{FilePath: "test.go", ErrorType: ErrorTypeNoMatch})
	if !tracker.ShouldRetry("test.go") {
		t.Error("expected ShouldRetry=true after 1 error")
	}

	// After max attempts, should not retry
	tracker.RecordError(ErrorRecord{FilePath: "test.go", ErrorType: ErrorTypeNoMatch})
	if tracker.ShouldRetry("test.go") {
		t.Error("expected ShouldRetry=false after max attempts")
	}
}

// TestGetAttemptCount tests attempt counting
func TestGetAttemptCount(t *testing.T) {
	tracker := NewTracker()

	if tracker.GetAttemptCount("test.go") != 0 {
		t.Error("expected 0 attempts initially")
	}

	tracker.RecordError(ErrorRecord{FilePath: "test.go", ErrorType: ErrorTypeNoMatch})
	if tracker.GetAttemptCount("test.go") != 1 {
		t.Error("expected 1 attempt after first error")
	}

	tracker.RecordError(ErrorRecord{FilePath: "test.go", ErrorType: ErrorTypeNoMatch})
	if tracker.GetAttemptCount("test.go") != 2 {
		t.Error("expected 2 attempts after second error")
	}
}

// TestRetryWindow tests the time-based retry window
func TestRetryWindow(t *testing.T) {
	tracker := NewTracker(WithRetryWindow(100 * time.Millisecond))

	// Record error
	tracker.RecordError(ErrorRecord{FilePath: "test.go", ErrorType: ErrorTypeNoMatch})

	// Should have 1 recent error
	if tracker.GetAttemptCount("test.go") != 1 {
		t.Fatal("expected 1 recent error")
	}

	// Wait for retry window to expire
	time.Sleep(150 * time.Millisecond)

	// Should have 0 recent errors (outside retry window)
	if tracker.GetAttemptCount("test.go") != 0 {
		t.Error("expected 0 recent errors after window expiry")
	}

	// History should still contain the error
	if len(tracker.GetHistory("test.go")) != 1 {
		t.Error("error should still be in history")
	}
}

// TestConcurrency tests concurrent access to tracker
func TestConcurrency(t *testing.T) {
	tracker := NewTracker()

	// Start multiple goroutines recording errors
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				tracker.RecordError(ErrorRecord{
					FilePath:  "test.go",
					ErrorType: ErrorTypeNoMatch,
				})
				tracker.GetAdvice("test.go", "fuzzy_replace")
				tracker.GetHistory("test.go")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 1000 total errors
	history := tracker.GetHistory("test.go")
	if len(history) != 1000 {
		t.Errorf("expected 1000 errors, got %d", len(history))
	}
}

// TestDiffHint tests diff generation for debugging
func TestDiffHint(t *testing.T) {
	tracker := NewTracker()

	searchText := `func Calculate(a, b int) int {
    return a + b
}`

	tracker.RecordError(ErrorRecord{
		FilePath:   "test.go",
		ErrorType:  ErrorTypeNoMatch,
		SearchText: searchText,
	})

	advice := tracker.GetAdvice("test.go", "fuzzy_replace")

	if advice.Diff == "" {
		t.Error("expected diff hint for error with search text")
	}

	if !contains(advice.Diff, "Search text attempted") {
		t.Error("expected 'Search text attempted' in diff hint")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
