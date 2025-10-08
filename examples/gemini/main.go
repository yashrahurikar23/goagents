package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yashrahurikar23/goagents/agent"
	"github.com/yashrahurikar23/goagents/core"
	"github.com/yashrahurikar23/goagents/llm/gemini"
	"github.com/yashrahurikar23/goagents/tools"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// Create Gemini client
	llm := gemini.New(
		gemini.WithAPIKey(apiKey),
		gemini.WithModel(gemini.ModelGemini15Flash),
		gemini.WithTemperature(0.7),
	)

	fmt.Println("✨ Google Gemini Example")
	fmt.Println("========================")
	fmt.Println()

	// Example 1: Simple completion
	fmt.Println("Example 1: Simple Completion")
	fmt.Println("-----------------------------")
	result, err := llm.Complete(context.Background(), "Explain quantum computing in one sentence.")
	if err != nil {
		log.Fatalf("Completion failed: %v", err)
	}
	fmt.Printf("Gemini: %s\n\n", result)

	// Example 2: ReAct agent with tools
	fmt.Println("Example 2: ReAct Agent with Calculator")
	fmt.Println("---------------------------------------")

	calc := tools.NewCalculator()
	reactAgent := agent.NewReActAgent(llm)
	reactAgent.AddTool(calc)

	response, err := reactAgent.Run(context.Background(),
		"Calculate the compound interest on $1000 at 5% for 3 years.")
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
		"I'm planning a trip to Japan. Remember this.")
	if err != nil {
		log.Fatalf("Conversation failed: %v", err)
	}
	fmt.Printf("You: I'm planning a trip to Japan. Remember this.\n")
	fmt.Printf("Gemini: %s\n\n", resp1.Content)

	// Second turn - test memory
	resp2, err := convAgent.Run(context.Background(),
		"What am I planning?")
	if err != nil {
		log.Fatalf("Conversation failed: %v", err)
	}
	fmt.Printf("You: What am I planning?\n")
	fmt.Printf("Gemini: %s\n\n", resp2.Content)

	// Example 4: Using different Gemini models
	fmt.Println("Example 4: Different Gemini Models")
	fmt.Println("----------------------------------")

	models := []struct {
		name  string
		model string
	}{
		{"Gemini 2.0 Flash", gemini.ModelGemini20Flash},
		{"Gemini 1.5 Flash", gemini.ModelGemini15Flash},
		{"Gemini 1.5 Flash 8B", gemini.ModelGemini15Flash8B},
		{"Gemini 1.5 Pro", gemini.ModelGemini15Pro},
	}

	for _, m := range models {
		client := gemini.New(
			gemini.WithAPIKey(apiKey),
			gemini.WithModel(m.model),
		)

		result, err := client.Complete(context.Background(),
			"Say hello in 3 words.")
		if err != nil {
			log.Printf("Model %s failed: %v", m.name, err)
			continue
		}

		fmt.Printf("%s: %s\n", m.name, result)
	}

	// Example 5: With system instruction
	fmt.Println()
	fmt.Println("Example 5: System Instructions")
	fmt.Println("-------------------------------")

	systemLLM := gemini.New(
		gemini.WithAPIKey(apiKey),
		gemini.WithModel(gemini.ModelGemini15Flash),
	)

	ctx := context.Background()
	messages := []core.Message{
		{Role: "system", Content: "You are a helpful math tutor. Always explain concepts clearly."},
		{Role: "user", Content: "What is a prime number?"},
	}

	resp, err := systemLLM.Chat(ctx, messages)
	if err != nil {
		log.Printf("Chat with system instruction failed: %v", err)
	} else {
		fmt.Printf("Gemini (as math tutor): %s\n", resp.Content)
	}

	fmt.Println()
	fmt.Println("✅ All examples completed successfully!")
	fmt.Println()
	fmt.Println("Note: Gemini offers generous free tier limits!")
	fmt.Println("Get your API key at: https://aistudio.google.com/apikey")
}
