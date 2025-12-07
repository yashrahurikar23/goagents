package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yashrahurikar23/goagents/internal/recovery"
)

// TestFileEditTool_RecoveryTracker_MaxAttempts tests that max retry attempts are enforced
func TestFileEditTool_RecoveryTracker_MaxAttempts(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func main() {
	fmt.Println("Hello")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create tool with custom tracker (2 max attempts for faster testing)
	tracker := recovery.NewTracker(recovery.WithMaxAttempts(2))
	tool, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithRecoveryTracker(tracker),
	)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := context.Background()

	// Attempt 1: Wrong search text (will fail)
	result1, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `func nonexistent() {}`,
		"replace":   `func something() {}`,
	})
	if err != nil {
		t.Fatalf("attempt 1 failed: %v", err)
	}

	resultMap1 := result1.(map[string]interface{})
	if resultMap1["success"].(bool) {
		t.Error("attempt 1 should fail")
	}

	attemptsRemaining1 := int(resultMap1["attempts_remaining"].(int))
	if attemptsRemaining1 != 1 {
		t.Errorf("after attempt 1, expected 1 attempt remaining, got %d", attemptsRemaining1)
	}

	// Attempt 2: Still wrong (will fail)
	result2, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `func alsoWrong() {}`,
		"replace":   `func something() {}`,
	})
	if err != nil {
		t.Fatalf("attempt 2 failed: %v", err)
	}

	resultMap2 := result2.(map[string]interface{})
	if resultMap2["success"].(bool) {
		t.Error("attempt 2 should fail")
	}

	attemptsRemaining2 := int(resultMap2["attempts_remaining"].(int))
	if attemptsRemaining2 != 0 {
		t.Errorf("after attempt 2, expected 0 attempts remaining, got %d", attemptsRemaining2)
	}

	// Attempt 3: Should be blocked
	result3, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `func whatever() {}`,
		"replace":   `func something() {}`,
	})
	if err != nil {
		t.Fatalf("attempt 3 failed: %v", err)
	}

	resultMap3 := result3.(map[string]interface{})
	if resultMap3["success"].(bool) {
		t.Error("attempt 3 should be blocked")
	}

	if !resultMap3["attempts_exceeded"].(bool) {
		t.Error("expected attempts_exceeded=true")
	}

	if resultMap3["error"] != "Maximum retry attempts reached" {
		t.Errorf("expected max attempts error, got: %v", resultMap3["error"])
	}
}

// TestFileEditTool_RecoveryTracker_SuccessClearsHistory tests that success clears error history
func TestFileEditTool_RecoveryTracker_SuccessClearsHistory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func Add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tracker := recovery.NewTracker()
	tool, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithRecoveryTracker(tracker),
	)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := context.Background()

	// Attempt 1: Fail (wrong text)
	_, _ = tool.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `func nonexistent() {}`,
		"replace":   `func something() {}`,
	})

	// Verify error was recorded
	if tracker.GetAttemptCount("test.go") != 1 {
		t.Error("expected 1 attempt recorded")
	}

	// Attempt 2: Succeed (correct text)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `return a + b`,
		"replace":   `return a * b`,
	})
	if err != nil {
		t.Fatalf("success attempt failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Error("expected success")
	}

	// Verify history was cleared
	if tracker.GetAttemptCount("test.go") != 0 {
		t.Error("expected history to be cleared after success")
	}
}

// TestFileEditTool_RecoveryTracker_Suggestions tests that suggestions are provided
func TestFileEditTool_RecoveryTracker_Suggestions(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func main() {
	fmt.Println("Hello")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(WithEditBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := context.Background()

	// Trigger a no-match error
	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `func nonexistent() {}`,
		"replace":   `func something() {}`,
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["success"].(bool) {
		t.Error("expected failure")
	}

	// Check that suggestions are provided
	suggestions, ok := resultMap["suggestions"].([]string)
	if !ok {
		t.Fatal("expected suggestions array")
	}

	if len(suggestions) == 0 {
		t.Error("expected at least one suggestion")
	}

	// Verify suggestions contain helpful advice
	hasReadFileSuggestion := false
	for _, suggestion := range suggestions {
		if containsSubstring(suggestion, "read") || containsSubstring(suggestion, "Read") {
			hasReadFileSuggestion = true
			break
		}
	}

	if !hasReadFileSuggestion {
		t.Error("expected suggestion to read file, got:", suggestions)
	}
}

// TestFileEditTool_RecoveryTracker_AlternativeApproach tests alternative approach suggestions
func TestFileEditTool_RecoveryTracker_AlternativeApproach(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func main() {
	fmt.Println("Hello")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(WithEditBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := context.Background()

	// Trigger 2 consecutive failures
	for i := 0; i < 2; i++ {
		_, _ = tool.Execute(ctx, map[string]interface{}{
			"operation": "fuzzy_replace",
			"file_path": "test.go",
			"search":    `func nonexistent() {}`,
			"replace":   `func something() {}`,
		})
	}

	// Third attempt should include alternative approach
	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `func stillWrong() {}`,
		"replace":   `func something() {}`,
	})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	alternative, ok := resultMap["alternative"].(string)
	if !ok || alternative == "" {
		t.Error("expected alternative approach after multiple failures")
	}

	// Should suggest line_replace
	if !containsSubstring(alternative, "line_replace") {
		t.Errorf("expected suggestion to use line_replace, got: %s", alternative)
	}
}

// TestFileEditTool_RecoveryTracker_PathValidation tests path validation error tracking
func TestFileEditTool_RecoveryTracker_PathValidation(t *testing.T) {
	tmpDir := t.TempDir()

	tracker := recovery.NewTracker()
	tool, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithRecoveryTracker(tracker),
	)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := context.Background()

	// Try path traversal (will cause path validation error)
	_, err = tool.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "../../../etc/passwd",
		"search":    `something`,
		"replace":   `something`,
	})

	// Should get error
	if err == nil {
		t.Fatal("expected path validation error")
	}

	// Check that error history was recorded with correct type
	history := tracker.GetHistory("../../../etc/passwd")
	if len(history) != 1 {
		t.Fatalf("expected 1 error recorded, got %d", len(history))
	}

	if history[0].ErrorType != recovery.ErrorTypePathValidation {
		t.Errorf("expected ErrorTypePathValidation, got %s", history[0].ErrorType)
	}
}

// TestFileEditTool_RecoveryTracker_SharedTracker tests sharing tracker across tools
func TestFileEditTool_RecoveryTracker_SharedTracker(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func main() {
	fmt.Println("Hello")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create shared tracker
	sharedTracker := recovery.NewTracker(recovery.WithMaxAttempts(2))

	// Create two tools sharing the same tracker
	tool1, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithRecoveryTracker(sharedTracker),
	)
	if err != nil {
		t.Fatalf("failed to create tool1: %v", err)
	}

	tool2, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithRecoveryTracker(sharedTracker),
	)
	if err != nil {
		t.Fatalf("failed to create tool2: %v", err)
	}

	ctx := context.Background()

	// Fail with tool1
	_, _ = tool1.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `func nonexistent() {}`,
		"replace":   `func something() {}`,
	})

	// Verify tool2 sees the same error count
	if sharedTracker.GetAttemptCount("test.go") != 1 {
		t.Error("shared tracker should show 1 attempt")
	}

	// Fail with tool2 (should hit max attempts)
	_, _ = tool2.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `func alsoWrong() {}`,
		"replace":   `func something() {}`,
	})

	// Both tools should now be blocked
	result, _ := tool1.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `func whatever() {}`,
		"replace":   `func something() {}`,
	})

	resultMap := result.(map[string]interface{})
	if !resultMap["attempts_exceeded"].(bool) {
		t.Error("expected both tools to be blocked by shared tracker")
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
