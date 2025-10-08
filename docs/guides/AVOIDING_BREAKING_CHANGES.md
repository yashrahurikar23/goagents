# 🛡️ Avoiding Breaking Changes: A Practical Guide

**Strategies and code examples for maintaining backward compatibility in GoAgents**

---

## Table of Contents

1. [Strategy 1: Functional Options Pattern](#strategy-1-functional-options-pattern)
2. [Strategy 2: Interface-Based Design](#strategy-2-interface-based-design)
3. [Strategy 3: Unexported Fields](#strategy-3-unexported-fields)
4. [Strategy 4: Deprecation Over Removal](#strategy-4-deprecation-over-removal)
5. [Strategy 5: Additive Changes Only](#strategy-5-additive-changes-only)
6. [Real-World Examples from GoAgents](#real-world-examples-from-goagents)

---

## Strategy 1: Functional Options Pattern

### The Problem

Traditional constructor functions break when you add new parameters:

```go
// v0.1.0
func NewAgent(llm LLM, maxIter int) *Agent {
    return &Agent{
        llm:     llm,
        maxIter: maxIter,
    }
}

// Users code:
agent := NewAgent(llm, 10)

// v0.2.0 - Want to add timeout
func NewAgent(llm LLM, maxIter int, timeout time.Duration) *Agent {
    // BREAKS ALL EXISTING CODE! 💥
}
```

### The Solution: Functional Options

```go
// agent/options.go

package agent

import "time"

// Option configures an agent
type Option func(*Agent)

// WithMaxIterations sets the maximum number of reasoning iterations
func WithMaxIterations(n int) Option {
    return func(a *Agent) {
        a.maxIterations = n
    }
}

// WithTimeout sets the execution timeout
func WithTimeout(d time.Duration) Option {
    return func(a *Agent) {
        a.timeout = d
    }
}

// WithVerbose enables verbose logging
func WithVerbose(v bool) Option {
    return func(a *Agent) {
        a.verbose = v
    }
}
```

```go
// agent/react.go

package agent

import (
    "context"
    "time"
)

// ReActAgent implements the ReAct pattern
type ReActAgent struct {
    llm           core.LLM
    tools         []core.Tool
    maxIterations int
    timeout       time.Duration
    verbose       bool
}

// NewReActAgent creates a new ReAct agent with optional configuration
func NewReActAgent(llm core.LLM, opts ...Option) *ReActAgent {
    // Set defaults
    a := &ReActAgent{
        llm:           llm,
        maxIterations: 10,
        timeout:       30 * time.Second,
        verbose:       false,
    }
    
    // Apply options
    for _, opt := range opts {
        opt(a)
    }
    
    return a
}
```

### Usage Examples

```go
// v0.1.0 - Simple usage (still works in v0.2.0+)
agent := agent.NewReActAgent(llm)

// v0.2.0 - With options
agent := agent.NewReActAgent(llm,
    agent.WithMaxIterations(20),
    agent.WithVerbose(true),
)

// v0.3.0 - Add more options without breaking
agent := agent.NewReActAgent(llm,
    agent.WithMaxIterations(20),
    agent.WithTimeout(60 * time.Second),  // NEW in v0.3.0
    agent.WithVerbose(true),
)

// v0.4.0 - Even more options
agent := agent.NewReActAgent(llm,
    agent.WithMaxIterations(20),
    agent.WithTimeout(60 * time.Second),
    agent.WithRetryPolicy(retryPolicy),   // NEW in v0.4.0
    agent.WithCallbacks(callbacks),       // NEW in v0.4.0
    agent.WithVerbose(true),
)
```

### Benefits

✅ **Backward compatible** - Old code keeps working  
✅ **Infinitely extensible** - Add options anytime  
✅ **Self-documenting** - Clear what each option does  
✅ **Optional configuration** - Sensible defaults  
✅ **Type-safe** - Compiler catches errors

---

## Strategy 2: Interface-Based Design

### The Problem

Concrete types lock you in:

```go
// ❌ BAD - Concrete type dependency
func NewAgent(llm *OpenAIClient) *Agent {
    return &Agent{llm: llm}
}

// Can't easily support Ollama or other providers
// Changing OpenAIClient breaks everyone
```

### The Solution: Interfaces

```go
// core/llm.go

package core

import "context"

// Message represents a chat message
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// Response represents an LLM response
type Response struct {
    Content      string
    TokensUsed   int
    FinishReason string
}

// LLM is the interface all language models must implement
type LLM interface {
    // Generate generates a response from messages
    Generate(ctx context.Context, messages []Message) (*Response, error)
    
    // Model returns the model name
    Model() string
}
```

```go
// llm/openai/client.go

package openai

import (
    "context"
    "github.com/yashrahurikar23/goagents/core"
)

// Client implements the LLM interface for OpenAI
type Client struct {
    apiKey string
    model  string
    // ... other internal fields
}

// Generate implements core.LLM
func (c *Client) Generate(ctx context.Context, messages []core.Message) (*core.Response, error) {
    // OpenAI-specific implementation
    return &core.Response{
        Content:    "response",
        TokensUsed: 100,
    }, nil
}

// Model implements core.LLM
func (c *Client) Model() string {
    return c.model
}

// New creates a new OpenAI client
func New(opts ...Option) *Client {
    c := &Client{
        model: "gpt-4",
    }
    for _, opt := range opts {
        opt(c)
    }
    return c
}
```

```go
// llm/ollama/client.go

package ollama

import (
    "context"
    "github.com/yashrahurikar23/goagents/core"
)

// Client implements the LLM interface for Ollama
type Client struct {
    baseURL string
    model   string
    // ... other internal fields
}

// Generate implements core.LLM
func (c *Client) Generate(ctx context.Context, messages []core.Message) (*core.Response, error) {
    // Ollama-specific implementation
    return &core.Response{
        Content:    "response",
        TokensUsed: 100,
    }, nil
}

// Model implements core.LLM
func (c *Client) Model() string {
    return c.model
}

// New creates a new Ollama client
func New(opts ...Option) *Client {
    c := &Client{
        baseURL: "http://localhost:11434",
        model:   "llama2",
    }
    for _, opt := range opts {
        opt(c)
    }
    return c
}
```

```go
// agent/react.go

package agent

import "github.com/yashrahurikar23/goagents/core"

// ✅ GOOD - Interface dependency
func NewReActAgent(llm core.LLM, opts ...Option) *ReActAgent {
    return &ReActAgent{llm: llm}
}
```

### Usage Examples

```go
// Works with OpenAI
openaiLLM := openai.New(openai.WithAPIKey(key))
agent1 := agent.NewReActAgent(openaiLLM)

// Works with Ollama
ollamaLLM := ollama.New(ollama.WithModel("llama2"))
agent2 := agent.NewReActAgent(ollamaLLM)

// Works with any future provider!
claudeLLM := anthropic.New(anthropic.WithAPIKey(key))
agent3 := agent.NewReActAgent(claudeLLM)
```

### Benefits

✅ **Provider independence** - Swap LLMs easily  
✅ **Easy testing** - Mock implementations  
✅ **Extensible** - Add providers without breaking  
✅ **Loose coupling** - Agents don't know about specifics

---

## Strategy 3: Unexported Fields

### The Problem

Exported fields lock your internal structure:

```go
// ❌ BAD - Exported fields
type Agent struct {
    LLM         core.LLM    // Can't change this!
    MaxIter     int         // Locked in!
    Tools       []core.Tool // Fixed!
}

// Users can directly access:
agent.MaxIter = 20
agent.Tools = append(agent.Tools, newTool)

// Now we can't refactor internal structure
```

### The Solution: Unexported Fields + Methods

```go
// agent/react.go

package agent

import "github.com/yashrahurikar23/goagents/core"

// ✅ GOOD - Unexported fields
type ReActAgent struct {
    llm           core.LLM      // lowercase = unexported
    tools         []core.Tool   // can change freely
    maxIterations int           // internal only
    verbose       bool          // private
    memory        *Memory       // can refactor
}

// Controlled access through methods
func (a *ReActAgent) AddTool(tool core.Tool) {
    a.tools = append(a.tools, tool)
}

func (a *ReActAgent) RemoveTool(name string) {
    // ... implementation with validation
}

func (a *ReActAgent) Tools() []core.Tool {
    // Return copy to prevent external modification
    tools := make([]core.Tool, len(a.tools))
    copy(tools, a.tools)
    return tools
}

func (a *ReActAgent) SetMaxIterations(n int) error {
    if n < 1 {
        return fmt.Errorf("max iterations must be positive")
    }
    a.maxIterations = n
    return nil
}

func (a *ReActAgent) MaxIterations() int {
    return a.maxIterations
}
```

### Future Refactoring (Non-Breaking!)

```go
// v0.2.0
type ReActAgent struct {
    llm           core.LLM
    tools         []core.Tool
    maxIterations int
}

// v0.3.0 - Change internal structure (no breaking change!)
type ReActAgent struct {
    llm           core.LLM
    toolRegistry  *ToolRegistry  // Changed from []core.Tool
    config        *Config        // Grouped settings
}

// Public API stays the same
func (a *ReActAgent) AddTool(tool core.Tool) {
    a.toolRegistry.Register(tool)  // Internal change
}

// Users' code still works!
```

### Benefits

✅ **Internal flexibility** - Refactor freely  
✅ **Validation** - Control through methods  
✅ **Encapsulation** - Hide complexity  
✅ **Non-breaking** - Change internals anytime

---

## Strategy 4: Deprecation Over Removal

### The Problem

Removing functions breaks existing code immediately:

```go
// v0.1.0
func (a *Agent) Run(input string) (*Response, error) {
    // ...
}

// v0.2.0 - Removed Run, added RunContext
// 💥 BREAKS EVERYONE!
func (a *Agent) RunContext(ctx context.Context, input string) (*Response, error) {
    // ...
}
```

### The Solution: Deprecate Gradually

```go
// agent/react.go

package agent

import (
    "context"
    "fmt"
)

// v0.2.0 - Add new method first
// RunContext executes the agent with context support for cancellation and timeouts.
//
// This is the recommended method for running agents. It provides:
//   - Cancellation support via context
//   - Timeout handling
//   - Better error propagation
//
// Example:
//     ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//     defer cancel()
//     response, err := agent.RunContext(ctx, "What is the weather?")
func (a *ReActAgent) RunContext(ctx context.Context, input string) (*Response, error) {
    // New implementation with context support
    return a.execute(ctx, input)
}

// Run executes the agent without context support.
//
// Deprecated: Use RunContext instead for better cancellation and timeout support.
// This method will be removed in v1.0.0.
//
// Migration example:
//     // Old:
//     response, err := agent.Run("question")
//
//     // New:
//     response, err := agent.RunContext(context.Background(), "question")
func (a *ReActAgent) Run(input string) (*Response, error) {
    // Call new method with background context
    return a.RunContext(context.Background(), input)
}
```

### Migration Timeline

```go
// v0.1.0 - Original
func (a *Agent) Run(input string) (*Response, error)

// v0.2.0 - Add new, keep old
func (a *Agent) Run(input string) (*Response, error)           // Works, no warning
func (a *Agent) RunContext(ctx context.Context, ...) (...)     // NEW!

// v0.3.0 - Deprecate old
func (a *Agent) Run(input string) (*Response, error)           // Deprecated warning
func (a *Agent) RunContext(ctx context.Context, ...) (...)     // Recommended

// v0.4.0 - Still keep old
func (a *Agent) Run(input string) (*Response, error)           // Still works
func (a *Agent) RunContext(ctx context.Context, ...) (...)     

// v1.0.0 - Can remove old (major version)
func (a *Agent) RunContext(ctx context.Context, ...) (...)     // Only this remains
```

### Deprecation Annotations

```go
// Use Go's standard deprecation notice format:

// Deprecated: Use NewMethod instead. This will be removed in v1.0.0.
func OldMethod() {}

// The word "Deprecated:" must be at the start of a comment line
// IDEs and linters will show warnings
```

### Benefits

✅ **Graceful migration** - Users have time to update  
✅ **Clear warnings** - IDE shows deprecation  
✅ **No sudden breaks** - Code keeps working  
✅ **Documentation** - Shows migration path

---

## Strategy 5: Additive Changes Only

### Safe: Adding Fields

```go
// v0.1.0
type Response struct {
    Content string
    Tokens  int
}

// v0.2.0 - Add fields (SAFE! ✅)
type Response struct {
    Content    string
    Tokens     int
    Cost       float64           // NEW - but doesn't break
    Model      string            // NEW
    Metadata   map[string]string // NEW
}

// Old code still works:
response, _ := agent.Run(ctx, "question")
fmt.Println(response.Content)  // ✅ Still works
```

### Unsafe: Changing Fields

```go
// v0.1.0
type Response struct {
    Content string
    Tokens  int  // ← int type
}

// v0.2.0 - Change type (BREAKS! ❌)
type Response struct {
    Content string
    Tokens  float64  // ← Changed to float64 💥
}

// Old code breaks:
var tokens int = response.Tokens  // ❌ Type mismatch!
```

### Safe: Adding Methods

```go
// v0.1.0
type Agent interface {
    Run(ctx context.Context, input string) (*Response, error)
}

// v0.2.0 - Can we add methods? IT DEPENDS!

// ❌ Adding to interface BREAKS implementations:
type Agent interface {
    Run(ctx context.Context, input string) (*Response, error)
    Stream(ctx context.Context, input string) (<-chan Token, error) // 💥 BREAKS!
}
// Now all implementations must add Stream() method

// ✅ Better: Create new interface
type StreamingAgent interface {
    Agent  // Embed old interface
    Stream(ctx context.Context, input string) (<-chan Token, error)
}

// Or: Add as separate interface
type Streamer interface {
    Stream(ctx context.Context, input string) (<-chan Token, error)
}

// Check if agent supports streaming
if streamer, ok := agent.(Streamer); ok {
    stream, _ := streamer.Stream(ctx, input)
}
```

### Safe: Adding Functions

```go
// v0.1.0
package tools

func NewCalculator() *Calculator { ... }
func NewHTTPTool() *HTTPTool { ... }

// v0.2.0 - Add functions (SAFE! ✅)
func NewCalculator() *Calculator { ... }
func NewHTTPTool() *HTTPTool { ... }
func NewFileTool() *FileTool { ... }      // NEW
func NewWebSearchTool() *WebSearchTool { ... } // NEW
```

### Benefits

✅ **Always safe** - Adding never breaks  
✅ **Easy to reason about** - Clear rules  
✅ **Predictable** - Users know what to expect

---

## Real-World Examples from GoAgents

### Example 1: Agent Options (Currently Implemented)

```go
// agent/options.go (v0.2.0)
package agent

// Already using functional options! ✅
type Option func(*baseAgent)

func WithMaxIterations(n int) Option {
    return func(a *baseAgent) {
        a.maxIterations = n
    }
}

func WithVerbose(v bool) Option {
    return func(a *baseAgent) {
        a.verbose = v
    }
}

// Can add more options later without breaking:
func WithTimeout(d time.Duration) Option {  // v0.3.0
    return func(a *baseAgent) {
        a.timeout = d
    }
}

func WithCallbacks(cb Callbacks) Option {  // v0.3.0
    return func(a *baseAgent) {
        a.callbacks = cb
    }
}
```

### Example 2: LLM Interface (Currently Implemented)

```go
// core/llm.go (v0.2.0)
package core

// Already using interface! ✅
type LLM interface {
    Generate(ctx context.Context, messages []Message) (*Response, error)
}

// Future: Add streaming without breaking
type StreamingLLM interface {
    LLM  // Embed existing interface
    GenerateStream(ctx context.Context, messages []Message) (<-chan Token, error)
}

// Usage:
if streamingLLM, ok := llm.(StreamingLLM); ok {
    // Use streaming
    stream, _ := streamingLLM.GenerateStream(ctx, messages)
} else {
    // Fallback to regular generation
    response, _ := llm.Generate(ctx, messages)
}
```

### Example 3: Memory Interface (Currently Implemented)

```go
// core/memory.go (v0.2.0)
package core

// Already using interface! ✅
type Memory interface {
    AddMessage(message Message)
    GetMessages() []Message
    Clear()
}

// Future: Add without breaking
type PersistentMemory interface {
    Memory  // Embed existing
    Save(ctx context.Context, path string) error
    Load(ctx context.Context, path string) error
}
```

### Example 4: Tool Interface (Future Enhancement)

```go
// core/tool.go (v0.3.0 proposal)
package core

// Current (simple)
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// Future: Add optional interfaces
type AsyncTool interface {
    Tool
    ExecuteAsync(ctx context.Context, args map[string]interface{}) (<-chan interface{}, error)
}

type CacheableTool interface {
    Tool
    CacheKey(args map[string]interface{}) string
}

// Tools can optionally implement these
```

---

## Testing Backward Compatibility

### Compatibility Test Suite

```go
// tests/compat_test.go

package tests

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/yashrahurikar23/goagents/agent"
    "github.com/yashrahurikar23/goagents/llm/ollama"
)

// TestBackwardCompatibility_v01_v02 ensures v0.1.0 code works in v0.2.0
func TestBackwardCompatibility_v01_v02(t *testing.T) {
    llm := ollama.New()
    
    // v0.1.0 style - simple creation
    agent := agent.NewReActAgent(llm)
    
    // Should still work
    response, err := agent.Run(context.Background(), "test")
    require.NoError(t, err)
    assert.NotNil(t, response)
}

// TestBackwardCompatibility_v02_WithOptions ensures options are optional
func TestBackwardCompatibility_v02_WithOptions(t *testing.T) {
    llm := ollama.New()
    
    // v0.2.0 style - with options
    agent := agent.NewReActAgent(llm,
        agent.WithMaxIterations(10),
        agent.WithVerbose(true),
    )
    
    response, err := agent.Run(context.Background(), "test")
    require.NoError(t, err)
    assert.NotNil(t, response)
}

// TestDeprecatedMethods_StillWork ensures deprecated methods continue to work
func TestDeprecatedMethods_StillWork(t *testing.T) {
    llm := ollama.New()
    agent := agent.NewReActAgent(llm)
    
    // Deprecated Run() should still work
    response, err := agent.Run("test")  // No context
    require.NoError(t, err)
    assert.NotNil(t, response)
    
    // New RunContext() should also work
    response2, err := agent.RunContext(context.Background(), "test")
    require.NoError(t, err)
    assert.NotNil(t, response2)
}
```

---

## Checklist: Before Releasing Changes

### ✅ Pre-Release Checklist

- [ ] **No function removals** - Only deprecations
- [ ] **No signature changes** - Keep existing parameters
- [ ] **No type changes** - Don't change exported types
- [ ] **Only additive changes** - Add, don't modify
- [ ] **Deprecation notices** - Document what's deprecated
- [ ] **Migration guide** - Show how to update
- [ ] **Tests** - All old tests still pass
- [ ] **Examples** - Update examples, keep old ones
- [ ] **CHANGELOG** - Document all changes
- [ ] **Version bump** - Correct semantic version

### Version Bump Guide

```
v0.x.y → v0.x.(y+1)  : Bug fixes only
v0.x.y → v0.(x+1).0  : New features (backward compatible)
v0.x.y → v1.0.0      : Ready for production, API stable
v1.x.y → v2.0.0      : Breaking changes (avoid!)
```

---

## Summary

### ✅ Do This

1. **Use functional options** for configuration
2. **Use interfaces** for dependencies
3. **Keep fields unexported** for flexibility
4. **Deprecate** before removing
5. **Add**, don't modify
6. **Document** everything
7. **Test** old code paths

### ❌ Avoid This

1. ~~Adding required parameters~~
2. ~~Changing function signatures~~
3. ~~Exporting internal fields~~
4. ~~Removing functions immediately~~
5. ~~Changing types~~
6. ~~Breaking without warning~~

### 🎯 Result

- **Happy users** - Code keeps working
- **Easy upgrades** - Clear migration paths
- **Professional** - Stable, reliable API
- **Flexible** - Can evolve without breaking

---

**GoAgents is already following most of these patterns! Keep it up! 🎉**
