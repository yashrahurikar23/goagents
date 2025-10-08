package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yashrahurikar23/goagents/core"
	"github.com/yashrahurikar23/goagents/llm/ollama"
)

func main() {
	client := ollama.New(
		ollama.WithModel("gemma3:270m"),
	)

	ctx := context.Background()

	// Test Complete
	fmt.Println("Testing Complete...")
	response, err := client.Complete(ctx, "Say hello in one word")
	if err != nil {
		log.Fatalf("Complete failed: %v", err)
	}
	fmt.Printf("Complete response: %s\n\n", response)

	// Test Chat
	fmt.Println("Testing Chat...")
	messages := []core.Message{
		core.UserMessage("What is 2+2? Answer with just the number."),
	}
	chatResp, err := client.Chat(ctx, messages)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}
	fmt.Printf("Chat response: %s\n", chatResp.Content)
	fmt.Printf("Meta: %+v\n\n", chatResp.Meta)

	// Test List Models
	fmt.Println("Testing ListModels...")
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Fatalf("ListModels failed: %v", err)
	}
	fmt.Printf("Found %d models\n", len(models.Models))
	for _, model := range models.Models[:3] {
		fmt.Printf("  - %s (%s)\n", model.Name, model.Details.ParameterSize)
	}
}
