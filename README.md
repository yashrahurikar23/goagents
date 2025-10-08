# 🚀 GoAgents - AI Agent Framework for Go

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.- **v0.3.0** 🚧 Anthropic Claude ✅, Google Gemini ✅, File Operations ✅, web search, streamingo/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tests](https://img.shields.io/badge/tests-113%20passing-success)](https://github.com/yashrahurikar23/goagents)

Production-ready AI agent framework for Go with support for multiple LLM providers and agent architectures.

**Let's Go, Agents!** 🚀

---

## ✨ Features

- 🤖 **3 Agent Types**: FunctionAgent, ReActAgent, ConversationalAgent
- 🔌 **4 LLM Providers**: OpenAI, Anthropic Claude, Google Gemini, Ollama (local AI)
- 🛠️ **Powerful Tools**: Calculator, HTTP client, File operations, easy custom tools
- 💾 **Memory Management**: 4 strategies for conversation history
- 🧪 **Fully Tested**: 165+ tests passing
- ⚡ **Production Ready**: Type-safe, concurrent, efficient
- 🌐 **Local AI Support**: Run offline with Ollama

---

## 📦 Installation

\`\`\`bash
go get github.com/yashrahurikar23/goagents@latest
\`\`\`

---

## 🚀 Quick Start

### Using Ollama (Local, Free)

\`\`\`go
package main

import (
    "context"
    "fmt"
    "github.com/yashrahurikar23/goagents/agent"
    "github.com/yashrahurikar23/goagents/llm/ollama"
    "github.com/yashrahurikar23/goagents/tools"
)

func main() {
    llm := ollama.New(ollama.WithModel("llama3.2:1b"))
    calc := tools.NewCalculator()
    
    agent := agent.NewReActAgent(llm)
    agent.AddTool(calc)
    
    response, _ := agent.Run(context.Background(), "What is 25 * 4 + 100?")
    fmt.Println("Agent:", response.Content)
}
\`\`\`

### Using OpenAI

\`\`\`go
package main

import (
    "context"
    "os"
    "github.com/yashrahurikar23/goagents/agent"
    "github.com/yashrahurikar23/goagents/llm/openai"
    "github.com/yashrahurikar23/goagents/tools"
)

func main() {
    llm := openai.New(
        openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
        openai.WithModel("gpt-4"),
    )
    
    httpTool := tools.NewHTTPTool()
    
    agent := agent.NewFunctionAgent(llm)
    agent.AddTool(httpTool)
    
    response, _ := agent.Run(ctx, "Fetch weather from wttr.in/Boston")
    fmt.Println(response.Content)
}
\`\`\`

---

## 🤖 Agent Types

### 1. FunctionAgent
OpenAI's native function calling. Best for production.

\`\`\`go
agent := agent.NewFunctionAgent(llm)
\`\`\`

### 2. ReActAgent
Reasoning + Action with transparent thought process.

\`\`\`go
agent := agent.NewReActAgent(llm)  // Works with Ollama!
\`\`\`

### 3. ConversationalAgent
Memory management with 4 strategies.

\`\`\`go
agent := agent.NewConversationalAgent(llm,
    agent.WithMemoryStrategy(agent.MemoryWindow),
    agent.WithMaxMessages(10),
)
\`\`\`

---

## 🛠️ Tools

### Built-in Tools

- **Calculator**: Math operations (add, subtract, multiply, divide, power, sqrt)
- **HTTP Client** (v0.2.0): REST API calls with auth, retries, JSON
- **File Operations** (NEW in v0.3.0): Secure file system access with read, write, list, delete, and safety features

### Custom Tools

\`\`\`go
type MyTool struct{}

func (t *MyTool) Name() string { return "mytool" }
func (t *MyTool) Description() string { return "Does something" }
func (t *MyTool) Schema() *core.ToolSchema { /* ... */ }
func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // Your logic here
}
\`\`\`

---

## 📚 Documentation

- **[Complete Docs](./docs/README.md)** - Full guides and API reference
- **[Agent Architectures](./docs/guides/AGENT_ARCHITECTURES.md)** - Detailed patterns
- **[Best Practices](./docs/guides/BEST_PRACTICES.md)** - Design guidelines
- **[API Reference](https://pkg.go.dev/github.com/yashrahurikar23/goagents)** - Go docs

---

## 📋 Examples

- **[ReAct with Ollama](./examples/react_ollama.go)** - Local AI
- **[Anthropic Claude](./examples/anthropic_claude/)** - Claude models
- **[Google Gemini](./examples/gemini/)** - Gemini models (free tier!)
- **[HTTP Tool](./examples/http_tool/)** - REST API integration

Run examples:
\`\`\`bash
go run examples/react_ollama.go
go run examples/anthropic_claude/main.go
go run examples/gemini/main.go
\`\`\`

---

## 🧪 Testing

\`\`\`bash
go test ./...              # All tests
go test ./agent/...        # Agent tests
go test -cover ./...       # With coverage
\`\`\`

---

## 🗺️ Roadmap

- **v0.1.0** ✅ Core framework, 3 agents, OpenAI + Ollama
- **v0.2.0** ✅ HTTP tool, examples
- **v0.3.0** � Anthropic Claude ✅, Google Gemini ✅, File ops, web search, streaming
- **v0.5.0** 🔮 RAG, vector DBs, multi-agent

[Full Roadmap](./docs/roadmap/V0.3.0_IMPLEMENTATION_CHECKLIST.md)

---

## 🤝 Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md)

---

## 📄 License

MIT License - see [LICENSE](./LICENSE)

---

## 💬 Community

- **GitHub**: [yashrahurikar23/goagents](https://github.com/yashrahurikar23/goagents)
- **Issues**: [Report bugs](https://github.com/yashrahurikar23/goagents/issues)
- **Discussions**: [Ask questions](https://github.com/yashrahurikar23/goagents/discussions)

---

**Built with ❤️ by [Yash Rahurikar](https://github.com/yashrahurikar23)**

**Star ⭐ if you find this useful!**
