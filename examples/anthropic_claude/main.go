package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yashrahurikar23/goagents/agent"
	"github.com/yashrahurikar23/goagents/llm/anthropic"
	"github.com/yashrahurikar23/goagents/tools"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create Anthropic Claude client
	llm := anthropic.New(
		anthropic.WithAPIKey(apiKey),
		anthropic.WithModel(anthropic.ModelClaude35Sonnet),
		anthropic.WithTemperature(0.7),
	)

	fmt.Println("🤖 Anthropic Claude Example")
	fmt.Println("============================")
	fmt.Println()

	// Example 1: Simple completion
	fmt.Println("Example 1: Simple Completion")
	fmt.Println("-----------------------------")
	result, err := llm.Complete(context.Background(), "Explain AI in one sentence.")
	if err != nil {
		log.Fatalf("Completion failed: %v", err)
	}
	fmt.Printf("Claude: %s\n\n", result)

	// Example 2: ReAct agent with tools
	fmt.Println("Example 2: ReAct Agent with Calculator")
	fmt.Println("---------------------------------------")

	calc := tools.NewCalculator()
	reactAgent := agent.NewReActAgent(llm)
	reactAgent.AddTool(calc)

	response, err := reactAgent.Run(context.Background(),
		"What is 15% of 230? Show your work.")
	if err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
	fmt.Printf("Agent: %s\n\n", response.Content)

	// Example 3: Conversational agent with memory
	fmt.Println("Example 3: Conversational Agent")
	fmt.Println("--------------------------------")

	convAgent := agent.NewConversationalAgent(llm)

	// First turn
	resp1, err := convAgent.Run(context.Background(),
		"My favorite color is blue. Remember this.")
	if err != nil {
		log.Fatalf("Conversation failed: %v", err)
	}
	fmt.Printf("You: My favorite color is blue. Remember this.\n")
	fmt.Printf("Claude: %s\n\n", resp1.Content)

	// Second turn - test memory
	resp2, err := convAgent.Run(context.Background(),
		"What is my favorite color?")
	if err != nil {
		log.Fatalf("Conversation failed: %v", err)
	}
	fmt.Printf("You: What is my favorite color?\n")
	fmt.Printf("Claude: %s\n\n", resp2.Content)

	// Example 4: Using different Claude models
	fmt.Println("Example 4: Different Claude Models")
	fmt.Println("----------------------------------")

	models := []string{
		anthropic.ModelClaude35Sonnet,
		anthropic.ModelClaude3Opus,
		anthropic.ModelClaude3Haiku,
	}

	for _, model := range models {
		client := anthropic.New(
			anthropic.WithAPIKey(apiKey),
			anthropic.WithModel(model),
		)

		result, err := client.Complete(context.Background(),
			"Say hello in 3 words.")
		if err != nil {
			log.Printf("Model %s failed: %v", model, err)
			continue
		}

		fmt.Printf("%s: %s\n", model, result)
	}

	fmt.Println()
	fmt.Println("✅ All examples completed successfully!")
}
