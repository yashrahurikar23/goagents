# Changelog

All notable changes to GoAgent will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2025-10-08

### Added

#### LLM Providers (NEW!)
- **Anthropic Claude Integration**
  - Complete Claude 3.5 Sonnet, 3 Opus, 3 Haiku support
  - Anthropic Messages API with 200K context window
  - System prompt handling and message conversion
  - Comprehensive error handling and metadata enrichment
  - 22 tests (17 unit + 5 integration)
  - Example with multiple Claude models
  
- **Google Gemini Integration**
  - Gemini 2.0 Flash, 1.5 Pro, 1.5 Flash support
  - Role mapping ("assistant" → "model" for Gemini compatibility)
  - Safety ratings and content filtering
  - PromptFeedback and blocked content handling
  - 28 tests (22 unit + 6 integration)
  - Example with free tier Gemini access

#### Tools (NEW!)
- **File Operations Tool**
  - 7 secure file operations (read, write, append, list, exists, delete, info)
  - **5-layer security protection:**
    1. Base directory enforcement
    2. Path traversal prevention (blocks "..")
    3. File size limits (default 10MB)
    4. Read-only mode support
    5. Safe file permissions (0644 files, 0755 directories)
  - 21 comprehensive tests covering operations, security, edge cases
  - Example with 8 usage scenarios
  - Detailed security documentation

#### Code Quality
- **Comprehensive Code Comments**
  - 745+ lines of WHY-focused documentation
  - Package-level docs for all modules (PURPOSE, DESIGN DECISIONS, KEY FEATURES)
  - Method-level docs explaining rationale and business logic
  - Security-critical code with detailed defense-in-depth explanations
  - Follows project code-comments-guidelines.md

#### Documentation
- **New Examples:**
  - `examples/anthropic_claude/` - Complete Claude integration example
  - `examples/gemini/` - Gemini with free tier guide
  - `examples/file_operations/` - 8 file operation scenarios
  
- **Archive Documentation:**
  - `CODE_COMMENTS_COMPLETE.md` - Documentation enhancement summary
  - `FILE_OPERATIONS_TOOL_COMPLETE.md` - File tool implementation details

### Technical Details
- **Total Tests:** 180+ (all passing) ⬆️ from 113
- **LLM Providers:** 4 (OpenAI, Ollama, Anthropic, Gemini) ⬆️ from 2
- **Tools:** 3 (Calculator, HTTP, File Operations) ⬆️ from 2
- **Code Comments:** ~745 lines of comprehensive WHY-focused documentation
- **Security:** Multi-layer file system protection with explicit documentation

### Changed
- Enhanced error messages across all LLM providers
- Improved metadata tracking for all providers (token usage, model info, finish reasons)
- Better type safety with pointer types for optional parameters

### Performance
- All tests pass in < 3 seconds
- Efficient message conversion and validation
- Token usage tracking for cost optimization
- File size limits prevent memory exhaustion

---

## [0.2.0] - 2025-10-08

### Added
- **HTTP Tool** with REST API support, authentication, retries
- Documentation reorganization (guides, archive)
- 13 HTTP tool tests

### Changed
- Cleaner directory structure
- Moved historical docs to `docs/archive/`

---

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
