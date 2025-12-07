package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yashrahurikar23/goagents/core"
)

// TestFileTool_BatchRead_Success tests successful batch reading of multiple files.
func TestFileTool_BatchRead_Success(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Create test files
	file1Content := "content of file 1"
	file2Content := "content of file 2"
	file3Content := "content of file 3"

	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(file1Content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(file2Content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte(file3Content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create tool
	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Execute batch_read
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_read",
		"file_paths": []interface{}{
			"file1.txt",
			"file2.txt",
			"file3.txt",
		},
	})

	if err != nil {
		t.Fatalf("batch_read failed: %v", err)
	}

	// Validate response structure
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map response, got %T", result)
	}

	// Check overall success
	if !resultMap["success"].(bool) {
		t.Error("Expected overall success to be true")
	}

	// Check counts
	if resultMap["total"].(int) != 3 {
		t.Errorf("Expected total=3, got %v", resultMap["total"])
	}
	if resultMap["successful"].(int) != 3 {
		t.Errorf("Expected successful=3, got %v", resultMap["successful"])
	}
	if resultMap["failed"].(int) != 0 {
		t.Errorf("Expected failed=0, got %v", resultMap["failed"])
	}

	// Check files array
	files := resultMap["files"].([]map[string]interface{})
	if len(files) != 3 {
		t.Fatalf("Expected 3 file results, got %d", len(files))
	}

	// Validate file1
	if !files[0]["success"].(bool) {
		t.Errorf("Expected file1 success, got error: %v", files[0]["error"])
	}
	if files[0]["content"].(string) != file1Content {
		t.Errorf("Expected file1 content %q, got %q", file1Content, files[0]["content"])
	}

	// Validate file2
	if !files[1]["success"].(bool) {
		t.Errorf("Expected file2 success, got error: %v", files[1]["error"])
	}
	if files[1]["content"].(string) != file2Content {
		t.Errorf("Expected file2 content %q, got %q", file2Content, files[1]["content"])
	}

	// Validate file3
	if !files[2]["success"].(bool) {
		t.Errorf("Expected file3 success, got error: %v", files[2]["error"])
	}
	if files[2]["content"].(string) != file3Content {
		t.Errorf("Expected file3 content %q, got %q", file3Content, files[2]["content"])
	}

	// Validate metadata is included
	if _, ok := files[0]["size"]; !ok {
		t.Error("Expected size metadata in file result")
	}
	if _, ok := files[0]["modified"]; !ok {
		t.Error("Expected modified metadata in file result")
	}
}

// TestFileTool_BatchRead_PartialSuccess tests batch read with some failing files.
func TestFileTool_BatchRead_PartialSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only 2 out of 3 files
	if err := os.WriteFile(filepath.Join(tmpDir, "exists1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "exists2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	// missing.txt doesn't exist

	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_read",
		"file_paths": []interface{}{
			"exists1.txt",
			"missing.txt",
			"exists2.txt",
		},
	})

	if err != nil {
		t.Fatalf("batch_read failed: %v", err)
	}

	resultMap := result.(map[string]interface{})

	// Should still report overall success (at least one file succeeded)
	if !resultMap["success"].(bool) {
		t.Error("Expected overall success to be true (partial success)")
	}

	// Check counts
	if resultMap["total"].(int) != 3 {
		t.Errorf("Expected total=3, got %v", resultMap["total"])
	}
	if resultMap["successful"].(int) != 2 {
		t.Errorf("Expected successful=2, got %v", resultMap["successful"])
	}
	if resultMap["failed"].(int) != 1 {
		t.Errorf("Expected failed=1, got %v", resultMap["failed"])
	}

	// Check individual results
	files := resultMap["files"].([]map[string]interface{})

	// File 1 should succeed
	if !files[0]["success"].(bool) {
		t.Error("Expected file 1 to succeed")
	}

	// File 2 should fail
	if files[1]["success"].(bool) {
		t.Error("Expected file 2 to fail")
	}
	if files[1]["error"] == nil {
		t.Error("Expected error message for missing file")
	}

	// File 3 should succeed
	if !files[2]["success"].(bool) {
		t.Error("Expected file 3 to succeed")
	}
}

// TestFileTool_BatchRead_PathTraversal tests security validation in batch read.
func TestFileTool_BatchRead_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid file
	if err := os.WriteFile(filepath.Join(tmpDir, "valid.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_read",
		"file_paths": []interface{}{
			"valid.txt",
			"../../etc/passwd", // Path traversal attempt
		},
	})

	if err != nil {
		t.Fatalf("batch_read failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	files := resultMap["files"].([]map[string]interface{})

	// Valid file should succeed
	if !files[0]["success"].(bool) {
		t.Error("Expected valid file to succeed")
	}

	// Path traversal should fail
	if files[1]["success"].(bool) {
		t.Error("Expected path traversal to fail")
	}
	errorMsg := files[1]["error"].(string)
	if !strings.Contains(errorMsg, "path traversal") {
		t.Errorf("Expected path traversal error, got: %s", errorMsg)
	}
}

// TestFileTool_BatchRead_SizeLimit tests per-file and cumulative size limits.
func TestFileTool_BatchRead_SizeLimit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create tool with small size limit
	tool, err := NewFileTool(WithBaseDir(tmpDir), WithMaxSize(100))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Create small file (within limit)
	smallContent := strings.Repeat("a", 50)
	if err := os.WriteFile(filepath.Join(tmpDir, "small.txt"), []byte(smallContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create large file (exceeds limit)
	largeContent := strings.Repeat("b", 200)
	if err := os.WriteFile(filepath.Join(tmpDir, "large.txt"), []byte(largeContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_read",
		"file_paths": []interface{}{
			"small.txt",
			"large.txt",
		},
	})

	if err != nil {
		t.Fatalf("batch_read failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	files := resultMap["files"].([]map[string]interface{})

	// Small file should succeed
	if !files[0]["success"].(bool) {
		t.Error("Expected small file to succeed")
	}

	// Large file should fail
	if files[1]["success"].(bool) {
		t.Error("Expected large file to fail")
	}
	errorMsg := files[1]["error"].(string)
	if !strings.Contains(errorMsg, "too large") {
		t.Errorf("Expected size limit error, got: %s", errorMsg)
	}
}

// TestFileTool_BatchRead_MaxFilesLimit tests the 20-file limit.
func TestFileTool_BatchRead_MaxFilesLimit(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Try to read 21 files
	filePaths := make([]interface{}, 21)
	for i := 0; i < 21; i++ {
		filePaths[i] = "file.txt" // Doesn't need to exist for this test
	}

	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation":  "batch_read",
		"file_paths": filePaths,
	})

	if err == nil {
		t.Fatal("Expected error for exceeding max files limit")
	}
	if !strings.Contains(err.Error(), "maximum 20 files") {
		t.Errorf("Expected max files error, got: %v", err)
	}
}

// TestFileTool_BatchRead_EmptyArray tests validation of empty file_paths array.
func TestFileTool_BatchRead_EmptyArray(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation":  "batch_read",
		"file_paths": []interface{}{},
	})

	if err == nil {
		t.Fatal("Expected error for empty file_paths array")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected empty array error, got: %v", err)
	}
}

// TestFileTool_BatchWrite_Success tests successful batch writing of multiple files.
func TestFileTool_BatchWrite_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Execute batch_write
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_write",
		"files": []interface{}{
			map[string]interface{}{
				"path":    "file1.txt",
				"content": "content of file 1",
			},
			map[string]interface{}{
				"path":    "subdir/file2.txt",
				"content": "content of file 2",
			},
			map[string]interface{}{
				"path":    "file3.txt",
				"content": "content of file 3",
			},
		},
	})

	if err != nil {
		t.Fatalf("batch_write failed: %v", err)
	}

	// Validate response
	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Error("Expected overall success to be true")
	}
	if resultMap["total"].(int) != 3 {
		t.Errorf("Expected total=3, got %v", resultMap["total"])
	}
	if resultMap["successful"].(int) != 3 {
		t.Errorf("Expected successful=3, got %v", resultMap["successful"])
	}

	// Verify files were actually created
	files := resultMap["files"].([]map[string]interface{})
	if len(files) != 3 {
		t.Fatalf("Expected 3 file results, got %d", len(files))
	}

	// Check file1
	content1, err := os.ReadFile(filepath.Join(tmpDir, "file1.txt"))
	if err != nil {
		t.Errorf("Failed to read file1: %v", err)
	}
	if string(content1) != "content of file 1" {
		t.Errorf("file1 content mismatch: got %q", string(content1))
	}

	// Check file2 (in subdirectory)
	content2, err := os.ReadFile(filepath.Join(tmpDir, "subdir", "file2.txt"))
	if err != nil {
		t.Errorf("Failed to read file2: %v", err)
	}
	if string(content2) != "content of file 2" {
		t.Errorf("file2 content mismatch: got %q", string(content2))
	}

	// Check file3
	content3, err := os.ReadFile(filepath.Join(tmpDir, "file3.txt"))
	if err != nil {
		t.Errorf("Failed to read file3: %v", err)
	}
	if string(content3) != "content of file 3" {
		t.Errorf("file3 content mismatch: got %q", string(content3))
	}
}

// TestFileTool_BatchWrite_ReadOnly tests that batch_write is blocked in read-only mode.
func TestFileTool_BatchWrite_ReadOnly(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileTool(WithBaseDir(tmpDir), WithAllowWrite(false))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_write",
		"files": []interface{}{
			map[string]interface{}{
				"path":    "file1.txt",
				"content": "content",
			},
		},
	})

	if err == nil {
		t.Fatal("Expected error for batch_write in read-only mode")
	}
	if !strings.Contains(err.Error(), "write operations are disabled") {
		t.Errorf("Expected read-only error, got: %v", err)
	}
}

// TestFileTool_BatchWrite_PartialSuccess tests batch write with some failures.
func TestFileTool_BatchWrite_PartialSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_write",
		"files": []interface{}{
			map[string]interface{}{
				"path":    "valid.txt",
				"content": "valid content",
			},
			map[string]interface{}{
				"path":    "../../etc/passwd", // Path traversal
				"content": "malicious",
			},
			map[string]interface{}{
				"path":    "also_valid.txt",
				"content": "also valid",
			},
		},
	})

	if err != nil {
		t.Fatalf("batch_write failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	files := resultMap["files"].([]map[string]interface{})

	// Check counts
	if resultMap["successful"].(int) != 2 {
		t.Errorf("Expected successful=2, got %v", resultMap["successful"])
	}
	if resultMap["failed"].(int) != 1 {
		t.Errorf("Expected failed=1, got %v", resultMap["failed"])
	}

	// First file should succeed
	if !files[0]["success"].(bool) {
		t.Error("Expected first file to succeed")
	}

	// Second file should fail (path traversal)
	if files[1]["success"].(bool) {
		t.Error("Expected second file to fail")
	}

	// Third file should succeed
	if !files[2]["success"].(bool) {
		t.Error("Expected third file to succeed")
	}

	// Verify successful files were created
	if _, err := os.Stat(filepath.Join(tmpDir, "valid.txt")); err != nil {
		t.Error("Expected valid.txt to be created")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "also_valid.txt")); err != nil {
		t.Error("Expected also_valid.txt to be created")
	}
}

// TestFileTool_BatchWrite_SizeLimit tests content size validation.
func TestFileTool_BatchWrite_SizeLimit(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileTool(WithBaseDir(tmpDir), WithMaxSize(100))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_write",
		"files": []interface{}{
			map[string]interface{}{
				"path":    "small.txt",
				"content": strings.Repeat("a", 50),
			},
			map[string]interface{}{
				"path":    "large.txt",
				"content": strings.Repeat("b", 200),
			},
		},
	})

	if err != nil {
		t.Fatalf("batch_write failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	files := resultMap["files"].([]map[string]interface{})

	// Small file should succeed
	if !files[0]["success"].(bool) {
		t.Error("Expected small file to succeed")
	}

	// Large file should fail
	if files[1]["success"].(bool) {
		t.Error("Expected large file to fail")
	}
	errorMsg := files[1]["error"].(string)
	if !strings.Contains(errorMsg, "too large") {
		t.Errorf("Expected size limit error, got: %s", errorMsg)
	}
}

// TestFileTool_BatchWrite_MaxFilesLimit tests the 20-file limit.
func TestFileTool_BatchWrite_MaxFilesLimit(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Try to write 21 files
	files := make([]interface{}, 21)
	for i := 0; i < 21; i++ {
		files[i] = map[string]interface{}{
			"path":    "file.txt",
			"content": "content",
		}
	}

	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_write",
		"files":     files,
	})

	if err == nil {
		t.Fatal("Expected error for exceeding max files limit")
	}
	if !strings.Contains(err.Error(), "maximum 20 files") {
		t.Errorf("Expected max files error, got: %v", err)
	}
}

// TestFileTool_BatchWrite_MissingFields tests validation of file objects.
func TestFileTool_BatchWrite_MissingFields(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Missing 'path' field
	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_write",
		"files": []interface{}{
			map[string]interface{}{
				"content": "content",
			},
		},
	})

	if err == nil {
		t.Fatal("Expected error for missing 'path' field")
	}
	if !strings.Contains(err.Error(), "missing 'path' field") {
		t.Errorf("Expected missing path error, got: %v", err)
	}

	// Missing 'content' field
	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"operation": "batch_write",
		"files": []interface{}{
			map[string]interface{}{
				"path": "file.txt",
			},
		},
	})

	if err == nil {
		t.Fatal("Expected error for missing 'content' field")
	}
	if !strings.Contains(err.Error(), "missing 'content' field") {
		t.Errorf("Expected missing content error, got: %v", err)
	}
}

// TestFileTool_BatchOperations_InSchema tests that batch operations appear in schema.
func TestFileTool_BatchOperations_InSchema(t *testing.T) {
	tmpDir := t.TempDir()

	// Read-write mode
	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	schema := tool.Schema()

	// Find operation parameter
	var operationParam *core.Parameter
	for i := range schema.Parameters {
		if schema.Parameters[i].Name == "operation" {
			operationParam = &schema.Parameters[i]
			break
		}
	}

	if operationParam == nil {
		t.Fatal("operation parameter not found in schema")
	}

	// Check that batch operations are in enum
	operations := operationParam.Enum
	hasBatchRead := false
	hasBatchWrite := false

	for _, op := range operations {
		if op == "batch_read" {
			hasBatchRead = true
		}
		if op == "batch_write" {
			hasBatchWrite = true
		}
	}

	if !hasBatchRead {
		t.Error("batch_read not found in schema operations")
	}
	if !hasBatchWrite {
		t.Error("batch_write not found in schema operations")
	}

	// Read-only mode should have batch_read but not batch_write
	toolReadOnly, err := NewFileTool(WithBaseDir(tmpDir), WithAllowWrite(false))
	if err != nil {
		t.Fatalf("Failed to create read-only tool: %v", err)
	}

	schemaReadOnly := toolReadOnly.Schema()
	var operationParamReadOnly *core.Parameter
	for i := range schemaReadOnly.Parameters {
		if schemaReadOnly.Parameters[i].Name == "operation" {
			operationParamReadOnly = &schemaReadOnly.Parameters[i]
			break
		}
	}

	operationsReadOnly := operationParamReadOnly.Enum
	hasBatchReadRO := false
	hasBatchWriteRO := false

	for _, op := range operationsReadOnly {
		if op == "batch_read" {
			hasBatchReadRO = true
		}
		if op == "batch_write" {
			hasBatchWriteRO = true
		}
	}

	if !hasBatchReadRO {
		t.Error("batch_read not found in read-only schema")
	}
	if hasBatchWriteRO {
		t.Error("batch_write should not be in read-only schema")
	}
}
