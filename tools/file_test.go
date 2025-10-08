package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestDir creates a temporary directory for testing
func setupTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "file-tool-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

// TestNewFileTool tests the constructor
func TestNewFileTool(t *testing.T) {
	tmpDir := setupTestDir(t)

	tool, err := NewFileTool(WithBaseDir(tmpDir))
	if err != nil {
		t.Fatalf("NewFileTool failed: %v", err)
	}

	if tool.Name() != "file_operations" {
		t.Errorf("Expected name 'file_operations', got '%s'", tool.Name())
	}

	if !strings.Contains(tool.Description(), tmpDir) {
		t.Errorf("Description should contain base directory")
	}

	schema := tool.Schema()
	if schema.Name != "file_operations" {
		t.Errorf("Expected schema name 'file_operations', got '%s'", schema.Name)
	}
}

// TestNewFileTool_NonexistentBaseDir tests creating tool with nonexistent base directory
func TestNewFileTool_NonexistentBaseDir(t *testing.T) {
	_, err := NewFileTool(WithBaseDir("/nonexistent/directory/xyz"))
	if err == nil {
		t.Error("Expected error for nonexistent base directory")
	}
}

// TestNewFileTool_DefaultBaseDir tests creating tool without specifying base directory
func TestNewFileTool_DefaultBaseDir(t *testing.T) {
	tool, err := NewFileTool()
	if err != nil {
		t.Fatalf("NewFileTool failed: %v", err)
	}

	cwd, _ := os.Getwd()
	if tool.baseDir != cwd {
		t.Errorf("Expected base dir to be current working directory")
	}
}

// TestFileTool_Options tests functional options
func TestFileTool_Options(t *testing.T) {
	tmpDir := setupTestDir(t)

	tests := []struct {
		name    string
		options []FileOption
		check   func(*FileTool) error
	}{
		{
			name:    "WithBaseDir",
			options: []FileOption{WithBaseDir(tmpDir)},
			check: func(f *FileTool) error {
				if f.baseDir != tmpDir {
					return nil
				}
				return nil
			},
		},
		{
			name:    "WithAllowWrite",
			options: []FileOption{WithBaseDir(tmpDir), WithAllowWrite(false)},
			check: func(f *FileTool) error {
				if f.allowWrite {
					t.Error("Expected allowWrite to be false")
				}
				return nil
			},
		},
		{
			name:    "WithMaxSize",
			options: []FileOption{WithBaseDir(tmpDir), WithMaxSize(1024)},
			check: func(f *FileTool) error {
				if f.maxSize != 1024 {
					t.Errorf("Expected maxSize 1024, got %d", f.maxSize)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := NewFileTool(tt.options...)
			if err != nil {
				t.Fatalf("NewFileTool failed: %v", err)
			}
			if err := tt.check(tool); err != nil {
				t.Error(err)
			}
		})
	}
}

// TestFileTool_ReadFile tests reading files
func TestFileTool_ReadFile(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	// Create a test file
	testContent := "Hello, World!"
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte(testContent), 0644)

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "read",
		"path":      "test.txt",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["content"] != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, resultMap["content"])
	}
}

// TestFileTool_ReadFile_NotFound tests reading nonexistent file
func TestFileTool_ReadFile_NotFound(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "read",
		"path":      "nonexistent.txt",
	})

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

// TestFileTool_ReadFile_TooLarge tests reading file exceeding max size
func TestFileTool_ReadFile_TooLarge(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir), WithMaxSize(10))

	// Create a file larger than max size
	testFile := filepath.Join(tmpDir, "large.txt")
	os.WriteFile(testFile, []byte("This is larger than 10 bytes"), 0644)

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "read",
		"path":      "large.txt",
	})

	if err == nil {
		t.Error("Expected error for file exceeding max size")
	}
}

// TestFileTool_WriteFile tests writing files
func TestFileTool_WriteFile(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	ctx := context.Background()
	testContent := "New content"

	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "write",
		"path":      "new-file.txt",
		"content":   testContent,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Error("Expected success to be true")
	}

	// Verify file was created
	data, err := os.ReadFile(filepath.Join(tmpDir, "new-file.txt"))
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}
	if string(data) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(data))
	}
}

// TestFileTool_WriteFile_ReadOnly tests writing in read-only mode
func TestFileTool_WriteFile_ReadOnly(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir), WithAllowWrite(false))

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "write",
		"path":      "test.txt",
		"content":   "content",
	})

	if err == nil {
		t.Error("Expected error for write in read-only mode")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Errorf("Expected 'disabled' in error, got: %v", err)
	}
}

// TestFileTool_AppendFile tests appending to files
func TestFileTool_AppendFile(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	// Create initial file
	testFile := filepath.Join(tmpDir, "append-test.txt")
	os.WriteFile(testFile, []byte("Initial content\n"), 0644)

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "append",
		"path":      "append-test.txt",
		"content":   "Appended content",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify content was appended
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "Initial content\nAppended content"
	if string(data) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(data))
	}
}

// TestFileTool_ListDirectory tests listing directory contents
func TestFileTool_ListDirectory(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	// Create some test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "list",
		"path":      ".",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	files := resultMap["files"].([]map[string]interface{})

	if len(files) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(files))
	}

	// Check that we have the expected files
	names := make(map[string]bool)
	for _, file := range files {
		names[file["name"].(string)] = true
	}

	if !names["file1.txt"] || !names["file2.txt"] || !names["subdir"] {
		t.Error("Missing expected files in directory listing")
	}
}

// TestFileTool_ListDirectory_NotADirectory tests listing a file instead of directory
func TestFileTool_ListDirectory_NotADirectory(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	// Create a file
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "list",
		"path":      "file.txt",
	})

	if err == nil {
		t.Error("Expected error for listing a file")
	}
}

// TestFileTool_FileExists tests checking file existence
func TestFileTool_FileExists(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	// Create a test file
	os.WriteFile(filepath.Join(tmpDir, "exists.txt"), []byte("content"), 0644)

	ctx := context.Background()

	// Test existing file
	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "exists",
		"path":      "exists.txt",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !result.(map[string]interface{})["exists"].(bool) {
		t.Error("Expected exists to be true")
	}

	// Test nonexistent file
	result, err = tool.Execute(ctx, map[string]interface{}{
		"operation": "exists",
		"path":      "nonexistent.txt",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.(map[string]interface{})["exists"].(bool) {
		t.Error("Expected exists to be false")
	}
}

// TestFileTool_DeleteFile tests deleting files
func TestFileTool_DeleteFile(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	// Create a test file
	testFile := filepath.Join(tmpDir, "delete-me.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "delete",
		"path":      "delete-me.txt",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.(map[string]interface{})["success"].(bool) {
		t.Error("Expected success to be true")
	}

	// Verify file was deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should have been deleted")
	}
}

// TestFileTool_DeleteFile_ReadOnly tests deleting in read-only mode
func TestFileTool_DeleteFile_ReadOnly(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir), WithAllowWrite(false))

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "delete",
		"path":      "test.txt",
	})

	if err == nil {
		t.Error("Expected error for delete in read-only mode")
	}
}

// TestFileTool_FileInfo tests getting file information
func TestFileTool_FileInfo(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	// Create a test file
	testContent := "test content"
	testFile := filepath.Join(tmpDir, "info-test.txt")
	os.WriteFile(testFile, []byte(testContent), 0644)

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "info",
		"path":      "info-test.txt",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["name"] != "info-test.txt" {
		t.Errorf("Expected name 'info-test.txt', got '%s'", resultMap["name"])
	}
	if resultMap["type"] != "file" {
		t.Errorf("Expected type 'file', got '%s'", resultMap["type"])
	}
	if resultMap["size"].(int64) != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), resultMap["size"])
	}
}

// TestFileTool_PathTraversal tests path traversal prevention
func TestFileTool_PathTraversal(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	ctx := context.Background()

	traversalPaths := []string{
		"../etc/passwd",
		"../../secret.txt",
		"subdir/../../etc/passwd",
	}

	for _, path := range traversalPaths {
		t.Run(path, func(t *testing.T) {
			_, err := tool.Execute(ctx, map[string]interface{}{
				"operation": "read",
				"path":      path,
			})

			if err == nil {
				t.Errorf("Expected error for path traversal: %s", path)
			}
			if !strings.Contains(err.Error(), "traversal") {
				t.Errorf("Expected 'traversal' in error, got: %v", err)
			}
		})
	}
}

// TestFileTool_InvalidOperation tests unknown operation
func TestFileTool_InvalidOperation(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "invalid_op",
		"path":      "test.txt",
	})

	if err == nil {
		t.Error("Expected error for invalid operation")
	}
	if !strings.Contains(err.Error(), "unknown operation") {
		t.Errorf("Expected 'unknown operation' in error, got: %v", err)
	}
}

// TestFileTool_MissingArguments tests missing required arguments
func TestFileTool_MissingArguments(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	ctx := context.Background()

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "missing operation",
			args: map[string]interface{}{
				"path": "test.txt",
			},
		},
		{
			name: "missing path",
			args: map[string]interface{}{
				"operation": "read",
			},
		},
		{
			name: "missing content for write",
			args: map[string]interface{}{
				"operation": "write",
				"path":      "test.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Execute(ctx, tt.args)
			if err == nil {
				t.Error("Expected error for missing arguments")
			}
		})
	}
}

// TestFileTool_CreateSubdirectory tests writing to subdirectory (auto-create)
func TestFileTool_CreateSubdirectory(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "write",
		"path":      "subdir/nested/file.txt",
		"content":   "nested content",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(filepath.Join(tmpDir, "subdir", "nested", "file.txt"))
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}
	if string(data) != "nested content" {
		t.Errorf("Expected 'nested content', got '%s'", string(data))
	}
}

// TestFileTool_ReadDirectory_ShouldFail tests that reading a directory fails
func TestFileTool_ReadDirectory_ShouldFail(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	// Create a subdirectory
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "read",
		"path":      "subdir",
	})

	if err == nil {
		t.Error("Expected error for reading a directory")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("Expected 'directory' in error, got: %v", err)
	}
}

// TestFileTool_DeleteDirectory_ShouldFail tests that deleting a directory fails
func TestFileTool_DeleteDirectory_ShouldFail(t *testing.T) {
	tmpDir := setupTestDir(t)
	tool, _ := NewFileTool(WithBaseDir(tmpDir))

	// Create a subdirectory
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "delete",
		"path":      "subdir",
	})

	if err == nil {
		t.Error("Expected error for deleting a directory")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("Expected 'directory' in error, got: %v", err)
	}
}
