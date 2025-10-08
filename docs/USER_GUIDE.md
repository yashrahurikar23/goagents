# 🚀 Using GoAgents in Your Project

**Quick guide to integrate GoAgents into your Go application**

---

## 📦 Step 1: Install the Package

In your project directory, run:

```bash
go get github.com/yashrahurikar23/goagents@latest
```

Or for a specific version:

```bash
go get github.com/yashrahurikar23/goagents@v0.2.0
```

This will:
- Download the package
- Add it to your `go.mod` file
- Download dependencies (if any)

---

## 📁 Step 2: Create Your Project Structure

```bash
# Create a new Go project (if you don't have one)
mkdir my-agent-app
cd my-agent-app

# Initialize Go module
go mod init github.com/yourusername/my-agent-app

# Install GoAgents
go get github.com/yashrahurikar23/goagents@latest

# Create main file
touch main.go
```

---

## 📝 Step 3: Write Your Code

### Example 1: Simple Agent with Ollama (Local, Free)

```go
// main.go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/yashrahurikar23/goagents/agent"
    "github.com/yashrahurikar23/goagents/llm/ollama"
    "github.com/yashrahurikar23/goagents/tools"
)

func main() {
    // 1. Create LLM client (Ollama - local and free)
    llm := ollama.New(
        ollama.WithModel("llama3.2:1b"),
        ollama.WithBaseURL("http://localhost:11434"),
    )
    
    // 2. Create tools
    calculator := tools.NewCalculator()
    
    // 3. Create agent
    myAgent := agent.NewReActAgent(llm)
    myAgent.AddTool(calculator)
    
    // 4. Run agent
    ctx := context.Background()
    response, err := myAgent.Run(ctx, "What is 25 * 4 + 100?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Agent:", response.Content)
}
```

**Prerequisites for Ollama:**
- Install Ollama: https://ollama.ai/download
- Pull a model: `ollama pull llama3.2:1b`
- Ollama runs on `http://localhost:11434` by default

---

### Example 2: Agent with OpenAI

```go
// main.go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/yashrahurikar23/goagents/agent"
    "github.com/yashrahurikar23/goagents/llm/openai"
    "github.com/yashrahurikar23/goagents/tools"
)

func main() {
    // 1. Create OpenAI client (requires API key)
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY environment variable required")
    }
    
    llm := openai.New(
        openai.WithAPIKey(apiKey),
        openai.WithModel("gpt-4"),
    )
    
    // 2. Create tools
    httpTool := tools.NewHTTPTool()
    calculator := tools.NewCalculator()
    
    // 3. Create FunctionAgent (best for OpenAI)
    myAgent := agent.NewFunctionAgent(llm)
    myAgent.AddTool(httpTool)
    myAgent.AddTool(calculator)
    
    // 4. Run agent
    ctx := context.Background()
    response, err := myAgent.Run(ctx, "Fetch the current time from worldtimeapi.org/api/timezone/America/New_York")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Agent:", response.Content)
}
```

**Prerequisites for OpenAI:**
- Get API key from https://platform.openai.com/api-keys
- Set environment variable: `export OPENAI_API_KEY=sk-...`

---

### Example 3: Conversational Agent with Memory

```go
// main.go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/yashrahurikar23/goagents/agent"
    "github.com/yashrahurikar23/goagents/llm/ollama"
)

func main() {
    llm := ollama.New(ollama.WithModel("llama3.2:1b"))
    
    // Create conversational agent with memory
    myAgent := agent.NewConversationalAgent(
        llm,
        agent.WithMemoryStrategy(agent.MemoryWindow), // Keep last N messages
        agent.WithMaxMessages(10),                    // Keep last 10 messages
    )
    
    ctx := context.Background()
    
    // First question
    response1, err := myAgent.Run(ctx, "My name is John")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Agent:", response1.Content)
    
    // Second question - agent remembers!
    response2, err := myAgent.Run(ctx, "What's my name?")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Agent:", response2.Content)
    // Output: Agent: Your name is John
}
```

---

## 🛠️ Step 4: Create Custom Tools

```go
// custom_tool.go
package main

import (
    "context"
    "fmt"
    
    "github.com/yashrahurikar23/goagents/core"
)

// WeatherTool fetches weather information
type WeatherTool struct{}

func NewWeatherTool() *WeatherTool {
    return &WeatherTool{}
}

func (w *WeatherTool) Name() string {
    return "weather"
}

func (w *WeatherTool) Description() string {
    return "Get current weather for a city"
}

func (w *WeatherTool) Schema() *core.ToolSchema {
    return &core.ToolSchema{
        Name:        "weather",
        Description: "Get current weather information for a city",
        Parameters: []core.Parameter{
            {
                Name:        "city",
                Type:        "string",
                Required:    true,
                Description: "Name of the city (e.g., 'Boston', 'New York')",
            },
        },
    }
}

func (w *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    city, ok := args["city"].(string)
    if !ok {
        return nil, fmt.Errorf("city parameter is required")
    }
    
    // TODO: Call actual weather API here
    // For demo, return mock data
    return map[string]interface{}{
        "city":        city,
        "temperature": 72,
        "condition":   "sunny",
        "humidity":    45,
    }, nil
}

// main.go
func main() {
    llm := ollama.New(ollama.WithModel("llama3.2:1b"))
    
    // Add custom tool
    weatherTool := NewWeatherTool()
    
    myAgent := agent.NewReActAgent(llm)
    myAgent.AddTool(weatherTool)
    
    ctx := context.Background()
    response, _ := myAgent.Run(ctx, "What's the weather in Boston?")
    fmt.Println(response.Content)
}
```

---

## ▶️ Step 5: Run Your Application

### Option 1: Direct Run
```bash
go run main.go
```

### Option 2: Build and Run
```bash
# Build executable
go build -o my-agent-app

# Run
./my-agent-app
```

### Option 3: With Environment Variables
```bash
# Set API key (for OpenAI)
export OPENAI_API_KEY=sk-your-key-here

# Run
go run main.go
```

---

## 📋 Complete Project Structure

After following the steps, your project will look like:

```
my-agent-app/
├── go.mod              # Generated by go mod init
├── go.sum              # Generated by go get
├── main.go             # Your main application
├── custom_tool.go      # (Optional) Your custom tools
├── .env                # (Optional) Environment variables
└── .gitignore          # (Optional) Git ignore file
```

**Example go.mod:**
```go
module github.com/yourusername/my-agent-app

go 1.22

require github.com/yashrahurikar23/goagents v0.2.0
```

---

## 🔧 Common Use Cases

### 1. Data Analysis Agent
```go
// Analyze CSV data, calculate statistics
agent := agent.NewFunctionAgent(llm)
agent.AddTool(tools.NewCalculator())
// Add your custom CSV reader tool
```

### 2. API Integration Agent
```go
// Interact with external APIs
agent := agent.NewFunctionAgent(llm)
agent.AddTool(tools.NewHTTPTool())
```

### 3. Research Assistant
```go
// Search and summarize information
agent := agent.NewReActAgent(llm)
agent.AddTool(tools.NewHTTPTool())
// Add web search tool (coming in v0.3.0)
```

### 4. Customer Support Bot
```go
// Chat with memory
agent := agent.NewConversationalAgent(llm,
    agent.WithMemoryStrategy(agent.MemorySummarize),
    agent.WithMaxMessages(50),
)
```

---

## 🐛 Troubleshooting

### Issue: `cannot find package`
**Solution:**
```bash
go mod tidy
go get github.com/yashrahurikar23/goagents@latest
```

### Issue: Ollama connection refused
**Solution:**
- Make sure Ollama is running: `ollama serve`
- Check it's on port 11434: `curl http://localhost:11434`
- Pull a model: `ollama pull llama3.2:1b`

### Issue: OpenAI API errors
**Solution:**
- Check API key is set: `echo $OPENAI_API_KEY`
- Verify API key at https://platform.openai.com/api-keys
- Check you have credits in your OpenAI account

### Issue: Agent not using tools
**Solution:**
- Make sure you called `agent.AddTool(tool)`
- Check tool description is clear
- Try with a more explicit question
- Use ReActAgent for debugging (shows thoughts)

---

## 📚 More Resources

- **[Full Documentation](https://github.com/yashrahurikar23/goagents/tree/main/docs)**
- **[API Reference](https://pkg.go.dev/github.com/yashrahurikar23/goagents)**
- **[Examples](https://github.com/yashrahurikar23/goagents/tree/main/examples)**
- **[Agent Architectures Guide](https://github.com/yashrahurikar23/goagents/blob/main/docs/guides/AGENT_ARCHITECTURES.md)**
- **[Best Practices](https://github.com/yashrahurikar23/goagents/blob/main/docs/guides/BEST_PRACTICES.md)**

---

## 💡 Quick Tips

1. **Start with Ollama** - Free and works offline
2. **Use ReActAgent** - Great for debugging (shows reasoning)
3. **Keep tools simple** - One tool, one purpose
4. **Test tools separately** - Before adding to agent
5. **Use context timeout** - Prevent hanging:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   response, err := agent.Run(ctx, "...")
   ```

---

## ✅ Checklist for Production

- [ ] Error handling for all agent calls
- [ ] Context timeout set appropriately
- [ ] API keys stored securely (not in code)
- [ ] Tool schemas well-documented
- [ ] Logging for debugging
- [ ] Tests for custom tools
- [ ] Rate limiting for API calls
- [ ] Graceful shutdown handling

---

## 🆘 Getting Help

- **Issues:** [Report bugs](https://github.com/yashrahurikar23/goagents/issues)
- **Discussions:** [Ask questions](https://github.com/yashrahurikar23/goagents/discussions)
- **Examples:** [See working code](https://github.com/yashrahurikar23/goagents/tree/main/examples)

---

**Happy Building!** 🚀

**Let's Go, Agents!** 🎉
