package openai_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yashrahurikar23/goagents/llm/openai"
)

// Example_basicCompletion demonstrates a simple text completion.
func Example_basicCompletion() {
	client := openai.New(
		openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
		openai.WithModel("gpt-4"),
	)

	response, err := client.Complete(context.Background(), "What is the capital of France?")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)
	// Output: Paris is the capital of France.
}

// Example_chatConversation demonstrates a multi-turn conversation.
func Example_chatConversation() {
	client := openai.New(
		openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
		openai.WithModel("gpt-4"),
	)

	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatMessage{
			openai.SystemMessage("You are a helpful assistant."),
			openai.UserMessage("What's the weather like today?"),
			openai.AssistantMessage("I don't have access to real-time weather data. Could you tell me your location?"),
			openai.UserMessage("I'm in San Francisco."),
		},
	}

	response, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.Choices[0].Message.Content)
}

// Example_streaming demonstrates streaming responses.
func Example_streaming() {
	client := openai.New(
		openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
		openai.WithModel("gpt-4"),
	)

	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatMessage{
			openai.UserMessage("Write a haiku about coding."),
		},
	}

	opts := openai.StreamOptions{
		OnChunk: func(chunk *openai.ChatCompletionStreamResponse) error {
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
				if content, ok := chunk.Choices[0].Delta.Content.(string); ok {
					fmt.Print(content)
				}
			}
			return nil
		},
		OnComplete: func() error {
			fmt.Println("\n[Stream complete]")
			return nil
		},
		OnError: func(err error) {
			log.Printf("Stream error: %v", err)
		},
	}

	if err := client.CreateChatCompletionStream(context.Background(), req, opts); err != nil {
		log.Fatal(err)
	}
}

// Example_functionCalling demonstrates function calling.
func Example_functionCalling() {
	client := openai.New(
		openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
		openai.WithModel("gpt-4"),
	)

	// Define a weather function
	weatherFunc := openai.NewFunction(
		"get_weather",
		"Get the current weather for a location",
		openai.JSONSchema(
			map[string]interface{}{
				"location": openai.PropertyString("The city and state, e.g. San Francisco, CA"),
				"unit":     openai.PropertyEnum("Temperature unit", []string{"celsius", "fahrenheit"}),
			},
			[]string{"location"},
		),
	)

	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatMessage{
			openai.UserMessage("What's the weather in Boston?"),
		},
		Tools: []openai.Tool{
			openai.NewTool(weatherFunc),
		},
	}

	response, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	// Check if model wants to call a function
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		toolCall := response.Choices[0].Message.ToolCalls[0]
		fmt.Printf("Function: %s\n", toolCall.Function.Name)
		fmt.Printf("Arguments: %s\n", toolCall.Function.Arguments)
	}
}

// Example_vision demonstrates image understanding.
func Example_vision() {
	client := openai.New(
		openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
		openai.WithModel("gpt-4-vision-preview"),
	)

	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatMessage{
			openai.UserMessageWithImage(
				"What's in this image?",
				"https://example.com/image.jpg",
				"high", // detail level
			),
		},
		MaxTokens: intPtr(300),
	}

	response, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.Choices[0].Message.Content)
}

// Example_jsonMode demonstrates JSON output mode.
func Example_jsonMode() {
	client := openai.New(
		openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
		openai.WithModel("gpt-4"),
	)

	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatMessage{
			openai.SystemMessage("You are a helpful assistant that outputs JSON."),
			openai.UserMessage("List 3 colors in JSON format with 'name' and 'hex' fields."),
		},
		ResponseFormat: &openai.ResponseFormat{
			Type: "json_object",
		},
	}

	response, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.Choices[0].Message.Content)
}

// Example_embeddings demonstrates creating embeddings.
func Example_embeddings() {
	client := openai.New(
		openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	req := openai.EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: []string{
			"The quick brown fox jumps over the lazy dog",
			"Machine learning is fascinating",
		},
	}

	response, err := client.CreateEmbedding(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	for i, emb := range response.Data {
		fmt.Printf("Embedding %d: %d dimensions\n", i, len(emb.Embedding))
	}
}

// Example_moderation demonstrates content moderation.
func Example_moderation() {
	client := openai.New(
		openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	req := openai.ModerationRequest{
		Input: "I want to hurt someone",
	}

	response, err := client.CreateModeration(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range response.Results {
		if result.Flagged {
			fmt.Println("Content flagged!")
			for category, flagged := range result.Categories {
				if flagged {
					fmt.Printf("- %s: %.2f\n", category, result.CategoryScores[category])
				}
			}
		}
	}
}

// Example_withOptions demonstrates various configuration options.
func Example_withOptions() {
	client := openai.New(
		openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
		openai.WithModel("gpt-4"),
		openai.WithBaseURL("https://api.openai.com/v1"),
		openai.WithTimeout(30*1000000000), // 30 seconds
		openai.WithMaxRetries(3),
	)

	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatMessage{
			openai.UserMessage("Hello!"),
		},
		Temperature:      floatPtr(0.7),
		MaxTokens:        intPtr(100),
		TopP:             floatPtr(1.0),
		PresencePenalty:  floatPtr(0.0),
		FrequencyPenalty: floatPtr(0.0),
	}

	response, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.Choices[0].Message.Content)
}

// Example_errorHandling demonstrates error handling.
func Example_errorHandling() {
	client := openai.New(
		openai.WithAPIKey("invalid-key"),
	)

	_, err := client.Complete(context.Background(), "Hello")
	if err != nil {
		if openai.IsRateLimitError(err) {
			fmt.Println("Rate limit exceeded, please retry later")
		} else if openai.IsTimeoutError(err) {
			fmt.Println("Request timed out")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
