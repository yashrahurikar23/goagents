// Example 3: Multiple Replacements in One File
//
// This example demonstrates making multiple code changes in a single
// operation. This is useful when you need to fix multiple related issues
// or apply a pattern across multiple locations in the same file.
//
// USE CASE: Add input validation to multiple calculator functions
//
// WHY MULTI_REPLACE?
// - Apply multiple fixes in one API call
// - Maintains consistency across changes
// - Atomic operation (all succeed or all fail)
// - Reduces back-and-forth iterations

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

	fmt.Println("Example 3: Multiple Replacements")
	fmt.Println("=================================\n")

	fmt.Println("SCENARIO: Add input validation to all math functions")
	fmt.Println("-----------------------------------------------------\n")

	// Define multiple replacements
	// Each replacement is a {search, replace} pair
	replacements := []map[string]interface{}{
		{
			"search": `func Sqrt(x float64) float64 {
    return math.Sqrt(x)
}`,
			"replace": `func Sqrt(x float64) (float64, error) {
    if x < 0 {
        return 0, fmt.Errorf("cannot take square root of negative number")
    }
    return math.Sqrt(x), nil
}`,
		},
		{
			"search": `func Divide(a, b float64) float64 {
    return a / b
}`,
			"replace": `func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}`,
		},
		{
			"search": `func Modulo(a, b int) int {
    return a % b
}`,
			"replace": `func Modulo(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("modulo by zero")
    }
    return a % b, nil
}`,
		},
	}

	// Perform multi-replace
	args := map[string]interface{}{
		"operation":    "multi_replace",
		"file_path":    "calculator.go",
		"replacements": replacements,
	}

	result, err := editTool.Execute(ctx, args)
	if err != nil {
		log.Fatalf("❌ Multi-replace failed: %v", err)
	}

	// Check result
	resultMap := result.(map[string]interface{})
	if resultMap["success"].(bool) {
		fmt.Println("✅ All replacements successful!")
		fmt.Println("-------------------------------")

		replacementCount := resultMap["replacements"].(int)
		fmt.Printf("Total replacements: %d\n", replacementCount)
		fmt.Printf("Backup created: calculator.go.backup\n")
		fmt.Println()

		// Show individual replacement results
		if results, ok := resultMap["results"].([]interface{}); ok {
			fmt.Println("Individual Results:")
			for i, r := range results {
				rMap := r.(map[string]interface{})
				line := rMap["line"].(int)
				confidence := rMap["confidence"].(float64) * 100
				fmt.Printf("  %d. Line %d - %.1f%% confidence\n", i+1, line, confidence)
			}
		}
	} else {
		fmt.Println("❌ Multi-replace failed")
		fmt.Printf("Error: %v\n", resultMap["error"])
	}

	fmt.Println("\n📊 Comparison:")
	fmt.Println("┌─────────────────────┬──────────────┬─────────────┐")
	fmt.Println("│ Approach            │ API Calls    │ Risk        │")
	fmt.Println("├─────────────────────┼──────────────┼─────────────┤")
	fmt.Println("│ Individual replaces │ 3 × 3 = 9    │ High        │")
	fmt.Println("│ Rewrite entire file │ 3            │ Very High   │")
	fmt.Println("│ Multi-replace       │ 1            │ Low         │")
	fmt.Println("└─────────────────────┴──────────────┴─────────────┘")

	fmt.Println("\n💡 Best Practices:")
	fmt.Println("- Use multi_replace for related changes in same file")
	fmt.Println("- Order matters: earlier replacements affect later ones")
	fmt.Println("- Each replacement gets fuzzy matching (10% tolerance)")
	fmt.Println("- Atomic: All succeed or all fail (can be configured)")
	fmt.Println("- Creates single backup before any changes")

	fmt.Println("\n⚠️  Important Notes:")
	fmt.Println("- Replacements applied in order (top to bottom)")
	fmt.Println("- Later replacements see results of earlier ones")
	fmt.Println("- If one fails, you can configure to continue or stop")
	fmt.Println("- Test on a copy first for complex multi-replace operations")
}
