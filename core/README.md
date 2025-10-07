# GoAgent Core Package

**Package:** `github.com/yashrahurikar23/goagents/core`

The core package provides fundamental types and interfaces for the GoAgent framework.

## Overview

This package contains:

- **Interfaces**: `LLM`, `Tool`, `Agent` - core abstractions
- **Types**: `Message`, `Response`, `ToolCall`, `ToolSchema` - data structures
- **Errors**: Framework-wide error types

## Key Principles

1. **Zero external dependencies** - Only uses Go stdlib
2. **Interface-based** - Allows multiple implementations
3. **Context-aware** - All I/O takes `context.Context`
4. **Type-safe** - Compile-time validation where possible

## Usage

```go
import "github.com/yashrahurikar23/goagents/core"

// Use interfaces to define dependencies
func processWithLLM(llm core.LLM, prompt string) error {
    resp, err := llm.Complete(context.Background(), prompt)
    if err != nil {
        return err
    }
    fmt.Println(resp)
    return nil
}
```

## Interfaces

### LLM

Language model providers implement this interface:

```go
type LLM interface {
    Chat(ctx context.Context, messages []Message) (*Response, error)
    Complete(ctx context.Context, prompt string) (string, error)
}
```

**Implementations:**
- `llm/openai` - OpenAI GPT models
- `llm/anthropic` - Anthropic Claude models (coming soon)
- `llm/ollama` - Local models via Ollama (coming soon)

### Tool

Tools implement this interface:

```go
type Tool interface {
    Name() string
    Description() string
    Schema() *ToolSchema
    Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}
```

**Built-in Tools:**
- `tools.Calculator` - Basic arithmetic
- `tools.HTTPClient` - HTTP requests (coming soon)
- `tools.WebSearch` - Web search (coming soon)

### Agent

Agents implement this interface:

```go
type Agent interface {
    Run(ctx context.Context, input string) (*Response, error)
    AddTool(tool Tool) error
    Reset() error
}
```

**Implementations:**
- `agent.FunctionAgent` - Simple function-calling agent
- `agent.ReActAgent` - Reasoning + Acting agent (coming soon)
- `agent.WorkflowAgent` - Workflow-based agent (coming soon)

## Types

### Message

Represents a chat message:

```go
msg := core.UserMessage("Hello, how are you?")
// or
msg := core.NewMessage("user", "Hello")
```

### Response

Response from LLM or agent:

```go
type Response struct {
    Content   string
    ToolCalls []ToolCall
    Meta      map[string]interface{}
}
```

### ToolCall

Record of a tool invocation:

```go
type ToolCall struct {
    ID       string
    Name     string
    Args     map[string]interface{}
    Result   interface{}
    Error    error
    Duration time.Duration
}
```

## Design Patterns

### Functional Options

All constructors use functional options:

```go
// Not part of core, but shows pattern
llm := openai.New(
    openai.WithAPIKey("sk-..."),
    openai.WithModel("gpt-4"),
    openai.WithTimeout(30*time.Second),
)
```

### Context for Cancellation

All I/O operations accept `context.Context`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := llm.Chat(ctx, messages)
```

### Error Wrapping

Errors preserve context:

```go
if err != nil {
    return fmt.Errorf("chat failed: %w", err)
}
```

## See Also

- [Getting Started](../GETTING_STARTED.md) - Implementation guide
- [Best Practices](../BEST_PRACTICES.md) - Design patterns
- [Examples](../examples/) - Working examples
