# GoAgent SDK/API Architecture

**Framework Name:** GoAgent (or alternative: AgenticGo, GoFlow, GoLlama)

**Goal:** Provide a Go-native AI agent framework that users can install as a package and use via clean, idiomatic APIs similar to LlamaIndex.

---

## 1. Package Structure & Import Paths

### Base Import Path
```
github.com/username/goagent
```

### Core Package Organization

```
goagent/
├── agent/              # Core agent implementations
│   ├── function.go     # FunctionAgent
│   ├── react.go        # ReActAgent
│   ├── workflow.go     # WorkflowAgent
│   └── types.go        # Agent interfaces
├── llm/                # LLM provider integrations
│   ├── openai/         # OpenAI client
│   ├── anthropic/      # Anthropic client
│   ├── ollama/         # Ollama local models
│   └── base.go         # LLM interface
├── tools/              # Tool implementations
│   ├── base.go         # Tool interface
│   ├── query_engine.go # QueryEngineTool
│   ├── function.go     # FunctionTool
│   └── builtin/        # Built-in tools
│       ├── search.go
│       ├── web.go
│       └── database.go
├── ingestion/          # Data loading & processing
│   ├── pipeline.go     # IngestionPipeline
│   ├── readers/        # Document readers
│   │   ├── pdf.go
│   │   ├── web.go
│   │   └── base.go
│   ├── parsers/        # Node parsers
│   │   ├── sentence.go
│   │   └── token.go
│   └── extractors/     # Metadata extractors
│       └── title.go
├── index/              # Index implementations
│   ├── vector.go       # VectorStoreIndex
│   ├── summary.go      # SummaryIndex
│   └── types.go        # Index interfaces
├── storage/            # Storage backends
│   ├── vectorstore/    # Vector databases
│   │   ├── qdrant/
│   │   ├── pinecone/
│   │   └── chroma/
│   ├── docstore/       # Document stores
│   │   ├── redis/
│   │   └── mongodb/
│   └── cache/          # Caching layer
├── retrieval/          # Retrieval components
│   ├── retriever.go    # Base retriever
│   ├── vector.go       # VectorRetriever
│   └── hybrid.go       # HybridRetriever
├── query/              # Query engines
│   ├── engine.go       # QueryEngine interface
│   ├── retriever.go    # RetrieverQueryEngine
│   └── router.go       # RouterQueryEngine
├── embeddings/         # Embedding models
│   ├── openai/
│   ├── huggingface/
│   └── base.go
├── workflow/           # Workflow orchestration
│   ├── workflow.go     # Workflow engine
│   ├── context.go      # Workflow context
│   └── events.go       # Event system
└── examples/           # Usage examples
    ├── quickstart/
    ├── agents/
    └── ingestion/
```

---

## 2. API Design Patterns

### Philosophy
- **Idiomatic Go:** Builder patterns, functional options, explicit error handling
- **Type Safety:** Interfaces for extensibility, structs for concrete types
- **Composability:** Small, focused packages that work together
- **Zero Magic:** Explicit initialization, no global state

### Core Interfaces

```go
package goagent

// Agent is the core interface for all agent types
type Agent interface {
    Run(ctx context.Context, input string, opts ...RunOption) (*Response, error)
    RunStream(ctx context.Context, input string, opts ...RunOption) (<-chan Event, error)
    AddTool(tool Tool) error
    Reset(ctx context.Context) error
}

// Tool interface for agent tools
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, input string) (*ToolOutput, error)
    Schema() *ToolSchema
}

// LLM interface for language model providers
type LLM interface {
    Complete(ctx context.Context, prompt string, opts ...LLMOption) (*Completion, error)
    Chat(ctx context.Context, messages []Message, opts ...LLMOption) (*ChatResponse, error)
    Stream(ctx context.Context, prompt string, opts ...LLMOption) (<-chan Token, error)
}

// Index interface for data indexing
type Index interface {
    AsQueryEngine(opts ...QueryEngineOption) QueryEngine
    AsRetriever(opts ...RetrieverOption) Retriever
    Insert(ctx context.Context, nodes []Node) error
    Delete(ctx context.Context, refDocID string) error
}

// QueryEngine interface for querying indexed data
type QueryEngine interface {
    Query(ctx context.Context, query string, opts ...QueryOption) (*QueryResponse, error)
}

// Retriever interface for retrieving relevant nodes
type Retriever interface {
    Retrieve(ctx context.Context, query string, opts ...RetrieveOption) ([]Node, error)
}
```

---

## 3. Usage Examples (User-Facing API)

### Example 1: Simple Agent with Tools

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/username/goagent/agent"
    "github.com/username/goagent/llm/openai"
    "github.com/username/goagent/tools/builtin"
)

func main() {
    ctx := context.Background()
    
    // Initialize LLM
    llm, err := openai.New(openai.WithAPIKey("sk-..."))
    if err != nil {
        panic(err)
    }
    
    // Create agent
    agent, err := agent.NewFunctionAgent(
        agent.WithLLM(llm),
        agent.WithTools(
            builtin.NewSearchTool(),
            builtin.NewWebScraperTool(),
        ),
    )
    if err != nil {
        panic(err)
    }
    
    // Run query
    response, err := agent.Run(ctx, "What's the latest news on AI?")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(response.Content)
}
```

### Example 2: RAG with Query Engine

```go
package main

import (
    "context"
    
    "github.com/username/goagent/ingestion"
    "github.com/username/goagent/ingestion/readers"
    "github.com/username/goagent/index"
    "github.com/username/goagent/storage/vectorstore/qdrant"
    "github.com/username/goagent/embeddings/openai"
)

func main() {
    ctx := context.Background()
    
    // Setup vector store
    vectorStore, _ := qdrant.New(
        qdrant.WithURL("localhost:6333"),
        qdrant.WithCollection("documents"),
    )
    
    // Setup embeddings
    embedModel, _ := openai.NewEmbedding(
        openai.WithModel("text-embedding-3-small"),
    )
    
    // Create ingestion pipeline
    pipeline := ingestion.NewPipeline(
        ingestion.WithReader(readers.NewPDFReader()),
        ingestion.WithParser(parsers.NewSentenceSplitter(
            parsers.WithChunkSize(512),
        )),
        ingestion.WithEmbedding(embedModel),
        ingestion.WithVectorStore(vectorStore),
    )
    
    // Ingest documents
    nodes, err := pipeline.Run(ctx, []string{"./data/*.pdf"})
    if err != nil {
        panic(err)
    }
    
    // Create index
    idx, _ := index.NewVectorStoreIndex(
        index.WithVectorStore(vectorStore),
        index.WithEmbedding(embedModel),
    )
    
    // Create query engine
    queryEngine := idx.AsQueryEngine(
        query.WithSimilarityTopK(3),
    )
    
    // Query
    response, _ := queryEngine.Query(ctx, "What are the main topics?")
    fmt.Println(response.Content)
}
```

### Example 3: Multi-Agent Workflow

```go
package main

import (
    "context"
    
    "github.com/username/goagent/agent"
    "github.com/username/goagent/workflow"
)

func main() {
    ctx := context.Background()
    
    // Create specialized agents
    researchAgent, _ := agent.NewFunctionAgent(
        agent.WithName("researcher"),
        agent.WithTools(searchTool, webTool),
    )
    
    writerAgent, _ := agent.NewFunctionAgent(
        agent.WithName("writer"),
        agent.WithTools(documentTool),
    )
    
    // Create workflow
    wf := workflow.New()
    
    // Define workflow steps
    wf.AddStep("research", func(ctx workflow.Context) error {
        result, err := researchAgent.Run(ctx, ctx.Input())
        if err != nil {
            return err
        }
        ctx.Set("research_data", result.Content)
        return ctx.SendEvent("write", result.Content)
    })
    
    wf.AddStep("write", func(ctx workflow.Context) error {
        research := ctx.Get("research_data").(string)
        result, err := writerAgent.Run(ctx, "Write report: "+research)
        if err != nil {
            return err
        }
        ctx.SetOutput(result.Content)
        return nil
    })
    
    // Execute workflow
    output, err := wf.Run(ctx, "Research and write about quantum computing")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(output)
}
```

---

## 4. Versioning Strategy

### Semantic Versioning (SemVer)
- **v0.x.y** - Pre-1.0: Breaking changes expected
- **v1.x.y** - Stable: Backwards compatibility guaranteed
- **v2.x.y** - Major version: Breaking changes with migration guide

### Module Versioning
```
github.com/username/goagent       → v1.x.x (core)
github.com/username/goagent/v2    → v2.x.x (major version)
```

### Compatibility Promise
- **Go Version:** Support last 2 major Go versions (currently 1.22, 1.23)
- **API Stability:** No breaking changes in minor/patch releases
- **Deprecation:** 2-release deprecation cycle with warnings

---

## 5. Documentation Strategy

### 1. **README.md** (Repository Root)
```markdown
# GoAgent - AI Agent Framework in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/username/goagent.svg)]
[![Go Report Card](https://goreportcard.com/badge/github.com/username/goagent)]
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)]

**GoAgent** is a production-ready AI agent framework built in Go, inspired by LlamaIndex.

## Features
- 🤖 Multiple agent types (Function, ReAct, Workflow)
- 🔧 Rich tool ecosystem (search, web, database, APIs)
- 📚 RAG with vector databases (Qdrant, Pinecone, Chroma)
- ⚡ High performance (3-6x faster than Python)
- 🔒 Type-safe APIs with compile-time validation
- 🚀 Production-ready with observability built-in

## Quick Start
```go
import "github.com/username/goagent/agent"

agent, _ := agent.NewFunctionAgent(
    agent.WithLLM(llm),
    agent.WithTools(searchTool),
)

response, _ := agent.Run(ctx, "What's the weather?")
```

## Documentation
- [Getting Started Guide](./docs/getting-started.md)
- [API Reference](https://pkg.go.dev/github.com/username/goagent)
- [Examples](./examples/)
- [Contributing Guide](./CONTRIBUTING.md)
```

### 2. **GoDoc Comments** (In-Code Documentation)
```go
// Package agent provides core agent implementations for building AI agents.
//
// Agents are the primary way to interact with language models and tools.
// This package includes:
//   - FunctionAgent: Simple function-calling agent
//   - ReActAgent: Reasoning + Acting agent
//   - WorkflowAgent: Multi-step workflow orchestration
//
// Example usage:
//
//	agent, err := agent.NewFunctionAgent(
//	    agent.WithLLM(llm),
//	    agent.WithTools(tool1, tool2),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	response, err := agent.Run(ctx, "your query")
package agent
```

### 3. **Documentation Site** (docs.goagent.dev)
Structure:
```
docs/
├── getting-started/
│   ├── installation.md
│   ├── quickstart.md
│   └── core-concepts.md
├── guides/
│   ├── agents/
│   ├── tools/
│   ├── rag/
│   └── workflows/
├── api-reference/
│   └── (auto-generated from GoDoc)
├── examples/
│   ├── simple-agent.md
│   ├── rag-chatbot.md
│   └── multi-agent.md
└── deployment/
    ├── docker.md
    ├── kubernetes.md
    └── monitoring.md
```

### 4. **Interactive Examples** (GitHub Repository)
```
examples/
├── 01-hello-agent/
│   └── main.go
├── 02-rag-query/
│   └── main.go
├── 03-tool-creation/
│   └── main.go
├── 04-multi-agent/
│   └── main.go
└── 05-production/
    ├── main.go
    └── docker-compose.yml
```

---

## 6. Installation & Distribution

### Go Module Installation
```bash
# Install latest version
go get github.com/username/goagent@latest

# Install specific version
go get github.com/username/goagent@v1.2.3

# Install with specific integrations
go get github.com/username/goagent/storage/vectorstore/qdrant
go get github.com/username/goagent/llm/anthropic
```

### Docker Images
```bash
# Official images
docker pull goagent/goagent:latest
docker pull goagent/goagent:v1.2.3
docker pull goagent/goagent:v1.2.3-alpine
```

### Binary Releases (GitHub Releases)
For CLI tools:
```bash
# Via install script
curl -sSL https://goagent.dev/install.sh | sh

# Via Homebrew
brew install goagent/tap/goagent

# Via Go install
go install github.com/username/goagent/cmd/goagent@latest
```

---

## 7. SDK API Exposure Patterns

### Pattern 1: **Functional Options** (Primary)
```go
// Flexible, extensible, idiomatic Go
agent, err := agent.NewFunctionAgent(
    agent.WithLLM(llm),
    agent.WithTools(tool1, tool2),
    agent.WithTimeout(30*time.Second),
    agent.WithRetries(3),
)
```

### Pattern 2: **Builder Pattern** (Alternative)
```go
// Fluent API for complex configurations
agent := agent.NewBuilder().
    LLM(llm).
    Tools(tool1, tool2).
    Timeout(30*time.Second).
    Build()
```

### Pattern 3: **Context-Based Configuration**
```go
// Runtime configuration via context
ctx := context.WithValue(ctx, agent.VerboseKey, true)
response, err := agent.Run(ctx, input)
```

### Pattern 4: **Type-Safe Tool Definition**
```go
// Define custom tool with schema
type MyTool struct {
    name string
}

func (t *MyTool) Name() string { return t.name }
func (t *MyTool) Description() string { return "My custom tool" }
func (t *MyTool) Schema() *ToolSchema {
    return &ToolSchema{
        Parameters: []Parameter{
            {Name: "query", Type: "string", Required: true},
        },
    }
}
func (t *MyTool) Execute(ctx context.Context, input string) (*ToolOutput, error) {
    // Implementation
}
```

---

## 8. Testing & Quality Assurance

### Unit Tests
```go
// All public APIs have tests
func TestFunctionAgent_Run(t *testing.T) {
    mockLLM := &MockLLM{}
    agent, _ := agent.NewFunctionAgent(
        agent.WithLLM(mockLLM),
    )
    
    response, err := agent.Run(context.Background(), "test")
    require.NoError(t, err)
    assert.Equal(t, "expected", response.Content)
}
```

### Integration Tests
```bash
# Test with real services via docker-compose
go test -tags=integration ./...
```

### Example Tests
```bash
# All examples must compile and run
cd examples/01-hello-agent && go run main.go
```

---

## 9. Observability & Monitoring

### Built-in Logging
```go
import "github.com/username/goagent/observability"

// Structured logging with slog
agent, _ := agent.NewFunctionAgent(
    agent.WithLogger(slog.Default()),
    agent.WithLogLevel(slog.LevelDebug),
)
```

### Metrics (Prometheus)
```go
// Automatic metric collection
import "github.com/username/goagent/observability/metrics"

metrics.RecordAgentLatency(duration)
metrics.IncrementToolCalls(toolName)
```

### Tracing (OpenTelemetry)
```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("goagent")
ctx, span := tracer.Start(ctx, "agent.Run")
defer span.End()

response, err := agent.Run(ctx, input)
```

---

## 10. Migration from LlamaIndex (Python)

### Side-by-Side Comparison

| LlamaIndex (Python) | GoAgent (Go) |
|---------------------|--------------|
| `from llama_index.core.agent import FunctionAgent` | `import "github.com/username/goagent/agent"` |
| `agent = FunctionAgent(tools=tools, llm=llm)` | `agent, err := agent.NewFunctionAgent(...)` |
| `response = agent.chat("query")` | `response, err := agent.Run(ctx, "query")` |
| `index = VectorStoreIndex.from_documents(docs)` | `idx, err := index.NewVectorStoreIndex(...)` |
| `query_engine = index.as_query_engine()` | `queryEngine := idx.AsQueryEngine()` |

### Migration Guide Document
```markdown
# Migrating from LlamaIndex to GoAgent

## Key Differences
1. **Error Handling:** Go uses explicit error returns
2. **Context:** Pass context.Context for cancellation/timeouts
3. **Configuration:** Functional options instead of kwargs
4. **Streaming:** Go channels instead of iterators

## Pattern Translations
[Detailed examples of common patterns]
```

---

## 11. Community & Contribution

### Repository Structure
```
.github/
├── workflows/
│   ├── ci.yml
│   ├── release.yml
│   └── docs.yml
├── ISSUE_TEMPLATE/
├── PULL_REQUEST_TEMPLATE.md
└── CODEOWNERS
```

### Contribution Flow
1. Fork repository
2. Create feature branch
3. Write tests + documentation
4. Submit PR with description
5. Automated CI runs tests
6. Code review + merge

---

## Summary

**GoAgent** provides a production-ready, Go-native AI agent framework with:
- **Clean API:** Idiomatic Go patterns (functional options, interfaces, explicit errors)
- **Easy Installation:** `go get github.com/username/goagent`
- **Comprehensive Docs:** GoDoc + website + examples
- **Type Safety:** Compile-time validation, no runtime surprises
- **Performance:** 3-6x faster than Python equivalents
- **Production Ready:** Logging, metrics, tracing built-in
- **LlamaIndex Compatibility:** Similar concepts, Go-native implementation

**Next Steps:**
1. Finalize framework name (GoAgent vs alternatives)
2. Create GitHub organization + repository
3. Implement core packages (agent, llm, tools)
4. Publish v0.1.0 with basic functionality
5. Build community via docs + examples
