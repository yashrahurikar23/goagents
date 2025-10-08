package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/yashrahurikar23/goagents/agent"
	"github.com/yashrahurikar23/goagents/llm/ollama"
	"github.com/yashrahurikar23/goagents/tools"
)

func main() {
	fmt.Println("🗂️  GoAgents - File Operations Tool Demo")
	fmt.Println("========================================")

	// Create a temporary workspace for this demo
	workspace, err := os.MkdirTemp("", "goagents-file-demo-*")
	if err != nil {
		log.Fatalf("Failed to create workspace: %v", err)
	}
	defer os.RemoveAll(workspace)
	fmt.Printf("📁 Working in: %s\n\n", workspace)

	// Initialize LLM (using Ollama with a small model)
	llm := ollama.New(
		ollama.WithModel("gemma3:270m"),
	)

	// Create File Tool
	fileTool, err := tools.NewFileTool(
		tools.WithBaseDir(workspace),
		tools.WithAllowWrite(true),
		tools.WithMaxSize(1024*1024), // 1MB max
	)
	if err != nil {
		log.Fatalf("Failed to create file tool: %v", err)
	}

	// Create ReAct agent with file tool
	reactAgent := agent.NewReActAgent(llm)
	if err := reactAgent.AddTool(fileTool); err != nil {
		log.Fatalf("Failed to add file tool: %v", err)
	}

	ctx := context.Background()

	// Example 1: Create and write to a file
	fmt.Println("Example 1: Create a Shopping List")
	fmt.Println("----------------------------------")
	response1, err := reactAgent.Run(ctx,
		"Create a file called 'shopping-list.txt' and write these items: milk, bread, eggs, cheese")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response1.Content)
	}
	fmt.Println()

	// Reset for next example
	reactAgent.Reset()

	// Example 2: Read and analyze file
	fmt.Println("Example 2: Read Shopping List")
	fmt.Println("-----------------------------")
	response2, err := reactAgent.Run(ctx,
		"Read the shopping-list.txt file and tell me how many items are on the list")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response2.Content)
	}
	fmt.Println()

	// Reset for next example
	reactAgent.Reset()

	// Example 3: Append to file
	fmt.Println("Example 3: Add More Items")
	fmt.Println("-------------------------")
	response3, err := reactAgent.Run(ctx,
		"Append 'apples' and 'bananas' to the shopping-list.txt file")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response3.Content)
	}
	fmt.Println()

	// Reset for next example
	reactAgent.Reset()

	// Example 4: List directory contents
	fmt.Println("Example 4: List Files in Workspace")
	fmt.Println("-----------------------------------")
	response4, err := reactAgent.Run(ctx,
		"List all files in the current directory and tell me their names")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response4.Content)
	}
	fmt.Println()

	// Reset for next example
	reactAgent.Reset()

	// Example 5: Check file existence
	fmt.Println("Example 5: Check If File Exists")
	fmt.Println("--------------------------------")
	response5, err := reactAgent.Run(ctx,
		"Check if a file called 'todo.txt' exists in the directory")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response5.Content)
	}
	fmt.Println()

	// Reset for next example
	reactAgent.Reset()

	// Example 6: Get file info
	fmt.Println("Example 6: Get File Information")
	fmt.Println("--------------------------------")
	response6, err := reactAgent.Run(ctx,
		"Get information about the shopping-list.txt file, including its size")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response6.Content)
	}
	fmt.Println()

	// Demonstrate security: attempt path traversal (should fail)
	fmt.Println("Example 7: Security Demo - Path Traversal Prevention")
	fmt.Println("----------------------------------------------------")
	reactAgent.Reset()
	response7, err := reactAgent.Run(ctx,
		"Try to read the file '../../../etc/passwd'")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response7.Content)
	}
	fmt.Println()

	// Demonstrate read-only mode
	fmt.Println("Example 8: Read-Only Mode")
	fmt.Println("-------------------------")

	// Create read-only file tool
	readOnlyTool, err := tools.NewFileTool(
		tools.WithBaseDir(workspace),
		tools.WithAllowWrite(false),
	)
	if err != nil {
		log.Fatalf("Failed to create read-only file tool: %v", err)
	}

	readOnlyAgent := agent.NewReActAgent(llm)
	if err := readOnlyAgent.AddTool(readOnlyTool); err != nil {
		log.Fatalf("Failed to add read-only file tool: %v", err)
	}

	response8, err := readOnlyAgent.Run(ctx,
		"Try to write 'test' to a file called 'test.txt'")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Agent: %s\n", response8.Content)
	}
	fmt.Println()

	// Show final workspace contents
	fmt.Println("Final Workspace Contents:")
	fmt.Println("------------------------")
	entries, err := os.ReadDir(workspace)
	if err != nil {
		log.Printf("Error reading workspace: %v\n", err)
	} else {
		for _, entry := range entries {
			info, _ := entry.Info()
			fmt.Printf("  - %s (%d bytes)\n", entry.Name(), info.Size())

			// Show shopping list contents
			if entry.Name() == "shopping-list.txt" {
				content, err := os.ReadFile(filepath.Join(workspace, entry.Name()))
				if err == nil {
					fmt.Printf("    Contents: %s\n", string(content))
				}
			}
		}
	}

	fmt.Println("\n✅ File Operations Tool Demo Complete!")
}
