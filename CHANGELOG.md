# Changelog

All notable changes to GoAgent will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-10-07

### Added

#### Core Package
- Core interfaces: `LLM`, `Tool`, `Agent`
- Type definitions: `Message`, `Response`, `ToolCall`, `ToolSchema`, `Parameter`
- Error types: `ErrInvalidArgument`, `ErrToolNotFound`, `ErrToolExecution`, `ErrLLMFailure`, `ErrTimeout`
- Helper functions: `NewMessage`, `SystemMessage`, `UserMessage`, `AssistantMessage`
- 42 comprehensive tests

#### Agent Package
- **FunctionAgent**: OpenAI native function calling with automatic tool execution
- **ReActAgent**: Reasoning + Acting pattern with transparent thought traces
- **ConversationalAgent**: Multi-turn conversations with 4 memory strategies
- Memory strategies: Window, Summarize, Selective, All
- Functional options pattern for configuration
- 43 comprehensive tests

#### LLM Providers
- **OpenAI Client**: Complete integration with GPT-3.5/GPT-4
  - Chat completions
  - Function calling support
  - Streaming (partial)
  - Error handling and retries
  
- **Ollama Client**: Local LLM support (NEW!)
  - Chat completions
  - Text generation
  - Streaming responses
  - Model management
  - Works with llama3.2, gemma3, qwen3, phi3, deepseek, and more
  - 15 integration tests

#### Tools
- Calculator tool with basic arithmetic operations (add, subtract, multiply, divide)
- Tool interface for easy custom tool creation

#### Testing
- Mock LLM client for testing
- Mock tool implementation
- 100+ tests passing across all packages
- Integration tests with real Ollama server

#### Documentation
- Comprehensive README with quick start examples
- Agent architecture guide (AGENT_ARCHITECTURES.md)
- Ollama client documentation (OLLAMA_CLIENT_COMPLETE.md)
- Packaging guide (PACKAGING_GUIDE.md)
- Release guide (READY_TO_SHIP.md)

### Technical Details
- Go 1.22.1+
- No external dependencies (stdlib only)
- Type-safe with generics where appropriate
- Concurrent-safe implementations
- Comprehensive error handling
- Production-ready code quality

### Performance
- All tests pass in < 2 seconds
- Low memory footprint
- Efficient token usage tracking
- Streaming support for long responses

## [Unreleased]

### Planned for v0.2.0
- HTTP tool for API calls
- File operations tool
- Web scraping tool
- More code examples
- Performance benchmarks

### Planned for v0.5.0
- RAG (Retrieval Augmented Generation)
- Vector store integrations (Qdrant, Pinecone, Weaviate)
- Multi-agent coordination
- Workflow system with events
- Additional LLM providers (Anthropic, Cohere)

### Planned for v1.0.0
- Complete documentation website
- Enterprise features
- Evaluation framework
- Advanced agent patterns
- Migration tools from LangChain/LlamaIndex

---

## Version History

- **v0.1.0** (2025-10-07): Initial release with core agents, OpenAI, and Ollama support

---

**Note:** This is the first public release. Please report any issues on [GitHub Issues](https://github.com/yashrahurikar23/goagents/issues).
