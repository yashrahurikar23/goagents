# GoAgent Best Practices & Design Principles

**Purpose:** Guidelines for building a clean, extensible, and scalable SDK that developers will love to use.

**Last Updated:** October 7, 2025

---

## Core Design Philosophy

### 1. **Simplicity Over Complexity**

**Principle:** Make simple things simple, complex things possible.

```go
// ✅ GOOD: Simple default, complexity optional
agent, err := agent.New(llm, tools...)

// ✅ GOOD: Advanced users can customize
agent, err := agent.New(llm, tools,
    agent.WithTimeout(30*time.Second),
    agent.WithRetries(3),
    agent.WithMemory(memory.NewBuffer()),
)

// ❌ BAD: Forcing complexity on simple use cases
agent, err := agent.New(&AgentConfig{
    LLM: llm,
    Tools: tools,
    Options: &AgentOptions{
        Timeout: time.Duration(30) * time.Second,
        Retries: 3,
        Memory: &MemoryConfig{Type: "buffer"},
    },
})
```

**Rules:**
- Zero-config defaults for 80% use cases
- Progressive disclosure of complexity
- Don't make users learn the entire API upfront

---

## Implementation Strategy

### Build Order: Bottom-Up Approach

**Why Bottom-Up?**
- ✅ Solid foundation prevents refactoring
- ✅ Test components in isolation
- ✅ Clear dependencies (no circular imports)
- ✅ Easy to understand and reason about

**Build Order:**

```
Phase 1: Foundation (Week 1-4)
├── 1. Core types & interfaces     ← Start here
├── 2. LLM clients (OpenAI)        ← Basic functionality
├── 3. Simple tools                ← Concrete implementation
└── 4. Basic agent (FunctionAgent) ← Integrate components

Phase 2: Data Layer (Week 5-8)
├── 5. Document readers
├── 6. Text splitters
├── 7. Embeddings
└── 8. Vector stores

Phase 3: Intelligence (Week 9-12)
├── 9. Retrievers
├── 10. Query engines
├── 11. RAG pipeline
└── 12. Advanced agents

Phase 4: Orchestration (Week 13+)
├── 13. Workflows
├── 14. Multi-agent
└── 15. Production features
```

---

## First Features to Build

### Sprint 1-2: Absolute Minimum (2 weeks)

**Goal:** One working example end-to-end

```go
// Target: This should work after 2 weeks
package main

import (
    "context"
    "fmt"
    "github.com/yourusername/goagent/agent"
    "github.com/yourusername/goagent/llm/openai"
    "github.com/yourusername/goagent/tools"
)

func main() {
    ctx := context.Background()
    
    // 1. LLM client
    llm, _ := openai.New(openai.WithAPIKey("sk-..."))
    
    // 2. Simple tool
    calculator := tools.NewCalculator()
    
    // 3. Agent
    agent, _ := agent.New(llm, calculator)
    
    // 4. Run
    response, _ := agent.Run(ctx, "What is 25 * 34?")
    fmt.Println(response.Content) // "850"
}
```

**Priority Order:**

1. **Core Interfaces** (Day 1-2)
   - `Agent` interface
   - `Tool` interface  
   - `LLM` interface
   - `Message`, `Response` types

2. **OpenAI Client** (Day 3-5)
   - Chat completion
   - Function calling
   - Error handling

3. **Simple Tool** (Day 6-7)
   - Calculator tool (pure Go, no deps)
   - Tool schema definition

4. **Function Agent** (Day 8-10)
   - Basic loop: LLM → Tool selection → Execute → Response
   - No streaming, no memory, no fancy features

5. **Example + Tests** (Day 11-14)
   - Working example app
   - Unit tests for each component
   - Integration test

---

## API Design Patterns

### Pattern 1: Functional Options (Primary)

**Why:** Idiomatic Go, extensible without breaking changes, self-documenting.

```go
// Core pattern for all constructors
type Option func(*Config) error

// Constructor
func New(required1, required2 string, opts ...Option) (*Thing, error) {
    cfg := &Config{
        // Sensible defaults
        Timeout: 30 * time.Second,
        Retries: 3,
    }
    
    for _, opt := range opts {
        if err := opt(cfg); err != nil {
            return nil, fmt.Errorf("invalid option: %w", err)
        }
    }
    
    // Validate required + optional
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return &Thing{config: cfg}, nil
}

// Options
func WithTimeout(d time.Duration) Option {
    return func(c *Config) error {
        if d <= 0 {
            return fmt.Errorf("timeout must be positive")
        }
        c.Timeout = d
        return nil
    }
}
```

**Benefits:**
- Backward compatible (add options without breaking API)
- Self-documenting (IDE autocomplete shows options)
- Validation at construction time

---

### Pattern 2: Interface Segregation

**Principle:** Small, focused interfaces > large, monolithic ones

```go
// ✅ GOOD: Small, focused interfaces
type Completer interface {
    Complete(ctx context.Context, prompt string) (string, error)
}

type Chatter interface {
    Chat(ctx context.Context, messages []Message) (*Response, error)
}

type Streamer interface {
    Stream(ctx context.Context, prompt string) (<-chan Token, error)
}

// Compose when needed
type LLM interface {
    Completer
    Chatter
    Streamer
}

// ❌ BAD: God interface
type LLM interface {
    Complete(...) (...)
    Chat(...) (...)
    Stream(...) (...)
    Embed(...) (...)
    FineTune(...) (...)
    // ... 20 more methods
}
```

**Benefits:**
- Easy to implement (can implement just what you need)
- Easy to test (mock only what you use)
- Easy to extend (add interfaces without breaking existing)

---

### Pattern 3: Dependency Injection

**Principle:** Pass dependencies explicitly, no global state

```go
// ✅ GOOD: Explicit dependencies
type Agent struct {
    llm   LLM
    tools []Tool
}

func NewAgent(llm LLM, tools ...Tool) *Agent {
    return &Agent{llm: llm, tools: tools}
}

// ❌ BAD: Global state
var defaultLLM LLM

func NewAgent(tools ...Tool) *Agent {
    return &Agent{llm: defaultLLM, tools: tools}
}
```

**Benefits:**
- Testable (inject mocks)
- Explicit (no hidden dependencies)
- Thread-safe (no shared state)

---

### Pattern 4: Context for Cancellation

**Principle:** All I/O operations take `context.Context` first parameter

```go
// ✅ GOOD: Context first parameter
func (a *Agent) Run(ctx context.Context, input string) (*Response, error) {
    // Can cancel, timeout, pass values
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Continue...
    }
}

// ❌ BAD: No context
func (a *Agent) Run(input string) (*Response, error) {
    // Can't cancel, no timeout
}
```

---

### Pattern 5: Error Wrapping

**Principle:** Wrap errors with context, preserve original

```go
// ✅ GOOD: Wrap with context
func (c *Client) Chat(ctx context.Context, msgs []Message) (*Response, error) {
    resp, err := c.doRequest(ctx, msgs)
    if err != nil {
        return nil, fmt.Errorf("chat request failed: %w", err)
    }
    return resp, nil
}

// ❌ BAD: Lose context
func (c *Client) Chat(ctx context.Context, msgs []Message) (*Response, error) {
    resp, err := c.doRequest(ctx, msgs)
    if err != nil {
        return nil, err // What failed? Where?
    }
    return resp, nil
}
```

---

## Package Organization

### Principle: Minimize Dependencies

**Goal:** Each package should be independently usable

```
goagent/
├── core/           # No dependencies (types, interfaces)
├── llm/            # Depends on: core
│   ├── openai/     # Depends on: core, llm
│   └── anthropic/  # Depends on: core, llm
├── tools/          # Depends on: core
│   └── builtin/    # Depends on: core, tools
├── agent/          # Depends on: core, llm, tools
└── examples/       # Depends on: everything
```

**Rules:**
1. **Core package has ZERO external dependencies** (only stdlib)
2. **Provider packages** (llm/openai) depend on core + interface package
3. **High-level packages** (agent) depend on abstractions, not implementations
4. **No circular dependencies** (use interfaces to break cycles)

---

### Package Naming

```go
// ✅ GOOD: Short, clear, no stuttering
import "github.com/user/goagent/agent"
import "github.com/user/goagent/tools"

agent.New(...)      // Not agent.NewAgent()
tools.Calculator()  // Not tools.NewCalculatorTool()

// ❌ BAD: Stuttering
import "github.com/user/goagent/agent"
agent.NewAgent()    // Stutter!

// ❌ BAD: Unclear
import "github.com/user/goagent/utils"
utils.Thing()       // What is this?
```

---

## Code Quality Standards

### 1. **Documentation**

**Rule:** Every exported symbol has a doc comment

```go
// ✅ GOOD: Complete documentation
// Agent represents an AI agent that can use tools to accomplish tasks.
// It coordinates between an LLM and a set of tools, deciding which tools
// to call based on the user's input.
//
// Example:
//   agent, err := agent.New(llm, searchTool, calculatorTool)
//   if err != nil {
//       log.Fatal(err)
//   }
//   response, err := agent.Run(ctx, "What is 2+2?")
type Agent struct {
    // ...
}

// ❌ BAD: Missing or vague
// Agent does agent stuff
type Agent struct {
    // ...
}
```

---

### 2. **Testing**

**Standards:**

```go
// Table-driven tests
func TestAgent_Run(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "simple calculation",
            input: "What is 2+2?",
            want:  "4",
        },
        {
            name:    "invalid input",
            input:   "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            agent := setupTestAgent(t)
            got, err := agent.Run(context.Background(), tt.input)
            
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.want, got.Content)
        })
    }
}
```

**Coverage Targets:**
- Core packages: 80%+ coverage
- Integration tests: Critical paths
- Examples: Must compile and run

---

### 3. **Error Messages**

**Principle:** Errors should be actionable

```go
// ✅ GOOD: Actionable error
return nil, fmt.Errorf("failed to connect to OpenAI API: %w. Check your API key at https://platform.openai.com/api-keys", err)

// ❌ BAD: Vague error
return nil, fmt.Errorf("connection error: %w", err)
```

---

## Extensibility Patterns

### 1. **Plugin Architecture**

**Goal:** Users can add their own LLMs, tools, vector stores

```go
// Define interface
type VectorStore interface {
    Add(ctx context.Context, vectors []Vector) error
    Search(ctx context.Context, query Vector, k int) ([]Result, error)
}

// Register pattern (optional)
var vectorStores = make(map[string]func() VectorStore)

func RegisterVectorStore(name string, factory func() VectorStore) {
    vectorStores[name] = factory
}

// Users can implement
type MyCustomStore struct {}

func (m *MyCustomStore) Add(ctx context.Context, vectors []Vector) error {
    // Custom implementation
}

func (m *MyCustomStore) Search(ctx context.Context, query Vector, k int) ([]Result, error) {
    // Custom implementation
}
```

---

### 2. **Middleware/Hooks Pattern**

**Goal:** Allow users to intercept and modify behavior

```go
// Middleware for LLM calls
type LLMMiddleware func(next LLM) LLM

// Example: Logging middleware
func WithLogging(logger *slog.Logger) LLMMiddleware {
    return func(next LLM) LLM {
        return &loggingLLM{
            next:   next,
            logger: logger,
        }
    }
}

type loggingLLM struct {
    next   LLM
    logger *slog.Logger
}

func (l *loggingLLM) Chat(ctx context.Context, msgs []Message) (*Response, error) {
    l.logger.Info("llm.chat.start", "messages", len(msgs))
    resp, err := l.next.Chat(ctx, msgs)
    l.logger.Info("llm.chat.end", "error", err)
    return resp, err
}

// Usage
llm := openai.New(apiKey)
llm = WithLogging(logger)(llm)
llm = WithCaching(cache)(llm)
```

---

### 3. **Builder Pattern for Complex Objects**

**When to use:** Complex objects with many optional fields

```go
// Builder for query engines
type QueryEngineBuilder struct {
    retriever Retriever
    topK      int
    reranker  Reranker
    // ... many optional fields
}

func NewQueryEngineBuilder(retriever Retriever) *QueryEngineBuilder {
    return &QueryEngineBuilder{
        retriever: retriever,
        topK:      5, // defaults
    }
}

func (b *QueryEngineBuilder) TopK(k int) *QueryEngineBuilder {
    b.topK = k
    return b
}

func (b *QueryEngineBuilder) Reranker(r Reranker) *QueryEngineBuilder {
    b.reranker = r
    return b
}

func (b *QueryEngineBuilder) Build() (*QueryEngine, error) {
    // Validation
    return &QueryEngine{/* ... */}, nil
}

// Usage
engine, err := NewQueryEngineBuilder(retriever).
    TopK(10).
    Reranker(reranker).
    Build()
```

---

## Scalability Patterns

### 1. **Concurrency with Worker Pools**

```go
// Process documents concurrently
type Pipeline struct {
    workers int
}

func (p *Pipeline) Process(ctx context.Context, docs []Document) ([]Node, error) {
    // Worker pool
    jobs := make(chan Document, len(docs))
    results := make(chan Node, len(docs))
    errors := make(chan error, 1)
    
    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < p.workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for doc := range jobs {
                node, err := p.processOne(ctx, doc)
                if err != nil {
                    select {
                    case errors <- err:
                    default:
                    }
                    return
                }
                results <- node
            }
        }()
    }
    
    // Send jobs
    go func() {
        for _, doc := range docs {
            jobs <- doc
        }
        close(jobs)
    }()
    
    // Wait and collect
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    var nodes []Node
    for node := range results {
        nodes = append(nodes, node)
    }
    
    select {
    case err := <-errors:
        return nil, err
    default:
        return nodes, nil
    }
}
```

---

### 2. **Streaming Results**

```go
// Stream tokens instead of waiting for complete response
func (a *Agent) RunStream(ctx context.Context, input string) (<-chan Event, error) {
    events := make(chan Event, 10) // Buffered
    
    go func() {
        defer close(events)
        
        // Process and send events
        for token := range a.llm.Stream(ctx, input) {
            select {
            case events <- Event{Type: "token", Data: token}:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return events, nil
}
```

---

### 3. **Resource Pooling**

```go
// Connection pool for vector stores
type Pool struct {
    conns chan *Connection
    size  int
}

func NewPool(size int, factory func() (*Connection, error)) (*Pool, error) {
    p := &Pool{
        conns: make(chan *Connection, size),
        size:  size,
    }
    
    for i := 0; i < size; i++ {
        conn, err := factory()
        if err != nil {
            return nil, err
        }
        p.conns <- conn
    }
    
    return p, nil
}

func (p *Pool) Get(ctx context.Context) (*Connection, error) {
    select {
    case conn := <-p.conns:
        return conn, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (p *Pool) Put(conn *Connection) {
    p.conns <- conn
}
```

---

## Anti-Patterns to Avoid

### ❌ 1. **God Objects**

Don't create objects that do everything

```go
// ❌ BAD
type Agent struct {
    // Agent stuff
    llm LLM
    tools []Tool
    
    // Vector store stuff
    vectorStore VectorStore
    embeddings Embeddings
    
    // HTTP server stuff
    router *http.ServeMux
    
    // Logging stuff
    logger *slog.Logger
    
    // Config stuff
    config *Config
    
    // ... 50 more fields
}
```

**Fix:** Composition over inheritance

```go
// ✅ GOOD
type Agent struct {
    llm   LLM
    tools []Tool
}

type RAGAgent struct {
    agent      *Agent
    index      Index
    queryEngine QueryEngine
}
```

---

### ❌ 2. **Premature Abstraction**

Don't create interfaces until you have 2+ implementations

```go
// ❌ BAD: Only one implementation
type CalculatorInterface interface {
    Add(a, b int) int
    Subtract(a, b int) int
}

type Calculator struct{}
func (c *Calculator) Add(a, b int) int { return a + b }
```

**Fix:** Start concrete, abstract when needed

```go
// ✅ GOOD: Start concrete
type Calculator struct{}
func (c *Calculator) Add(a, b int) int { return a + b }

// Add interface later when you have multiple implementations
```

---

### ❌ 3. **Package Cycles**

```go
// ❌ BAD: Cycle
package agent
import "goagent/tools"

package tools
import "goagent/agent"  // Cycle!
```

**Fix:** Introduce interface in common package

```go
// ✅ GOOD
package core
type Agent interface {...}

package agent
import "goagent/core"
// Implements core.Agent

package tools
import "goagent/core"
// Uses core.Agent interface
```

---

## Performance Guidelines

### 1. **Avoid Allocations in Hot Paths**

```go
// ❌ BAD: Allocates on every call
func (a *Agent) process(input string) string {
    buffer := bytes.NewBuffer(nil) // Allocation!
    buffer.WriteString(input)
    return buffer.String()
}

// ✅ GOOD: Reuse buffer
type Agent struct {
    buffer *bytes.Buffer
}

func (a *Agent) process(input string) string {
    a.buffer.Reset()
    a.buffer.WriteString(input)
    return a.buffer.String()
}
```

---

### 2. **Use sync.Pool for Temporary Objects**

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process(input string) string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    buf.WriteString(input)
    return buf.String()
}
```

---

### 3. **Benchmark Critical Paths**

```go
func BenchmarkAgent_Run(b *testing.B) {
    agent := setupTestAgent(b)
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = agent.Run(ctx, "test input")
    }
}
```

---

## Summary: Golden Rules

1. **Start Simple** - Build minimum viable features first
2. **Bottom-Up** - Foundation before fancy features
3. **Interfaces** - Small, focused, implementation-agnostic
4. **Options** - Functional options for extensibility
5. **Context** - Always pass context for cancellation
6. **Errors** - Wrap with context, make actionable
7. **Documentation** - Every export has examples
8. **Testing** - Table-driven tests, 80%+ coverage
9. **No Magic** - Explicit > implicit
10. **Composition** - Small pieces, loosely joined

---

## Next Steps

1. **Review this doc** before starting any feature
2. **Reference during PRs** to maintain consistency
3. **Update** as we learn what works
4. **Share** with contributors

---

**Remember:** Good code is code that others can understand, extend, and maintain. We're building for the long term! 🚀
