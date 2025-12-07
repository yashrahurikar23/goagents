// Example 5: Line-Based Replace (When You Know Line Numbers)
//
// This example demonstrates line_replace operation, which is useful when
// you know exactly which lines to replace. This is faster than fuzzy
// matching and useful for AI agents that track line numbers.
//
// USE CASE: Replace specific lines when line numbers are known
//
// WHY LINE_REPLACE?
// - Faster than fuzzy matching (no search needed)
// - Precise control over what changes
// - Good for AI agents that track line numbers
// - No ambiguity about which code to replace

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yashrahurikar23/goagents/tools"
)

func main() {
	ctx := context.Background()

	// Create the file edit tool
	editTool, err := tools.NewFileEditTool(
		tools.WithEditBaseDir("testdata"),
		tools.WithEditAllowWrite(true),
		tools.WithCreateBackup(true),
	)
	if err != nil {
		log.Fatalf("Failed to create edit tool: %v", err)
	}

	fmt.Println("Example 5: Line-Based Replace")
	fmt.Println("==============================\n")

	fmt.Println("SCENARIO: Replace lines 10-15 with improved implementation")
	fmt.Println("-----------------------------------------------------------\n")

	// When to use line_replace:
	// 1. You know the exact line numbers
	// 2. You want to replace a specific line range
	// 3. You want faster execution (no fuzzy search)
	// 4. The code block is unique enough that line numbers are reliable

	args := map[string]interface{}{
		"operation":  "line_replace",
		"file_path":  "calculator.go",
		"line_start": 10, // Starting line (1-indexed)
		"line_end":   15, // Ending line (inclusive)
		"replace": `func Multiply(a, b float64) float64 {
    // Improved implementation with overflow check
    result := a * b
    if math.IsInf(result, 0) {
        fmt.Println("Warning: Result is infinity")
        return 0
    }
    return result
}`,
	}

	result, err := editTool.Execute(ctx, args)
	if err != nil {
		log.Fatalf("❌ Line replace failed: %v", err)
	}

	// Check result
	resultMap := result.(map[string]interface{})
	if resultMap["success"].(bool) {
		fmt.Println("✅ Line replace successful!")
		fmt.Println("---------------------------")
		fmt.Printf("Lines replaced: %d-%d\n", 10, 15)
		fmt.Printf("Backup created: calculator.go.backup\n")

		if message, ok := resultMap["message"].(string); ok {
			fmt.Printf("Result: %s\n", message)
		}
	} else {
		fmt.Println("❌ Line replace failed")
		fmt.Printf("Error: %v\n", resultMap["error"])
	}

	fmt.Println("\n📊 Comparison: Line Replace vs Fuzzy Replace")
	fmt.Println("┌─────────────────┬────────────────┬─────────────────┐")
	fmt.Println("│ Feature         │ Line Replace   │ Fuzzy Replace   │")
	fmt.Println("├─────────────────┼────────────────┼─────────────────┤")
	fmt.Println("│ Speed           │ Faster         │ Slower          │")
	fmt.Println("│ Precision       │ Exact lines    │ Best match      │")
	fmt.Println("│ Requires        │ Line numbers   │ Code sample     │")
	fmt.Println("│ Whitespace OK   │ No             │ Yes             │")
	fmt.Println("│ Error recovery  │ Limited        │ Excellent       │")
	fmt.Println("│ Best for        │ Known lines    │ Unknown lines   │")
	fmt.Println("└─────────────────┴────────────────┴─────────────────┘")

	fmt.Println("\n💡 When to Use Line Replace:")
	fmt.Println("✅ You have accurate line numbers")
	fmt.Println("✅ File hasn't changed since line numbers were determined")
	fmt.Println("✅ You want maximum speed")
	fmt.Println("✅ Replacing entire function or code block")
	fmt.Println("✅ AI agent tracks line numbers from previous reads")

	fmt.Println("\n❌ When NOT to Use Line Replace:")
	fmt.Println("- File may have changed (line numbers outdated)")
	fmt.Println("- Don't have line numbers")
	fmt.Println("- Want whitespace tolerance")
	fmt.Println("- Need fuzzy matching")
	fmt.Println("- Want better error recovery")

	fmt.Println("\n🔄 Typical Workflow:")
	fmt.Println("1. 📖 Read file to get current content")
	fmt.Println("2. 🔍 Parse and identify line numbers")
	fmt.Println("3. ✏️  Use line_replace with exact line range")
	fmt.Println("4. ✅ Verify the change")

	fmt.Println("\n⚠️  Important Notes:")
	fmt.Println("- Line numbers are 1-indexed (first line is 1, not 0)")
	fmt.Println("- line_end is inclusive (lines 10-15 replaces 6 lines)")
	fmt.Println("- If file changes between read and replace, line numbers may be wrong")
	fmt.Println("- Consider using line_anchor with fuzzy_replace as alternative")
}
