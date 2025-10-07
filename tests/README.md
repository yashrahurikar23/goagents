# Testing Infrastructure

This directory contains all testing utilities, mocks, and test suites for the goagent framework.

**⚠️ IMPORTANT:** This directory is **NOT** part of the production package. Users installing `github.com/yashrahurikar/goagents` will NOT get these files. This is development-only infrastructure.

## Directory Structure

```
tests/
├── mocks/          # Mock implementations for testing
│   ├── llm_mock.go       # Mock LLM for unit tests
│   ├── tool_mock.go      # Mock Tool for unit tests
│   └── http_mock.go      # HTTP server mocks for OpenAI client tests
├── testutil/       # Test utilities and helpers
│   ├── helpers.go        # Common test assertions and utilities
│   └── fixtures.go       # Test data loaders
├── fixtures/       # Test data files (JSON responses, etc.)
│   └── openai/           # OpenAI API response samples
├── integration/    # Integration tests (multiple components)
│   └── agent_llm_test.go # Agent + LLM + Tool integration
└── e2e/           # End-to-end tests (real API calls)
    └── openai_test.go    # Tests with real OpenAI API
```

## Production vs Testing Code

### Production Code (shipped to users)
```
goagent/
├── core/           ✅ Part of package
├── llm/            ✅ Part of package
├── tools/          ✅ Part of package
├── agent/          ✅ Part of package
└── examples/       ✅ Part of package (documentation)
```

### Testing Code (NOT shipped)
```
goagent/
└── tests/          ❌ NOT part of package (dev only)
```

## Why Separate?

1. **Package Size:** Users don't download test infrastructure
2. **Clear Boundaries:** Testing code can import production code, but not vice versa
3. **No Pollution:** Test utilities don't leak into production API
4. **Flexibility:** Can use different patterns/dependencies in tests
5. **Import Paths:** Production code has clean import paths without `_test` suffix

## Test Organization

### Unit Tests (in production packages)
```
core/types_test.go          # Tests core/types.go
llm/openai/client_test.go   # Tests llm/openai/client.go
```

**Why in production packages:**
- Can test private functions/methods
- Conventional Go testing location
- Access to package internals

### Mocks and Helpers (in tests/)
```
tests/mocks/llm_mock.go     # Mock implementations
tests/testutil/helpers.go   # Test utilities
```

**Why separate:**
- Not needed by package users
- Can be shared across all test files
- Keeps production packages clean

### Integration Tests (in tests/)
```
tests/integration/agent_test.go
```

**Why separate:**
- Tests multiple packages together
- No natural "home" package
- Prevents import cycles

### E2E Tests (in tests/)
```
tests/e2e/openai_test.go
```

**Why separate:**
- Requires API keys
- Not run in CI by default
- Expensive/slow tests

## Import Patterns

### ✅ Correct: Tests importing production code
```go
// tests/mocks/llm_mock.go
package mocks

import "github.com/yashrahurikar/goagents/core"

type MockLLM struct {
    // Implements core.LLM
}
```

### ✅ Correct: Tests importing mocks
```go
// llm/openai/client_test.go
package openai

import "github.com/yashrahurikar/goagents/tests/mocks"

func TestClient(t *testing.T) {
    mockLLM := mocks.NewMockLLM()
    // ...
}
```

### ❌ Wrong: Production code importing tests
```go
// core/interfaces.go
package core

import "github.com/yashrahurikar/goagents/tests/mocks" // ❌ NEVER DO THIS
```

## Running Tests

```bash
# Run all unit tests (fast)
go test ./core/... ./llm/... ./tools/... ./agent/...

# Run all tests including integration (slower)
go test ./... -v

# Run only integration tests
go test ./tests/integration/... -v

# Run E2E tests (requires OPENAI_API_KEY)
OPENAI_API_KEY=sk-xxx go test ./tests/e2e/... -v

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Best Practices

### 1. Mocks Should Be Minimal
```go
// ✅ Good: Simple, focused mock
type MockLLM struct {
    ChatFunc func(ctx context.Context, messages []Message) (*Response, error)
}

// ❌ Bad: Complex mock with lots of state
type MockLLM struct {
    responses []Response
    errors []error
    callCount int
    history [][]Message
    config map[string]interface{}
}
```

### 2. Test Utilities Should Be Reusable
```go
// ✅ Good: Reusable assertion
func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

// ❌ Bad: Specific to one test
func AssertChatResponseContainsCalculation(t *testing.T, resp *Response) {
    // Too specific
}
```

### 3. Fixtures Should Be Versioned
```
tests/fixtures/openai/
├── chat_response_v1.json
├── chat_response_function_call_v1.json
└── error_rate_limit_v1.json
```

### 4. Integration Tests Should Be Isolated
```go
// ✅ Good: Each test is independent
func TestAgentWithCalculator(t *testing.T) {
    agent := agent.New()
    agent.AddTool(calculator.New())
    // Test in isolation
}

// ❌ Bad: Tests share state
var sharedAgent *agent.Agent

func TestAgentAdd(t *testing.T) {
    sharedAgent.Run(...) // ❌ Affects other tests
}
```

## What's Next

1. **Phase 1:** Create mocks (MockLLM, MockTool, MockHTTP)
2. **Phase 2:** Create test utilities (assertions, fixtures)
3. **Phase 3:** Write unit tests for core package
4. **Phase 4:** Write unit tests for OpenAI client
5. **Phase 5:** Write integration tests
6. **Phase 6:** Write E2E tests

---

**Remember:** All code in `tests/` is for **development only** and will **NOT** be shipped to users! 🔒
