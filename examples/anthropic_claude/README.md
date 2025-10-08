# Anthropic Claude Example

This example demonstrates using Anthropic's Claude models with the GoAgents framework.

## Features Demonstrated

- ✅ Simple text completion
- ✅ ReAct agent with tools (calculator)
- ✅ Conversational agent with memory
- ✅ Using different Claude models (Opus, Sonnet, Haiku)

## Prerequisites

You need an Anthropic API key. Get one at: https://console.anthropic.com/

## Setup

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

## Running the Example

```bash
go run main.go
```

## Expected Output

```
🤖 Anthropic Claude Example
============================

Example 1: Simple Completion
-----------------------------
Claude: AI is a field of computer science that enables machines to perform tasks that typically require human intelligence.

Example 2: ReAct Agent with Calculator
---------------------------------------
Agent: Let me calculate 15% of 230 for you...
[calculation steps]
The answer is 34.5

Example 3: Conversational Agent
--------------------------------
You: My favorite color is blue. Remember this.
Claude: I'll remember that your favorite color is blue!

You: What is my favorite color?
Claude: Your favorite color is blue, as you mentioned earlier.

Example 4: Different Claude Models
----------------------------------
claude-3-5-sonnet-20241022: Hello there, friend!
claude-3-opus-20240229: Hi, how are you?
claude-3-haiku-20240307: Hello, nice to meet you!

✅ All examples completed successfully!
```

## Available Claude Models

- **Claude 3.5 Sonnet** (`ModelClaude35Sonnet`) - Best balance of intelligence and speed
- **Claude 3 Opus** (`ModelClaude3Opus`) - Most capable model
- **Claude 3 Sonnet** (`ModelClaude3Sonnet`) - Good balance
- **Claude 3 Haiku** (`ModelClaude3Haiku`) - Fastest and most compact
- **Claude 3.5 Haiku** (`ModelClaude35Haiku`) - Improved Haiku

## Configuration Options

```go
llm := anthropic.New(
    anthropic.WithAPIKey("your-key"),
    anthropic.WithModel(anthropic.ModelClaude35Sonnet),
    anthropic.WithTemperature(0.7),        // 0.0 - 1.0
    anthropic.WithMaxTokens(2048),         // Max tokens to generate
    anthropic.WithTopP(0.9),               // Nucleus sampling
    anthropic.WithTopK(50),                // Top-k sampling
)
```

## Learn More

- [Anthropic Documentation](https://docs.anthropic.com/)
- [Claude Models Overview](https://docs.anthropic.com/claude/docs/models-overview)
- [GoAgents Documentation](../../docs/README.md)
