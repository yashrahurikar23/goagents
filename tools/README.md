# GoAgents Common Tools

This package provides commonly-needed tool implementations that work with the `core.Tool` interface.

## Philosophy

GoAgents follows a **two-tier architecture**:

1. **Core Package** (`goagents/core/`)
   - Defines the `Tool` interface
   - No tool implementations
   - Pure abstractions

2. **Tools Package** (`goagents/tools/`) ← **You are here**
   - Provides common tool implementations
   - Generic, reusable across domains
   - Optional - users can ignore and build their own

## What Goes Here?

### ✅ Include: Generic, commonly-needed operations
- File operations (read, write, list, delete)
- HTTP requests (GET, POST, PUT, DELETE)
- JSON/YAML parsing
- Text processing utilities
- Time/date operations

### ❌ Exclude: Domain-specific operations
- SQL queries (database-specific)
- Kubernetes operations (DevOps-specific)
- Code compilation (language-specific)
- Git operations (SCM-specific)

Domain-specific tools should be implemented by users in their own applications.

## Usage

### Option 1: Use Built-in Tools As-Is

```go
import (
    "github.com/yashrahurikar23/goagents/agent"
    "github.com/yashrahurikar23/goagents/tools"
)

// Create agent
reactAgent := agent.NewReActAgent(llmClient)

// Use built-in file tools
fileRead := tools.NewFileReadTool("/path/to/workspace", 10240)
fileWrite := tools.NewFileWriteTool("/path/to/workspace", 10240, true)

reactAgent.AddTool(fileRead)
reactAgent.AddTool(fileWrite)
```

### Option 2: Extend Built-in Tools

```go
import "github.com/yashrahurikar23/goagents/tools"

// Extend with custom behavior
type ValidatedFileWriteTool struct {
    *tools.FileWriteTool  // Embed built-in tool
    validator FileValidator
}

func (t *ValidatedFileWriteTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // Add custom validation
    if err := t.validator.Validate(args); err != nil {
        return nil, err
    }
    // Call embedded tool
    return t.FileWriteTool.Execute(ctx, args)
}
```

### Option 3: Build Your Own Tools

```go
import "github.com/yashrahurikar23/goagents/core"

// Implement core.Tool interface from scratch
type MyCustomTool struct {
    // Your fields
}

func (t *MyCustomTool) Name() string { return "my_tool" }
func (t *MyCustomTool) Description() string { return "..." }
func (t *MyCustomTool) Schema() *core.ToolSchema { return ... }
func (t *MyCustomTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    // Your implementation
}
```

## Available Tools

### File Operations (`file.go`)

#### FileReadTool
Reads file contents with size limits and security checks.

```go
tool := tools.NewFileReadTool(
    workingDir,  // Base directory
    10240,       // Max chars (10KB)
)
```

**Parameters:**
- `file_path` (string, required): Relative path to file

**Security:**
- Path traversal prevention
- Working directory scoping
- Size limits

#### FileWriteTool
Creates or modifies files with automatic backups.

```go
tool := tools.NewFileWriteTool(
    workingDir,    // Base directory
    10240,         // Max content chars (10KB)
    true,          // Create backups
)
```

**Parameters:**
- `file_path` (string, required): Relative path to file
- `content` (string, required): File content to write

**Features:**
- Automatic backup of existing files
- Directory creation if needed
- Rollback on error

#### FileListTool
Lists files and directories with metadata.

```go
tool := tools.NewFileListTool(workingDir)
```

**Parameters:**
- `directory` (string, optional): Relative path to directory (default: ".")

**Returns:**
- File names, sizes, and types

## Security

All file tools implement security best practices:

- **Path Traversal Prevention**: Blocks `../`, absolute paths, symlink attacks
- **Working Directory Scoping**: Can only access files within configured directory
- **Size Limits**: Prevents memory exhaustion
- **Safe Defaults**: Secure by default configuration

## Testing

Run tests:
```bash
cd goagents/tools
go test -v
go test -race
go test -cover
```

## Examples

See `examples/` directory in the repository for complete examples of using these tools with different agent types.

## Design Principles

1. **Generic First**: Tools should work across multiple domains
2. **Secure by Default**: All tools include security validations
3. **Extensible**: Easy to extend via embedding or composition
4. **Well-Documented**: Clear parameter schemas and error messages
5. **Well-Tested**: Comprehensive tests including security edge cases

## Contributing

When adding new tools to this package, ask:

1. **Is it generic?** Would multiple types of agents need this?
2. **Is it safe?** Does it include proper security validations?
3. **Is it tested?** Does it have comprehensive test coverage?
4. **Is it documented?** Are parameters and behavior clear?

If yes to all four, it belongs here. If no, it should be in a user's application-specific code.
