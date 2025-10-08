# 🎨 API Design Guide for GoAgents

**Best practices for maintaining a stable, user-friendly API**

---

## Table of Contents

1. [Constructor Naming Convention](#constructor-naming-convention)
2. [Avoiding Breaking Changes](#avoiding-breaking-changes)
3. [Go API Design Principles](#go-api-design-principles)
4. [Versioning Strategy](#versioning-strategy)

---

## Constructor Naming Convention

### The `New*` Prefix Debate

**Question:** Should we rename `NewReActAgent` to just `ReActAgent`?

### ❌ Recommendation: **KEEP the `New` Prefix**

**Why Go Uses `New*`:**

1. **Standard Library Convention**
   ```go
   // Go standard library examples
   http.NewRequest()
   bytes.NewBuffer()
   strings.NewReader()
   context.Background()  // Exception (factory function)
   ```

2. **Distinguishes Constructor from Type**
   ```go
   // ✅ CLEAR: Constructor vs Type
   agent := agent.NewReActAgent(llm)  // Constructor function
   var a agent.ReActAgent             // Type declaration
   
   // ❌ CONFUSING: Both look like types
   agent := agent.ReActAgent(llm)     // Is this a constructor?
   var a agent.ReActAgent             // Or a type?
   ```

3. **Ecosystem Consistency**
   ```go
   // Popular Go packages all use New*
   gin.New()
   gorm.New()
   redis.NewClient()
   grpc.NewServer()
   zerolog.New()
   ```

4. **User Expectations**
   - Go developers **expect** `New*` constructors
   - Deviating causes confusion
   - Hurts adoption

### ✅ When to Skip `New` Prefix

**Only skip `New` for:**

1. **Factory Functions** (return different types)
   ```go
   // Returns different types based on config
   func Agent(config Config) interface{} {
       if config.Type == "react" {
           return NewReActAgent(config)
       }
       return NewFunctionAgent(config)
   }
   ```

2. **Builder Pattern Starters**
   ```go
   // Builder pattern
   agent := agent.Builder().
       WithLLM(llm).
       WithTools(tools).
       Build()
   ```

3. **Simple Value Constructors**
   ```go
   // Simple value wrappers (no allocation)
   func Timeout(d time.Duration) TimeoutOption {
       return TimeoutOption{duration: d}
   }
   ```

### Alternative: Support Both (Not Recommended)

If you **really** want both styles:

```go
// agent/react.go

// NewReActAgent creates a new ReAct agent (recommended).
func NewReActAgent(llm core.LLM, opts ...Option) *ReActAgent {
    // implementation
}

// ReActAgent is an alias for NewReActAgent.
// Deprecated: Use NewReActAgent for consistency with Go conventions.
func ReActAgent(llm core.LLM, opts ...Option) *ReActAgent {
    return NewReActAgent(llm, opts...)
}
```

**Problems with this approach:**
- ❌ Confusing documentation
- ❌ Two ways to do the same thing
- ❌ Still non-idiomatic
- ❌ Adds maintenance burden

### 📊 Community Examples

**Packages that use `New*`:**
- Docker SDK: `docker.NewClient()`
- Kubernetes client: `kubernetes.NewForConfig()`
- AWS SDK: `s3.New()`
- gRPC: `grpc.NewServer()`
- Gin: `gin.New()`
- Echo: `echo.New()`
- Chi: `chi.NewRouter()`

**Packages that DON'T use `New*`:**
- Cobra: `cobra.Command{}` (struct literal, not constructor)
- Viper: `viper.Get()` (singleton pattern)

### 🎯 Final Verdict

**Keep `New*` prefix for:**
- ✅ Consistency with Go ecosystem
- ✅ Clear distinction between type and constructor
- ✅ User expectations
- ✅ Better documentation

**Your current API is perfect:**
```go
agent := agent.NewReActAgent(llm)
calc := tools.NewCalculator()
```

**Don't change it!** 🎯

---

## Avoiding Breaking Changes

### Understanding Semantic Versioning

```
vMAJOR.MINOR.PATCH
```

- **MAJOR** (v1 → v2): Breaking changes allowed
- **MINOR** (v0.2 → v0.3): New features, no breaks
- **PATCH** (v0.2.0 → v0.2.1): Bug fixes only

### Before v1.0.0 (Current: v0.2.0)

**Good news:** Breaking changes are EXPECTED in v0.x.x

- ✅ You can evolve the API freely
- ✅ Users know it's still maturing
- ✅ Document breaking changes in CHANGELOG

**But still minimize breaks:**
- Makes upgrades easier
- Builds user trust
- Reduces migration pain

### After v1.0.0

**Strict compatibility rules:**

```go
// v1.x.x - Cannot change these without v2.0.0
type Agent interface {
    Run(ctx context.Context, input string) (*Response, error)
}

// Adding this requires v2.0.0 (breaking change)
type Agent interface {
    Run(ctx context.Context, input string) (*Response, error)
    RunStream(ctx context.Context, input string) (<-chan Token, error)  // NEW
}
```

---

## Strategies to Avoid Breaking Changes

### 1. Use Interfaces, Not Concrete Types

**✅ GOOD - Extensible**
```go
// Define interface
type LLM interface {
    Generate(ctx context.Context, messages []Message) (*Response, error)
}

// Implementation
type OpenAIClient struct {
    apiKey string
    // Internal fields can change
}

func (c *OpenAIClient) Generate(ctx context.Context, messages []Message) (*Response, error) {
    // implementation
}

// Users depend on interface, not struct
func NewAgent(llm LLM) *Agent {
    return &Agent{llm: llm}
}
```

**Why it works:**
- Can add methods to interface in v2.x.x
- Can change internal implementation freely
- Users only depend on interface

**❌ BAD - Rigid**
```go
// Users depend on concrete type
func NewAgent(llm *OpenAIClient) *Agent {
    return &Agent{llm: llm}
}

// Can't change OpenAIClient without breaking users
type OpenAIClient struct {
    APIKey string  // Exported - can't change!
}
```

---

### 2. Functional Options Pattern

**✅ GOOD - Infinitely Extensible**
```go
type Option func(*Agent)

func WithMaxIterations(n int) Option {
    return func(a *Agent) {
        a.maxIterations = n
    }
}

func WithVerbose(v bool) Option {
    return func(a *Agent) {
        a.verbose = v
    }
}

// Constructor accepts variable options
func NewReActAgent(llm core.LLM, opts ...Option) *ReActAgent {
    a := &ReActAgent{
        llm:           llm,
        maxIterations: 10,  // defaults
        verbose:       false,
    }
    
    for _, opt := range opts {
        opt(a)
    }
    
    return a
}

// Usage - can add new options without breaking
agent := agent.NewReActAgent(llm,
    agent.WithMaxIterations(20),
    agent.WithVerbose(true),
    agent.WithTimeout(30*time.Second),  // Can add later!
)
```

**Why it works:**
- Add new options anytime (non-breaking)
- Backward compatible (old code still works)
- Clear, self-documenting API

**❌ BAD - Breaks on Every Addition**
```go
// Adding parameter breaks all existing code!
func NewReActAgent(llm core.LLM, maxIter int, verbose bool) *ReActAgent {
    // ...
}

// Later want to add timeout...
// This breaks EVERYONE
func NewReActAgent(llm core.LLM, maxIter int, verbose bool, timeout time.Duration) *ReActAgent {
    // ...
}
```

---

### 3. Unexported Struct Fields

**✅ GOOD - Internal Flexibility**
```go
type ReActAgent struct {
    llm         core.LLM      // unexported - can change freely
    tools       []core.Tool   // unexported
    maxIter     int           // unexported
    verbose     bool          // unexported
}

// Access through methods only
func (a *ReActAgent) AddTool(tool core.Tool) {
    a.tools = append(a.tools, tool)
}

// Can change internal structure without breaking users
```

**Why it works:**
- Internal fields can change anytime
- Only public API matters
- Users can't depend on internal state

**❌ BAD - Locked In**
```go
type ReActAgent struct {
    LLM         core.LLM      // EXPORTED - can't change!
    Tools       []core.Tool   // EXPORTED - locked in!
    MaxIter     int           // EXPORTED - fixed!
}

// Users directly access fields
agent.MaxIter = 20  // Now we can't change the field name or type
```

---

### 4. Deprecation, Not Removal

**✅ GOOD - Graceful Migration**
```go
// Version 1: Original method
func (a *Agent) Run(input string) (*Response, error) {
    return a.RunContext(context.Background(), input)
}

// Version 2: Want to add context
// DON'T remove old method - deprecate it!

// Run is deprecated. Use RunContext instead.
//
// Deprecated: Use RunContext for better cancellation support.
func (a *Agent) Run(input string) (*Response, error) {
    return a.RunContext(context.Background(), input)
}

// RunContext is the new recommended method.
func (a *Agent) RunContext(ctx context.Context, input string) (*Response, error) {
    // new implementation
}
```

**Migration path:**
1. v0.2.0: Add `RunContext()`, keep `Run()`
2. v0.3.0: Mark `Run()` as deprecated
3. v0.4.0: Still keep `Run()` (works, just warns)
4. v1.0.0: Can remove `Run()` (major version)

**Why it works:**
- Users have time to migrate
- No sudden breakage
- Clear migration path
- Linters warn about deprecated usage

**❌ BAD - Immediate Break**
```go
// v0.2.0: Has Run()
func (a *Agent) Run(input string) (*Response, error)

// v0.3.0: Removed Run(), added RunContext()
// BREAKS ALL EXISTING CODE!
func (a *Agent) RunContext(ctx context.Context, input string) (*Response, error)
```

---

### 5. Extend, Don't Replace

**✅ GOOD - Additive Changes**
```go
// v0.2.0
type Response struct {
    Content string
    Tokens  int
}

// v0.3.0 - Add fields (non-breaking)
type Response struct {
    Content    string
    Tokens     int
    Cost       float64  // NEW - but doesn't break old code
    Model      string   // NEW
}

// Old code still works
response, _ := agent.Run(ctx, "question")
fmt.Println(response.Content)  // Works as before
```

**❌ BAD - Breaking Changes**
```go
// v0.2.0
type Response struct {
    Content string
    Tokens  int
}

// v0.3.0 - Changed field type (BREAKING!)
type Response struct {
    Content string
    Tokens  float64  // Changed int -> float64 (BREAKS!)
}

// v0.3.0 - Removed field (BREAKING!)
type Response struct {
    Content string
    // Removed Tokens field (BREAKS!)
}
```

---

### 6. Version Packages for Major Changes

**For truly breaking changes, use module versioning:**

```go
// v1 - Original API
import "github.com/yashrahurikar23/goagents/agent"

agent := agent.NewReActAgent(llm)
```

```go
// v2 - Breaking changes in separate import path
import "github.com/yashrahurikar23/goagents/v2/agent"

agent := agent.NewReActAgent(llm, config)  // Different signature
```

**Both versions can coexist:**
```go
import (
    v1agent "github.com/yashrahurikar23/goagents/agent"
    v2agent "github.com/yashrahurikar23/goagents/v2/agent"
)

// Can use both in same project during migration
oldAgent := v1agent.NewReActAgent(llm)
newAgent := v2agent.NewReActAgent(llm, config)
```

---

## Compatibility Checklist

### ✅ Safe Changes (MINOR version)

**These DON'T break compatibility:**

- ✅ Add new functions
- ✅ Add new types
- ✅ Add new methods to interfaces (with default implementations)
- ✅ Add new fields to structs (if unexported or optional)
- ✅ Add new options to functional options
- ✅ Add new packages
- ✅ Fix bugs
- ✅ Improve performance
- ✅ Add documentation
- ✅ Deprecate (but keep) old functions

**Example:**
```go
// v0.2.0
package tools

func NewCalculator() *Calculator
func NewHTTPTool() *HTTPTool

// v0.3.0 - Add new tool (safe!)
func NewFileTool() *FileTool  // NEW - doesn't break anything
```

---

### ❌ Breaking Changes (MAJOR version)

**These REQUIRE major version bump (v1 → v2):**

- ❌ Remove public functions
- ❌ Remove public methods
- ❌ Change function signatures
- ❌ Change method signatures
- ❌ Remove fields from exported structs
- ❌ Change field types in exported structs
- ❌ Rename exported identifiers
- ❌ Change interface definitions (add/remove methods)
- ❌ Change package names
- ❌ Remove packages

**Example:**
```go
// v0.2.0
func (a *Agent) Run(input string) (*Response, error)

// v0.3.0 - BREAKING CHANGE!
func (a *Agent) Run(ctx context.Context, input string) (*Response, error)
//                ^^^^^^^^^^^^^^^^^^^ Added parameter = breaking!

// Requires v1.0.0 → v2.0.0
```

---

## Go Module Compatibility Rules

### The Go 1 Compatibility Promise

**Go standard library guarantees:**
> "Code that compiled under Go 1.0 will continue to compile under future Go 1.x releases"

**You should aim for similar guarantees:**

```go
// v0.x.x - Breaking changes OK (pre-release)
github.com/yashrahurikar23/goagents v0.2.0

// v1.x.x - Must maintain compatibility
github.com/yashrahurikar23/goagents v1.0.0
github.com/yashrahurikar23/goagents v1.1.0  // Can add features
github.com/yashrahurikar23/goagents v1.2.0  // Can add features
// All v1.x.x are compatible with each other

// v2.x.x - Breaking changes allowed
github.com/yashrahurikar23/goagents/v2 v2.0.0
```

### Minimum Version Selection

**Go modules use MVS (Minimum Version Selection):**

```go
// go.mod
module myapp

require (
    github.com/yashrahurikar23/goagents v0.2.0
)
```

**If you release v0.2.1 with bug fixes:**
- User can upgrade without code changes
- `go get -u` upgrades safely
- No breaking changes

**If you release v0.3.0 with breaking changes:**
- User must explicitly upgrade
- Must update code
- Should read CHANGELOG

---

## Documentation Best Practices

### 1. Godoc Comments

**Every exported identifier needs a comment:**

```go
// NewReActAgent creates a new ReAct (Reasoning + Acting) agent.
//
// The agent combines reasoning and acting capabilities using the ReAct pattern.
// It iteratively reasons about the task, decides which tool to use, executes
// the tool, and uses the result to inform the next reasoning step.
//
// Parameters:
//   - llm: The language model to use for reasoning
//   - opts: Optional configuration options
//
// Example:
//     llm := ollama.New()
//     agent := agent.NewReActAgent(llm,
//         agent.WithMaxIterations(10),
//         agent.WithVerbose(true),
//     )
//
// See also: NewFunctionAgent, NewConversationalAgent
func NewReActAgent(llm core.LLM, opts ...Option) *ReActAgent {
    // ...
}
```

### 2. Deprecation Notices

**Use proper deprecation format:**

```go
// Run executes the agent with the given input.
//
// Deprecated: Use RunContext instead for better cancellation support.
// This method will be removed in v1.0.0.
func (a *Agent) Run(input string) (*Response, error) {
    return a.RunContext(context.Background(), input)
}
```

### 3. Compatibility Notes

**Document version changes:**

```go
// Response represents the agent's output.
//
// Added in v0.1.0.
// Field 'Cost' added in v0.3.0.
// Field 'Metadata' added in v0.4.0.
type Response struct {
    Content  string            // v0.1.0+
    Tokens   int               // v0.1.0+
    Cost     float64           // v0.3.0+
    Metadata map[string]string // v0.4.0+
}
```

---

## Testing for Compatibility

### 1. Keep Old Tests

**Don't delete tests when refactoring:**

```go
// TestNewReActAgent_Legacy tests the original API
func TestNewReActAgent_Legacy(t *testing.T) {
    // Test v0.1.0 API still works
    agent := agent.NewReActAgent(llm)
    _, err := agent.Run(ctx, "question")
    require.NoError(t, err)
}

// TestNewReActAgent_WithOptions tests new options
func TestNewReActAgent_WithOptions(t *testing.T) {
    // Test v0.2.0 API
    agent := agent.NewReActAgent(llm,
        agent.WithMaxIterations(20),
    )
    _, err := agent.Run(ctx, "question")
    require.NoError(t, err)
}
```

### 2. Compatibility Test Suite

```go
// compat_test.go

// TestBackwardCompatibility ensures old code still works
func TestBackwardCompatibility(t *testing.T) {
    // Test that v0.1.0 style code still works in v0.2.0
    testV01API(t)
    testV02API(t)
}

func testV01API(t *testing.T) {
    // Original API from v0.1.0
    llm := ollama.New()
    agent := agent.NewReActAgent(llm)
    response, err := agent.Run(context.Background(), "test")
    require.NoError(t, err)
    require.NotNil(t, response)
}

func testV02API(t *testing.T) {
    // New API from v0.2.0
    llm := ollama.New()
    agent := agent.NewReActAgent(llm,
        agent.WithMaxIterations(10),
    )
    response, err := agent.Run(context.Background(), "test")
    require.NoError(t, err)
    require.NotNil(t, response)
}
```

---

## CHANGELOG.md

**Always maintain a changelog:**

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nothing yet

## [0.3.0] - 2025-10-15

### Added
- Streaming support via `RunStream()` method
- New `FileTool` for file operations
- Cost tracking with `WithCostTracking()` option

### Changed
- Improved error messages in all agents

### Deprecated
- `Run()` method (use `RunContext()` instead)

### Fixed
- Memory leak in conversational agent

## [0.2.0] - 2025-10-08

### Added
- HTTP tool for REST API calls
- Examples for HTTP tool usage

## [0.1.0] - 2025-10-08

### Added
- Initial release
- Three agent types: Function, ReAct, Conversational
- OpenAI and Ollama LLM providers
- Calculator tool
```

---

## Summary

### Question 1: Remove `New*` Prefix?

**❌ Don't remove it**
- Goes against Go conventions
- Confuses Go developers
- Reduces clarity
- Current API is idiomatic

**Verdict:** Keep `NewReActAgent`, `NewCalculator`, etc.

---

### Question 2: Avoid Breaking Changes?

**✅ Key strategies:**
1. Use interfaces, not concrete types
2. Functional options pattern
3. Unexported struct fields
4. Deprecate, don't remove
5. Extend, don't replace
6. Version packages for major changes

**✅ Safe changes:**
- Add functions, types, methods
- Add optional fields
- Add options
- Fix bugs

**❌ Breaking changes:**
- Remove/rename exports
- Change signatures
- Change types
- Remove fields

**Version carefully:**
- v0.x.x: Breaking changes OK
- v1.x.x: Maintain compatibility
- v2.x.x: Breaking changes allowed

---

### Question 3: Advanced Features?

**See:** `docs/ADVANCED_FEATURES_ROADMAP.md`

**Top 3 priorities:**
1. **RAG** (v0.5.0) - Vector DBs, embeddings, document loaders
2. **Streaming** (v0.3.0) - Real-time token output
3. **Observability** (v0.4.0) - Tracing, cost tracking, debugging

**These make GoAgents competitive with LangChain/LlamaIndex!**

---

## Next Steps

1. ✅ Keep current API (don't rename `New*`)
2. ✅ Follow compatibility guidelines going forward
3. 🔄 Start planning v0.3.0 features (streaming, more tools)
4. 🔄 Build RAG support for v0.5.0
5. 🔄 Add comprehensive observability

**Your API design is already solid!** 🎯
