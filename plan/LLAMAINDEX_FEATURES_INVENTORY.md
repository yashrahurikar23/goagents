# LlamaIndex Complete Feature Inventory

**Purpose:** Comprehensive documentation of all LlamaIndex features to understand the full scope of building a similar framework in Go.

**Last Updated:** October 7, 2025

---

## 1. Core Components

### 1.1 Agents

| Feature | Description | Priority | Complexity |
|---------|-------------|----------|------------|
| **FunctionAgent** | Simple agent that uses LLM function calling to select and execute tools | High | Medium |
| **ReActAgent** | Reasoning + Acting agent that follows thought-action-observation loop | High | Medium |
| **OpenAIAgent** | OpenAI-specific agent using native function calling API | Medium | Low |
| **WorkflowAgent** | Agent integrated with workflow system for complex orchestration | High | High |
| **StructuredPlannerAgent** | Plans query steps into structured DAG before execution | Medium | High |
| **Custom Agents** | Base class for building custom agent types | Low | Medium |

**Key APIs:**
- `agent.run(query)` - Synchronous execution
- `agent.chat(message)` - Conversational interface
- `agent.stream_chat(message)` - Streaming responses
- `agent.reset()` - Clear conversation history
- `agent.add_tool(tool)` - Dynamic tool addition

---

### 1.2 Tools

| Feature | Description | Priority | Complexity |
|---------|-------------|----------|------------|
| **QueryEngineTool** | Wrap query engine as tool for agent | High | Low |
| **FunctionTool** | Wrap Python function as tool with schema | High | Low |
| **LoadAndSearchTool** | Load large data + index + search (for big datasets) | Medium | Medium |
| **ToolSpec** | Base specification for defining tool collections | Medium | Low |
| **RetrieverTool** | Expose retriever as tool | Medium | Low |
| **OnDemandLoaderTool** | Load data on-demand when tool is called | Low | Medium |

**Built-in Tools:**
- Web search (Google, Bing, DuckDuckGo, Brave)
- Web scraping (Trafilatura, BeautifulSoup, Selenium)
- Database query (SQL, NoSQL)
- API calls (REST, GraphQL)
- File operations (read, write, list)
- Wolfram Alpha
- Code execution (Python, JavaScript)
- Wikipedia search
- ArXiv paper search
- Gmail, Slack, Discord integrations

---

### 1.3 LLM Integrations

| Provider | Models | Streaming | Function Calling |
|----------|--------|-----------|------------------|
| **OpenAI** | GPT-3.5, GPT-4, GPT-4-turbo | ✅ | ✅ |
| **Anthropic** | Claude 3 (Opus, Sonnet, Haiku) | ✅ | ✅ |
| **Azure OpenAI** | GPT-3.5, GPT-4 | ✅ | ✅ |
| **Google** | Gemini Pro, PaLM 2 | ✅ | ✅ |
| **Cohere** | Command, Command-R | ✅ | ❌ |
| **Hugging Face** | Any HF model | ✅ | ❌ |
| **Ollama** | Llama 2/3, Mistral, etc. | ✅ | ⚠️ |
| **AWS Bedrock** | Claude, Llama 2, Titan | ✅ | ⚠️ |
| **Replicate** | Open source models | ✅ | ❌ |
| **Together AI** | Multiple open source | ✅ | ❌ |
| **Anyscale** | Llama 2, Mistral | ✅ | ❌ |

**Key APIs:**
- `llm.complete(prompt)` - Text completion
- `llm.chat(messages)` - Chat completion
- `llm.stream_complete(prompt)` - Streaming completion
- `llm.stream_chat(messages)` - Streaming chat

---

### 1.4 Embedding Models

| Provider | Models | Dimension | Cost |
|----------|--------|-----------|------|
| **OpenAI** | text-embedding-3-small, text-embedding-3-large | 1536, 3072 | Low |
| **Hugging Face** | BAAI/bge-small, sentence-transformers | 384-1024 | Free |
| **Cohere** | embed-english-v3.0, embed-multilingual | 1024 | Medium |
| **Google** | textembedding-gecko | 768 | Medium |
| **Voyage AI** | voyage-01, voyage-02 | 1024 | Medium |
| **Azure OpenAI** | text-embedding-ada-002 | 1536 | Low |
| **Bedrock** | Amazon Titan Embeddings | 1536 | Low |

**Key Features:**
- Batch embedding generation
- Async embedding
- Caching support
- Similarity functions (cosine, euclidean, dot product)

---

## 2. Data Loading & Ingestion

### 2.1 Document Readers

| Reader | Formats | Use Case | Priority |
|--------|---------|----------|----------|
| **SimpleDirectoryReader** | All files in directory | General file loading | High |
| **PDFReader** | PDF documents | Document processing | High |
| **DocxReader** | Word documents | Document processing | High |
| **HTMLReader** | HTML files | Web content | High |
| **MarkdownReader** | Markdown files | Documentation | High |
| **JSONReader** | JSON files | Structured data | Medium |
| **CSVReader** | CSV files | Tabular data | Medium |
| **DatabaseReader** | SQL databases | Database integration | High |
| **NotionReader** | Notion pages | Notion integration | Medium |
| **GoogleDocsReader** | Google Docs | Google Workspace | Medium |
| **ConfluenceReader** | Confluence pages | Wiki/docs | Medium |
| **SlackReader** | Slack messages | Chat history | Medium |
| **DiscordReader** | Discord messages | Chat history | Low |
| **WebPageReader** | Web pages (BeautifulSoup) | Web scraping | High |
| **TrafilaturaReader** | Web pages (clean extraction) | Article extraction | High |
| **SeleniumReader** | Dynamic web pages | JavaScript sites | Medium |
| **YoutubeTranscriptReader** | YouTube videos | Video transcripts | Low |
| **GitHubReader** | GitHub repos | Code repositories | Medium |
| **S3Reader** | AWS S3 files | Cloud storage | Medium |
| **GCSReader** | Google Cloud Storage | Cloud storage | Low |
| **MongoReader** | MongoDB collections | NoSQL database | Medium |
| **ChromaReader** | Chroma collections | Vector DB import | Low |
| **PineconeReader** | Pinecone indexes | Vector DB import | Low |

**Total Readers:** 100+ official integrations

---

### 2.2 Node Parsers (Text Splitting)

| Parser | Strategy | Best For | Priority |
|--------|----------|----------|----------|
| **SentenceSplitter** | Sentence boundaries | General text | High |
| **TokenTextSplitter** | Token limits (GPT tokens) | LLM context limits | High |
| **CodeSplitter** | AST-based splitting | Source code | Medium |
| **SemanticSplitter** | Semantic similarity | Coherent chunks | High |
| **SimpleFileNodeParser** | File-level nodes | File metadata | Low |
| **HTMLNodeParser** | HTML structure | Web content | Medium |
| **MarkdownNodeParser** | Markdown sections | Documentation | Medium |
| **JSONNodeParser** | JSON structure | Structured data | Low |

**Key Features:**
- Configurable chunk size & overlap
- Metadata preservation
- Relationship tracking (parent/child nodes)
- Character/token counting

---

### 2.3 Ingestion Pipeline

| Feature | Description | Priority |
|---------|-------------|----------|
| **Sequential Transformations** | Apply parsers, extractors, embeddings in order | High |
| **Parallel Processing** | Process multiple documents concurrently | High |
| **Caching** | Cache embeddings to avoid re-computation | High |
| **Document Store** | Track processed documents (Redis, MongoDB) | Medium |
| **Deduplication** | Skip already-processed documents | High |
| **Vector Store Integration** | Direct ingestion into vector DB | High |
| **Metadata Extraction** | Auto-extract titles, summaries, keywords | Medium |
| **Error Handling** | Graceful failure handling per document | High |

**Example Pipeline:**
```
Documents → Parser → Metadata Extractor → Embeddings → Vector Store
```

---

### 2.4 Metadata Extractors

| Extractor | Extracts | Priority |
|-----------|----------|----------|
| **TitleExtractor** | Document titles from content | Medium |
| **KeywordExtractor** | Keywords/tags using LLM | Low |
| **SummaryExtractor** | Document summaries | Medium |
| **QuestionsAnsweredExtractor** | Q&A pairs from content | Low |
| **EntityExtractor** | Named entities (people, places, etc.) | Low |

---

## 3. Indexing & Storage

### 3.1 Index Types

| Index Type | Use Case | Priority | Complexity |
|------------|----------|----------|------------|
| **VectorStoreIndex** | Semantic search via embeddings | High | Medium |
| **SummaryIndex** | Summarization of all documents | Medium | Low |
| **TreeIndex** | Hierarchical tree structure | Low | High |
| **KeywordTableIndex** | Keyword-based lookup | Medium | Low |
| **KnowledgeGraphIndex** | Graph relationships | Low | High |
| **DocumentSummaryIndex** | Per-doc summaries + retrieval | Medium | Medium |
| **SQLTableIndex** | Query structured SQL data | High | Medium |
| **PandasIndex** | Query pandas DataFrames | Medium | Medium |

---

### 3.2 Vector Store Integrations

| Vector Store | Cloud/Local | Priority | Notes |
|--------------|-------------|----------|-------|
| **Qdrant** | Both | High | Rust-based, fast |
| **Pinecone** | Cloud | High | Managed service |
| **Chroma** | Local | High | Embedded, easy setup |
| **Weaviate** | Both | Medium | GraphQL API |
| **Milvus** | Both | Medium | Scalable |
| **FAISS** | Local | High | Facebook AI, in-memory |
| **Redis** | Both | Medium | Redis Vector Search |
| **Elasticsearch** | Both | Medium | Full-text + vector |
| **Postgres (pgvector)** | Local | High | SQL + vectors |
| **MongoDB Atlas** | Cloud | Medium | Managed MongoDB |
| **LanceDB** | Local | Low | Embedded, Rust |
| **Zilliz** | Cloud | Low | Managed Milvus |
| **Azure Cognitive Search** | Cloud | Medium | Azure integration |
| **Google Vertex AI** | Cloud | Low | GCP integration |

**Total:** 40+ vector store integrations

---

### 3.3 Document Stores

| Store | Use Case | Priority |
|-------|----------|----------|
| **SimpleDocumentStore** | In-memory | High |
| **MongoDocumentStore** | MongoDB persistence | Medium |
| **RedisDocumentStore** | Redis persistence | Medium |
| **FirestoreDocumentStore** | Google Firestore | Low |
| **DynamoDBDocumentStore** | AWS DynamoDB | Low |

---

### 3.4 Storage Context

**Features:**
- Unified interface for all storage components
- Persist/load indexes from disk
- Coordinate vector store, doc store, index store
- Enable multi-index scenarios

---

## 4. Retrieval

### 4.1 Retriever Types

| Retriever | Strategy | Priority |
|-----------|----------|----------|
| **VectorIndexRetriever** | Vector similarity search | High |
| **KeywordTableRetriever** | Keyword matching | Medium |
| **TreeRetriever** | Hierarchical traversal | Low |
| **KGRetriever** | Knowledge graph traversal | Low |
| **BM25Retriever** | BM25 ranking (sparse) | Medium |
| **HybridRetriever** | Dense + sparse fusion | High |
| **VectorIndexAutoRetriever** | Auto-filter by metadata | High |
| **RecursiveRetriever** | Multi-level retrieval | Medium |
| **RouterRetriever** | Route to best retriever | Medium |

---

### 4.2 Retrieval Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| **Default** | Retrieve embeddings | General |
| **Similarity** | Top-K similar nodes | Semantic search |
| **MMR** | Maximum Marginal Relevance | Diversity |
| **Hybrid** | Combine dense + sparse | Best results |

---

### 4.3 Advanced Retrieval Features

| Feature | Description | Priority |
|---------|-------------|----------|
| **Metadata Filtering** | Filter by attributes (date, author, etc.) | High |
| **Re-ranking** | Re-order results with separate model | Medium |
| **Query Transformation** | Rewrite query for better retrieval | High |
| **Hypothetical Document Embeddings (HyDE)** | Generate hypothetical answer → embed → search | Medium |
| **Multi-Query Retrieval** | Generate multiple queries from one | Medium |
| **Auto-Retrieval** | LLM generates filters + queries | High |

---

## 5. Query Engines

### 5.1 Query Engine Types

| Engine | Purpose | Priority |
|--------|---------|----------|
| **RetrieverQueryEngine** | Retrieve nodes → synthesize answer | High |
| **TreeQueryEngine** | Hierarchical querying | Low |
| **RouterQueryEngine** | Route to specialized engines | High |
| **SubQuestionQueryEngine** | Break into sub-questions | High |
| **SQLQueryEngine** | Natural language → SQL | High |
| **PandasQueryEngine** | Query pandas DataFrames | Medium |
| **JSONQueryEngine** | Query JSON data | Low |
| **CitationQueryEngine** | Include source citations | High |
| **RetrieverRouterQueryEngine** | Route to best retriever | Medium |
| **MultiStepQueryEngine** | Multi-step reasoning | Medium |

---

### 5.2 Response Synthesis

| Mode | Strategy | Priority |
|------|----------|----------|
| **Refine** | Iteratively refine answer with each chunk | Medium |
| **Compact** | Fit multiple chunks into context | High |
| **Tree Summarize** | Hierarchical summarization | Medium |
| **Simple Concatenate** | Concat all chunks + single LLM call | Low |
| **Accumulate** | Generate answer per chunk, return all | Low |

---

### 5.3 Query Transformations

| Transformation | Description | Priority |
|----------------|-------------|----------|
| **HyDE** | Generate hypothetical doc → search | Medium |
| **Multi-Step** | Decompose into steps | Medium |
| **Step Decompose** | Break into sequential steps | Medium |
| **Single-Step** | Query as-is | High |

---

## 6. Workflows

### 6.1 Workflow System

| Feature | Description | Priority |
|---------|-------------|----------|
| **Event-Driven Architecture** | Async event passing between steps | High |
| **Context Management** | Shared state across steps | High |
| **Step Definition** | Define workflow steps with decorators | High |
| **Conditional Branching** | Route based on conditions | High |
| **Parallel Execution** | Execute multiple branches concurrently | Medium |
| **Human-in-the-Loop** | Wait for human input | Medium |
| **Error Handling** | Retry, fallback strategies | High |
| **Streaming** | Stream events during execution | Medium |

---

### 6.2 Multi-Agent Patterns

| Pattern | Description | Priority |
|---------|-------------|----------|
| **Hand-off** | Agent hands task to another agent | High |
| **Planner + Executor** | Planner agent creates plan, executor agents run | High |
| **Hierarchical** | Manager agent coordinates worker agents | Medium |
| **Parallel** | Multiple agents work concurrently | Medium |
| **Sequential** | Agents process in sequence | High |
| **Debate** | Multiple agents debate before consensus | Low |

---

## 7. Evaluation & Observability

### 7.1 Evaluation

| Feature | Metric | Priority |
|---------|--------|----------|
| **Relevancy Evaluation** | Is answer relevant? | High |
| **Faithfulness Evaluation** | Is answer faithful to sources? | High |
| **Context Relevancy** | Are retrieved contexts relevant? | Medium |
| **Answer Correctness** | Compare to ground truth | High |
| **Semantic Similarity** | Embedding similarity to reference | Medium |
| **Custom Evaluators** | Define custom metrics | Medium |

---

### 7.2 Observability

| Feature | Description | Priority |
|---------|-------------|----------|
| **Callback Handlers** | Hook into execution events | High |
| **LlamaDebugHandler** | Debug all LLM calls | High |
| **OpenTelemetry Integration** | Distributed tracing | Medium |
| **WandB Integration** | Log to Weights & Biases | Low |
| **Langfuse Integration** | LLM observability platform | Medium |
| **Token Counting** | Track token usage | High |
| **Latency Tracking** | Measure response times | High |

---

## 8. Advanced Features

### 8.1 Chat & Memory

| Feature | Description | Priority |
|---------|-------------|----------|
| **Chat History** | Maintain conversation context | High |
| **Memory Buffers** | Store recent messages | High |
| **Summarization Memory** | Summarize old messages | Medium |
| **Entity Memory** | Track entities across conversation | Low |
| **Vector Memory** | Semantic memory retrieval | Medium |

---

### 8.2 Structured Outputs

| Feature | Description | Priority |
|---------|-------------|----------|
| **Pydantic Output Parsing** | Parse LLM output to Pydantic models | High |
| **JSON Mode** | Force JSON output | High |
| **Function Calling** | Structured function arguments | High |
| **Output Parsers** | Custom parsing logic | Medium |

---

### 8.3 Fine-tuning & Optimization

| Feature | Description | Priority |
|---------|-------------|----------|
| **Prompt Optimization** | Optimize prompts automatically | Low |
| **Fine-tune Embeddings** | Train custom embedding models | Low |
| **Knowledge Distillation** | Use GPT-4 to train GPT-3.5 | Low |

---

### 8.4 Multi-Modal

| Feature | Description | Priority |
|---------|-------------|----------|
| **Image Understanding** | GPT-4-Vision, Claude 3 Vision | Medium |
| **PDF with Images** | Extract images from PDFs | Medium |
| **Audio Transcription** | Whisper integration | Low |

---

## 9. Deployment & Production

### 9.1 Serving

| Feature | Description | Priority |
|---------|-------------|----------|
| **FastAPI Integration** | Deploy as REST API | High |
| **LlamaIndex Server** | Built-in HTTP server | Medium |
| **Streaming Endpoints** | Server-sent events | High |
| **Authentication** | API key auth | Medium |

---

### 9.2 Caching

| Feature | Description | Priority |
|---------|-------------|----------|
| **LLM Caching** | Cache LLM responses | High |
| **Embedding Caching** | Cache embeddings | High |
| **Redis Cache** | Production-ready caching | High |
| **In-memory Cache** | Development caching | High |

---

## 10. Feature Summary by Priority

### Must-Have (MVP)
1. **Agents:** FunctionAgent, ReActAgent
2. **LLMs:** OpenAI, Anthropic, Ollama
3. **Embeddings:** OpenAI, HuggingFace
4. **Tools:** QueryEngineTool, FunctionTool, Web Search, Database
5. **Readers:** PDF, Web, Markdown, CSV, Database
6. **Parsers:** SentenceSplitter, TokenTextSplitter
7. **Indexes:** VectorStoreIndex
8. **Vector Stores:** Qdrant, Pinecone, Chroma, Postgres
9. **Retrievers:** VectorIndexRetriever, HybridRetriever
10. **Query Engines:** RetrieverQueryEngine, SQLQueryEngine
11. **Ingestion Pipeline:** Sequential transformations, caching
12. **Observability:** Logging, token counting

### High Priority (Phase 2)
1. **Agents:** WorkflowAgent
2. **Multi-Agent:** Hand-off, Planner+Executor patterns
3. **Advanced Retrieval:** Auto-retrieval, HyDE, metadata filtering
4. **Query Engines:** SubQuestionQueryEngine, RouterQueryEngine
5. **Workflows:** Event system, context management
6. **Evaluation:** Relevancy, faithfulness metrics
7. **Chat Memory:** Conversation history, memory buffers
8. **Structured Outputs:** Pydantic parsing, JSON mode
9. **More Readers:** Google Docs, Notion, Confluence, GitHub
10. **Document Store:** Redis, MongoDB persistence

### Medium Priority (Phase 3)
1. **Agents:** Custom agent framework
2. **More LLMs:** Google, Cohere, AWS Bedrock
3. **More Vector Stores:** Weaviate, Milvus, Elasticsearch
4. **Advanced Query Engines:** Multi-step, citation engine
5. **Re-ranking:** Cohere re-rank, cross-encoder models
6. **Metadata Extractors:** Title, summary, entity extraction
7. **More Readers:** Slack, Discord, YouTube
8. **Multi-Modal:** Image understanding, audio transcription
9. **Deployment:** FastAPI server, streaming endpoints

### Low Priority (Future)
1. **Advanced Indexes:** TreeIndex, KnowledgeGraphIndex
2. **Fine-tuning:** Prompt optimization, custom embeddings
3. **Specialized Tools:** Code execution, complex integrations
4. **Advanced Patterns:** Debate agents, hierarchical agents

---

## Total Scope Estimate

| Category | Features | Effort (Weeks) |
|----------|----------|----------------|
| **Core (MVP)** | 30 features | 12-16 weeks |
| **High Priority** | 25 features | 8-12 weeks |
| **Medium Priority** | 20 features | 8-10 weeks |
| **Low Priority** | 15 features | 6-8 weeks |
| **Total** | ~90 features | **34-46 weeks** (8-11 months) |

---

## Conclusion

LlamaIndex is a **massive framework** with 90+ core features across 10 major categories. For GoAgent:

1. **MVP Focus:** Implement 30 must-have features first (12-16 weeks)
2. **Iterative Approach:** Ship v0.1 early, gather feedback
3. **80/20 Rule:** 30 core features cover 80% of use cases
4. **Community:** Build integrations as community contributions
5. **Differentiation:** Leverage Go's strengths (performance, concurrency, safety)

**Recommendation:** Start with MVP, prove value, then expand systematically.
