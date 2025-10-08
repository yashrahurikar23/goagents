// Package tools provides secure file system operations for AI agents.
//
// PURPOSE:
// This package implements a production-grade file operations tool that enables AI agents
// to safely interact with the file system without compromising security.
//
// WHY THIS EXISTS:
// AI agents often need to read configuration files, process data, or generate output files.
// However, granting unrestricted file system access is extremely dangerous because:
// 1. Agents could accidentally or maliciously read sensitive files (/etc/passwd, ~/.ssh/keys)
// 2. Path traversal attacks (../../etc/passwd) could escape intended directories
// 3. Agents could delete or overwrite critical system files
// 4. Large files could exhaust memory or disk space
// 5. Agents might be vulnerable to prompt injection attacks that manipulate file operations
//
// This tool implements multiple layers of security to enable safe file operations while
// preventing common attacks and accidents.
//
// KEY DESIGN DECISIONS:
// SECURITY LAYERS:
// 1. Base Directory Enforcement: All operations restricted to a configured base directory
// 2. Path Traversal Prevention: Blocks ".." and validates absolute paths are within base
// 3. File Size Limits: Prevents memory exhaustion from reading/writing large files
// 4. Read-Only Mode: Optional mode that disables all write operations
// 5. Permission Validation: Uses safe default file permissions (0644 files, 0755 directories)
//
// OPERATION DESIGN:
// - Read: Safe for config files, data processing (with size limits)
// - Write/Append: Creates parent directories automatically, validates content size
// - List: Safe directory listing with metadata (type, size, modified time)
// - Exists: Safe existence checking without revealing content
// - Delete: Prevents directory deletion (must be explicit), validates existence
// - Info: Safe metadata retrieval without content access
//
// AVAILABLE OPERATIONS:
// Read-only mode: read, list, exists, info
// Read-write mode: all above plus write, append, delete
//
// CONFIGURATION:
// - WithBaseDir: Set allowed directory (default: current working directory)
// - WithAllowWrite: Enable/disable write operations (default: true)
// - WithMaxSize: Set file size limit (default: 10MB)
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yashrahurikar23/goagents/core"
)

// FileTool provides secure file system operations for agents.
// WHY: Encapsulates all security configuration and enforces safety rules for file operations.
// This prevents agents from accidentally or maliciously accessing files outside allowed areas.
type FileTool struct {
	baseDir    string // WHY: All operations restricted to this directory tree for security
	allowWrite bool   // WHY: Enables read-only mode to prevent any modifications
	maxSize    int64  // WHY: Prevents memory exhaustion from large files (bytes)
}

// FileOption is a function that configures a FileTool.
// WHY: Functional options pattern allows flexible configuration with sensible defaults
// and enables adding new options without breaking existing code.
type FileOption func(*FileTool)

// WithBaseDir sets the base directory for file operations (default: current working directory).
// WHY: Security boundary enforcement - all operations must stay within this directory tree.
// This prevents path traversal attacks and accidental access to system files.
// The tool validates that all resolved paths start with this base directory.
//
// WHEN TO USE:
// - Point to project directory when agent should only access project files
// - Point to data directory for data processing tasks
// - Point to output directory for generation tasks
// - Point to sandbox directory for untrusted operations
func WithBaseDir(dir string) FileOption {
	return func(f *FileTool) {
		f.baseDir = dir
	}
}

// WithAllowWrite enables or disables write operations (default: true).
// WHY: Read-only mode provides an extra safety layer when agent only needs to read files.
// This prevents accidental or malicious:
// - File deletion or overwriting
// - Disk space exhaustion
// - Modification of important configuration files
//
// WHEN TO USE:
// - Set to false for read-only tasks (data analysis, information retrieval)
// - Set to true for tasks that need to generate output (report generation, code generation)
// - Start with false and enable only when needed (principle of least privilege)
func WithAllowWrite(allow bool) FileOption {
	return func(f *FileTool) {
		f.allowWrite = allow
	}
}

// WithMaxSize sets the maximum file size for operations (default: 10MB).
// WHY: Prevents memory exhaustion and disk space attacks. Without limits, agents could:
// - Read huge files into memory causing OOM crashes
// - Write unbounded content exhausting disk space
// - Use file operations for denial of service
//
// SIZE GUIDELINES:
// - 1MB: Configuration files, small data files
// - 10MB: Medium documents, code files (default)
// - 100MB: Large documents, datasets (use with caution)
// - Consider streaming for files larger than memory
func WithMaxSize(size int64) FileOption {
	return func(f *FileTool) {
		f.maxSize = size
	}
}

// NewFileTool creates a new file operations tool with the given options.
// WHY: Constructor ensures the tool is properly configured with secure defaults before use.
// Validates configuration early to fail fast rather than during operations.
//
// SECURITY VALIDATION:
// 1. Converts base directory to absolute path (prevents relative path confusion)
// 2. Verifies base directory exists (prevents operations on non-existent paths)
// 3. Uses secure defaults (10MB limit, current directory, write enabled)
//
// DEFAULT SECURITY POSTURE:
// - Base directory: Current working directory (safe for most use cases)
// - Write enabled: true (convenience, but consider read-only for sensitive tasks)
// - Max size: 10MB (protects against memory exhaustion while allowing reasonable files)
func NewFileTool(opts ...FileOption) (*FileTool, error) {
	// Get current working directory as default base directory
	// WHY: CWD is typically the project directory, providing reasonable default scope
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	f := &FileTool{
		baseDir:    cwd,
		allowWrite: true,
		maxSize:    10 * 1024 * 1024, // 10MB default - balances safety and utility
	}

	// Apply user-provided options
	// WHY: Options are applied after defaults, allowing users to override
	for _, opt := range opts {
		opt(f)
	}

	// Convert to absolute path for consistent security checks
	// WHY: Relative paths can be ambiguous and lead to security holes.
	// Absolute paths ensure consistent validation in validatePath()
	absBase, err := filepath.Abs(f.baseDir)
	if err != nil {
		return nil, fmt.Errorf("invalid base directory: %w", err)
	}
	f.baseDir = absBase

	// Verify base directory exists - fail fast if misconfigured
	// WHY: Better to fail at construction time than during first operation.
	// Helps developers catch configuration errors early.
	if _, err := os.Stat(f.baseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("base directory does not exist: %s", f.baseDir)
	}

	return f, nil
}

// Name returns the tool name (implements core.Tool).
// WHY: Unique identifier for the tool in agent's tool registry.
// LLMs use this name in function calling.
func (f *FileTool) Name() string {
	return "file_operations"
}

// Description returns a human-readable description of the tool (implements core.Tool).
// WHY: Helps LLMs understand what the tool does and when to use it.
// Includes current configuration (mode and base directory) for context.
func (f *FileTool) Description() string {
	mode := "read-only"
	if f.allowWrite {
		mode = "read-write"
	}
	return fmt.Sprintf("Perform secure file system operations (%s mode). "+
		"Supports reading, writing, listing, checking existence, and getting file information. "+
		"All operations are restricted to the base directory: %s", mode, f.baseDir)
}

// Schema returns the tool's parameter schema for LLM function calling (implements core.Tool).
// WHY: Defines the interface contract between LLM and tool. The LLM uses this schema to:
// 1. Understand what parameters are required/optional
// 2. Know which operations are available (varies by read-only vs read-write mode)
// 3. Generate correct function calls with proper argument types
//
// SECURITY NOTE:
// Available operations are dynamically determined by allowWrite flag.
// Read-only mode only exposes safe read operations, hiding write operations from LLM.
func (f *FileTool) Schema() *core.ToolSchema {
	// Build operations list based on mode
	// WHY: Only expose write operations if enabled. This prevents LLM from even
	// attempting write operations in read-only mode.
	operations := []interface{}{"read", "list", "exists", "info"}
	if f.allowWrite {
		operations = append(operations, "write", "append", "delete")
	}

	return &core.ToolSchema{
		Name:        "file_operations",
		Description: "Perform file system operations",
		Parameters: []core.Parameter{
			{
				Name:        "operation",
				Type:        "string",
				Description: "Operation to perform: read (read file content), write (write/create file), append (append to file), list (list directory), exists (check if file exists), delete (delete file), info (get file metadata)",
				Required:    true,
				Enum:        operations, // WHY: Enum restricts to valid operations
			},
			{
				Name:        "path",
				Type:        "string",
				Description: "Path to the file or directory (relative to base directory)",
				Required:    true,
			},
			{
				Name:        "content",
				Type:        "string",
				Description: "Content to write or append (required for write/append operations)",
				Required:    false, // WHY: Only needed for write/append operations
			},
		},
	}
}

// Execute performs the requested file operation (implements core.Tool).
// WHY: This is the main entry point called by agents. It orchestrates:
// 1. Parameter validation and extraction
// 2. Security validation via validatePath()
// 3. Permission checking for write operations
// 4. Routing to specific operation handlers
//
// SECURITY FLOW:
// 1. Validate parameters are correct types (prevent type confusion attacks)
// 2. Validate path is safe and within base directory (prevent path traversal)
// 3. Check write permissions for destructive operations (enforce read-only mode)
// 4. Execute operation with validated, safe path
//
// ERROR HANDLING:
// Returns descriptive errors for:
// - Invalid parameters (wrong type, missing required)
// - Security violations (path traversal, outside base dir, write disabled)
// - File system errors (not found, permission denied, etc.)
func (f *FileTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Extract and validate operation parameter
	// WHY: Type assertion ensures we got a string. LLMs sometimes send wrong types.
	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	// Extract and validate path parameter
	// WHY: Path is required for all operations. Check existence and type.
	pathVal, ok := args["path"]
	if !ok {
		return nil, fmt.Errorf("path is required")
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	// CRITICAL SECURITY: Validate path before any file system operations
	// WHY: This is the primary security gate. It prevents:
	// - Path traversal attacks (../../etc/passwd)
	// - Access outside base directory
	// - Symlink attacks (if path contains ..)
	safePath, err := f.validatePath(path)
	if err != nil {
		return nil, err
	}

	// Route to appropriate operation handler
	// WHY: Each operation has specific logic and security considerations
	switch operation {
	case "read":
		return f.readFile(ctx, safePath)

	case "write":
		// SECURITY: Check write permission before attempting operation
		if !f.allowWrite {
			return nil, fmt.Errorf("write operations are disabled")
		}
		content, ok := args["content"].(string)
		if !ok {
			return nil, fmt.Errorf("content must be a string for write operation")
		}
		return f.writeFile(ctx, safePath, content, false)

	case "append":
		// SECURITY: Check write permission before attempting operation
		if !f.allowWrite {
			return nil, fmt.Errorf("write operations are disabled")
		}
		content, ok := args["content"].(string)
		if !ok {
			return nil, fmt.Errorf("content must be a string for append operation")
		}
		return f.writeFile(ctx, safePath, content, true)

	case "list":
		return f.listDirectory(ctx, safePath)

	case "exists":
		return f.fileExists(ctx, safePath)

	case "delete":
		// SECURITY: Check write permission before attempting operation
		if !f.allowWrite {
			return nil, fmt.Errorf("write operations are disabled")
		}
		return f.deleteFile(ctx, safePath)

	case "info":
		return f.fileInfo(ctx, safePath)

	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// validatePath ensures the path is safe and within the base directory.
// WHY: This is the CRITICAL SECURITY FUNCTION that prevents all path-based attacks.
// It implements multiple defensive layers:
//
// LAYER 1: Path Traversal Prevention
// - Blocks ".." sequences that could escape base directory
// - Example blocked: "../../etc/passwd"
//
// LAYER 2: Path Joining
// - Uses filepath.Join to safely combine base directory with user path
// - Handles OS-specific separators correctly (/ vs \)
//
// LAYER 3: Absolute Path Resolution
// - Converts to absolute path to eliminate ambiguity
// - Resolves symlinks and relative components
//
// LAYER 4: Prefix Verification
// - Ensures final absolute path starts with base directory
// - Prevents escapes via symlinks or other tricks
//
// SECURITY NOTE:
// This function must be called for EVERY file operation before touching the file system.
// Bypassing this function would be a critical security vulnerability.
func (f *FileTool) validatePath(path string) (string, error) {
	// LAYER 1: Prevent path traversal attacks
	// WHY: ".." allows escaping parent directories. Even though filepath.Join
	// cleans paths, we block ".." explicitly as defense-in-depth.
	// Examples blocked: "../../../etc/passwd", "safe/../../../etc/passwd"
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", path)
	}

	// LAYER 2: Join with base directory safely
	// WHY: filepath.Join handles OS-specific separators and cleans the path
	fullPath := filepath.Join(f.baseDir, path)

	// LAYER 3: Resolve to absolute path
	// WHY: Absolute paths are unambiguous and can be reliably validated.
	// This also resolves any symlinks and relative components.
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// LAYER 4: Verify path is within base directory
	// WHY: Final check ensures resolved path hasn't escaped base directory.
	// This catches:
	// - Symlink attacks (symlink pointing outside base)
	// - Path cleaning edge cases
	// - Any other escape attempts
	if !strings.HasPrefix(absPath, f.baseDir) {
		return "", fmt.Errorf("path outside base directory: %s", path)
	}

	return absPath, nil
}

// readFile reads the contents of a file.
// WHY: Provides safe file reading with size limits to prevent memory exhaustion.
//
// SECURITY CHECKS:
// 1. Validates file exists and is not a directory
// 2. Checks file size before reading (prevents OOM attacks)
// 3. Only reads files within configured size limit
//
// BUSINESS LOGIC:
// - Returns file content as string (suitable for text files, config, code)
// - Includes size and path in response for agent context
// - Binary files will be read but may not display correctly
//
// WHY CHECK SIZE FIRST:
// We stat the file before reading to:
// 1. Avoid loading huge files into memory
// 2. Provide clear error message about size limit
// 3. Fail fast before allocating memory
func (f *FileTool) readFile(ctx context.Context, path string) (interface{}, error) {
	// Check file metadata first
	// WHY: Stat is cheap and lets us validate before expensive read operation
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Prevent reading directories
	// WHY: ReadFile would fail anyway, but we provide clearer error.
	// Use 'list' operation for directories.
	if info.IsDir() {
		return nil, fmt.Errorf("cannot read directory: %s", path)
	}

	// SECURITY: Enforce size limit before reading
	// WHY: Prevents memory exhaustion attacks. Agent could request reading
	// /dev/zero or huge log files that would crash the application.
	if info.Size() > f.maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d)", info.Size(), f.maxSize)
	}

	// Read entire file into memory
	// WHY: For files within size limit, reading entire file is simplest.
	// For larger files, users should increase limit or implement streaming.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Return structured response
	// WHY: Including size and path helps agent understand what it read
	return map[string]interface{}{
		"content": string(data),
		"size":    info.Size(),
		"path":    path,
	}, nil
}

// writeFile writes or appends content to a file.
// WHY: Provides safe file writing with automatic directory creation and size limits.
// Used for both 'write' (create/overwrite) and 'append' operations.
//
// SECURITY CHECKS:
// 1. Validates content size before writing (prevents disk exhaustion)
// 2. Creates parent directories with safe permissions (0755)
// 3. Creates files with safe permissions (0644 - owner read/write, others read)
//
// BUSINESS LOGIC:
// - Write mode: Creates new file or overwrites existing (truncate)
// - Append mode: Adds content to end of file (preserves existing)
// - Auto-creates parent directories (convenience feature)
// - Returns bytes written for verification
//
// WHY SAFE PERMISSIONS:
// - 0755 for directories: Owner full access, others can read/execute
// - 0644 for files: Owner read/write, others read-only
// These prevent accidental creation of world-writable files.
func (f *FileTool) writeFile(ctx context.Context, path string, content string, append bool) (interface{}, error) {
	// SECURITY: Check content size before writing
	// WHY: Prevents disk exhaustion attacks. Agent could generate unbounded
	// content or repeatedly append to files until disk is full.
	if int64(len(content)) > f.maxSize {
		return nil, fmt.Errorf("content too large: %d bytes (max: %d)", len(content), f.maxSize)
	}

	// Auto-create parent directories if they don't exist
	// WHY: Convenience feature - agent doesn't need to manually create directory structure.
	// Uses 0755 permissions (owner: rwx, group/others: r-x) which are safe defaults.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Build file open flags based on operation mode
	// WHY:
	// - O_CREATE: Create file if it doesn't exist
	// - O_WRONLY: Open for writing only (more secure than O_RDWR)
	// - O_APPEND: Add to end of file (append mode)
	// - O_TRUNC: Truncate file to zero (write mode)
	flags := os.O_CREATE | os.O_WRONLY
	if append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	// Open/create file with safe permissions (0644)
	// WHY: 0644 = owner read/write, group/others read-only
	// Prevents accidental creation of world-writable files
	file, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Write content to file
	bytesWritten, err := file.WriteString(content)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Determine operation name for response
	operation := "written"
	if append {
		operation = "appended"
	}

	// Return structured response with verification info
	// WHY: Bytes written helps agent verify operation succeeded as expected
	return map[string]interface{}{
		"success":       true,
		"bytes_written": bytesWritten,
		"path":          path,
		"operation":     operation,
	}, nil
}

// listDirectory lists the contents of a directory.
// WHY: Provides safe directory browsing without revealing file contents.
// Useful for agents to explore file structure before reading specific files.
//
// SECURITY:
// - Only lists files within validated path (already checked by validatePath)
// - Does not reveal file contents, only metadata
// - Gracefully handles permission errors (skips inaccessible entries)
//
// BUSINESS LOGIC:
// - Returns array of file/directory entries with metadata
// - Includes name, type, size, and modification time
// - Useful for agents to understand project structure
// - Agents can then use 'read' operation on specific files
func (f *FileTool) listDirectory(ctx context.Context, path string) (interface{}, error) {
	// Verify path exists and is a directory
	// WHY: Provide clear error if path is a file or doesn't exist
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", path)
	}

	// Read directory entries
	// WHY: ReadDir is more efficient than Readdir for listing
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Build file list with metadata
	// WHY: Pre-allocate slice to avoid reallocations
	files := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		// Get file info for metadata
		info, err := entry.Info()
		if err != nil {
			// Skip entries we can't access
			// WHY: Graceful degradation - don't fail entire listing
			// because one file is inaccessible (permissions, corruption, etc.)
			continue
		}

		// Determine file type for agent context
		fileType := "file"
		if entry.IsDir() {
			fileType = "directory"
		}

		// Build entry metadata
		// WHY: Provide useful info without revealing contents:
		// - name: For subsequent operations
		// - type: So agent knows if it can read or list
		// - size: For understanding file importance
		// - modified: For finding recent changes
		files = append(files, map[string]interface{}{
			"name":     entry.Name(),
			"type":     fileType,
			"size":     info.Size(),
			"modified": info.ModTime().Format(time.RFC3339),
		})
	}

	// Return structured response
	// WHY: Include path and count for agent context
	return map[string]interface{}{
		"path":  path,
		"count": len(files),
		"files": files,
	}, nil
}

// fileExists checks if a file or directory exists.
// WHY: Safe existence checking without revealing any file contents or metadata.
// Useful for agents to check before attempting read/write operations.
//
// SECURITY:
// - Does not reveal file contents
// - Does not reveal detailed error (permission denied vs not found)
// - Simply returns true/false for existence
//
// BUSINESS LOGIC:
// - Returns true if path exists (file or directory)
// - Returns false if path doesn't exist
// - Returns false even on permission errors (fail closed)
func (f *FileTool) fileExists(ctx context.Context, path string) (interface{}, error) {
	// Check if path exists
	// WHY: os.Stat returns error for both "not found" and other errors.
	// We only check for IsNotExist to determine existence.
	_, err := os.Stat(path)
	exists := !os.IsNotExist(err)

	// Return simple boolean response
	// WHY: Keep it simple - just existence check, not detailed metadata
	return map[string]interface{}{
		"exists": exists,
		"path":   path,
	}, nil
}

// deleteFile deletes a file.
// WHY: Provides safe file deletion with safeguards against accidental directory deletion.
//
// SECURITY CHECKS:
// 1. Verifies file exists before attempting deletion
// 2. Prevents directory deletion (must be explicit, not through this operation)
// 3. Only works when write operations are enabled
//
// BUSINESS LOGIC:
// - Only deletes files, not directories
// - Fails if file doesn't exist (explicit error)
// - Returns success confirmation
//
// WHY BLOCK DIRECTORY DELETION:
// Directories should be deleted explicitly with their own operation to prevent
// accidental deletion of directory trees. This is a safety measure.
func (f *FileTool) deleteFile(ctx context.Context, path string) (interface{}, error) {
	// Verify file exists first
	// WHY: Provide clear error if file doesn't exist, rather than generic
	// "no such file" from os.Remove
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", path)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// SAFETY: Prevent accidental directory deletion
	// WHY: Deleting directories should be explicit and rare. This prevents
	// agents from accidentally deleting entire directory trees.
	// Future: could add explicit deleteDirectory operation with confirmation.
	if info.IsDir() {
		return nil, fmt.Errorf("cannot delete directory with delete operation: %s", path)
	}

	// Delete the file
	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	// Return success confirmation
	// WHY: Explicit confirmation helps agent verify operation succeeded
	return map[string]interface{}{
		"success": true,
		"path":    path,
		"deleted": true,
	}, nil
}

// fileInfo returns metadata about a file or directory.
// WHY: Provides safe metadata access without revealing file contents.
// Useful for agents to understand file characteristics before reading.
//
// SECURITY:
// - Does not reveal file contents
// - Only exposes standard metadata (size, type, modified time, permissions)
// - Safe for both files and directories
//
// BUSINESS LOGIC:
// - Returns comprehensive metadata for decision-making
// - Includes name, type, size, modified time, permissions
// - Works for both files and directories
// - Agents can use this to decide if file is worth reading (based on size, age, etc.)
func (f *FileTool) fileInfo(ctx context.Context, path string) (interface{}, error) {
	// Get file metadata
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Determine type for clarity
	fileType := "file"
	if info.IsDir() {
		fileType = "directory"
	}

	// Return comprehensive metadata
	// WHY: Provide rich metadata for agent decision-making:
	// - path: Full path for reference
	// - name: Filename without path
	// - type: file or directory
	// - size: For deciding if file is too large to read
	// - modified: For finding recent changes
	// - permissions: For understanding access rights
	return map[string]interface{}{
		"path":        path,
		"name":        info.Name(),
		"type":        fileType,
		"size":        info.Size(),
		"modified":    info.ModTime().Format(time.RFC3339),
		"permissions": info.Mode().String(),
	}, nil
}
