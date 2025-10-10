# FreeCodeCamp Approach: Building an AI Coding Agent with Python and Gemini

**Source**: [FreeCodeCamp Tutorial](https://www.freecodecamp.org/news/build-an-ai-coding-agent-with-python-and-gemini/)

**Author**: Lane Wagner (Boot.dev)

**Date**: October 2, 2025

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Implementation Details](#implementation-details)
4. [Security Measures](#security-measures)
5. [Key Takeaways](#key-takeaways)
6. [Code Examples](#code-examples)

---

## Overview

### What It Builds

A command-line tool that:
- Accepts coding tasks in natural language
- Autonomously scans, reads, writes, and executes code
- Iterates until the task is complete
- Works with Google's free Gemini API

### Example Usage

```bash
python main.py "fix my calculator app, it's not starting correctly"
```

Output:
```
# Calling function: get_files_info
# Calling function: get_file_content
# Calling function: write_file
# Calling function: run_python_file
# Calling function: write_file
# Calling function: run_python_file
# Final response:
Great! The calculator app now seems to be working correctly.
```

### Design Philosophy

- **Simple but effective**: Minimal complexity, focuses on core functionality
- **Educational**: Designed to teach how AI agents work under the hood
- **Practical**: Solves real problems like bug fixing
- **Transparent**: Verbose mode shows all decision-making

---

## Architecture

### High-Level Flow

```
┌─────────────────────────────────────────────────────────┐
│  User Prompt: "Fix bug in calculator"                   │
└───────────────────┬─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│  System Prompt + Function Declarations                  │
│  (Tells LLM what functions it can call)                 │
└───────────────────┬─────────────────────────────────────┘
                    ↓
        ┌───────────────────────┐
        │  Gemini-2.0-Flash     │
        │  (LLM Decision Engine) │
        └──────────┬────────────┘
                   ↓
    ┌──────────────┴──────────────┐
    │  Function Call Decision      │
    │  get_files_info(".")         │
    └──────────────┬──────────────┘
                   ↓
    ┌──────────────────────────────┐
    │  Execute Function            │
    │  Return: ["main.py", ...]    │
    └──────────────┬───────────────┘
                   ↓
    ┌──────────────────────────────┐
    │  Add to Conversation History │
    └──────────────┬───────────────┘
                   ↓
        ┌──────────────────┐
        │  Next LLM Call   │
        │  (Loop up to 20x) │
        └──────────┬────────┘
                   ↓
        ┌──────────────────┐
        │  Final Response  │
        │  or Max Iterations│
        └──────────────────┘
```

### Core Components

1. **Gemini Client**: Google AI API wrapper
2. **Function Definitions**: Schema describing available tools
3. **Conversation Manager**: Maintains message history
4. **Function Executor**: Routes function calls to implementations
5. **Agent Loop**: Iterative execution with max iterations

---

## Implementation Details

### 1. Functions (Tools)

#### get_files_info

```python
def get_files_info(working_directory, directory="."):
    """Lists files with metadata."""
    abs_working_dir = os.path.abspath(working_directory)
    target_dir = os.path.abspath(os.path.join(working_directory, directory))
    
    # Security: Ensure within working directory
    if not target_dir.startswith(abs_working_dir):
        return f'Error: Cannot list "{directory}" as it is outside the permitted working directory'
    
    # List files with sizes
    files_info = []
    for filename in os.listdir(target_dir):
        filepath = os.path.join(target_dir, filename)
        file_size = os.path.getsize(filepath)
        is_dir = os.path.isdir(filepath)
        files_info.append(f"- {filename}: file_size={file_size} bytes, is_dir={is_dir}")
    
    return "\n".join(files_info)
```

**Key Features:**
- Returns formatted string (LLM-friendly)
- Security: Path validation
- Metadata: Size and type info

#### get_file_content

```python
def get_file_content(working_directory, file_path):
    """Reads file content with truncation."""
    abs_file_path = os.path.abspath(os.path.join(working_directory, file_path))
    
    # Security checks
    if not abs_file_path.startswith(abs_working_dir):
        return f'Error: Cannot read "{file_path}" as it is outside the permitted working directory'
    
    if not os.path.isfile(abs_file_path):
        return f'Error: File not found or is not a regular file: "{file_path}"'
    
    # Read with truncation
    with open(abs_file_path, "r") as f:
        content = f.read(MAX_CHARS)  # MAX_CHARS = 10000
        if os.path.getsize(abs_file_path) > MAX_CHARS:
            content += f'[...File "{file_path}" truncated at {MAX_CHARS} characters]'
    
    return content
```

**Key Features:**
- Truncation at 10,000 characters (prevent token overflow)
- Clear truncation message
- Error handling returns strings (not exceptions)

#### write_file

```python
def write_file(working_directory, file_path, content):
    """Writes/overwrites file content."""
    abs_file_path = os.path.abspath(os.path.join(working_directory, file_path))
    
    # Security check
    if not abs_file_path.startswith(abs_working_dir):
        return f'Error: Cannot write to "{file_path}" as it is outside the permitted working directory'
    
    # Create directory if needed
    if not os.path.exists(abs_file_path):
        os.makedirs(os.path.dirname(abs_file_path), exist_ok=True)
    
    # Write file
    with open(abs_file_path, "w") as f:
        f.write(content)
    
    return f'Successfully wrote to "{file_path}" ({len(content)} characters written)'
```

**Key Features:**
- Auto-creates directories
- Returns success confirmation (feedback to LLM)
- Character count in response

#### run_python_file

```python
def run_python_file(working_directory, file_path, args=None):
    """Executes Python file with timeout."""
    abs_file_path = os.path.abspath(os.path.join(working_directory, file_path))
    
    # Security checks
    if not abs_file_path.startswith(abs_working_dir):
        return f'Error: Cannot execute "{file_path}" as it is outside the permitted working directory'
    
    if not file_path.endswith(".py"):
        return f'Error: "{file_path}" is not a Python file.'
    
    # Execute with timeout
    commands = ["python", abs_file_path]
    if args:
        commands.extend(args)
    
    result = subprocess.run(
        commands,
        capture_output=True,
        text=True,
        timeout=30,  # 30-second timeout
        cwd=abs_working_dir
    )
    
    # Format output
    output = []
    if result.stdout:
        output.append(f"STDOUT:\n{result.stdout}")
    if result.stderr:
        output.append(f"STDERR:\n{result.stderr}")
    if result.returncode != 0:
        output.append(f"Process exited with code {result.returncode}")
    
    return "\n".join(output) if output else "No output produced."
```

**Key Features:**
- 30-second timeout (prevents infinite loops)
- Captures stdout and stderr
- Returns exit code if non-zero
- Passes optional arguments

### 2. Function Declarations (Schema)

```python
from google.genai import types

schema_get_files_info = types.FunctionDeclaration(
    name="get_files_info",
    description="Lists files in the specified directory along with their sizes, constrained to the working directory.",
    parameters=types.Schema(
        type=types.Type.OBJECT,
        properties={
            "directory": types.Schema(
                type=types.Type.STRING,
                description="The directory to list files from, relative to the working directory. If not provided, lists files in the working directory itself.",
            ),
        },
    ),
)

schema_get_file_content = types.FunctionDeclaration(
    name="get_file_content",
    description=f"Reads and returns the first {MAX_CHARS} characters of the content from a specified file within the working directory.",
    parameters=types.Schema(
        type=types.Type.OBJECT,
        properties={
            "file_path": types.Schema(
                type=types.Type.STRING,
                description="The path to the file whose content should be read, relative to the working directory.",
            ),
        },
        required=["file_path"],
    ),
)

# Similar declarations for write_file and run_python_file...

available_functions = types.Tool(
    function_declarations=[
        schema_get_files_info,
        schema_get_file_content,
        schema_run_python_file,
        schema_write_file,
    ]
)
```

**Key Points:**
- Clear descriptions help LLM choose correctly
- Parameters specify types and descriptions
- `required` array enforces mandatory parameters

### 3. System Prompt

```python
system_prompt = """
You are a helpful AI coding agent.

When a user asks a question or makes a request, make a function call plan. You can perform the following operations:
- List files and directories
- Read file contents
- Execute Python files with optional arguments
- Write or overwrite files

All paths you provide should be relative to the working directory. You do not need to specify the working directory in your function calls as it is automatically injected for security reasons.
"""
```

**Key Elements:**
- Clear role definition
- Explicit capabilities list
- Security note (working directory handled automatically)

### 4. Agent Loop

```python
def generate_content_loop(client, messages, verbose, max_iterations=20):
    for iteration in range(max_iterations):
        try:
            # Call LLM with conversation history
            response = client.models.generate_content(
                model="gemini-2.0-flash-001",
                contents=messages,
                config=types.GenerateContentConfig(
                    tools=[available_functions],
                    system_instruction=system_prompt
                ),
            )
            
            # Add model's response to conversation
            for candidate in response.candidates:
                messages.append(candidate.content)
            
            # Check if we have final text response
            if response.text:
                print("Final response:")
                print(response.text)
                break
            
            # Handle function calls
            if response.function_calls:
                function_responses = []
                for function_call_part in response.function_calls:
                    # Execute function
                    function_call_result = call_function(function_call_part, verbose)
                    function_responses.append(function_call_result.parts[0])
                
                # Add function results to conversation
                if function_responses:
                    messages.append(types.Content(role="user", parts=function_responses))
        
        except Exception as e:
            print(f"Error: {e}")
            break
    else:
        print(f"Reached maximum iterations ({max_iterations}). Agent may not have completed the task.")
```

**Key Features:**
- Max 20 iterations (prevents infinite loops)
- Maintains full conversation history
- Handles multiple function calls per iteration
- Graceful error handling
- Exit when final text response received

### 5. Message Structure

```python
# Initial message
messages = [
    types.Content(role="user", parts=[types.Part(text=user_prompt)]),
]

# After LLM response (function call)
# System automatically adds:
messages.append(candidate.content)  # LLM's function call

# After function execution
# We add:
messages.append(types.Content(
    role="user",  # Note: Function results sent as "user" role
    parts=[function_response]
))
```

**Conversation Example:**
```
[User] "Fix the calculator bug"
[Model] I want to call get_files_info({"directory": "."})
[User/Tool] Result: ["calculator.py", "tests.py"]
[Model] I want to call get_file_content({"file_path": "calculator.py"})
[User/Tool] Result: [file contents]
[Model] I want to call write_file({"file_path": "calculator.py", "content": "..."})
[User/Tool] Result: "Successfully wrote 150 characters"
[Model] "Fixed the division by zero bug in calculator.py"
```

---

## Security Measures

### 1. Working Directory Scoping

**Problem**: LLM could access sensitive files anywhere on the system

**Solution**: All functions validate paths stay within working directory

```python
abs_working_dir = os.path.abspath(working_directory)
abs_file_path = os.path.abspath(os.path.join(working_directory, file_path))

if not abs_file_path.startswith(abs_working_dir):
    return f'Error: Cannot access "{file_path}" as it is outside the permitted working directory'
```

**Protection Against:**
- `../../../etc/passwd` (path traversal)
- `/absolute/path/to/sensitive/file`
- Symbolic links outside working directory

### 2. Execution Timeout

**Problem**: Infinite loops or hanging processes

**Solution**: 30-second timeout on all code execution

```python
result = subprocess.run(
    commands,
    capture_output=True,
    text=True,
    timeout=30,  # Hard stop after 30 seconds
    cwd=abs_working_dir
)
```

### 3. File Size Limits

**Problem**: Reading huge files could exhaust tokens/memory

**Solution**: Truncate at 10,000 characters

```python
MAX_CHARS = 10000

content = f.read(MAX_CHARS)
if os.path.getsize(abs_file_path) > MAX_CHARS:
    content += f'[...File "{file_path}" truncated at {MAX_CHARS} characters]'
```

### 4. Iteration Limits

**Problem**: Agent could loop forever

**Solution**: Max 20 iterations

```python
for iteration in range(max_iterations):  # max_iterations = 20
    # ... agent logic ...
else:
    print(f"Reached maximum iterations ({max_iterations}).")
```

### 5. Python-Only Execution

**Problem**: Arbitrary shell command execution

**Solution**: Only `.py` files allowed

```python
if not file_path.endswith(".py"):
    return f'Error: "{file_path}" is not a Python file.'
```

**Limitations Acknowledged:**
- Still risky for production use
- For learning purposes only
- Don't distribute to untrusted users

---

## Key Takeaways

### What Works Well

✅ **Simplicity**: Minimal moving parts, easy to understand
✅ **Function Calling**: Natural way for LLM to use tools
✅ **Conversation History**: Context accumulates naturally
✅ **Gemini Free Tier**: Generous limits for learning
✅ **Verbose Mode**: Great for debugging and learning

### Limitations

❌ **No Advanced Planning**: Purely reactive, no lookahead
❌ **Context Window**: Can fill up with large files
❌ **Security**: Not production-ready (acknowledged)
❌ **Error Recovery**: Limited backtracking capability
❌ **Single Language**: Python-only execution

### Lessons for GoAgents

1. **Clear Function Descriptions**: Help LLM choose correctly
2. **String-Based Returns**: Always return strings from tools (never raise exceptions)
3. **Security by Default**: Validate all paths, limit execution
4. **Feedback Messages**: Return detailed success/error messages
5. **Iteration Limits**: Always set max iterations
6. **Token Awareness**: Truncate large inputs

### Cost Efficiency

- **Typical Run**: 5-10 function calls
- **Cost**: ~$0.50-$1.00 per task (GPT-4 pricing)
- **Gemini Free Tier**: 15 requests/minute, 1500 requests/day
- **Optimization**: Cache working directory listing, minimize file reads

---

## Code Examples

### Complete Main Flow

```python
import sys
import os
from google import genai
from google.genai import types
from dotenv import load_dotenv

def main():
    load_dotenv()
    
    # Parse arguments
    verbose = "--verbose" in sys.argv
    user_prompt = " ".join([arg for arg in sys.argv[1:] if not arg.startswith("--")])
    
    if not user_prompt:
        print('Usage: python main.py "your prompt here" [--verbose]')
        sys.exit(1)
    
    # Initialize client
    api_key = os.environ.get("GEMINI_API_KEY")
    client = genai.Client(api_key=api_key)
    
    # Start conversation
    messages = [
        types.Content(role="user", parts=[types.Part(text=user_prompt)]),
    ]
    
    # Run agent loop
    generate_content_loop(client, messages, verbose)

if __name__ == "__main__":
    main()
```

### Function Executor

```python
def call_function(function_call_part, verbose=False):
    if verbose:
        print(f" - Calling function: {function_call_part.name}({function_call_part.args})")
    else:
        print(f" - Calling function: {function_call_part.name}")
    
    # Map function names to implementations
    function_map = {
        "get_files_info": get_files_info,
        "get_file_content": get_file_content,
        "run_python_file": run_python_file,
        "write_file": write_file,
    }
    
    function_name = function_call_part.name
    if function_name not in function_map:
        return types.Content(
            role="tool",
            parts=[
                types.Part.from_function_response(
                    name=function_name,
                    response={"error": f"Unknown function: {function_name}"},
                )
            ],
        )
    
    # Execute function
    args = dict(function_call_part.args)
    args["working_directory"] = WORKING_DIR  # Inject working directory
    function_result = function_map[function_name](**args)
    
    # Return result
    return types.Content(
        role="tool",
        parts=[
            types.Part.from_function_response(
                name=function_name,
                response={"result": function_result},
            )
        ],
    )
```

---

## Comparison to GoAgents

| Feature | FreeCodeCamp Approach | GoAgents Equivalent |
|---------|----------------------|---------------------|
| **LLM Provider** | Gemini (hardcoded) | OpenAI, Anthropic, Gemini, Ollama |
| **Agent Pattern** | Custom loop | ReAct Agent, Function Agent |
| **File Operations** | Custom functions | `tools.FileOperationsTool` |
| **Security** | Path validation, timeouts | 5-layer security system |
| **Function Calling** | Gemini function calling | Universal function calling interface |
| **Language** | Python | Go |
| **Testing** | Manual | Comprehensive test suite |

### What GoAgents Adds

- **Provider Abstraction**: Switch LLMs without changing agent code
- **Enhanced Security**: Read-only mode, file size limits, permission checks
- **Agent Patterns**: Pre-built ReAct, Function, Conversational agents
- **Production Ready**: Comprehensive error handling, logging
- **Type Safety**: Go's type system prevents many runtime errors
- **Extensibility**: Easy to add new tools and agents

---

## Next Steps

- See `03_DEEPSENSE_APPROACH.md` for more advanced planning strategies
- See `05_GOAGENTS_IMPLEMENTATION.md` for how to implement this with GoAgents
- See `06_BEST_PRACTICES.md` for production considerations
