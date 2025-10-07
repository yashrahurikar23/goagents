# Contributing to GoAgent

Thanks for your interest in contributing to GoAgent! 🎉

## How to Contribute

### Reporting Bugs

1. Check if the issue already exists in [GitHub Issues](https://github.com/yashrahurikar/goagents/issues)
2. Use the Bug Report template
3. Include:
   - Clear description of the bug
   - Steps to reproduce
   - Expected vs actual behavior
   - Code sample (if applicable)
   - Environment details (Go version, OS, GoAgent version)

### Suggesting Features

1. Check if the feature is already requested
2. Use the Feature Request template
3. Explain:
   - The problem you're trying to solve
   - Your proposed solution
   - Alternative approaches you considered
   - Example use cases

### Pull Requests

1. **Fork the repository**
   ```bash
   git clone https://github.com/yashrahurikar/goagents.git
   cd goagent
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

3. **Make your changes**
   - Write clean, idiomatic Go code
   - Follow existing code style
   - Add tests for new functionality
   - Update documentation if needed

4. **Run tests**
   ```bash
   # Run all tests
   go test ./...
   
   # Run with race detector
   go test -race ./...
   
   # Check coverage
   go test -cover ./...
   ```

5. **Commit your changes**
   ```bash
   git add .
   git commit -m "Add: description of your changes"
   ```
   
   Use conventional commit messages:
   - `Add:` for new features
   - `Fix:` for bug fixes
   - `Docs:` for documentation changes
   - `Test:` for test additions
   - `Refactor:` for code refactoring
   - `Chore:` for maintenance tasks

6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Open a Pull Request**
   - Go to the original repository
   - Click "New Pull Request"
   - Select your branch
   - Fill in the PR template
   - Wait for review

## Code Style Guidelines

### Go Best Practices

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` to format code
- Run `golint` before committing
- Keep functions small and focused
- Use meaningful variable names
- Add comments for exported functions

### Example

```go
// NewCalculator creates a new calculator tool that performs basic arithmetic operations.
// It implements the core.Tool interface and can be used with any agent type.
func NewCalculator() *Calculator {
    return &Calculator{}
}
```

### Testing

- Write tests for all new functionality
- Use table-driven tests when appropriate
- Mock external dependencies
- Aim for high test coverage

Example test:

```go
func TestCalculator_Execute(t *testing.T) {
    tests := []struct {
        name    string
        args    map[string]interface{}
        want    float64
        wantErr bool
    }{
        {
            name: "addition",
            args: map[string]interface{}{
                "operation": "add",
                "a":         2.0,
                "b":         3.0,
            },
            want:    5.0,
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            calc := NewCalculator()
            result, err := calc.Execute(context.Background(), tt.args)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if result != tt.want {
                t.Errorf("Execute() = %v, want %v", result, tt.want)
            }
        })
    }
}
```

## Project Structure

```
github.com/yashrahurikar/goagents/
├── agent/              # Agent implementations
│   ├── function.go     # FunctionAgent
│   ├── react.go        # ReActAgent
│   └── conversational.go # ConversationalAgent
├── llm/                # LLM provider clients
│   ├── openai/         # OpenAI integration
│   └── ollama/         # Ollama integration
├── core/               # Core interfaces and types
│   ├── interfaces.go   # LLM, Tool, Agent interfaces
│   ├── types.go        # Message, Response, etc.
│   └── errors.go       # Error types
├── tools/              # Tool implementations
├── examples/           # Example applications
└── tests/              # Test utilities
```

## Development Setup

### Prerequisites

- Go 1.22 or higher
- Git
- (Optional) Ollama for local LLM testing

### Setup

```bash
# Clone the repository
git clone https://github.com/yashrahurikar/goagents.git
cd goagent

# Install dependencies (should be none!)
go mod tidy

# Run tests
go test ./...

# Run specific tests
go test ./agent -v
go test ./llm/ollama -v
```

## Areas We Need Help

### High Priority

- [ ] Additional LLM providers (Anthropic, Cohere, Google)
- [ ] More tools (HTTP client, file operations, web scraper)
- [ ] Performance benchmarks
- [ ] More examples and tutorials

### Medium Priority

- [ ] Vector store integrations (for RAG)
- [ ] Document loaders
- [ ] Multi-agent coordination
- [ ] Workflow system

### Low Priority

- [ ] Advanced agent patterns
- [ ] Fine-tuning support
- [ ] Evaluation framework

## Questions?

- **General questions:** Use [GitHub Discussions](https://github.com/yashrahurikar/goagents/discussions)
- **Bug reports:** Use [GitHub Issues](https://github.com/yashrahurikar/goagents/issues)
- **Feature requests:** Use [GitHub Issues](https://github.com/yashrahurikar/goagents/issues)
- **Twitter:** [@yashrahurikar](https://twitter.com/yashrahurikar)

## Code of Conduct

Please be respectful and professional in all interactions. See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for details.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to GoAgent! 🚀**
