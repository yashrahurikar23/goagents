# Error Recovery Package

The `internal/recovery` package provides error tracking and intelligent retry logic for file editing operations. It helps AI agents recover from repeated failures by analyzing error patterns and suggesting alternative approaches.

## Why This Package Exists

When AI agents attempt to edit code files, they can encounter various failures:
- **No Match Errors**: The search text isn't found in the file (70% of errors)
- **Low Confidence Matches**: Fuzzy matching finds a match but below the 85% threshold
- **Path Validation Errors**: Security checks prevent accessing certain paths
- **File Access Errors**: File doesn't exist or permissions prevent access

Without error tracking, agents might:
- Retry the same failing operation indefinitely
- Not learn from previous failures
- Miss obvious alternative approaches
- Provide poor feedback to users

This package solves these problems by:
1. **Tracking error history** per file and operation
2. **Limiting retry attempts** (default: 3 attempts in 5 minutes)
3. **Providing context-aware suggestions** based on error type
4. **Suggesting alternative approaches** after multiple failures
5. **Thread-safe concurrent access** for multi-agent scenarios

## Key Concepts

### Error Types

The package categorizes errors into 5 types:

```go
const (
    ErrorTypeNoMatch        // Search string not found
    ErrorTypeLowConfidence  // Match confidence below 85%
    ErrorTypePathValidation // Path security validation failed
    ErrorTypeFileAccess     // File read/write error
    ErrorTypeOther          // Other errors
)
```

### Retry Window

Errors are tracked within a **retry window** (default: 5 minutes). Old errors outside the window don't count toward the retry limit. This allows agents to retry files after enough time has passed.

### Recovery Advice

After each error, the tracker provides structured advice:

```go
type RecoveryAdvice struct {
    ShouldRetry           bool     // Whether retry is allowed
    AttemptsRemaining     int      // How many attempts left
    Suggestions           []string // Context-specific suggestions
    AlternativeApproach   string   // Alternative method to try
    Diff                  string   // Visual diff of what was attempted
}
```

## Basic Usage

### 1. Create a Tracker

```go
import "github.com/yashrahurikar23/goagents/internal/recovery"

// Default: 3 attempts in 5 minutes
tracker := recovery.NewTracker()

// Custom configuration
tracker := recovery.NewTracker(
    recovery.WithMaxAttempts(5),
    recovery.WithRetryWindow(10 * time.Minute),
)
```

### 2. Record Errors

```go
tracker.RecordError(recovery.ErrorRecord{
    FilePath:   "calculator.go",
    Operation:  "fuzzy_replace",
    ErrorType:  recovery.ErrorTypeNoMatch,
    ErrorMsg:   "search text not found",
    SearchText: "func Add(a, b int) int {\n    return a + b\n}",
    Confidence: 0.0,
    LineAnchor: 42,
})
```

### 3. Get Advice Before Retry

```go
advice := tracker.GetAdvice("calculator.go", "fuzzy_replace")

if !advice.ShouldRetry {
    return fmt.Errorf("max retry attempts reached")
}

fmt.Printf("Attempts remaining: %d\n", advice.AttemptsRemaining)
fmt.Println("Suggestions:")
for _, suggestion := range advice.Suggestions {
    fmt.Printf("  - %s\n", suggestion)
}

if advice.AlternativeApproach != "" {
    fmt.Printf("Alternative: %s\n", advice.AlternativeApproach)
}
```

## Integration with FileEditTool

The recovery tracker is designed to integrate with `tools/file_edit.go`:

```go
type FileEditTool struct {
    baseDir      string
    matcher      *fuzzy.Matcher
    tracker      *recovery.Tracker  // Add tracker
    allowWrite   bool
    maxFileSize  int64
    createBackup bool
}

func (f *FileEditTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    filePath := args["file_path"].(string)
    
    // Check if retry is allowed
    advice := f.tracker.GetAdvice(filePath, "fuzzy_replace")
    if !advice.ShouldRetry {
        return map[string]interface{}{
            "success": false,
            "error": "Max retry attempts reached",
            "suggestions": advice.Suggestions,
            "alternative": advice.AlternativeApproach,
        }, nil
    }
    
    // Attempt the operation
    result, err := f.fuzzyReplace(ctx, args)
    
    // Record error if failed
    if err != nil || !result["success"].(bool) {
        f.tracker.RecordError(recovery.ErrorRecord{
            FilePath:   filePath,
            Operation:  "fuzzy_replace",
            ErrorType:  determineErrorType(err, result),
            ErrorMsg:   fmt.Sprintf("%v", err),
            SearchText: args["search"].(string),
            Confidence: getConfidence(result),
        })
        
        // Add recovery advice to response
        advice = f.tracker.GetAdvice(filePath, "fuzzy_replace")
        result["suggestions"] = advice.Suggestions
        result["attempts_remaining"] = advice.AttemptsRemaining
        result["alternative"] = advice.AlternativeApproach
    } else {
        // Success! Clear error history for this file
        f.tracker.ClearHistory(filePath)
    }
    
    return result, err
}
```

## Error-Specific Suggestions

The tracker provides context-aware suggestions based on error type:

### No Match Errors

```
1. Read the file first using read_file operation to see exact content
2. Check for typos in search text - even one character matters
3. Use line_anchor parameter to narrow search area
4. Consider using line_replace with specific line numbers
5. Verify the search text still exists (code may have changed)
```

### Low Confidence Errors

```
1. Last match had 75.0% confidence (need 85%+)
2. Ensure indentation (tabs vs spaces) matches the file
3. Check line endings match (\\n vs \\r\\n)
4. You're very close - double-check whitespace carefully
```

### Path Validation Errors

```
1. Use relative paths from workspace base directory
2. Avoid '..' directory traversal attempts
3. Do not use absolute paths like '/etc/passwd'
```

### File Access Errors

```
1. Verify file exists using file_exists operation
2. Check file permissions - may not have read/write access
3. Ensure file path is correct (case-sensitive on Unix)
```

## Alternative Approaches

After 2+ consecutive failures, the tracker suggests alternatives:

- **fuzzy_replace failing** → "Try using line_replace instead: specify exact line numbers"
- **multi_replace failing** → "Consider breaking into smaller, more targeted edits"
- **file_access errors** → "Use list_directory to verify file exists"

## Statistics and Monitoring

Track overall error patterns:

```go
stats := tracker.GetStats()

fmt.Printf("Total errors: %d\n", stats["total_errors"])
fmt.Printf("Files tracked: %d\n", stats["files_tracked"])

errorsByType := stats["errors_by_type"].(map[recovery.ErrorType]int)
fmt.Printf("No match errors: %d\n", errorsByType[recovery.ErrorTypeNoMatch])
fmt.Printf("Low confidence: %d\n", errorsByType[recovery.ErrorTypeLowConfidence])
```

## Convenience Methods

### Quick Retry Check

```go
if !tracker.ShouldRetry("calculator.go") {
    return fmt.Errorf("max attempts reached")
}
```

### Get Attempt Count

```go
attempts := tracker.GetAttemptCount("calculator.go")
fmt.Printf("Current attempts: %d/%d\n", attempts, maxAttempts)
```

### Clear History

```go
// Clear specific file (on success)
tracker.ClearHistory("calculator.go")

// Clear all files (reset session)
tracker.ClearAllHistory()
```

## Thread Safety

The tracker is **fully thread-safe** using `sync.RWMutex`. Multiple agents can:
- Record errors concurrently
- Get advice concurrently
- Access history concurrently
- Clear history concurrently

Tested with 10 goroutines × 100 operations = 1000 concurrent operations.

## Performance

- **RecordError**: ~1-2μs (creates timestamp, appends to slice)
- **GetAdvice**: ~10-20μs (reads recent errors, generates suggestions)
- **GetHistory**: ~5μs (returns copy of slice)
- **GetStats**: ~50μs (iterates all errors)

All operations are fast enough for real-time agent use.

## Testing

Run the comprehensive test suite:

```bash
go test ./internal/recovery/... -v -cover
```

**Test Coverage: 89.7%** (19 tests)

Tests include:
- Basic error recording and retrieval
- Multi-attempt tracking
- Retry window expiration
- Error-type specific suggestions
- Alternative approach generation
- Statistics calculation
- Thread safety (1000 concurrent operations)
- Edge cases (empty history, max attempts, etc.)

## Configuration Options

### WithMaxAttempts

```go
// Allow up to 5 retry attempts
tracker := recovery.NewTracker(
    recovery.WithMaxAttempts(5),
)
```

**Default**: 3 attempts  
**Use case**: Increase for complex operations, decrease for strict failure policies

### WithRetryWindow

```go
// Track errors for 10 minutes
tracker := recovery.NewTracker(
    recovery.WithRetryWindow(10 * time.Minute),
)
```

**Default**: 5 minutes  
**Use case**: Longer windows for slow operations, shorter for fast iteration

## Real-World Example

```go
tracker := recovery.NewTracker()

// Attempt 1: No match
tracker.RecordError(recovery.ErrorRecord{
    FilePath:   "calculator.go",
    Operation:  "fuzzy_replace",
    ErrorType:  recovery.ErrorTypeNoMatch,
    SearchText: "func Add(a, b int) int {\n    return a + b\n}",
})

advice := tracker.GetAdvice("calculator.go", "fuzzy_replace")
// Suggests: "Read the file first to see exact content"
// AttemptsRemaining: 2

// Attempt 2: Low confidence (agent reads file, tries again with correct spacing)
tracker.RecordError(recovery.ErrorRecord{
    FilePath:   "calculator.go",
    Operation:  "fuzzy_replace",
    ErrorType:  recovery.ErrorTypeLowConfidence,
    SearchText: "func Add(a, b int) int {\n  return a + b\n}",
    Confidence: 0.78,
})

advice = tracker.GetAdvice("calculator.go", "fuzzy_replace")
// Suggests: "78% confidence - check indentation matches exactly"
// Alternative: "Try using line_replace with specific line numbers"
// AttemptsRemaining: 1

// Attempt 3: Agent tries line_replace approach instead
// SUCCESS! Clear history
tracker.ClearHistory("calculator.go")
```

## Future Enhancements

Potential future improvements:
1. **Pattern learning**: Track which alternative approaches work best
2. **Cross-file insights**: Learn from similar errors in other files
3. **Confidence trends**: Track if confidence is improving/degrading
4. **Success rate metrics**: Track overall success rate per operation
5. **Export/import history**: Persist error history across sessions

## Dependencies

- `github.com/sergi/go-diff/diffmatchpatch` - For diff generation (visual debugging)
- Standard library: `sync`, `time`, `strings`, `fmt`

## See Also

- `internal/fuzzy` - Fuzzy string matching for file editing
- `tools/file_edit.go` - FileEditTool that uses this package
- `/IMPLEMENTATION_ARCHITECTURE.md` - Overall architecture design
