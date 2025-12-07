// Package recovery provides error tracking and retry logic for file editing operations.
// It helps agents recover from repeated failures by tracking error patterns and
// suggesting alternative approaches.
package recovery

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// ErrorType categorizes the kind of error that occurred
type ErrorType string

const (
	ErrorTypeNoMatch        ErrorType = "no_match"        // Search string not found
	ErrorTypeLowConfidence  ErrorType = "low_confidence"  // Match confidence below threshold
	ErrorTypePathValidation ErrorType = "path_validation" // Path security validation failed
	ErrorTypeFileAccess     ErrorType = "file_access"     // File read/write error
	ErrorTypeOther          ErrorType = "other"           // Other errors
)

// ErrorRecord represents a single error occurrence
type ErrorRecord struct {
	Timestamp  time.Time
	FilePath   string
	Operation  string
	ErrorType  ErrorType
	ErrorMsg   string
	SearchText string // The search text that failed (if applicable)
	Confidence float64
	LineAnchor int
	Attempt    int // Attempt number (1, 2, 3...)
}

// RecoveryAdvice provides suggestions for handling repeated failures
type RecoveryAdvice struct {
	ShouldRetry         bool
	AttemptsRemaining   int
	Suggestions         []string
	AlternativeApproach string
	Diff                string // Visual diff showing what was attempted
}

// Tracker tracks errors and provides recovery advice
type Tracker struct {
	mu            sync.RWMutex
	errors        map[string][]ErrorRecord // key: filePath
	maxAttempts   int
	retryWindow   time.Duration // Time window for counting consecutive failures
	diffGenerator *diffmatchpatch.DiffMatchPatch
}

// NewTracker creates a new error recovery tracker
func NewTracker(options ...TrackerOption) *Tracker {
	t := &Tracker{
		errors:        make(map[string][]ErrorRecord),
		maxAttempts:   3,
		retryWindow:   5 * time.Minute,
		diffGenerator: diffmatchpatch.New(),
	}

	for _, opt := range options {
		opt(t)
	}

	return t
}

// TrackerOption is a functional option for configuring the Tracker
type TrackerOption func(*Tracker)

// WithMaxAttempts sets the maximum number of retry attempts
func WithMaxAttempts(max int) TrackerOption {
	return func(t *Tracker) {
		if max > 0 {
			t.maxAttempts = max
		}
	}
}

// WithRetryWindow sets the time window for counting consecutive failures
func WithRetryWindow(window time.Duration) TrackerOption {
	return func(t *Tracker) {
		if window > 0 {
			t.retryWindow = window
		}
	}
}

// RecordError records an error occurrence
func (t *Tracker) RecordError(record ErrorRecord) {
	t.mu.Lock()
	defer t.mu.Unlock()

	record.Timestamp = time.Now()

	// Get recent errors for this file
	recentErrors := t.getRecentErrorsLocked(record.FilePath)
	record.Attempt = len(recentErrors) + 1

	t.errors[record.FilePath] = append(t.errors[record.FilePath], record)
}

// GetAdvice provides recovery advice based on error history
func (t *Tracker) GetAdvice(filePath string, operation string) RecoveryAdvice {
	t.mu.RLock()
	defer t.mu.RUnlock()

	recentErrors := t.getRecentErrorsLocked(filePath)

	advice := RecoveryAdvice{
		ShouldRetry:       len(recentErrors) < t.maxAttempts,
		AttemptsRemaining: t.maxAttempts - len(recentErrors),
		Suggestions:       []string{},
	}

	// No errors yet, can proceed
	if len(recentErrors) == 0 {
		return advice
	}

	// Analyze error patterns and provide suggestions
	advice.Suggestions = t.generateSuggestionsLocked(recentErrors, operation)
	advice.AlternativeApproach = t.suggestAlternativeLocked(recentErrors, operation)

	// Generate diff if we have search text
	lastError := recentErrors[len(recentErrors)-1]
	if lastError.SearchText != "" {
		advice.Diff = t.generateDiffHintLocked(lastError)
	}

	return advice
}

// GetHistory returns the error history for a file
func (t *Tracker) GetHistory(filePath string) []ErrorRecord {
	t.mu.RLock()
	defer t.mu.RUnlock()

	errors := t.errors[filePath]
	if errors == nil {
		return []ErrorRecord{}
	}

	// Return a copy to avoid external modification
	result := make([]ErrorRecord, len(errors))
	copy(result, errors)
	return result
}

// ClearHistory clears error history for a file
func (t *Tracker) ClearHistory(filePath string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.errors, filePath)
}

// ClearAllHistory clears all error history
func (t *Tracker) ClearAllHistory() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.errors = make(map[string][]ErrorRecord)
}

// GetStats returns statistics about tracked errors
func (t *Tracker) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	totalErrors := 0
	fileCount := len(t.errors)
	errorsByType := make(map[ErrorType]int)

	for _, records := range t.errors {
		totalErrors += len(records)
		for _, record := range records {
			errorsByType[record.ErrorType]++
		}
	}

	return map[string]interface{}{
		"total_errors":    totalErrors,
		"files_tracked":   fileCount,
		"errors_by_type":  errorsByType,
		"max_attempts":    t.maxAttempts,
		"retry_window_ms": t.retryWindow.Milliseconds(),
	}
}

// getRecentErrorsLocked returns errors within the retry window (must be called with lock held)
func (t *Tracker) getRecentErrorsLocked(filePath string) []ErrorRecord {
	allErrors := t.errors[filePath]
	if allErrors == nil {
		return []ErrorRecord{}
	}

	cutoff := time.Now().Add(-t.retryWindow)
	recent := []ErrorRecord{}

	for _, err := range allErrors {
		if err.Timestamp.After(cutoff) {
			recent = append(recent, err)
		}
	}

	return recent
}

// generateSuggestionsLocked generates context-specific suggestions (must be called with lock held)
func (t *Tracker) generateSuggestionsLocked(errors []ErrorRecord, operation string) []string {
	if len(errors) == 0 {
		return []string{}
	}

	suggestions := []string{}
	lastError := errors[len(errors)-1]

	switch lastError.ErrorType {
	case ErrorTypeNoMatch:
		suggestions = append(suggestions,
			"Read the file first using read_file operation to see the exact content and formatting",
			"Check for typos in the search text - even a single character difference will prevent matching",
		)

		if lastError.LineAnchor == 0 {
			suggestions = append(suggestions,
				"Use line_anchor parameter to narrow the search area and improve performance",
			)
		}

		if len(errors) >= 2 {
			suggestions = append(suggestions,
				"Consider using line_replace operation with specific line numbers instead of fuzzy matching",
				"The code structure may have changed - verify the search text still exists in the file",
			)
		}

	case ErrorTypeLowConfidence:
		suggestions = append(suggestions,
			fmt.Sprintf("Last match had %.1f%% confidence (need 85%%+). The search text is close but not quite right", lastError.Confidence*100),
			"Ensure indentation (tabs vs spaces) matches the actual file",
			"Check that line endings match (\\n vs \\r\\n)",
		)

		if lastError.Confidence > 0.70 {
			suggestions = append(suggestions,
				"Confidence is high - you're very close. Double-check whitespace and indentation carefully",
			)
		}

	case ErrorTypePathValidation:
		suggestions = append(suggestions,
			"Use relative paths from the workspace base directory",
			"Avoid '..' directory traversal attempts",
			"Do not use absolute paths like '/etc/passwd'",
		)

	case ErrorTypeFileAccess:
		suggestions = append(suggestions,
			"Verify the file exists using file_exists operation",
			"Check file permissions - you may not have read/write access",
			"Ensure the file path is correct (case-sensitive on Unix systems)",
		)
	}

	// Add attempt-specific advice
	if len(errors) >= 2 {
		suggestions = append(suggestions,
			fmt.Sprintf("This is attempt %d of %d - consider trying a different approach", len(errors), t.maxAttempts),
		)
	}

	return suggestions
}

// suggestAlternativeLocked suggests an alternative approach (must be called with lock held)
func (t *Tracker) suggestAlternativeLocked(errors []ErrorRecord, operation string) string {
	if len(errors) < 2 {
		return ""
	}

	lastError := errors[len(errors)-1]

	switch lastError.ErrorType {
	case ErrorTypeNoMatch, ErrorTypeLowConfidence:
		if operation == "fuzzy_replace" {
			return "Try using line_replace instead: specify exact line numbers (line_start, line_end) to replace a specific range"
		}
		return "Consider breaking the change into smaller, more targeted edits"

	case ErrorTypeFileAccess:
		return "Use list_directory operation to verify the file exists and check available files"

	default:
		return ""
	}
}

// generateDiffHintLocked generates a visual diff to help debug matching issues (must be called with lock held)
func (t *Tracker) generateDiffHintLocked(record ErrorRecord) string {
	if record.SearchText == "" {
		return ""
	}

	// This is a placeholder - in real usage, we'd compare against actual file content
	// For now, just show what was attempted
	lines := strings.Split(record.SearchText, "\n")
	if len(lines) > 10 {
		// Truncate long searches
		truncated := make([]string, 0, 9)
		truncated = append(truncated, lines[:5]...)
		truncated = append(truncated, "... (truncated) ...")
		truncated = append(truncated, lines[len(lines)-3:]...)
		lines = truncated
	}

	var result strings.Builder
	result.WriteString("Search text attempted:\n")
	for i, line := range lines {
		result.WriteString(fmt.Sprintf("  %d: %s\n", i+1, line))
	}

	return result.String()
}

// ShouldRetry is a convenience method to quickly check if retry is allowed
func (t *Tracker) ShouldRetry(filePath string) bool {
	advice := t.GetAdvice(filePath, "")
	return advice.ShouldRetry
}

// GetAttemptCount returns the number of recent attempts for a file
func (t *Tracker) GetAttemptCount(filePath string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.getRecentErrorsLocked(filePath))
}
