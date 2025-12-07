// Example 2: Batch Read Operations
//
// This example demonstrates reading multiple files in a single API call.
// This is one of the biggest performance improvements in Phase 1.
//
// USE CASE: Read implementation file + test file for analysis
//
// WHY BATCH READ?
// - Reduces API calls by 50-95% depending on file count
// - Faster overall execution
// - Lower costs at scale
// - Same security validation per file
//
// SAVINGS:
// - 2 files: 50% reduction (2 calls → 1 call)
// - 5 files: 80% reduction (5 calls → 1 call)
// - 20 files: 95% reduction (20 calls → 1 call)

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yashrahurikar23/goagents/tools"
)

func main() {
	ctx := context.Background()

	// Create the file tool with read-only access
	fileTool, err := tools.NewFileTool(
		tools.WithBaseDir("testdata"),
		tools.WithAllowWrite(false), // Read-only for this example
	)
	if err != nil {
		log.Fatalf("Failed to create file tool: %v", err)
	}

	fmt.Println("Example 2: Batch Read Operations")
	fmt.Println("=================================\n")

	// OLD WAY: Multiple separate read calls
	fmt.Println("❌ OLD WAY: Multiple API Calls")
	fmt.Println("-------------------------------")
	fmt.Println("API Call 1: read_file('calculator.go')")
	fmt.Println("API Call 2: read_file('calculator_test.go')")
	fmt.Println("API Call 3: read_file('utils.go')")
	fmt.Println("Total: 3 API calls\n")

	// NEW WAY: Single batch read call
	fmt.Println("✅ NEW WAY: Single Batch API Call")
	fmt.Println("---------------------------------")

	args := map[string]interface{}{
		"operation": "batch_read",
		"file_paths": []interface{}{
			"calculator.go",
			"calculator_test.go",
			"utils.go",
		},
	}

	result, err := fileTool.Execute(ctx, args)
	if err != nil {
		log.Fatalf("❌ Batch read failed: %v", err)
	}

	// Parse result
	resultMap := result.(map[string]interface{})

	fmt.Printf("✅ Batch read complete!\n")
	fmt.Printf("Total files: %v\n", resultMap["total"])
	fmt.Printf("Successful: %v\n", resultMap["successful"])
	fmt.Printf("Failed: %v\n", resultMap["failed"])
	fmt.Printf("Total size: %v bytes\n", resultMap["total_size"])
	fmt.Println()

	// Access individual file results
	files := resultMap["files"].([]interface{})
	for i, f := range files {
		fileMap := f.(map[string]interface{})
		fileName := fileMap["path"].(string)
		success := fileMap["success"].(bool)

		fmt.Printf("File %d: %s\n", i+1, fileName)
		if success {
			size := fileMap["size"].(int64)
			modified := fileMap["modified"].(string)
			fmt.Printf("  ✅ Success - %d bytes, modified: %s\n", size, modified)

			// Content is available in fileMap["content"]
			// Process it as needed...
		} else {
			errorMsg := fileMap["error"].(string)
			fmt.Printf("  ❌ Failed - %s\n", errorMsg)
		}
	}

	fmt.Println("\n📊 Performance Comparison:")
	fmt.Println("┌──────────────┬────────────┬──────────────┬──────────┐")
	fmt.Println("│ Files        │ Old Way    │ New Way      │ Savings  │")
	fmt.Println("├──────────────┼────────────┼──────────────┼──────────┤")
	fmt.Println("│ 2 files      │ 2 calls    │ 1 call       │ 50%      │")
	fmt.Println("│ 5 files      │ 5 calls    │ 1 call       │ 80%      │")
	fmt.Println("│ 10 files     │ 10 calls   │ 1 call       │ 90%      │")
	fmt.Println("│ 20 files     │ 20 calls   │ 1 call       │ 95%      │")
	fmt.Println("└──────────────┴────────────┴──────────────┴──────────┘")

	fmt.Println("\n💡 Tips:")
	fmt.Println("- Max 20 files per batch (returns error if exceeded)")
	fmt.Println("- Cumulative size limit: 50MB (5x single file limit)")
	fmt.Println("- Partial success: Some files can succeed even if others fail")
	fmt.Println("- Each file validated for security (path traversal prevention)")
}
