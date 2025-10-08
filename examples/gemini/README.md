# Google Gemini Example

This example demonstrates using Google's Gemini models with the GoAgents framework.

## Features Demonstrated

- ✅ Simple text completion
- ✅ ReAct agent with tools (calculator)
- ✅ Conversational agent with memory
- ✅ Using different Gemini models (Pro, Flash, Flash 8B)
- ✅ System instructions

## Prerequisites

You need a Google Gemini API key. Get one at: <https://aistudio.google.com/apikey>

**Note:** Gemini offers a generous free tier! Perfect for development and experimentation.

## Setup

```bash
export GEMINI_API_KEY="your-api-key-here"
```

## Running the Example

```bash
go run main.go
```

## Expected Output

```text
✨ Google Gemini Example
========================

Example 1: Simple Completion
-----------------------------
Gemini: Quantum computing harnesses quantum mechanics to perform calculations exponentially faster than classical computers for specific problems.

Example 2: ReAct Agent with Calculator
---------------------------------------
Agent: Let me calculate the compound interest...
[calculation steps]
The compound interest is approximately $157.63

Example 3: Conversational Agent
--------------------------------
You: I'm planning a trip to Japan. Remember this.
Gemini: That sounds exciting! I'll remember that you're planning a trip to Japan.

You: What am I planning?
Gemini: You're planning a trip to Japan!

Example 4: Different Gemini Models
----------------------------------
Gemini 2.0 Flash: Hello, how are you?
Gemini 1.5 Flash: Hello there today!
Gemini 1.5 Flash 8B: Hello, nice day!
Gemini 1.5 Pro: Hello, welcome here!

Example 5: System Instructions
-------------------------------
Gemini (as math tutor): A prime number is a positive integer greater than 1 that has no positive integer divisors other than 1 and itself...

✅ All examples completed successfully!

Note: Gemini offers generous free tier limits!
Get your API key at: https://aistudio.google.com/apikey
```

## Available Gemini Models

- **Gemini 2.0 Flash** (`ModelGemini20Flash`) - Latest experimental model
- **Gemini 1.5 Flash** (`ModelGemini15Flash`) - Fast and efficient (recommended)
- **Gemini 1.5 Flash 8B** (`ModelGemini15Flash8B`) - Smaller, faster variant
- **Gemini 1.5 Pro** (`ModelGemini15Pro`) - Most capable
- **Gemini Pro** (`ModelGeminiPro`) - Standard model
- **Gemini Pro Vision** (`ModelGeminiProVision`) - Multimodal (images + text)

## Configuration Options

```go
llm := gemini.New(
    gemini.WithAPIKey("your-key"),
    gemini.WithModel(gemini.ModelGemini15Flash),
    gemini.WithTemperature(0.7),        // 0.0 - 1.0
    gemini.WithMaxTokens(2048),         // Max tokens to generate
    gemini.WithTopP(0.9),               // Nucleus sampling
    gemini.WithTopK(40),                // Top-k sampling
)
```

## Pricing

Gemini offers a **generous free tier**:

- **Gemini 1.5 Flash**: 15 requests/minute, 1 million tokens/minute (free)
- **Gemini 1.5 Pro**: 2 requests/minute, 32K tokens/minute (free)

Perfect for development and small-scale applications!

## Learn More

- [Google Gemini Documentation](https://ai.google.dev/gemini-api/docs)
- [Gemini Models Overview](https://ai.google.dev/gemini-api/docs/models)
- [GoAgents Documentation](../../docs/README.md)
