# OpenAI Client

**Package:** `github.com/yashrahurikar/goagents/llm/openai`

A comprehensive OpenAI API client supporting all major features: chat completions, streaming, function calling, vision, embeddings, moderation, and more.

## Features

✅ **Chat Completions** - Multi-turn conversations with GPT models  
✅ **Streaming** - Real-time token streaming with SSE  
✅ **Function Calling** - Tool use and function execution  
✅ **Vision** - Image understanding with GPT-4 Vision  
✅ **Embeddings** - Text embeddings for semantic search  
✅ **Moderation** - Content policy violation detection  
✅ **JSON Mode** - Structured JSON output  
✅ **Retry Logic** - Automatic exponential backoff  
✅ **Error Handling** - Comprehensive error types  
✅ **Context Support** - Cancellation and timeouts  

## Installation

```bash
go get github.com/yashrahurikar/goagents
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/yashrahurikar/goagents/llm/openai"
)

func main() {
    client := openai.New(
        openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
        openai.WithModel("gpt-4"),
    )
    
    response, err := client.Complete(context.Background(), "What is Go?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response)
}
```

## Configuration

### Client Options

```go
client := openai.New(
    openai.WithAPIKey("sk-..."),           // Required: Your API key
    openai.WithModel("gpt-4"),              // Default model
    openai.WithBaseURL("https://..."),      // Custom API endpoint
    openai.WithTimeout(30*time.Second),     // HTTP timeout
    openai.WithMaxRetries(3),               // Retry attempts
    openai.WithHTTPClient(customClient),    // Custom HTTP client
)
```

### Available Models

- **GPT-4**: `gpt-4`, `gpt-4-turbo-preview`
- **GPT-4 Vision**: `gpt-4-vision-preview`
- **GPT-3.5**: `gpt-3.5-turbo`, `gpt-3.5-turbo-16k`
- **GPT-5**: `gpt-5` (latest flagship model)
- **Embeddings**: `text-embedding-ada-002`, `text-embedding-3-small`, `text-embedding-3-large`
- **Moderation**: `omni-moderation-latest`

## Usage Guide

### 1. Simple Completion

```go
response, err := client.Complete(ctx, "Explain quantum computing in one sentence")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
```

### 2. Chat Conversation

```go
req := openai.ChatCompletionRequest{
    Messages: []openai.ChatMessage{
        openai.SystemMessage("You are a helpful coding assistant."),
        openai.UserMessage("How do I sort a slice in Go?"),
    },
    Temperature: floatPtr(0.7),
    MaxTokens:   intPtr(500),
}

resp, err := client.CreateChatCompletion(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Choices[0].Message.Content)
```

### 3. Streaming Responses

```go
req := openai.ChatCompletionRequest{
    Messages: []openai.ChatMessage{
        openai.UserMessage("Write a story about a robot"),
    },
}

opts := openai.StreamOptions{
    OnChunk: func(chunk *openai.ChatCompletionStreamResponse) error {
        if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
            if content, ok := chunk.Choices[0].Delta.Content.(string); ok {
                fmt.Print(content) // Print each token as it arrives
            }
        }
        return nil
    },
    OnComplete: func() error {
        fmt.Println("\n✓ Complete")
        return nil
    },
    OnError: func(err error) {
        log.Printf("Stream error: %v", err)
    },
}

err := client.CreateChatCompletionStream(ctx, req, opts)
if err != nil {
    log.Fatal(err)
}
```

### 4. Function Calling

```go
// Define a weather function
weatherFunc := openai.NewFunction(
    "get_weather",
    "Get current weather for a location",
    openai.JSONSchema(
        map[string]interface{}{
            "location": openai.PropertyString("City and state, e.g. Boston, MA"),
            "unit":     openai.PropertyEnum("Temperature unit", []string{"celsius", "fahrenheit"}),
        },
        []string{"location"}, // Required parameters
    ),
)

req := openai.ChatCompletionRequest{
    Messages: []openai.ChatMessage{
        openai.UserMessage("What's the weather in Tokyo?"),
    },
    Tools: []openai.Tool{
        openai.NewTool(weatherFunc),
    },
}

resp, err := client.CreateChatCompletion(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Handle function call
if len(resp.Choices[0].Message.ToolCalls) > 0 {
    toolCall := resp.Choices[0].Message.ToolCalls[0]
    
    // Parse arguments
    var args map[string]interface{}
    json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
    
    // Execute function (your implementation)
    result := getWeather(args["location"].(string), args["unit"].(string))
    
    // Send result back to model
    req.Messages = append(req.Messages,
        resp.Choices[0].Message,
        openai.ToolMessage(toolCall.ID, result),
    )
    
    // Get final response
    finalResp, _ := client.CreateChatCompletion(ctx, req)
    fmt.Println(finalResp.Choices[0].Message.Content)
}
```

### 5. Vision (Image Understanding)

```go
// With image URL
req := openai.ChatCompletionRequest{
    Model: "gpt-4-vision-preview",
    Messages: []openai.ChatMessage{
        openai.UserMessageWithImage(
            "What's in this image? Describe in detail.",
            "https://example.com/image.jpg",
            "high", // detail level: low, high, or auto
        ),
    },
    MaxTokens: intPtr(500),
}

// With base64 encoded image
base64Image := "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
req := openai.ChatCompletionRequest{
    Model: "gpt-4-vision-preview",
    Messages: []openai.ChatMessage{
        {
            Role: "user",
            Content: []openai.ContentPart{
                {Type: "text", Text: "Describe this image"},
                {
                    Type: "image_url",
                    ImageURL: &openai.ImageURL{URL: base64Image, Detail: "high"},
                },
            },
        },
    },
}

resp, err := client.CreateChatCompletion(ctx, req)
```

### 6. JSON Mode

```go
req := openai.ChatCompletionRequest{
    Messages: []openai.ChatMessage{
        openai.SystemMessage("You output valid JSON."),
        openai.UserMessage("List 3 programming languages with their year of creation."),
    },
    ResponseFormat: &openai.ResponseFormat{
        Type: "json_object",
    },
}

resp, err := client.CreateChatCompletion(ctx, req)
// Response will be valid JSON
```

### 7. Embeddings

```go
req := openai.EmbeddingRequest{
    Model: "text-embedding-ada-002",
    Input: []string{
        "The quick brown fox",
        "Machine learning is powerful",
        "Go is a great language",
    },
}

resp, err := client.CreateEmbedding(ctx, req)
if err != nil {
    log.Fatal(err)
}

for i, emb := range resp.Data {
    fmt.Printf("Embedding %d: %d dimensions\n", i, len(emb.Embedding))
    // Use emb.Embedding for similarity search, clustering, etc.
}
```

### 8. Content Moderation

```go
// Text moderation
req := openai.ModerationRequest{
    Model: "omni-moderation-latest",
    Input: "Some text to check...",
}

resp, err := client.CreateModeration(ctx, req)
if err != nil {
    log.Fatal(err)
}

for _, result := range resp.Results {
    if result.Flagged {
        fmt.Println("⚠️  Content flagged!")
        for category, flagged := range result.Categories {
            if flagged {
                score := result.CategoryScores[category]
                fmt.Printf("  - %s: %.2f%%\n", category, score*100)
            }
        }
    }
}

// Multimodal moderation (text + image)
req := openai.ModerationRequest{
    Model: "omni-moderation-latest",
    Input: []openai.ModerationInput{
        {Type: "text", Text: "Some text..."},
        {
            Type: "image_url",
            ImageURL: &openai.ImageURL{URL: "https://example.com/image.jpg"},
        },
    },
}
```

### 9. List Available Models

```go
resp, err := client.ListModels(ctx)
if err != nil {
    log.Fatal(err)
}

for _, model := range resp.Data {
    fmt.Printf("%s (owned by: %s)\n", model.ID, model.OwnedBy)
}
```

## Advanced Features

### Custom Parameters

```go
req := openai.ChatCompletionRequest{
    Model:            "gpt-4",
    Messages:         messages,
    Temperature:      floatPtr(0.8),      // Creativity (0-2)
    MaxTokens:        intPtr(1000),        // Max response length
    TopP:             floatPtr(0.95),      // Nucleus sampling
    N:                intPtr(3),            // Number of completions
    Stop:             []string{"\n\n"},    // Stop sequences
    PresencePenalty:  floatPtr(0.6),       // Penalize repetition
    FrequencyPenalty: floatPtr(0.5),       // Penalize frequency
    LogitBias:        map[string]float64{  // Token bias
        "1234": -100, // Prevent token
        "5678": 100,  // Encourage token
    },
    User:             "user-123",          // User ID for tracking
    Seed:             intPtr(42),           // Deterministic output
    ReasoningEffort:  "high",              // GPT-5: reasoning depth
    Verbosity:        "medium",            // GPT-5: output verbosity
}
```

### Error Handling

```go
resp, err := client.Complete(ctx, "Hello")
if err != nil {
    // Check specific error types
    if openai.IsRateLimitError(err) {
        fmt.Println("⏱️  Rate limited, retry after delay")
        time.Sleep(time.Minute)
        // Retry...
    } else if openai.IsTimeoutError(err) {
        fmt.Println("⏰ Request timed out")
    } else if oaiErr, ok := err.(*openai.OpenAIError); ok {
        fmt.Printf("OpenAI Error: %s (HTTP %d)\n", oaiErr.Message, oaiErr.StatusCode)
        fmt.Printf("Type: %s, Code: %v\n", oaiErr.Type, oaiErr.Code)
    } else {
        fmt.Printf("Other error: %v\n", err)
    }
}
```

### Context and Cancellation

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.Complete(ctx, "Long prompt...")

// Manual cancellation
ctx, cancel := context.WithCancel(context.Background())

go func() {
    time.Sleep(5 * time.Second)
    cancel() // Cancel after 5 seconds
}()

resp, err := client.Complete(ctx, "...")
```

## Core Integration

This client implements the `core.LLM` interface:

```go
import "github.com/yashrahurikar/goagents/core"

func useWithCore(llm core.LLM) {
    messages := []core.Message{
        core.UserMessage("Hello!"),
    }
    
    resp, err := llm.Chat(context.Background(), messages)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(resp.Content)
    
    // Check for tool calls
    for _, tc := range resp.ToolCalls {
        fmt.Printf("Tool: %s, Args: %v\n", tc.Name, tc.Args)
    }
}

// Use OpenAI client as core.LLM
client := openai.New(openai.WithAPIKey("sk-..."))
useWithCore(client)
```

## Best Practices

### 1. **Always use context**
```go
ctx := context.Background()
resp, err := client.Complete(ctx, prompt)
```

### 2. **Handle errors properly**
```go
if err != nil {
    if openai.IsRateLimitError(err) {
        // Implement backoff
    }
    return err
}
```

### 3. **Use streaming for long responses**
```go
// Better UX with streaming
err := client.CreateChatCompletionStream(ctx, req, opts)
```

### 4. **Set reasonable timeouts**
```go
ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
defer cancel()
```

### 5. **Reuse client instances**
```go
// Create once, reuse many times
var client = openai.New(openai.WithAPIKey(apiKey))
```

### 6. **Use pointer helpers for optional fields**
```go
func intPtr(i int) *int { return &i }
func floatPtr(f float64) *float64 { return &f }
```

## Token Usage and Pricing

Track token usage from responses:

```go
resp, err := client.CreateChatCompletion(ctx, req)
if resp.Usage != nil {
    fmt.Printf("Tokens used:\n")
    fmt.Printf("  Prompt: %d\n", resp.Usage.PromptTokens)
    fmt.Printf("  Completion: %d\n", resp.Usage.CompletionTokens)
    fmt.Printf("  Total: %d\n", resp.Usage.TotalTokens)
}
```

## Limitations

- **Vision models**: Require `MaxTokens` to be explicitly set
- **Streaming**: `Usage` information not available in stream
- **Function calling**: Arguments are returned as JSON strings
- **Rate limits**: Implement exponential backoff for production
- **Token limits**: Different models have different context windows

## See Also

- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)
- [Core Package](../core/README.md) - Core interfaces and types
- [Examples](./examples_test.go) - Complete working examples
- [Best Practices](../../BEST_PRACTICES.md) - SDK design patterns
