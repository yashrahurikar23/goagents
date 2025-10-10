# Architecture Patterns for Coding Agents

This document describes common architectural patterns used in successful coding agent implementations.

## Table of Contents

1. [Overview](#overview)
2. [Pattern 1: Iterative Function Calling](#pattern-1-iterative-function-calling)
3. [Pattern 2: Upfront Planning with Execution](#pattern-2-upfront-planning-with-execution)
4. [Pattern 3: Incremental Planning](#pattern-3-incremental-planning)
5. [Pattern 4: Multi-Agent Collaboration](#pattern-4-multi-agent-collaboration)
6. [Comparison Matrix](#comparison-matrix)

## Overview

Coding agents can be architected in several ways, each with distinct advantages and trade-offs. The choice depends on:
- Task complexity
- Available context window size
- Cost constraints
- Required flexibility
- Codebase size

---

## Pattern 1: Iterative Function Calling

**Used by**: FreeCodeCamp tutorial, basic GPT Engineer implementations

### Architecture

```
User Input → LLM (with function definitions) → Function Call Decision
                ↑                                        ↓
                └────────── Execute Function ───────────┘
                           (repeat until done)
```

### How It Works

1. User provides high-level goal
2. LLM has access to predefined functions (tools)
3. LLM decides which function to call with what arguments
4. Function executes, result returned to LLM
5. LLM decides next action based on result
6. Repeats until task complete or max iterations reached

### Functions Typically Provided

```
- get_files_info(directory)         # List directory contents
- get_file_content(file_path)       # Read file contents
- write_file(file_path, content)    # Write/overwrite file
- run_python_file(file_path, args)  # Execute code
- finish()                          # Signal completion
```

### Example Flow

```
User: "Fix the bug in calculator.py where division by zero crashes"

1. LLM → get_files_info(".")
   Result: ["calculator.py", "test_calculator.py"]

2. LLM → get_file_content("calculator.py")
   Result: [file contents with bug]

3. LLM → write_file("calculator.py", [fixed content])
   Result: "Successfully wrote 150 characters"

4. LLM → run_python_file("test_calculator.py")
   Result: "All tests passed"

5. LLM → "Bug fixed! Added zero division check."
```

### Pros
✅ Simple to implement
✅ Flexible - adapts to unexpected situations
✅ No complex planning logic needed
✅ Works well for small, focused tasks

### Cons
❌ Can get lost without clear direction
❌ May waste iterations exploring
❌ No lookahead - purely reactive
❌ Higher token usage (many LLM calls)

### Best For
- Bug fixes
- Small feature additions
- Exploratory tasks
- Simple refactoring

---

## Pattern 2: Upfront Planning with Execution

**Used by**: Deepsense.ai Approach A, MetaGPT

### Architecture

```
User Input → Planner Agent → Complete Plan (JSON)
                                    ↓
              ┌─────────────────────┴─────────────────────┐
              ↓                     ↓                     ↓
         Task 1 Agent          Task 2 Agent          Task 3 Agent
         (CodeWriter)          (TestWriter)          (Refactorer)
              ↓                     ↓                     ↓
         File 1.py             test_1.py             File 1.py (v2)
```

### How It Works

1. **Planning Phase**: Planner agent generates complete task breakdown
2. **Execution Phase**: Each task executed sequentially by specialized agents
3. **Context Management**: Previous task outputs added to context for next task

### Planning Structure

```json
{
  "tasks": [
    {
      "id": 1,
      "type": "write_code",
      "description": "Implement data loading function",
      "output_file": "data_loader.py",
      "prerequisites": []
    },
    {
      "id": 2,
      "type": "write_code",
      "description": "Implement model training",
      "output_file": "train.py",
      "prerequisites": ["data_loader.load_data()"]
    },
    {
      "id": 3,
      "type": "write_tests",
      "description": "Write unit tests",
      "output_file": "test_data_loader.py",
      "prerequisites": []
    }
  ]
}
```

### Agent Roles

1. **Planner**: Creates task breakdown
2. **CodeWriter**: Writes implementation code
3. **TestWriter**: Writes unit tests
4. **CodeRefactorer**: Improves code quality
5. **AdditionalFilesWriter**: Creates README, requirements.txt, etc.

### Pros
✅ Clear structure and organization
✅ Efficient - fewer LLM calls
✅ Specialized agents for specific tasks
✅ Predictable execution flow
✅ Good for larger projects

### Cons
❌ Inflexible - can't adapt if plan is wrong
❌ Planning failures cascade
❌ May generate invalid plans
❌ Struggles with unexpected requirements

### Best For
- Well-defined projects
- Multi-file applications
- Data science pipelines
- Projects with clear requirements

---

## Pattern 3: Incremental Planning

**Used by**: Deepsense.ai Approach B, some AutoGPT variants

### Architecture

```
User Input → Initial State
                ↓
      ┌─────────┴─────────┐
      ↓                   ↓
Action Agent       Check: Complete?
(decides next action)     ↓ No
      ↓                   ↓
Execute Action ───────────┘
(CREATE, EDIT, DELETE, PEEK, FINISH)
      ↓ FINISH
   Complete
```

### How It Works

1. Agent starts with initial problem description
2. Decides next action based on current state and history
3. Executes action (create file, edit file, run code, etc.)
4. Updates state with action result
5. Repeats until agent chooses FINISH

### Available Actions

```
- CREATE <filename>    # Create new file
- EDIT <filename>      # Modify existing file
- DELETE <filename>    # Remove file
- PEEK <filename>      # Read file (adds to context)
- RUN <filename>       # Execute code
- FINISH               # Task complete
```

### Example Flow

```
State 0: "Build a calculator CLI"

Action 1: CREATE calculator.py
State 1: calculator.py exists (empty)

Action 2: EDIT calculator.py
State 2: calculator.py has basic add/subtract

Action 3: PEEK calculator.py
State 3: Context now includes calculator.py code

Action 4: EDIT calculator.py
State 4: calculator.py now has multiply/divide

Action 5: CREATE test_calculator.py
State 5: test file created

Action 6: RUN test_calculator.py
State 6: Tests fail - division by zero bug

Action 7: EDIT calculator.py
State 7: Bug fixed

Action 8: RUN test_calculator.py
State 8: All tests pass

Action 9: FINISH
```

### Pros
✅ Very flexible - adapts to discoveries
✅ Mimics human developer workflow
✅ Can recover from mistakes
✅ Explores and adjusts strategy

### Cons
❌ Prone to infinite loops
❌ May take inefficient paths
❌ Needs careful action selection
❌ Higher token usage
❌ Requires robust state management

### Best For
- Exploratory development
- Debugging unknown issues
- Adaptive refactoring
- Research projects

---

## Pattern 4: Multi-Agent Collaboration

**Used by**: MetaGPT, advanced implementations

### Architecture

```
                    Project Manager
                    (orchestrates)
                           ↓
        ┌──────────────────┼──────────────────┐
        ↓                  ↓                  ↓
    Architect          Engineer          QA Engineer
 (designs system)   (writes code)      (writes tests)
        ↓                  ↓                  ↓
   design_doc.md      src/*.py          tests/*.py
                           ↓
                    Code Reviewer
                    (reviews & suggests)
```

### How It Works

1. **Project Manager**: Breaks down requirements into work items
2. **Architect**: Designs system structure, defines interfaces
3. **Engineer**: Implements features based on design
4. **QA Engineer**: Writes comprehensive tests
5. **Code Reviewer**: Reviews code quality, suggests improvements
6. **Iterative Refinement**: Agents collaborate until quality bar met

### Communication Protocol

Agents communicate through structured documents:

```
design_doc.md       # Architect's output
implementation.md   # Engineer's plan
code_review.md      # Reviewer's feedback
test_results.txt    # QA's output
```

### Pros
✅ Mimics real software teams
✅ Built-in quality checks
✅ Separation of concerns
✅ Each agent specialized
✅ Produces documentation

### Cons
❌ Complex orchestration
❌ Higher cost (many LLM calls)
❌ Longer execution time
❌ Coordination overhead
❌ Overkill for simple tasks

### Best For
- Complex applications
- Production-grade code
- Team collaboration simulation
- High-quality requirements

---

## Comparison Matrix

| Pattern | Flexibility | Efficiency | Complexity | Best Task Size | Token Cost |
|---------|-------------|------------|------------|----------------|------------|
| **Iterative Function Calling** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐ | Small | Medium |
| **Upfront Planning** | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Medium-Large | Low |
| **Incremental Planning** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐ | Small-Medium | High |
| **Multi-Agent Collaboration** | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | Large | Very High |

---

## Hybrid Approaches

Real-world implementations often combine patterns:

### Example: Planning + Iterative

```
1. Generate high-level plan (5-7 major tasks)
2. For each task, use iterative function calling
3. Allow plan modification if task fails
```

### Example: Multi-Agent + Incremental

```
1. Project Manager assigns tasks to agents
2. Each agent uses incremental planning for their task
3. Agents communicate results back to manager
```

---

## Choosing the Right Pattern

### Decision Tree

```
Is task well-defined and clear?
├─ Yes → Is it large (10+ files)?
│         ├─ Yes → Multi-Agent Collaboration
│         └─ No  → Upfront Planning
│
└─ No  → Is exploration needed?
          ├─ Yes → Incremental Planning
          └─ No  → Iterative Function Calling
```

### Quick Guide

**Use Iterative Function Calling when:**
- Task is small and focused
- Requirements are clear but simple
- Fast results needed
- Cost-conscious

**Use Upfront Planning when:**
- Building complete applications
- Requirements are comprehensive
- Efficiency is priority
- Multiple files needed

**Use Incremental Planning when:**
- Requirements are unclear
- Debugging complex issues
- Exploratory development
- Need maximum flexibility

**Use Multi-Agent Collaboration when:**
- Large, complex projects
- Quality is paramount
- Have budget for high token usage
- Need comprehensive documentation

---

## Implementation Tips

### All Patterns

1. **Set iteration limits**: Prevent infinite loops (typically 10-20 iterations)
2. **Monitor costs**: Track token usage per task
3. **Add logging**: Debug agent decision-making
4. **Use timeouts**: Prevent hanging executions
5. **Validate outputs**: Check files are syntactically valid

### Pattern-Specific

**Iterative:**
- Provide clear function descriptions
- Use few-shot examples in system prompt
- Return detailed error messages from tools

**Planning:**
- Use structured output (JSON with schema)
- Validate plan before execution
- Add retry logic for plan generation

**Incremental:**
- Limit context window carefully
- Implement action history pruning
- Add "warmup" steps before FINISH is available

**Multi-Agent:**
- Define clear interfaces between agents
- Use message passing protocols
- Implement handoff validation

---

## Next Steps

- See `02_FREECODECAMP_APPROACH.md` for a detailed Iterative Function Calling implementation
- See `03_DEEPSENSE_APPROACH.md` for Upfront Planning and Incremental Planning implementations
- See `05_GOAGENTS_IMPLEMENTATION.md` for how to implement these patterns with GoAgents
