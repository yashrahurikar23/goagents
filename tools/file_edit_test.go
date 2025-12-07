package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFileEditTool_FuzzyReplace_ExactMatch tests fuzzy replacement with exact match
func TestFileEditTool_FuzzyReplace_ExactMatch(t *testing.T) {
	// Setup temporary directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// Create test file
	content := `package main

func main() {
	fmt.Println("Hello, World!")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create tool
	tool, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithEditAllowWrite(true),
	)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Execute fuzzy replace
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `fmt.Println("Hello, World!")`,
		"replace":   `fmt.Println("Hello, GoAgents!")`,
	})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Verify result
	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Errorf("expected success=true, got false: %v", resultMap)
	}

	confidence := resultMap["confidence"].(float64)
	if confidence != 1.0 {
		t.Errorf("expected confidence=1.0 for exact match, got %f", confidence)
	}

	// Verify file content
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(newContent), "Hello, GoAgents!") {
		t.Errorf("file not updated correctly: %s", string(newContent))
	}
}

// TestFileEditTool_FuzzyReplace_WhitespaceDifference tests fuzzy matching with whitespace
func TestFileEditTool_FuzzyReplace_WhitespaceDifference(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "calculator.go")

	// Original code (2 spaces indentation)
	content := `func Calculate(a, b int) int {
  return a + b
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(WithEditBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// LLM's search block (4 spaces indentation)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "calculator.go",
		"search": `func Calculate(a, b int) int {
    return a + b
}`,
		"replace": `func Calculate(a, b int) int {
    return a * b
}`,
	})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Errorf("fuzzy match should succeed despite whitespace difference: %v", resultMap)
	}

	confidence := resultMap["confidence"].(float64)
	if confidence < 0.90 {
		t.Errorf("expected high confidence (>0.90), got %f", confidence)
	}

	// Verify multiplication is there
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(newContent), "a * b") {
		t.Errorf("file not updated correctly: %s", string(newContent))
	}
}

// TestFileEditTool_FuzzyReplace_NoMatch tests handling of no match
func TestFileEditTool_FuzzyReplace_NoMatch(t *testing.T) {
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

	// Try to replace something that doesn't exist
	result, err := tool.Execute(context.Background(), map[string]interface{}{
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
		t.Error("expected success=false for no match")
	}

	if resultMap["error"] == nil {
		t.Error("expected error message for no match")
	}
}

// TestFileEditTool_MultiReplace tests multiple replacements
func TestFileEditTool_MultiReplace(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "math.go")

	content := `package math

func Add(a, b int) int {
	return a + b
}

func Subtract(a, b int) int {
	return a - b
}

func Multiply(a, b int) int {
	return a * b
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(WithEditBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Replace multiple functions
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "multi_replace",
		"file_path": "math.go",
		"replacements": []interface{}{
			map[string]interface{}{
				"search":  "return a + b",
				"replace": "result := a + b\n\treturn result",
			},
			map[string]interface{}{
				"search":  "return a - b",
				"replace": "result := a - b\n\treturn result",
			},
			map[string]interface{}{
				"search":  "return a * b",
				"replace": "result := a * b\n\treturn result",
			},
		},
	})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Errorf("expected success=true: %v", resultMap)
	}

	successCount := int(resultMap["success_count"].(int))
	if successCount != 3 {
		t.Errorf("expected 3 successful replacements, got %d", successCount)
	}

	// Verify file content
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expectedCount := strings.Count(string(newContent), "result :=")
	if expectedCount != 3 {
		t.Errorf("expected 3 'result :=' occurrences, got %d", expectedCount)
	}
}

// TestFileEditTool_LineReplace tests line range replacement
func TestFileEditTool_LineReplace(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.txt")

	content := `line 1
line 2
line 3
line 4
line 5
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(WithEditBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Replace lines 2-4
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation":  "line_replace",
		"file_path":  "config.txt",
		"line_start": float64(2),
		"line_end":   float64(4),
		"replace":    "new middle content",
	})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Errorf("expected success=true: %v", resultMap)
	}

	// Verify file content
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	lines := strings.Split(string(newContent), "\n")
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}

	if lines[0] != "line 1" {
		t.Errorf("line 1 should be preserved, got: %s", lines[0])
	}

	if lines[1] != "new middle content" {
		t.Errorf("line 2 should be replaced, got: %s", lines[1])
	}

	if lines[2] != "line 5" {
		t.Errorf("line 5 should be preserved, got: %s", lines[2])
	}
}

// TestFileEditTool_PreviewMode tests read-only mode
func TestFileEditTool_PreviewMode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	originalContent := `package main

func main() {
	fmt.Println("Original")
}
`
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithEditAllowWrite(false), // Preview mode
	)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Try to edit
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `fmt.Println("Original")`,
		"replace":   `fmt.Println("Modified")`,
	})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Errorf("expected success=true in preview: %v", resultMap)
	}

	if !resultMap["preview"].(bool) {
		t.Error("expected preview=true")
	}

	// Verify file is unchanged
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != originalContent {
		t.Error("file should not be modified in preview mode")
	}
}

// TestFileEditTool_PathTraversal tests security - path traversal prevention
func TestFileEditTool_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileEditTool(WithEditBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Try to access parent directory
	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "../../../etc/passwd",
		"search":    "something",
		"replace":   "something",
	})

	if err == nil {
		t.Error("expected error for path traversal attempt")
	}

	if !strings.Contains(err.Error(), "traversal") && !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("expected path traversal error, got: %v", err)
	}
}

// TestFileEditTool_AbsolutePath tests security - absolute path prevention
func TestFileEditTool_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileEditTool(WithEditBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Try to use absolute path
	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "/etc/passwd",
		"search":    "something",
		"replace":   "something",
	})

	if err == nil {
		t.Error("expected error for absolute path")
	}

	if !strings.Contains(err.Error(), "absolute") {
		t.Errorf("expected absolute path error, got: %v", err)
	}
}

// TestFileEditTool_FileSize tests file size limit
func TestFileEditTool_FileSize(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Create file larger than limit
	largeContent := strings.Repeat("x", 2*1024*1024) // 2MB
	if err := os.WriteFile(testFile, []byte(largeContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithEditMaxFileSize(1024*1024), // 1MB limit
	)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Try to edit large file
	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "large.txt",
		"search":    "xxx",
		"replace":   "yyy",
	})

	if err == nil {
		t.Error("expected error for file too large")
	}

	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected file size error, got: %v", err)
	}
}

// TestFileEditTool_Backup tests backup file creation
func TestFileEditTool_Backup(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	originalContent := `package main

func main() {
	fmt.Println("Original")
}
`
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithCreateBackup(true),
	)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Edit file
	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "test.go",
		"search":    `fmt.Println("Original")`,
		"replace":   `fmt.Println("Modified")`,
	})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Verify backup exists
	backupFile := testFile + ".bak"
	backupContent, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("backup file not created: %v", err)
	}

	if string(backupContent) != originalContent {
		t.Error("backup content doesn't match original")
	}

	// Verify main file is modified
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(newContent), "Modified") {
		t.Error("main file not updated")
	}
}

// TestFileEditTool_RealWorldCode tests with actual Go code
func TestFileEditTool_RealWorldCode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "calculator.go")

	// Real Go code with a bug
	buggyCode := `package calculator

import "fmt"

func Divide(a, b float64) float64 {
	return a / b
}
`
	if err := os.WriteFile(testFile, []byte(buggyCode), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(WithEditBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Fix the bug (add division by zero check)
	fixedCode := `func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}`

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "calculator.go",
		"search": `func Divide(a, b float64) float64 {
	return a / b
}`,
		"replace": fixedCode,
	})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Errorf("expected success: %v", resultMap)
	}

	// Verify fix is applied
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(newContent), "division by zero") {
		t.Error("bug fix not applied")
	}

	if !strings.Contains(string(newContent), "fmt.Errorf") {
		t.Error("error return not added")
	}
}

// TestFileEditTool_LineAnchor tests line anchor optimization
func TestFileEditTool_LineAnchor(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.go")

	// Create large file with target at specific line
	var content strings.Builder
	content.WriteString("package main\n\n")
	for i := 1; i <= 100; i++ {
		content.WriteString(fmt.Sprintf("// Comment line %d\n", i))
	}
	content.WriteString("func targetFunction() {\n")
	content.WriteString("\treturn 42\n")
	content.WriteString("}\n")
	for i := 101; i <= 200; i++ {
		content.WriteString(fmt.Sprintf("// Comment line %d\n", i))
	}

	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(WithEditBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Use line anchor near target (around line 102)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation":   "fuzzy_replace",
		"file_path":   "large.go",
		"line_anchor": float64(105),
		"search":      "return 42",
		"replace":     "return 100",
	})

	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Errorf("expected success with line anchor: %v", resultMap)
	}

	// Verify replacement
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(newContent), "return 100") {
		t.Error("replacement not applied")
	}
}

// TestFileEditTool_IntegrationScenario tests a realistic agent workflow
func TestFileEditTool_IntegrationScenario(t *testing.T) {
	tmpDir := t.TempDir()

	// Scenario: Agent fixes bugs in a calculator
	calcFile := filepath.Join(tmpDir, "calculator.go")
	calcContent := `package calculator

func Add(a, b int) int {
	return a + b
}

func Divide(a, b int) int {
	return a / b
}

func Multiply(a, b int) int {
	return a * b
}
`
	if err := os.WriteFile(calcFile, []byte(calcContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool, err := NewFileEditTool(
		WithEditBaseDir(tmpDir),
		WithCreateBackup(true),
	)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	ctx := context.Background()

	// Step 1: Fix divide by zero bug
	result1, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "fuzzy_replace",
		"file_path": "calculator.go",
		"search": `func Divide(a, b int) int {
	return a / b
}`,
		"replace": `func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}`,
	})
	if err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	if !result1.(map[string]interface{})["success"].(bool) {
		t.Error("step 1 should succeed")
	}

	// Step 2: Add input validation to all functions (multi-replace)
	result2, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "multi_replace",
		"file_path": "calculator.go",
		"replacements": []interface{}{
			map[string]interface{}{
				"search":  "package calculator",
				"replace": "package calculator\n\nimport \"errors\"",
			},
		},
	})
	if err != nil {
		t.Fatalf("step 2 failed: %v", err)
	}
	if !result2.(map[string]interface{})["success"].(bool) {
		t.Error("step 2 should succeed")
	}

	// Verify final state
	finalContent, err := os.ReadFile(calcFile)
	if err != nil {
		t.Fatalf("failed to read final file: %v", err)
	}

	finalStr := string(finalContent)
	if !strings.Contains(finalStr, "division by zero") {
		t.Error("division by zero check not added")
	}
	if !strings.Contains(finalStr, "import \"errors\"") {
		t.Error("import not added")
	}

	// Verify backup was created
	if _, err := os.Stat(calcFile + ".bak"); os.IsNotExist(err) {
		t.Error("backup file not created")
	}
}
