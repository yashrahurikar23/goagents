package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/yashrahurikar23/goagents/agent"
	"github.com/yashrahurikar23/goagents/llm/ollama"
	"github.com/yashrahurikar23/goagents/tools"
)

func main() {
	fmt.Println("🌐 HTTP Tool Example - Agent making API calls")
	fmt.Println(strings.Repeat("=", 50))

	// Create Ollama client (make sure Ollama is running locally)
	// Download model: ollama pull llama3.2:1b
	llm := ollama.New(
		ollama.WithModel("llama3.2:1b"),
		ollama.WithTemperature(0.1),
	)

	// Create HTTP tool for making web requests
	httpTool := tools.NewHTTPTool()

	// Create ReAct agent with HTTP tool
	reactAgent := agent.NewReActAgent(llm)
	if err := reactAgent.AddTool(httpTool); err != nil {
		log.Fatal(err)
	}

	// Example 1: Fetch a random fact from an API
	fmt.Println("\n📡 Example 1: Fetching a random fact")
	fmt.Println(strings.Repeat("-", 50))

	response1, err := reactAgent.Run(
		context.Background(),
		"Use the HTTP tool to fetch a random fact from https://uselessfacts.jsph.pl/api/v2/facts/random. Tell me the fact you received.",
	)
	if err != nil {
		log.Printf("Error in Example 1: %v\n", err)
	} else {
		fmt.Printf("\n🤖 Agent: %s\n", response1.Content)
	}

	// Example 2: Check if a website is reachable
	fmt.Println("\n\n🔍 Example 2: Checking website status")
	fmt.Println(strings.Repeat("-", 50))

	response2, err := reactAgent.Run(
		context.Background(),
		"Use the HTTP tool to check if https://httpbin.org/status/200 is working. Report the status code you receive.",
	)
	if err != nil {
		log.Printf("Error in Example 2: %v\n", err)
	} else {
		fmt.Printf("\n🤖 Agent: %s\n", response2.Content)
	}

	// Example 3: Get JSON data from an API
	fmt.Println("\n\n📊 Example 3: Fetching JSON data")
	fmt.Println(strings.Repeat("-", 50))

	response3, err := reactAgent.Run(
		context.Background(),
		"Use the HTTP tool to GET data from https://jsonplaceholder.typicode.com/posts/1. Tell me the title and body of the post.",
	)
	if err != nil {
		log.Printf("Error in Example 3: %v\n", err)
	} else {
		fmt.Printf("\n🤖 Agent: %s\n", response3.Content)
	}

	// Example 4: POST data to an API
	fmt.Println("\n\n📤 Example 4: Posting data to an API")
	fmt.Println(strings.Repeat("-", 50))

	response4, err := reactAgent.Run(
		context.Background(),
		`Use the HTTP tool to POST this data to https://jsonplaceholder.typicode.com/posts:
		{
			"title": "GoAgents Test",
			"body": "Testing HTTP tool",
			"userId": 1
		}
		Tell me the response you get.`,
	)
	if err != nil {
		log.Printf("Error in Example 4: %v\n", err)
	} else {
		fmt.Printf("\n🤖 Agent: %s\n", response4.Content)
	}

	fmt.Println("\n\n✅ All examples completed!")
	fmt.Println("\n💡 Try your own examples:")
	fmt.Println("   - Fetch weather data from an API")
	fmt.Println("   - Check GitHub repository stars")
	fmt.Println("   - Query a REST API with parameters")
	fmt.Println("   - Send webhooks to external services")
}
