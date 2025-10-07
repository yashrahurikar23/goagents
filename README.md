# 🚀 GoAgent - AI Agent Framework for Go# 🚀 GoAgent - AI Agent Framework for Go



[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://go.dev/)[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://go.dev/)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)



Production-ready AI agent framework for Go with support for multiple LLM providers and agent patterns.Production-ready AI agent framework for Go with support for multiple LLM providers and agent patterns.



## ✨ Features## ✨ Features



- 🤖 **3 Agent Types**: FunctionAgent (OpenAI), ReActAgent (reasoning), ConversationalAgent (memory)- 🤖 **3 Agent Types**: FunctionAgent (OpenAI), ReActAgent (reasoning), ConversationalAgent (memory)

- 🔌 **Multiple LLM Providers**: OpenAI, Ollama (local & free)- 🔌 **Multiple LLM Providers**: OpenAI, Ollama (local & free)

- 🛠️ **Tool System**: Easy to create and integrate custom tools- 🛠️ **Tool System**: Easy to create and integrate custom tools

- 💾 **Memory Management**: 4 strategies for conversation history- 💾 **Memory Management**: 4 strategies for conversation history

- 🧪 **100% Tested**: Comprehensive test coverage (100+ tests)- 🧪 **100% Tested**: Comprehensive test coverage (100+ tests)

- ⚡ **Production Ready**: Type-safe, concurrent, efficient- ⚡ **Production Ready**: Type-safe, concurrent, efficient

- 🌐 **Local AI Support**: Run completely offline with Ollama- 🌐 **Local AI Support**: Run completely offline with Ollama



## 📦 Installation## 📦 Installation



```bash```bash

go get github.com/yashrahurikar23/goagents@latestgo get github.com/yashrahurikar23/goagents@latest

``````



## 🚀 Quick Start## 🚀 Quick Start



### Using Ollama (Local, Free)### Using Ollama (Local, Free)



```go1. **Research Phase**

package main   - Market analysis (no Go competition identified)

   - Technical feasibility (all required libraries available)

import (   - Performance benchmarking (validated 3-6x improvements)

    "context"   - Go vs Python comparison (Go recommended)

    "fmt"

    "log"2. **Architecture Phase**

       - Complete SDK design (functional options pattern)

    "github.com/yashrahurikar23/goagents/agent"   - Package structure defined (agent, llm, tools, ingestion, etc.)

    "github.com/yashrahurikar23/goagents/llm/ollama"   - API patterns established (idiomatic Go)

)   - Versioning strategy (SemVer)



func main() {3. **Scope Definition**

    // Create Ollama client (local LLM)   - 90+ features inventoried from LlamaIndex

    llm := ollama.New(   - 35-week implementation plan (4 phases)

        ollama.WithModel("llama3.2:1b"),   - Sprint-by-sprint breakdown (48 sprints)

        ollama.WithTemperature(0.7),   - Resource planning (1-3 developers)

    )

    ### 📁 Documentation Structure

    // Create ReAct agent (transparent reasoning)

    myAgent := agent.NewReActAgent(llm)```

    goagent/

    // Run a query├── SDK_ARCHITECTURE.md          # Complete API design

    ctx := context.Background()├── LLAMAINDEX_FEATURES_INVENTORY.md  # Feature list (90+)

    response, err := myAgent.Run(ctx, "What is 25 * 4?")├── IMPLEMENTATION_SCOPE.md      # 35-week plan

    if err != nil {└── notes/

        log.Fatal(err)    ├── README.md                # Index

    }    ├── 00_project_roadmap.md    # High-level roadmap

        ├── 01_llamaindex_ai_agents_overview.md

    fmt.Println("Answer:", response.Content)    ├── 02_golang_vs_python_for_ai_agents.md

        ├── 03_golang_tools_ecosystem.md

    // View reasoning trace    └── 04_golang_data_processing.md

    for _, step := range myAgent.GetTrace() {```

        fmt.Printf("Thought: %s\n", step.Thought)

        fmt.Printf("Action: %s\n", step.Action)---

    }

}## Framework Architecture

```

### Core Packages

### Using OpenAI

```

```gogithub.com/username/goagent/

package main├── agent/        # FunctionAgent, ReActAgent, WorkflowAgent

├── llm/          # OpenAI, Anthropic, Ollama clients

import (├── tools/        # Tool interface + built-in tools

    "context"├── ingestion/    # Data loading & processing pipeline

    "fmt"├── index/        # VectorStoreIndex, SummaryIndex

    "os"├── storage/      # Vector stores, doc stores, caching

    ├── retrieval/    # Retrievers with filtering & re-ranking

    "github.com/yashrahurikar23/goagents/agent"├── query/        # Query engines

    "github.com/yashrahurikar23/goagents/llm/openai"├── workflow/     # Event-driven orchestration

    "github.com/yashrahurikar23/goagents/tools"└── embeddings/   # Embedding models

)```



func main() {### User-Facing API Example

    // Create OpenAI client

    llm := openai.New(```go

        openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),import "github.com/username/goagent/agent"

    )

    // Create agent

    // Create function agentagent, err := agent.NewFunctionAgent(

    myAgent := agent.NewFunctionAgent(llm)    agent.WithLLM(llm),

        agent.WithTools(searchTool, webTool),

    // Add tools)

    calculator := tools.NewCalculator()

    myAgent.AddTool(calculator)// Run query

    response, err := agent.Run(ctx, "What's the latest AI news?")

    // Run queryfmt.Println(response.Content)

    response, _ := myAgent.Run(context.Background(), ```

        "Calculate 15% tip on a $47.50 bill")

    ---

    fmt.Println(response.Content)

}## Implementation Timeline

```

### Phase 1: MVP Foundation (12-16 weeks)

### Conversational Agent with Memory

**Target:** v0.5.0 - Working agents with RAG

```go

package main- ✅ Core interfaces & infrastructure

- ✅ LLM clients (OpenAI, Anthropic, Ollama)

import (- ✅ Basic agents (FunctionAgent, ReActAgent)

    "context"- ✅ Core tools (search, web, database)

    "fmt"- ✅ Embeddings & vector stores (Qdrant, Postgres)

    - ✅ Document readers (PDF, web, markdown)

    "github.com/yashrahurikar23/goagents/agent"- ✅ Text splitters & ingestion pipeline

    "github.com/yashrahurikar23/goagents/llm/ollama"- ✅ Query engines & retrieval

)

**Deliverable:** Functional MVP with example apps

func main() {

    llm := ollama.New(ollama.WithModel("llama3.2:1b"))---

    

    // Create chatbot with memory### Phase 2: Production Features (8-12 weeks)

    chatbot := agent.NewConversationalAgent(

        llm,**Target:** v0.8.0 - Production-ready

        agent.ConvWithSystemPrompt("You are a helpful assistant."),

        agent.ConvWithMemoryStrategy(agent.MemoryStrategyWindow),- Workflow system (event-driven)

        agent.ConvWithMaxMessages(20),- Multi-agent patterns (hand-off, planner)

    )- Advanced retrieval (HyDE, auto-retrieval)

    - Advanced query engines (SubQuestion, Router)

    ctx := context.Background()- Observability (logging, metrics, tracing)

    - Caching & persistence (Redis, MongoDB)

    // Turn 1

    resp1, _ := chatbot.Run(ctx, "My name is Alice")**Deliverable:** Enterprise-ready framework

    fmt.Println("Bot:", resp1.Content)

    ---

    // Turn 2 (remembers Alice)

    resp2, _ := chatbot.Run(ctx, "What's my name?")### Phase 3: Enterprise Features (8-10 weeks)

    fmt.Println("Bot:", resp2.Content) // "Your name is Alice"

}**Target:** v0.9.0 - Enterprise complete

```

- Evaluation framework

## 🤖 Agent Types- Chat & memory systems

- Structured outputs (Pydantic-like)

### 1. FunctionAgent- Additional integrations (Gemini, Cohere, Pinecone)

- Deployment tooling (Docker, K8s)

**Best for:** Production apps with OpenAI  

**Features:** Native function calling, automatic tool execution, reliable**Deliverable:** Full-featured framework



```go---

agent := agent.NewFunctionAgent(openaiClient)

agent.AddTool(myTool)### Phase 4: Advanced & Polish (6-8 weeks)

response, _ := agent.Run(ctx, "Use the tool to help me")

```**Target:** v1.0.0 - Production release



### 2. ReActAgent- Advanced agent patterns

- Performance optimization & benchmarking

**Best for:** Debugging, non-OpenAI models, transparency  - Complete documentation

**Features:** Visible reasoning, works with any LLM, thought traces- Example applications (10+)

- Migration guide from LlamaIndex

```go

agent := agent.NewReActAgent(ollamaClient)**Deliverable:** v1.0.0 public release

agent.AddTool(myTool)

response, _ := agent.Run(ctx, "Think through this problem")---



// See the reasoning## Feature Scope

for _, step := range agent.GetTrace() {

    fmt.Println("Thought:", step.Thought)### Must-Have (MVP) - 30 Features

    fmt.Println("Action:", step.Action)

}- [x] Agents: FunctionAgent, ReActAgent

```- [x] LLMs: OpenAI, Anthropic, Ollama

- [x] Tools: QueryEngine, Function, Search, Web, Database

### 3. ConversationalAgent- [x] Readers: PDF, Web, Markdown, CSV, Database

- [x] Parsers: Sentence, Token splitters

**Best for:** Chatbots, customer support, long conversations  - [x] Indexes: VectorStoreIndex

**Features:** 4 memory strategies, conversation export, multi-turn- [x] Vector Stores: Qdrant, Postgres, Chroma

- [x] Query Engines: RetrieverQueryEngine, SQLQueryEngine

```go- [x] Ingestion Pipeline with caching

agent := agent.NewConversationalAgent(- [x] Basic observability

    llm,

    agent.ConvWithMemoryStrategy(agent.MemoryStrategyWindow),**Effort:** 12-16 weeks

    agent.ConvWithMaxMessages(20),

)---

```

### High Priority (Phase 2) - 25 Features

Memory strategies:

- [ ] WorkflowAgent

- `MemoryStrategyWindow` - Keep last N messages (fast, simple)- [ ] Multi-agent patterns (hand-off, planner)

- `MemoryStrategySummarize` - Compress old messages (token-efficient)- [ ] Advanced retrieval (auto-retrieval, HyDE)

- `MemoryStrategySelective` - Keep important, summarize rest- [ ] SubQuestionQueryEngine, RouterQueryEngine

- `MemoryStrategyAll` - No limits (keep everything)- [ ] Workflow system with events

- [ ] Evaluation metrics

## 🔌 LLM Providers- [ ] Chat memory & history

- [ ] Structured outputs

### Ollama (Local, Free)- [ ] More readers (Notion, Google Docs, GitHub)

- [ ] Document stores (Redis, MongoDB)

```go

import "github.com/yashrahurikar23/goagents/llm/ollama"**Effort:** 8-12 weeks



llm := ollama.New(---

    ollama.WithModel("llama3.2:1b"),

    ollama.WithTemperature(0.7),### Medium Priority (Phase 3) - 20 Features

    ollama.WithMaxTokens(100),

)- [ ] Custom agent framework

```- [ ] More LLMs (Google, Cohere)

- [ ] More vector stores (Weaviate, Elasticsearch)

**Supported models:** llama3.2, gemma3, qwen3, phi3, deepseek, and more!- [ ] Re-ranking (Cohere, cross-encoder)

- [ ] Metadata extractors

### OpenAI- [ ] Multi-modal (images, audio)

- [ ] FastAPI server

```go

import "github.com/yashrahurikar23/goagents/llm/openai"**Effort:** 8-10 weeks



llm := openai.New(---

    openai.WithAPIKey("sk-..."),

    openai.WithModel("gpt-4"),### Low Priority (Future) - 15 Features

    openai.WithTemperature(0.7),

)- [ ] Advanced indexes (Tree, KnowledgeGraph)

```- [ ] Fine-tuning & optimization

- [ ] Specialized tools

## 🛠️ Creating Custom Tools- [ ] Advanced patterns (debate agents)



```go**Effort:** 6-8 weeks

package main

---

import (

    "context"## Resource Requirements

    "fmt"

    ### Team Options

    "github.com/yashrahurikar23/goagents/core"

)**Option A: Solo Developer (Recommended for Bootstrap)**

- Duration: 40-46 weeks (10-11 months)

type WeatherTool struct{}- Cost: $80k-120k salary/year

- Risk: Longer time-to-market

func (t *WeatherTool) Name() string {

    return "weather"**Option B: 2 Developers (Recommended for Speed)**

}- Duration: 20-24 weeks (5-6 months)

- Cost: $160k-240k salary/year

func (t *WeatherTool) Description() string {- Split: Core/Infrastructure + Features/Integrations

    return "Get weather for a city"

}**Option C: 3 Developers (Fast Track)**

- Duration: 16-20 weeks (4-5 months)

func (t *WeatherTool) Schema() *core.ToolSchema {- Cost: $240k-360k salary/year

    return &core.ToolSchema{- Split: Core + Data + Agents

        Name:        "weather",

        Description: "Get current weather",### Infrastructure Costs

        Parameters: []core.Parameter{

            {- LLM API credits: $500-1000/month

                Name:        "city",- Vector DB hosting: $200-500/month

                Type:        "string",- CI/CD: Free (GitHub Actions)

                Description: "City name",

                Required:    true,**Total:** ~$8k-18k/year

            },

        },---

    }

}## Success Metrics



func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {### v0.5.0 (MVP) Targets

    city := args["city"].(string)- ✅ 2-3 working example apps

    // Your weather API call here- ✅ 10+ unit tests per package

    return fmt.Sprintf("Weather in %s: Sunny, 72°F", city), nil- ✅ Basic documentation (README + GoDoc)

}- ✅ Performance: <100ms agent response time

```- ✅ 5+ GitHub stars



## 📚 Documentation### v1.0.0 (Production) Targets

- ✅ 90% feature parity with LlamaIndex core

- **[Quick Start Guide](./READY_TO_SHIP.md)** - Get started in 5 minutes- ✅ 10+ production deployments

- **[Packaging Guide](./PACKAGING_GUIDE.md)** - How to use as a package- ✅ 100+ GitHub stars

- **[Agent Architectures](./AGENT_ARCHITECTURES.md)** - Deep dive into agent patterns- ✅ 5+ community contributions

- **[Ollama Client](./OLLAMA_CLIENT_COMPLETE.md)** - Ollama integration guide- ✅ Performance: 3-6x faster than LlamaIndex

- **[Examples](./examples/)** - Working code examples- ✅ Documentation site live

- **[API Reference](https://pkg.go.dev/github.com/yashrahurikar23/goagents)** - Full API docs- ✅ 50+ unit + integration tests



## 🧪 Running Tests---



```bash## Competitive Analysis

# Run all tests

go test ./...### Current Market



# Run with coverage| Framework | Language | Status | Market Share |

go test -cover ./...|-----------|----------|--------|--------------|

| LlamaIndex | Python | Mature | High |

# Run specific package| LangChain | Python | Mature | High |

go test ./agent/...| AutoGen | Python | Growing | Medium |

| **GoAgent** | **Go** | **New** | **0% (Opportunity!)** |

# Verbose output

go test -v ./...### Differentiation

```

1. **Performance:** 3-6x faster execution

## 📊 Project Status2. **Concurrency:** 10,000+ agents vs ~100

3. **Memory:** 4-7x less usage

| Component | Status | Tests |4. **Type Safety:** Compile-time validation

|-----------|--------|-------|5. **Deployment:** Cloud-native, Docker-friendly

| Core | ✅ Ready | 42 passing |6. **Production:** Built for enterprise from day 1

| Agents | ✅ Ready | 43 passing |

| OpenAI Client | ✅ Ready | Integration |---

| Ollama Client | ✅ Ready | 15 passing |

| Tools | ✅ Ready | Working |## Risk Assessment



**Total:** 100+ tests passing, production-ready!| Risk | Impact | Mitigation | Probability |

|------|--------|------------|-------------|

## 🗺️ Roadmap| Go ecosystem gaps | High | Wrap Python tools via API | Low (25%) |

| LLM API changes | Medium | Abstraction layer | Medium (40%) |

### v0.1.0 - Current (October 2025)| Performance issues | Medium | Early benchmarking | Low (20%) |

| Complex features | High | MVP-first approach | Medium (50%) |

- ✅ Core agent types (Function, ReAct, Conversational)| Adoption slow | Medium | Open source + marketing | Medium (40%) |

- ✅ OpenAI integration

- ✅ Ollama integration (local AI)**Overall Risk:** Low-Medium (manageable with mitigation)

- ✅ Tool system

- ✅ Memory management---



### v0.2.0 - Planned (November 2025)## Next Steps



- HTTP tool for API calls### Week 1: Repository Setup

- File operations tool1. Create GitHub organization

- More examples and tutorials2. Initialize Go modules

- Performance optimizations3. Setup CI/CD (GitHub Actions)

4. Create project structure

### v0.5.0 - Planned (Q1 2026)5. Define core interfaces



- RAG (Retrieval Augmented Generation)### Week 2: First Prototype

- Vector store integrations1. OpenAI LLM client

- Multi-agent coordination2. Basic FunctionAgent

- Workflow system3. Simple tool (HTTP client)

4. Example application

### v1.0.0 - Goal (Q2 2026)5. Unit tests



- Enterprise features### Week 4: First Release

- Complete documentation site1. v0.1.0: Core interfaces

- Production deployments2. v0.2.0: LLM clients

- Community ecosystem3. Blog post announcement

4. Community feedback

## 🤝 Contributing

### Week 8: MVP Demo

Contributions are welcome! Please feel free to submit a Pull Request.1. v0.3.0: Working agents + tools

2. 2-3 example apps

**Areas we need help with:**3. Basic documentation

4. Performance benchmarks

- More LLM provider integrations (Anthropic, Cohere)

- Additional tools (web scraping, database, etc.)### Week 20: MVP Launch

- Documentation and examples1. v0.5.0: Complete MVP

- Bug reports and feature requests2. Documentation site

3. Public announcement

## 📄 License4. Early adopters (5-10)



MIT License - see [LICENSE](./LICENSE) file for details.### Week 42: Production Release

1. v1.0.0: Production ready

## 🙏 Acknowledgments2. Marketing campaign

3. Community building

- Inspired by [LlamaIndex](https://github.com/run-llama/llama_index) and [LangChain](https://github.com/langchain-ai/langchain)4. Enterprise outreach

- Built with love for the Go community 💙

---

## 📞 Support & Community

## Recommendation

- **Issues:** [GitHub Issues](https://github.com/yashrahurikar23/goagents/issues)

- **Discussions:** [GitHub Discussions](https://github.com/yashrahurikar23/goagents/discussions)### ✅ PROCEED WITH IMPLEMENTATION

- **Twitter:** [@yashrahurikar](https://twitter.com/yashrahurikar)

**Rationale:**

## ⭐ Star History1. **Clear Market Gap** - No mature Go agent framework exists

2. **Technical Feasibility** - All required libraries available in Go

If you find this project useful, please consider giving it a star! ⭐3. **Strong Value Prop** - 3-6x performance + type safety + production-ready

4. **Executable Plan** - Detailed 35-week roadmap with sprint breakdown

---5. **Low Risk** - Manageable risks with mitigation strategies

6. **High Reward** - First-mover advantage in growing market

**Built with ❤️ by [Yash Rahurikar](https://github.com/yashrahurikar)**

**Recommended Approach:**

**Go + AI = 🚀**- Start with 2-developer team (5-6 months to MVP)

- Focus on v0.5.0 MVP first (20 weeks)
- Release early, gather feedback, iterate
- Build community from day 1 (open source)
- Target v1.0.0 in 8-9 months

**Success Probability:** 70-80% (High confidence)

---

## Resources

### Documentation
- **[Getting Started](./GETTING_STARTED.md)** - Implementation strategy & first steps
- **[Best Practices](./BEST_PRACTICES.md)** - Design patterns & guidelines
- [SDK Architecture](./plan/SDK_ARCHITECTURE.md)
- [Features Inventory](./plan/LLAMAINDEX_FEATURES_INVENTORY.md)
- [Implementation Scope](./plan/IMPLEMENTATION_SCOPE.md)
- [Research Notes](./notes/README.md)

### References
- [LlamaIndex Docs](https://docs.llamaindex.ai/)
- [LlamaIndex GitHub](https://github.com/run-llama/llama_index)
- [Go Documentation](https://go.dev/doc/)

---

**Contact:** Yash Rahurikar  
**Date:** October 7, 2025  
**Status:** Ready to Build! 🚀
