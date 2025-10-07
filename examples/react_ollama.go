package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yashrahurikar/goagents/agent"
	"github.com/yashrahurikar/goagents/llm/ollama"
	"github.com/yashrahurikar/goagents/tools"
)

func main() {
	fmt.Println("🚀 Testing ReActAgent with Ollama\n")

	// Create Ollama client with a capable model
	llm := ollama.New(
		ollama.WithModel("llama3.2:1b"), // Use a slightly larger model for better reasoning
		ollama.WithTemperature(0.1),     // Low temperature for focused reasoning
	)

	// Create ReAct agent
	reactAgent := agent.NewReActAgent(llm)

	// Add calculator tool
	calc := tools.NewCalculator()
	if err := reactAgent.AddTool(calc); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Test 1: Simple calculation
	fmt.Println("Test 1: Simple Calculation")
	fmt.Println("Question: What is 25 * 4?")
	fmt.Println("---")

	response, err := reactAgent.Run(ctx, "What is 25 * 4?")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Answer: %s\n", response.Content)
	}

	// Show reasoning trace
	trace := reactAgent.GetTrace()
	fmt.Println("\nReasoning Trace:")
	for i, step := range trace {
		fmt.Printf("\nIteration %d:\n", i+1)
		fmt.Printf("  Thought: %s\n", step.Thought)
		fmt.Printf("  Action: %s\n", step.Action)
		fmt.Printf("  Observation: %s\n", step.Observation)
	}

	// Reset for next test
	reactAgent.Reset()

	fmt.Println("\n============================================================\n")

	// Test 2: Multi-step calculation
	fmt.Println("Test 2: Multi-Step Calculation")
	fmt.Println("Question: Calculate (10 + 5) * 2")
	fmt.Println("---")

	response2, err := reactAgent.Run(ctx, "Calculate (10 + 5) * 2. First add 10 and 5, then multiply the result by 2.")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Answer: %s\n", response2.Content)
	}

	// Show reasoning trace
	trace2 := reactAgent.GetTrace()
	fmt.Println("\nReasoning Trace:")
	for i, step := range trace2 {
		fmt.Printf("\nIteration %d:\n", i+1)
		fmt.Printf("  Thought: %s\n", step.Thought)
		fmt.Printf("  Action: %s\n", step.Action)
		fmt.Printf("  Observation: %s\n", step.Observation)
	}

	fmt.Println("\n============================================================")
	fmt.Println("✅ ReActAgent with Ollama test complete!")
}
