# File Operations Tool Example

This example demonstrates the **File Operations Tool** in GoAgents, which allows AI agents to safely interact with the file system.

## Features

The File Operations Tool provides:

- **Read** - Read file contents
- **Write** - Create or overwrite files
- **Append** - Add content to existing files
- **List** - List directory contents
- **Exists** - Check if files exist
- **Delete** - Remove files
- **Info** - Get file metadata (size, type, permissions, modified time)

## Security Features

The tool includes multiple security layers:

1. **Path Traversal Prevention** - Blocks `../` attempts to escape base directory
2. **Base Directory Enforcement** - All operations confined to specified directory
3. **File Size Limits** - Prevents reading/writing excessively large files
4. **Read-Only Mode** - Optional mode that disables write operations
5. **Directory Protection** - Prevents accidental deletion of directories

## Setup

### Prerequisites

- Go 1.22 or later
- Ollama running locally with a small model (e.g., `gemma3:270m`)

### Install Ollama Model

```bash
ollama pull gemma3:270m
```

## Running the Example

```bash
cd examples/file_operations
go run main.go
```

## What It Does

The example demonstrates 8 scenarios:

1. **Create Shopping List** - Agent creates a file and writes shopping items
2. **Read and Count** - Agent reads the file and counts items
3. **Append Items** - Agent adds more items to the list
4. **List Directory** - Agent lists all files in the workspace
5. **Check Existence** - Agent checks if a file exists
6. **Get File Info** - Agent retrieves file metadata
7. **Security Demo** - Shows path traversal prevention in action
8. **Read-Only Mode** - Demonstrates write operation blocking

## Expected Output

```
🗂️  GoAgents - File Operations Tool Demo
========================================

📁 Working in: /tmp/goagents-file-demo-xxxxx

Example 1: Create a Shopping List
----------------------------------
Agent: I've created shopping-list.txt with your items: milk, bread, eggs, cheese

Example 2: Read Shopping List
-----------------------------
Agent: The shopping list contains 4 items

Example 3: Add More Items
-------------------------
Agent: I've added apples and bananas to the shopping list

...
```

## Configuration Options

### File Tool Options

```go
fileTool, err := tools.NewFileTool(
    tools.WithBaseDir("/path/to/workspace"),  // Set base directory
    tools.WithAllowWrite(true),               // Enable/disable writes
    tools.WithMaxSize(1024*1024),             // Max file size (1MB)
)
```

### Common Configurations

**Full Access (Development)**
```go
fileTool, _ := tools.NewFileTool(
    tools.WithBaseDir("./workspace"),
    tools.WithAllowWrite(true),
    tools.WithMaxSize(10 * 1024 * 1024), // 10MB
)
```

**Read-Only (Production)**
```go
fileTool, _ := tools.NewFileTool(
    tools.WithBaseDir("/data/readonly"),
    tools.WithAllowWrite(false),
    tools.WithMaxSize(5 * 1024 * 1024), // 5MB
)
```

**Restricted Writing (Sandbox)**
```go
fileTool, _ := tools.NewFileTool(
    tools.WithBaseDir("/tmp/sandbox"),
    tools.WithAllowWrite(true),
    tools.WithMaxSize(100 * 1024), // 100KB
)
```

## Use Cases

### Documentation Assistant
```go
// Agent can read project files, analyze structure, write documentation
agent.Run(ctx, "Read README.md and create a CONTRIBUTING.md file")
```

### Log Analyzer
```go
// Agent can analyze log files (read-only mode recommended)
agent.Run(ctx, "Read server.log and summarize errors from the last hour")
```

### Code Generator
```go
// Agent can write code files based on specifications
agent.Run(ctx, "Create a Go file called 'handler.go' with a basic HTTP handler")
```

### File Organizer
```go
// Agent can list, read, and organize files
agent.Run(ctx, "List all .txt files and create an index.md with their names")
```

## Safety Considerations

1. **Always set a base directory** - Never use root `/` as base
2. **Use read-only mode when possible** - Especially for production environments
3. **Set appropriate file size limits** - Prevent memory exhaustion
4. **Monitor agent operations** - Log file access for security audits
5. **Test path traversal protection** - Verify security before deployment

## Integration with Other Tools

The File Tool works great with other GoAgents tools:

```go
// Combine with Calculator for data processing
agent.AddTool(fileTool)
agent.AddTool(calculatorTool)
agent.Run(ctx, "Read numbers.txt and calculate their sum")

// Combine with HTTP for data fetching
agent.AddTool(fileTool)
agent.AddTool(httpTool)
agent.Run(ctx, "Fetch data from API and save it to data.json")
```

## Troubleshooting

### "path traversal not allowed"
- The agent attempted to access files outside the base directory
- This is expected security behavior
- Configure the base directory to include needed files

### "write operations are disabled"
- The tool is in read-only mode
- Set `WithAllowWrite(true)` if writes are needed
- Verify this is intentional for security

### "file too large"
- File exceeds the configured max size
- Increase `WithMaxSize()` if appropriate
- Consider streaming or chunking for large files

### "base directory does not exist"
- The specified base directory doesn't exist
- Create the directory before initializing the tool
- Verify the path is correct

## Learn More

- [GoAgents Documentation](../../docs/README.md)
- [Tool Development Guide](../../docs/guides/BEST_PRACTICES.md)
- [Security Best Practices](../../docs/guides/BEST_PRACTICES.md#security)
- [ReAct Agent Guide](../../docs/guides/AGENT_ARCHITECTURES.md#react-agent)

## License

MIT License - See [LICENSE](../../LICENSE) for details
