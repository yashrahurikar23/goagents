# GoAgent Testing Strategy

**Last Updated:** October 7, 2025  
**Status:** Comprehensive Testing Plan  
**Goal:** 100% automated testing coverage with CI/CD integration

---

## 📋 Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Testing Layers](#testing-layers)
3. [Test Organization](#test-organization)
4. [Unit Testing Strategy](#unit-testing-strategy)
5. [Integration Testing Strategy](#integration-testing-strategy)
6. [Mock Strategy](#mock-strategy)
7. [End-to-End Testing](#end-to-end-testing)
8. [Performance Testing](#performance-testing)
9. [CI/CD Integration](#cicd-integration)
10. [Test Automation](#test-automation)
11. [Implementation Plan](#implementation-plan)

---

## 🎯 Testing Philosophy

### Core Principles

1. **Fast Feedback** - Unit tests run in milliseconds
2. **Deterministic** - Tests produce same results every time
3. **Isolated** - Tests don't depend on external services by default
4. **Comprehensive** - Cover all code paths and edge cases
5. **Maintainable** - Tests are easy to understand and update

### Testing Pyramid

```
        /\
       /  \      E2E Tests (5%)
      /____\     - Full integration
     /      \    - Real API calls
    /        \   - Slow, expensive
   /__________\  
  /            \ Integration Tests (20%)
 /              \ - Multiple components
/________________\ - Mocked external APIs
\                / 
 \              /  Unit Tests (75%)
  \            /   - Single functions
   \__________/    - Fast, isolated
```

---

## 🏗️ Testing Layers

### Layer 1: Unit Tests (75% of tests)

**Purpose:** Test individual functions/methods in isolation  
**Speed:** < 1ms per test  
**Dependencies:** None (all mocked)

**What to test:**
- Core package types and interfaces
- OpenAI client methods (with HTTP mocks)
- Tool implementations
- Agent logic
- Error handling
- Edge cases

### Layer 2: Integration Tests (20% of tests)

**Purpose:** Test component interactions  
**Speed:** 10-100ms per test  
**Dependencies:** Multiple components, mocked external APIs

**What to test:**
- LLM client + Tool integration
- Agent + LLM + Tool pipeline
- Data flow between components
- Error propagation
- Retry mechanisms

### Layer 3: End-to-End Tests (5% of tests)

**Purpose:** Test complete workflows with real APIs  
**Speed:** 1-10s per test  
**Dependencies:** Real OpenAI API (requires key)

**What to test:**
- Complete agent workflows
- Real API responses
- Error scenarios with real API
- Performance benchmarks

---

## 📁 Test Organization

### Directory Structure

```
goagent/
├── core/
│   ├── interfaces.go
│   ├── types.go
│   ├── errors.go
│   ├── interfaces_test.go       ← Unit tests
│   ├── types_test.go            ← Unit tests
│   └── errors_test.go           ← Unit tests
├── llm/openai/
│   ├── client.go
│   ├── types.go
│   ├── client_test.go           ← Unit tests (mocked HTTP)
│   ├── types_test.go            ← Unit tests
│   ├── integration_test.go      ← Integration tests
│   └── testdata/                ← Test fixtures
│       ├── responses/
│       │   ├── chat_success.json
│       │   ├── chat_error.json
│       │   └── streaming.txt
│       └── requests/
│           └── example_requests.json
├── tools/
│   ├── calculator.go
│   ├── calculator_test.go       ← Unit tests
│   ├── http.go
│   └── http_test.go             ← Unit tests
├── agent/
│   ├── function.go
│   ├── function_test.go         ← Unit tests
│   └── integration_test.go      ← Integration tests
├── tests/
│   ├── e2e/
│   │   ├── quickstart_test.go   ← E2E tests
│   │   ├── rag_test.go
│   │   └── README.md
│   ├── integration/
│   │   ├── agent_llm_test.go
│   │   └── README.md
│   └── mocks/
│       ├── llm_mock.go
│       ├── tool_mock.go
│       └── http_mock.go
└── .github/
    └── workflows/
        └── test.yml             ← CI/CD pipeline
```

### Naming Conventions

- **Unit tests:** `*_test.go` in same package
- **Integration tests:** `integration_test.go`
- **E2E tests:** `tests/e2e/*_test.go`
- **Mocks:** `tests/mocks/*_mock.go`
- **Test data:** `testdata/` directory

---

## 🧪 Unit Testing Strategy

### Core Package Tests

#### 1. Test Core Types (`core/types_test.go`)

```go
package core_test

import (
    "testing"
    "github.com/yashrahurikar23/goagents/core"
)

func TestNewMessage(t *testing.T) {
    tests := []struct {
        name     string
        role     string
        content  string
        want     core.Message
    }{
        {
            name:    "user message",
            role:    "user",
            content: "hello",
            want:    core.Message{Role: "user", Content: "hello"},
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := core.NewMessage(tt.role, tt.content)
            if got.Role != tt.want.Role || got.Content != tt.want.Content {
                t.Errorf("NewMessage() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestUserMessage(t *testing.T) {
    msg := core.UserMessage("test content")
    
    if msg.Role != "user" {
        t.Errorf("UserMessage() role = %v, want user", msg.Role)
    }
    if msg.Content != "test content" {
        t.Errorf("UserMessage() content = %v, want test content", msg.Content)
    }
}

// Test Message with metadata
func TestMessageWithMetadata(t *testing.T) {
    msg := core.NewMessage("assistant", "response")
    msg.Meta = map[string]interface{}{
        "model": "gpt-4",
        "tokens": 100,
    }
    
    if msg.Meta["model"] != "gpt-4" {
        t.Errorf("Expected model=gpt-4, got %v", msg.Meta["model"])
    }
}
```

#### 2. Test Error Types (`core/errors_test.go`)

```go
package core_test

import (
    "errors"
    "testing"
    "github.com/yashrahurikar23/goagents/core"
)

func TestErrLLMFailure(t *testing.T) {
    originalErr := errors.New("connection timeout")
    err := &core.ErrLLMFailure{
        Provider: "openai",
        Err:      originalErr,
    }
    
    // Test error message
    want := "LLM \"openai\" request failed: connection timeout"
    if err.Error() != want {
        t.Errorf("Error() = %v, want %v", err.Error(), want)
    }
    
    // Test error unwrapping
    if !errors.Is(err, originalErr) {
        t.Error("Expected error to wrap original error")
    }
}

func TestErrToolNotFound(t *testing.T) {
    err := &core.ErrToolNotFound{ToolName: "calculator"}
    want := "tool not found: calculator"
    
    if err.Error() != want {
        t.Errorf("Error() = %v, want %v", err.Error(), want)
    }
}
```

### OpenAI Client Tests

#### 3. Test with Mocked HTTP (`llm/openai/client_test.go`)

```go
package openai_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/yashrahurikar23/goagents/llm/openai"
)

// Mock HTTP server for testing
func setupMockServer(t *testing.T, response string, statusCode int) *httptest.Server {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        if r.Header.Get("Authorization") != "Bearer test-key" {
            t.Error("Missing or incorrect authorization header")
        }
        
        w.WriteHeader(statusCode)
        w.Write([]byte(response))
    })
    
    return httptest.NewServer(handler)
}

func TestCreateChatCompletion_Success(t *testing.T) {
    // Mock successful response
    mockResponse := `{
        "id": "chatcmpl-123",
        "object": "chat.completion",
        "created": 1677652288,
        "model": "gpt-4",
        "choices": [{
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "Hello! How can I help you?"
            },
            "finish_reason": "stop"
        }],
        "usage": {
            "prompt_tokens": 10,
            "completion_tokens": 20,
            "total_tokens": 30
        }
    }`
    
    server := setupMockServer(t, mockResponse, http.StatusOK)
    defer server.Close()
    
    // Create client pointing to mock server
    client := openai.New(
        openai.WithAPIKey("test-key"),
        openai.WithBaseURL(server.URL),
    )
    
    // Test request
    req := openai.ChatCompletionRequest{
        Model: "gpt-4",
        Messages: []openai.ChatMessage{
            openai.UserMessage("Hello"),
        },
    }
    
    resp, err := client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        t.Fatalf("CreateChatCompletion() error = %v", err)
    }
    
    // Validate response
    if resp.Model != "gpt-4" {
        t.Errorf("Model = %v, want gpt-4", resp.Model)
    }
    if len(resp.Choices) != 1 {
        t.Fatalf("Expected 1 choice, got %d", len(resp.Choices))
    }
    if content, ok := resp.Choices[0].Message.Content.(string); !ok || content != "Hello! How can I help you?" {
        t.Errorf("Unexpected content: %v", resp.Choices[0].Message.Content)
    }
    if resp.Usage.TotalTokens != 30 {
        t.Errorf("TotalTokens = %d, want 30", resp.Usage.TotalTokens)
    }
}

func TestCreateChatCompletion_RateLimit(t *testing.T) {
    mockResponse := `{
        "error": {
            "message": "Rate limit exceeded",
            "type": "rate_limit_error",
            "code": "rate_limit_exceeded"
        }
    }`
    
    server := setupMockServer(t, mockResponse, http.StatusTooManyRequests)
    defer server.Close()
    
    client := openai.New(
        openai.WithAPIKey("test-key"),
        openai.WithBaseURL(server.URL),
        openai.WithMaxRetries(0), // Disable retries for test
    )
    
    req := openai.ChatCompletionRequest{
        Model:    "gpt-4",
        Messages: []openai.ChatMessage{openai.UserMessage("Test")},
    }
    
    _, err := client.CreateChatCompletion(context.Background(), req)
    
    // Should return error
    if err == nil {
        t.Fatal("Expected error, got nil")
    }
    
    // Should be rate limit error
    if !openai.IsRateLimitError(err) {
        t.Errorf("Expected rate limit error, got %v", err)
    }
}

func TestCreateChatCompletion_RetryLogic(t *testing.T) {
    attemptCount := 0
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attemptCount++
        
        // Fail first 2 attempts, succeed on 3rd
        if attemptCount < 3 {
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte(`{"error": {"message": "Server error"}}`))
            return
        }
        
        // Success response
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{
            "choices": [{"message": {"role": "assistant", "content": "Success"}}]
        }`))
    })
    
    server := httptest.NewServer(handler)
    defer server.Close()
    
    client := openai.New(
        openai.WithAPIKey("test-key"),
        openai.WithBaseURL(server.URL),
        openai.WithMaxRetries(3),
    )
    
    req := openai.ChatCompletionRequest{
        Model:    "gpt-4",
        Messages: []openai.ChatMessage{openai.UserMessage("Test")},
    }
    
    resp, err := client.CreateChatCompletion(context.Background(), req)
    
    // Should succeed after retries
    if err != nil {
        t.Fatalf("Expected success after retries, got error: %v", err)
    }
    
    // Should have retried 3 times total
    if attemptCount != 3 {
        t.Errorf("Expected 3 attempts, got %d", attemptCount)
    }
    
    if content, ok := resp.Choices[0].Message.Content.(string); !ok || content != "Success" {
        t.Errorf("Unexpected content: %v", resp.Choices[0].Message.Content)
    }
}

func TestComplete_ContextCancellation(t *testing.T) {
    // Server that delays response
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Simulate slow response
        <-r.Context().Done()
    })
    
    server := httptest.NewServer(handler)
    defer server.Close()
    
    client := openai.New(
        openai.WithAPIKey("test-key"),
        openai.WithBaseURL(server.URL),
    )
    
    // Create context with immediate cancellation
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately
    
    _, err := client.Complete(ctx, "Test")
    
    // Should return context canceled error
    if err == nil {
        t.Fatal("Expected context canceled error, got nil")
    }
    if !errors.Is(err, context.Canceled) {
        t.Errorf("Expected context.Canceled, got %v", err)
    }
}
```

#### 4. Test Streaming (`llm/openai/streaming_test.go`)

```go
package openai_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "github.com/yashrahurikar23/goagents/llm/openai"
)

func TestCreateChatCompletionStream(t *testing.T) {
    // Mock SSE stream
    sseResponse := `data: {"id":"1","choices":[{"delta":{"content":"Hello"}}]}

data: {"id":"1","choices":[{"delta":{"content":" world"}}]}

data: {"id":"1","choices":[{"delta":{"content":"!"}}]}

data: [DONE]

`
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/event-stream")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(sseResponse))
    })
    
    server := httptest.NewServer(handler)
    defer server.Close()
    
    client := openai.New(
        openai.WithAPIKey("test-key"),
        openai.WithBaseURL(server.URL),
    )
    
    // Accumulate chunks
    var chunks []string
    completed := false
    
    streamOpts := openai.StreamOptions{
        OnChunk: func(chunk *openai.ChatCompletionStreamResponse) error {
            if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
                if content, ok := chunk.Choices[0].Delta.Content.(string); ok {
                    chunks = append(chunks, content)
                }
            }
            return nil
        },
        OnComplete: func() error {
            completed = true
            return nil
        },
    }
    
    req := openai.ChatCompletionRequest{
        Model:    "gpt-4",
        Messages: []openai.ChatMessage{openai.UserMessage("Hi")},
    }
    
    err := client.CreateChatCompletionStream(context.Background(), req, streamOpts)
    if err != nil {
        t.Fatalf("Stream error: %v", err)
    }
    
    // Verify chunks
    want := []string{"Hello", " world", "!"}
    if len(chunks) != len(want) {
        t.Fatalf("Got %d chunks, want %d", len(chunks), len(want))
    }
    for i, chunk := range chunks {
        if chunk != want[i] {
            t.Errorf("Chunk %d = %v, want %v", i, chunk, want[i])
        }
    }
    
    // Verify completion called
    if !completed {
        t.Error("OnComplete was not called")
    }
    
    // Verify full message
    full := strings.Join(chunks, "")
    if full != "Hello world!" {
        t.Errorf("Full message = %v, want 'Hello world!'", full)
    }
}
```

### Tool Tests

#### 5. Test Calculator Tool (`tools/calculator_test.go`)

```go
package tools_test

import (
    "context"
    "testing"
    "github.com/yashrahurikar23/goagents/tools"
)

func TestCalculator_Add(t *testing.T) {
    calc := tools.NewCalculator()
    
    tests := []struct {
        name    string
        args    map[string]interface{}
        want    float64
        wantErr bool
    }{
        {
            name:    "add positive numbers",
            args:    map[string]interface{}{"a": 5.0, "b": 3.0, "operation": "add"},
            want:    8.0,
            wantErr: false,
        },
        {
            name:    "add negative numbers",
            args:    map[string]interface{}{"a": -5.0, "b": 3.0, "operation": "add"},
            want:    -2.0,
            wantErr: false,
        },
        {
            name:    "missing argument",
            args:    map[string]interface{}{"a": 5.0, "operation": "add"},
            want:    0,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := calc.Execute(context.Background(), tt.args)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr {
                got, ok := result.(float64)
                if !ok {
                    t.Fatalf("Result is not float64: %T", result)
                }
                if got != tt.want {
                    t.Errorf("Execute() = %v, want %v", got, tt.want)
                }
            }
        })
    }
}

func TestCalculator_DivideByZero(t *testing.T) {
    calc := tools.NewCalculator()
    
    args := map[string]interface{}{
        "a":         10.0,
        "b":         0.0,
        "operation": "divide",
    }
    
    _, err := calc.Execute(context.Background(), args)
    
    if err == nil {
        t.Fatal("Expected error for division by zero, got nil")
    }
    
    // Should be ErrToolExecution
    if _, ok := err.(*core.ErrToolExecution); !ok {
        t.Errorf("Expected ErrToolExecution, got %T", err)
    }
}

func TestCalculator_Schema(t *testing.T) {
    calc := tools.NewCalculator()
    schema := calc.Schema()
    
    if schema.Name != "calculator" {
        t.Errorf("Name = %v, want calculator", schema.Name)
    }
    
    if len(schema.Parameters) == 0 {
        t.Error("Expected parameters to be defined")
    }
    
    // Verify required parameters
    requiredParams := []string{"a", "b", "operation"}
    for _, param := range schema.Parameters {
        if param.Required {
            found := false
            for _, required := range requiredParams {
                if param.Name == required {
                    found = true
                    break
                }
            }
            if !found {
                t.Errorf("Unexpected required parameter: %s", param.Name)
            }
        }
    }
}
```

---

## 🔗 Integration Testing Strategy

### Agent + LLM + Tool Integration

#### 6. Integration Test (`tests/integration/agent_llm_test.go`)

```go
package integration_test

import (
    "context"
    "testing"
    "github.com/yashrahurikar23/goagents/agent"
    "github.com/yashrahurikar23/goagents/llm/openai"
    "github.com/yashrahurikar23/goagents/tools"
    "github.com/yashrahurikar23/goagents/tests/mocks"
)

func TestFunctionAgent_WithMockedLLM(t *testing.T) {
    // Create mock LLM that returns tool call
    mockLLM := mocks.NewMockLLM()
    mockLLM.SetResponse(&core.Response{
        Content: "",
        ToolCalls: []core.ToolCall{
            {
                ID:   "call_123",
                Name: "calculator",
                Args: map[string]interface{}{
                    "a":         5.0,
                    "b":         3.0,
                    "operation": "add",
                },
            },
        },
    })
    
    // Create real calculator tool
    calc := tools.NewCalculator()
    
    // Create agent with mock LLM and real tool
    agent := agent.NewFunctionAgent(mockLLM)
    err := agent.AddTool(calc)
    if err != nil {
        t.Fatalf("Failed to add tool: %v", err)
    }
    
    // Run agent
    response, err := agent.Run(context.Background(), "What is 5 + 3?")
    if err != nil {
        t.Fatalf("Agent run failed: %v", err)
    }
    
    // Verify tool was called
    if !mockLLM.WasCalled() {
        t.Error("Expected LLM to be called")
    }
    
    // Verify response contains result
    // (Implementation depends on how FunctionAgent formats responses)
    if response.Content == "" {
        t.Error("Expected non-empty response")
    }
}
```

---

## 🎭 Mock Strategy

### Mock LLM (`tests/mocks/llm_mock.go`)

```go
package mocks

import (
    "context"
    "github.com/yashrahurikar23/goagents/core"
)

// MockLLM is a mock implementation of core.LLM for testing
type MockLLM struct {
    responses []*core.Response
    calls     [][]core.Message
    errors    []error
    callCount int
}

func NewMockLLM() *MockLLM {
    return &MockLLM{
        responses: []*core.Response{},
        calls:     [][]core.Message{},
        errors:    []error{},
    }
}

// SetResponse sets the response to return
func (m *MockLLM) SetResponse(resp *core.Response) {
    m.responses = append(m.responses, resp)
}

// SetError sets an error to return
func (m *MockLLM) SetError(err error) {
    m.errors = append(m.errors, err)
}

// Chat implements core.LLM
func (m *MockLLM) Chat(ctx context.Context, messages []core.Message) (*core.Response, error) {
    m.calls = append(m.calls, messages)
    
    if m.callCount < len(m.errors) && m.errors[m.callCount] != nil {
        err := m.errors[m.callCount]
        m.callCount++
        return nil, err
    }
    
    if m.callCount < len(m.responses) {
        resp := m.responses[m.callCount]
        m.callCount++
        return resp, nil
    }
    
    return &core.Response{Content: "Mock response"}, nil
}

// Complete implements core.LLM
func (m *MockLLM) Complete(ctx context.Context, prompt string) (string, error) {
    resp, err := m.Chat(ctx, []core.Message{{Role: "user", Content: prompt}})
    if err != nil {
        return "", err
    }
    return resp.Content, nil
}

// WasCalled returns true if Chat was called
func (m *MockLLM) WasCalled() bool {
    return len(m.calls) > 0
}

// GetCalls returns all Chat calls
func (m *MockLLM) GetCalls() [][]core.Message {
    return m.calls
}

// Reset resets the mock
func (m *MockLLM) Reset() {
    m.responses = []*core.Response{}
    m.calls = [][]core.Message{}
    m.errors = []error{}
    m.callCount = 0
}
```

### Mock Tool (`tests/mocks/tool_mock.go`)

```go
package mocks

import (
    "context"
    "github.com/yashrahurikar23/goagents/core"
)

// MockTool is a mock implementation of core.Tool for testing
type MockTool struct {
    name        string
    description string
    schema      *core.ToolSchema
    result      interface{}
    err         error
    calls       []map[string]interface{}
}

func NewMockTool(name string) *MockTool {
    return &MockTool{
        name:        name,
        description: "Mock tool for testing",
        schema: &core.ToolSchema{
            Name:        name,
            Description: "Mock tool",
            Parameters:  []core.Parameter{},
        },
        calls: []map[string]interface{}{},
    }
}

// SetResult sets the result to return
func (m *MockTool) SetResult(result interface{}) {
    m.result = result
}

// SetError sets an error to return
func (m *MockTool) SetError(err error) {
    m.err = err
}

// Name implements core.Tool
func (m *MockTool) Name() string {
    return m.name
}

// Description implements core.Tool
func (m *MockTool) Description() string {
    return m.description
}

// Schema implements core.Tool
func (m *MockTool) Schema() *core.ToolSchema {
    return m.schema
}

// Execute implements core.Tool
func (m *MockTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    m.calls = append(m.calls, args)
    
    if m.err != nil {
        return nil, m.err
    }
    
    if m.result != nil {
        return m.result, nil
    }
    
    return "mock result", nil
}

// GetCalls returns all Execute calls
func (m *MockTool) GetCalls() []map[string]interface{} {
    return m.calls
}

// WasCalled returns true if Execute was called
func (m *MockTool) WasCalled() bool {
    return len(m.calls) > 0
}
```

---

## 🌐 End-to-End Testing

### E2E Test with Real API (`tests/e2e/quickstart_test.go`)

```go
// +build e2e

package e2e_test

import (
    "context"
    "os"
    "testing"
    "github.com/yashrahurikar23/goagents/agent"
    "github.com/yashrahurikar23/goagents/llm/openai"
    "github.com/yashrahurikar23/goagents/tools"
)

func TestQuickstart_RealAPI(t *testing.T) {
    // Skip if no API key
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("Skipping E2E test: OPENAI_API_KEY not set")
    }
    
    // Create real OpenAI client
    client := openai.New(
        openai.WithAPIKey(apiKey),
        openai.WithModel("gpt-4"),
    )
    
    // Create real calculator tool
    calc := tools.NewCalculator()
    
    // Create agent
    agent := agent.NewFunctionAgent(client)
    err := agent.AddTool(calc)
    if err != nil {
        t.Fatalf("Failed to add tool: %v", err)
    }
    
    // Test calculation
    response, err := agent.Run(context.Background(), "What is 25 * 4?")
    if err != nil {
        t.Fatalf("Agent run failed: %v", err)
    }
    
    // Verify response contains "100"
    if !strings.Contains(response.Content, "100") {
        t.Errorf("Expected response to contain '100', got: %v", response.Content)
    }
    
    t.Logf("Agent response: %s", response.Content)
}

func TestStreaming_RealAPI(t *testing.T) {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("Skipping E2E test: OPENAI_API_KEY not set")
    }
    
    client := openai.New(
        openai.WithAPIKey(apiKey),
        openai.WithModel("gpt-4"),
    )
    
    var chunks []string
    completed := false
    
    streamOpts := openai.StreamOptions{
        OnChunk: func(chunk *openai.ChatCompletionStreamResponse) error {
            if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
                if content, ok := chunk.Choices[0].Delta.Content.(string); ok {
                    chunks = append(chunks, content)
                }
            }
            return nil
        },
        OnComplete: func() error {
            completed = true
            return nil
        },
    }
    
    req := openai.ChatCompletionRequest{
        Model:    "gpt-4",
        Messages: []openai.ChatMessage{
            openai.UserMessage("Say hello"),
        },
    }
    
    err := client.CreateChatCompletionStream(context.Background(), req, streamOpts)
    if err != nil {
        t.Fatalf("Stream failed: %v", err)
    }
    
    // Verify received chunks
    if len(chunks) == 0 {
        t.Error("Expected to receive chunks")
    }
    
    if !completed {
        t.Error("Stream did not complete")
    }
    
    full := strings.Join(chunks, "")
    t.Logf("Streamed response: %s", full)
}
```

---

## ⚡ Performance Testing

### Benchmark Tests (`llm/openai/benchmark_test.go`)

```go
package openai_test

import (
    "context"
    "testing"
    "github.com/yashrahurikar23/goagents/llm/openai"
    "github.com/yashrahurikar23/goagents/tests/mocks"
)

func BenchmarkCreateChatCompletion(b *testing.B) {
    mockHTTP := mocks.NewMockHTTPServer()
    defer mockHTTP.Close()
    
    client := openai.New(
        openai.WithAPIKey("test-key"),
        openai.WithBaseURL(mockHTTP.URL()),
    )
    
    req := openai.ChatCompletionRequest{
        Model:    "gpt-4",
        Messages: []openai.ChatMessage{
            openai.UserMessage("Benchmark test"),
        },
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.CreateChatCompletion(context.Background(), req)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkMessageCreation(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = openai.UserMessage("Test message")
    }
}
```

---

## 🤖 CI/CD Integration

### GitHub Actions Workflow (`.github/workflows/test.yml`)

```yaml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  # Job 1: Unit Tests (Fast, runs on every commit)
  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [ '1.21', '1.22' ]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run unit tests
        run: go test -v -short -race -coverprofile=coverage.txt -covermode=atomic ./...
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.txt
          flags: unittests
  
  # Job 2: Integration Tests (Slower, with mocks)
  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Run integration tests
        run: go test -v -run Integration ./...
  
  # Job 3: E2E Tests (Slowest, requires API key)
  e2e-tests:
    name: E2E Tests
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Run E2E tests
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: go test -v -tags=e2e ./tests/e2e/...
  
  # Job 4: Lint & Static Analysis
  lint:
    name: Lint
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
  
  # Job 5: Build verification
  build:
    name: Build
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Build
        run: go build -v ./...
      
      - name: Build examples
        run: |
          cd examples/quickstart && go build -v .
```

### Makefile for Local Testing

```makefile
# Makefile for GoAgent testing

.PHONY: test test-unit test-integration test-e2e test-all coverage lint build clean

# Run all unit tests (fast)
test-unit:
	@echo "Running unit tests..."
	go test -v -short -race ./...

# Run integration tests (moderate speed)
test-integration:
	@echo "Running integration tests..."
	go test -v -run Integration ./...

# Run E2E tests (slow, requires API key)
test-e2e:
	@echo "Running E2E tests..."
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY not set"; \
		exit 1; \
	fi
	go test -v -tags=e2e ./tests/e2e/...

# Run all tests
test-all: test-unit test-integration test-e2e

# Default: run unit tests
test: test-unit

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	go test -v -short -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Build all packages
build:
	@echo "Building..."
	go build -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f coverage.out coverage.html
	go clean -cache

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Watch and auto-test on file changes (requires entr)
watch:
	@echo "Watching for changes..."
	find . -name "*.go" | entr -c make test-unit
```

---

## 📊 Test Coverage Goals

### Coverage Targets

| Package | Target | Priority |
|---------|--------|----------|
| `core` | 90%+ | Critical |
| `llm/openai` | 85%+ | High |
| `tools` | 85%+ | High |
| `agent` | 80%+ | High |
| Overall | 85%+ | High |

### Tracking Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View HTML report
go tool cover -html=coverage.out

# View coverage by package
go tool cover -func=coverage.out

# Check total coverage
go test -cover ./...
```

---

## 🚀 Implementation Plan

### Week 1: Core & OpenAI Tests

#### Day 1-2: Core Package Tests
- [ ] `core/types_test.go` - Test all type constructors
- [ ] `core/errors_test.go` - Test all error types
- [ ] `core/interfaces_test.go` - Test interface contracts

#### Day 3-5: OpenAI Client Tests
- [ ] `llm/openai/client_test.go` - HTTP mocks, retry logic
- [ ] `llm/openai/streaming_test.go` - Streaming tests
- [ ] `llm/openai/integration_test.go` - Component integration
- [ ] `llm/openai/testdata/` - Create test fixtures

#### Day 6-7: Mock Infrastructure
- [ ] `tests/mocks/llm_mock.go` - Mock LLM implementation
- [ ] `tests/mocks/tool_mock.go` - Mock Tool implementation
- [ ] `tests/mocks/http_mock.go` - HTTP server mocks

### Week 2: Tools & Agent Tests

#### Day 1-2: Tool Tests
- [ ] `tools/calculator_test.go` - Complete calculator tests
- [ ] `tools/http_test.go` - HTTP tool tests
- [ ] Test all operations and edge cases

#### Day 3-5: Agent Tests
- [ ] `agent/function_test.go` - FunctionAgent unit tests
- [ ] `tests/integration/agent_llm_test.go` - Integration tests
- [ ] Test tool execution pipeline

#### Day 6-7: E2E Tests
- [ ] `tests/e2e/quickstart_test.go` - Full workflow
- [ ] `tests/e2e/rag_test.go` - RAG pipeline (if implemented)
- [ ] Performance benchmarks

### Week 3: CI/CD & Polish

#### Day 1-3: CI/CD Setup
- [ ] `.github/workflows/test.yml` - GitHub Actions
- [ ] Codecov integration
- [ ] Badge generation

#### Day 4-5: Documentation
- [ ] Test documentation
- [ ] Coverage reports
- [ ] Contributing guide

#### Day 6-7: Performance
- [ ] Benchmark tests
- [ ] Memory profiling
- [ ] Optimization

---

## 📚 Testing Best Practices

### 1. Test Naming
```go
// Good: Descriptive test names
func TestCalculator_Add_PositiveNumbers(t *testing.T)
func TestOpenAI_CreateChatCompletion_RetryOnRateLimit(t *testing.T)

// Bad: Vague test names
func TestAdd(t *testing.T)
func TestAPI(t *testing.T)
```

### 2. Table-Driven Tests
```go
func TestCalculator_Operations(t *testing.T) {
    tests := []struct {
        name    string
        op      string
        a, b    float64
        want    float64
        wantErr bool
    }{
        {"add", "add", 5, 3, 8, false},
        {"subtract", "subtract", 5, 3, 2, false},
        {"divide by zero", "divide", 5, 0, 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### 3. Test Helpers
```go
// Helper for common setup
func setupTestClient(t *testing.T) *openai.Client {
    t.Helper()
    return openai.New(
        openai.WithAPIKey("test-key"),
        openai.WithBaseURL(mockServer.URL),
    )
}
```

### 4. Cleanup
```go
func TestFeature(t *testing.T) {
    server := setupMockServer()
    defer server.Close() // Always cleanup
    
    t.Cleanup(func() {
        // Additional cleanup
    })
}
```

---

## 🎯 Success Metrics

### Code Coverage
- [x] Core package: 90%+
- [ ] OpenAI client: 85%+
- [ ] Tools: 85%+
- [ ] Agent: 80%+
- [ ] Overall: 85%+

### Test Execution Time
- Unit tests: < 5 seconds
- Integration tests: < 30 seconds
- E2E tests: < 2 minutes
- Full suite: < 3 minutes

### CI/CD
- All PRs must pass tests
- Coverage must not decrease
- Linting must pass
- Build must succeed

---

## 📖 Resources

### Go Testing
- [Official Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Testing Best Practices](https://go.dev/doc/effective_go#testing)

### Tools
- [testify](https://github.com/stretchr/testify) - Assertion library
- [gomock](https://github.com/golang/mock) - Mock generation
- [httptest](https://pkg.go.dev/net/http/httptest) - HTTP testing

### CI/CD
- [GitHub Actions](https://docs.github.com/en/actions)
- [Codecov](https://about.codecov.io/)
- [golangci-lint](https://golangci-lint.run/)

---

## ✅ Next Steps

1. **Review this document** and provide feedback
2. **Start with Week 1** - Core & OpenAI tests
3. **Set up CI/CD** early for continuous validation
4. **Track coverage** with Codecov or similar
5. **Iterate and improve** based on findings

---

**Status:** Ready to implement comprehensive testing!  
**Goal:** 85%+ coverage with full automation  
**Timeline:** 3 weeks for complete test suite

Let's build rock-solid tests! 🚀
