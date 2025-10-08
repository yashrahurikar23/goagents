# Implementation Strategy & First Steps

**Date:** October 7, 2025  
**Status:** Ready to Start Development

---

## Why Bottom-Up Approach?

After analyzing LlamaIndex's comprehensive feature set, I recommend **starting from the bottom up**:

### Bottom-Up Benefits

1. **Solid Foundation** - No need to refactor core types later
2. **Clear Dependencies** - Each layer builds on the previous (no circular imports)
3. **Testable** - Can test each component in isolation
4. **Predictable** - Know exactly what you're building at each step
5. **Confidence** - Each layer "just works" before moving up

### Top-Down Risks

- Constantly refactoring as you learn requirements
- Circular dependency hell
- Abstract interfaces that don't match reality
- Hard to test (need mocks for everything)

---

## Build Order: The First 4 Weeks

### Week 1: Core Foundation

**Goal:** Types and interfaces that everything else builds on

```
Day 1-2: Core Types
├── Message (role, content)
├── Response (content, metadata)
├── ToolCall (name, args, result)
└── Error types

Day 3-4: Core Interfaces
├── LLM interface (Chat, Complete)
├── Tool interface (Name, Description, Execute)
└── Agent interface (Run, Reset)

Day 5: Package Structure
├── Set up go.mod
├── Create directory structure
└── CI/CD (GitHub Actions)
```

**Deliverable:** `goagent/core` package with all interfaces

---

### Week 2: LLM Client

**Goal:** Working OpenAI integration (80% of users use this)

```
Day 1-2: OpenAI Client Setup
├── API client structure
├── HTTP client with retries
└── Error handling

Day 3-5: Chat Completion
├── Messages → API request
├── Parse response
├── Function calling support
└── Unit tests

Day 6-7: Streaming
├── SSE parsing
└── Channel-based streaming
```

**Deliverable:** Can call OpenAI and get responses

---

### Week 3: Tools & Agent

**Goal:** First working agent!

```
Day 1-2: Simple Tools
├── Calculator (pure Go, no deps)
├── HTTP client tool
└── Tool schema/registration

Day 3-5: Function Agent
├── LLM → Tool selection
├── Execute tool
├── Format response
└── Basic error handling

Day 6-7: Integration & Examples
├── End-to-end example
├── Integration tests
└── Documentation
```

**Deliverable:** Working agent that can use tools!

---

### Week 4: Polish & Extend

**Goal:** Make it production-ready

```
Day 1-3: More Tools
├── Web search (SerpAPI)
├── Web scraper (Colly)
└── Database tool (PostgreSQL)

Day 4-5: Agent Features
├── Conversation memory
├── Streaming responses
└── Better error messages

Day 6-7: Release Prep
├── README with examples
├── GoDoc comments
├── v0.1.0 release
```

**Deliverable:** `goagent v0.1.0` - First public release!

---

## First Sprint Tasks (Next 2 Weeks)

### Sprint 1: Minimum Viable Product

**Goal:** One complete example working

#### Task 1: Project Setup (4 hours)

```bash
# Initialize Go module
cd goagent
go mod init github.com/yourusername/goagent

# Create structure
mkdir -p {core,llm/openai,tools,agent,examples/quickstart}

# Setup GitHub Actions
# ... CI configuration
```

#### Task 2: Core Types (8 hours)

**File:** `core/types.go`

```go
package core

import (
    "context"
    "time"
)

// Message represents a chat message
type Message struct {
    Role    string                 // "user", "assistant", "system"
    Content string                 // Text content
    Name    string                 // Optional speaker name
    Meta    map[string]interface{} // Additional metadata
}

// Response from LLM or agent
type Response struct {
    Content   string                 // Main response text
    ToolCalls []ToolCall             // Tools that were called
    Meta      map[string]interface{} // Metadata (tokens, latency, etc)
}

// ToolCall represents a function/tool invocation
type ToolCall struct {
    ID     string                 // Unique call ID
    Name   string                 // Tool name
    Args   map[string]interface{} // Arguments
    Result interface{}            // Execution result
}

// ToolSchema defines a tool's interface
type ToolSchema struct {
    Name        string
    Description string
    Parameters  []Parameter
}

// Parameter for a tool
type Parameter struct {
    Name        string
    Type        string // "string", "number", "boolean", "object"
    Description string
    Required    bool
}
```

#### Task 3: Core Interfaces (8 hours)

**File:** `core/interfaces.go`

```go
package core

import "context"

// LLM is the interface for language model providers
type LLM interface {
    // Chat sends messages and gets a response
    Chat(ctx context.Context, messages []Message) (*Response, error)
    
    // Complete sends a prompt and gets completion
    Complete(ctx context.Context, prompt string) (string, error)
}

// Tool is something an agent can use to accomplish tasks
type Tool interface {
    // Name returns the tool's unique identifier
    Name() string
    
    // Description explains what the tool does
    Description() string
    
    // Schema returns the tool's parameter schema
    Schema() *ToolSchema
    
    // Execute runs the tool with given arguments
    Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// Agent coordinates between LLM and tools
type Agent interface {
    // Run executes the agent with given input
    Run(ctx context.Context, input string) (*Response, error)
    
    // AddTool registers a new tool
    AddTool(tool Tool) error
    
    // Reset clears conversation history
    Reset() error
}
```

#### Task 4: OpenAI Client (24 hours)

**File:** `llm/openai/client.go`

```go
package openai

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "github.com/yourusername/goagent/core"
)

type Client struct {
    apiKey  string
    model   string
    baseURL string
    http    *http.Client
}

type Option func(*Client)

func New(opts ...Option) (*Client, error) {
    c := &Client{
        model:   "gpt-4o-mini",
        baseURL: "https://api.openai.com/v1",
        http: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
    
    for _, opt := range opts {
        opt(c)
    }
    
    if c.apiKey == "" {
        return nil, fmt.Errorf("API key required")
    }
    
    return c, nil
}

func WithAPIKey(key string) Option {
    return func(c *Client) {
        c.apiKey = key
    }
}

func WithModel(model string) Option {
    return func(c *Client) {
        c.model = model
    }
}

// Chat implements core.LLM
func (c *Client) Chat(ctx context.Context, messages []core.Message) (*core.Response, error) {
    // Convert to OpenAI format
    reqMessages := make([]map[string]string, len(messages))
    for i, msg := range messages {
        reqMessages[i] = map[string]string{
            "role":    msg.Role,
            "content": msg.Content,
        }
    }
    
    reqBody := map[string]interface{}{
        "model":    c.model,
        "messages": reqMessages,
    }
    
    body, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", 
        c.baseURL+"/chat/completions", bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.apiKey)
    
    resp, err := c.http.Do(req)
    if err != nil {
        return nil, fmt.Errorf("http request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    
    var result struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    
    if len(result.Choices) == 0 {
        return nil, fmt.Errorf("no response from API")
    }
    
    return &core.Response{
        Content: result.Choices[0].Message.Content,
    }, nil
}

// Complete implements core.LLM
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
    resp, err := c.Chat(ctx, []core.Message{
        {Role: "user", Content: prompt},
    })
    if err != nil {
        return "", err
    }
    return resp.Content, nil
}
```

#### Task 5: Simple Tool (8 hours)

**File:** `tools/calculator.go`

```go
package tools

import (
    "context"
    "fmt"
    "github.com/yourusername/goagent/core"
)

type Calculator struct{}

func NewCalculator() *Calculator {
    return &Calculator{}
}

func (c *Calculator) Name() string {
    return "calculator"
}

func (c *Calculator) Description() string {
    return "Performs basic arithmetic operations (add, subtract, multiply, divide)"
}

func (c *Calculator) Schema() *core.ToolSchema {
    return &core.ToolSchema{
        Name:        "calculator",
        Description: c.Description(),
        Parameters: []core.Parameter{
            {
                Name:        "operation",
                Type:        "string",
                Description: "Operation to perform: add, subtract, multiply, divide",
                Required:    true,
            },
            {
                Name:        "a",
                Type:        "number",
                Description: "First number",
                Required:    true,
            },
            {
                Name:        "b",
                Type:        "number",
                Description: "Second number",
                Required:    true,
            },
        },
    }
}

func (c *Calculator) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    op, ok := args["operation"].(string)
    if !ok {
        return nil, fmt.Errorf("operation must be a string")
    }
    
    a, ok := args["a"].(float64)
    if !ok {
        return nil, fmt.Errorf("a must be a number")
    }
    
    b, ok := args["b"].(float64)
    if !ok {
        return nil, fmt.Errorf("b must be a number")
    }
    
    switch op {
    case "add":
        return a + b, nil
    case "subtract":
        return a - b, nil
    case "multiply":
        return a * b, nil
    case "divide":
        if b == 0 {
            return nil, fmt.Errorf("division by zero")
        }
        return a / b, nil
    default:
        return nil, fmt.Errorf("unknown operation: %s", op)
    }
}
```

#### Task 6: Basic Agent (16 hours)

**File:** `agent/function.go`

```go
package agent

import (
    "context"
    "fmt"
    "github.com/yourusername/goagent/core"
)

// FunctionAgent is a simple agent that uses tools
type FunctionAgent struct {
    llm   core.LLM
    tools map[string]core.Tool
}

// New creates a new FunctionAgent
func New(llm core.LLM, tools ...core.Tool) (*FunctionAgent, error) {
    if llm == nil {
        return nil, fmt.Errorf("LLM is required")
    }
    
    agent := &FunctionAgent{
        llm:   llm,
        tools: make(map[string]core.Tool),
    }
    
    for _, tool := range tools {
        if err := agent.AddTool(tool); err != nil {
            return nil, err
        }
    }
    
    return agent, nil
}

// AddTool registers a tool
func (a *FunctionAgent) AddTool(tool core.Tool) error {
    if tool == nil {
        return fmt.Errorf("tool cannot be nil")
    }
    
    name := tool.Name()
    if _, exists := a.tools[name]; exists {
        return fmt.Errorf("tool %s already registered", name)
    }
    
    a.tools[name] = tool
    return nil
}

// Run executes the agent
func (a *FunctionAgent) Run(ctx context.Context, input string) (*core.Response, error) {
    // Create system prompt with tool descriptions
    systemPrompt := "You are a helpful assistant with access to these tools:\n"
    for _, tool := range a.tools {
        systemPrompt += fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description())
    }
    systemPrompt += "\nTo use a tool, respond with: USE_TOOL: <tool_name> <args_as_json>"
    
    messages := []core.Message{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: input},
    }
    
    // Call LLM
    resp, err := a.llm.Chat(ctx, messages)
    if err != nil {
        return nil, fmt.Errorf("llm chat: %w", err)
    }
    
    // TODO: Parse tool calls and execute
    // For MVP, just return LLM response
    
    return resp, nil
}

// Reset clears history
func (a *FunctionAgent) Reset() error {
    // No history yet in MVP
    return nil
}
```

#### Task 7: Example Application (4 hours)

**File:** `examples/quickstart/main.go`

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/yourusername/goagent/agent"
    "github.com/yourusername/goagent/llm/openai"
    "github.com/yourusername/goagent/tools"
)

func main() {
    ctx := context.Background()
    
    // 1. Create LLM client
    apiKey := os.Getenv("OPENAI_API_KEY")
    llm, err := openai.New(
        openai.WithAPIKey(apiKey),
        openai.WithModel("gpt-4o-mini"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. Create tools
    calculator := tools.NewCalculator()
    
    // 3. Create agent
    agent, err := agent.New(llm, calculator)
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. Run agent
    response, err := agent.Run(ctx, "What is 25 * 34?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Agent:", response.Content)
}
```

---

## Success Criteria

After 2 weeks, you should have:

✅ Clean package structure  
✅ Core types & interfaces defined  
✅ Working OpenAI client  
✅ At least 1 working tool  
✅ Basic FunctionAgent  
✅ One complete example running  
✅ Unit tests for each component  
✅ README with usage instructions

---

## What NOT to Build (Yet)

Don't get distracted by these (save for later):

❌ Streaming (Week 3)  
❌ Memory/history (Week 3)  
❌ Multiple agents (Week 10+)  
❌ RAG (Week 5-8)  
❌ Workflows (Week 9+)  
❌ Advanced tools (Week 3-4)  
❌ Multiple LLM providers (Week 3)  
❌ Observability (Week 8+)

**Focus:** One working example end-to-end!

---

## Next Steps

1. **Read BEST_PRACTICES.md** - Understand patterns
2. **Set up go.mod** - Initialize project
3. **Create core package** - Types & interfaces
4. **Implement OpenAI client** - LLM integration
5. **Build calculator tool** - First tool
6. **Create FunctionAgent** - Tie it together
7. **Write example** - Prove it works
8. **Test & document** - Make it usable

**Let's build!** 🚀
