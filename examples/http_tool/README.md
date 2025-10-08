# HTTP Tool Example

This example demonstrates how to use the HTTP tool with a GoAgents agent to make API calls.

## Prerequisites

1. **Ollama installed and running**
   ```bash
   # Install Ollama (if not already installed)
   curl -fsSL https://ollama.com/install.sh | sh
   
   # Pull the model
   ollama pull llama3.2:1b
   
   # Start Ollama (if not running)
   ollama serve
   ```

2. **Internet connection** - The example makes real HTTP requests to public APIs

## Features Demonstrated

- ✅ **GET requests** - Fetching data from APIs
- ✅ **POST requests** - Sending data to APIs
- ✅ **JSON handling** - Parsing JSON responses
- ✅ **Status code checking** - Verifying request success
- ✅ **Agent reasoning** - ReActAgent using HTTP tool

## Usage

```bash
cd examples/http_tool
go run main.go
```

## What It Does

### Example 1: Random Fact API
- Fetches a random fact from https://uselessfacts.jsph.pl
- Demonstrates GET request with JSON response

### Example 2: Status Check
- Checks if a website is reachable
- Demonstrates status code handling

### Example 3: JSON API
- Gets data from JSONPlaceholder API
- Demonstrates parsing JSON response data

### Example 4: POST Request
- Posts data to an API endpoint
- Demonstrates sending JSON body

## Example Output

```
🌐 HTTP Tool Example - Agent making API calls
==================================================

📡 Example 1: Fetching a random fact
--------------------------------------------------
Thought: I need to fetch a random fact from the API
Action: http
Action Input: {"method": "GET", "url": "https://uselessfacts.jsph.pl/api/v2/facts/random"}
Observation: {"status_code": 200, "body": {"text": "...", ...}}

🤖 Agent: I fetched a random fact: "..."

🔍 Example 2: Checking website status
--------------------------------------------------
...
```

## APIs Used

All APIs are **free and public** (no API keys required):

- [Useless Facts API](https://uselessfacts.jsph.pl/) - Random facts
- [HTTPBin](https://httpbin.org/) - HTTP testing service
- [JSONPlaceholder](https://jsonplaceholder.typicode.com/) - Fake REST API

## Customization

### Use OpenAI Instead of Ollama

Replace the LLM client:

```go
import "github.com/yashrahurikar23/goagents/llm/openai"

llm := openai.New(
    openai.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
    openai.WithModel("gpt-4"),
)
```

### Configure HTTP Tool Options

```go
httpTool := tools.NewHTTPTool(
    tools.WithTimeout(10 * time.Second),
    tools.WithMaxRetries(5),
    tools.WithUserAgent("MyAgent/1.0"),
)
```

### Try Other APIs

Weather API example:
```go
response, err := reactAgent.Run(
    context.Background(),
    "Use the HTTP tool to fetch weather data from https://api.open-meteo.com/v1/forecast?latitude=40.7128&longitude=-74.0060&current_weather=true",
)
```

GitHub API example:
```go
response, err := reactAgent.Run(
    context.Background(),
    "Use the HTTP tool to fetch information about the golang/go repository from https://api.github.com/repos/golang/go",
)
```

## Troubleshooting

### Ollama not found
```
Error: failed to connect to Ollama
Solution: Make sure Ollama is installed and running (ollama serve)
```

### Model not found
```
Error: model not found
Solution: Pull the model first (ollama pull llama3.2:1b)
```

### Network error
```
Error: request failed
Solution: Check your internet connection
```

### Timeout error
```
Error: context deadline exceeded
Solution: Increase timeout with WithTimeout option
```

## Learn More

- [HTTP Tool Documentation](../../tools/http.go)
- [ReAct Agent Documentation](../../agent/react.go)
- [Ollama Documentation](https://ollama.com/)
- [GoAgents README](../../README.md)

## Next Steps

1. **Try with your own APIs** - Weather, news, cryptocurrency, etc.
2. **Add authentication** - Use headers for API keys
3. **Handle errors** - Check response status codes
4. **Combine tools** - Use HTTP tool with calculator, file tools, etc.

---

**Let's Go, Agents!** 🚀
