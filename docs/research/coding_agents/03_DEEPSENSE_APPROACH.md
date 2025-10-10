# Deepsense.ai Approach: Creating Code Writing Agents for Data Science

**Source**: [Deepsense.ai Blog](https://deepsense.ai/blog/creating-your-own-code-writing-agent-how-to-get-results-fast-and-avoid-the-most-common-pitfalls/)

**Authors**: Alan Konarski, Maks Operlejn, Patryk Kowalski

**Date**: October 30, 2023

---

## Table of Contents

1. [Overview](#overview)
2. [Approach A: Upfront Planning](#approach-a-upfront-planning)
3. [Approach B: Incremental Planning](#approach-b-incremental-planning)
4. [Context Management](#context-management)
5. [Evaluation](#evaluation)
6. [Lessons Learned](#lessons-learned)

---

## Overview

### Motivation

Deepsense.ai set out to build coding agents specifically for **Data Science tasks** like:

- Image classification (Computer Vision)
- Sentiment analysis (NLP)
- Tabular data modeling (Spotify genre classification)

They found that existing solutions (AutoGPT, GPT Engineer) performed poorly on DS-specific tasks, so they built their own.

### Research Questions

1. **Planning Strategy**: Should agents plan everything upfront or plan as they go?
2. **Context Management**: How to handle large codebases within token limits?
3. **Specialization**: Are specialized agents better than generalists?
4. **Quality vs. Cost**: What's the optimal balance?

### Benchmarks

They evaluated agents on 7 small-scale DS projects:

- License plate detection (CV)
- Text summarizer (NLP)
- Spotify genre classification (tabular)
- Others (not fully detailed)

**Evaluation Criteria:**

**Quantitative:**

- Can code execute without errors?
- Cost to generate
- Time to complete

**Qualitative (Human Review):**

- Effort to fix issues
- Solution completeness
- Code style and best practices

---

## Approach A: Upfront Planning

### Evolution Through 3 Iterations

#### Iteration 1: Baseline

**Architecture:**

```
User Input → Planner Agent → Complete Plan (JSON)
                                    ↓
                    ┌───────────────┴───────────────┐
                    ↓                               ↓
              Task 1: Write Code             Task 2: Write Code
                    ↓                               ↓
                 file1.py                        file2.py
```

**How It Works:**

1. User provides problem description
2. Planner generates complete task list in JSON
3. CodeWriter executes each task sequentially
4. Uses LangChain's FileManagementToolkit

**Task Types:**

- `write_code`: Generate implementation code
- `finish`: Mark completion

**Problems Encountered:**

❌ **No Context Sharing**: Tasks treated as independent
- Model can't import previously written functions
- No awareness of prior work

❌ **Unreliable FileManagementToolkit**: LangChain's toolkit called wrong tools frequently

❌ **Invalid JSON**: Models often returned malformed JSON, causing crashes

---

#### Iteration 2: Context-Aware

**Improvements:**

1. **Custom File Management**: Replaced LangChain's toolkit with custom implementation
2. **Full File Context**: Inject entire previously-generated files into prompt
3. **New Specialized Roles**:
   - **TestWriter**: Writes pytest unit tests
   - **CodeRefactorer**: Improves code quality
   - **AdditionalFilesWriter**: Creates README, requirements.txt, etc.

**Architecture:**

```
User Input → Planner → Complete Plan
                            ↓
        ┌───────────────────┼───────────────────┐
        ↓                   ↓                   ↓
   CodeWriter          TestWriter         CodeRefactorer
   (+ file context)   (+ code context)    (+ all context)
        ↓                   ↓                   ↓
    src/*.py            tests/*.py         src/*.py (v2)
```

**Context Injection:**

```python
# Example: When generating train.py
context = """
Previously generated files:

--- data_loader.py ---
def load_data(path):
    # ... code ...
    return dataframe

--- config.py ---
MODEL_TYPE = "random_forest"
BATCH_SIZE = 32
"""

prompt = f"{context}\n\nTask: Write training script using load_data()"
```

**Problems Encountered:**

❌ **Context Length Explosion**: Full files quickly exceed 16k token limit

❌ **High Costs**: Large prompts expensive (GPT-4 pricing)

❌ **Slow Generation**: Large context → slow processing

❌ **Occasional Invalid JSON**: Plan generation still fails sometimes

---

#### Iteration 3: Vector Embeddings for Context

**Major Innovation**: Instead of full files, use vector embeddings to retrieve only relevant code chunks.

**Architecture:**

```
Code Files → Code Splitter → Chunks → Embeddings → FAISS Vector Store
                                                           ↓
Task → Prerequisites → Query Embeddings → Similarity Search
                                                ↓
                                    Relevant Chunks Only
                                                ↓
                                        LLM Prompt (optimized)
```

**Code Splitter (Custom Implementation):**

Instead of naive splitting:

```python
# BAD: Character-based splitting
chunks = [code[i:i+1000] for i in range(0, len(code), 1000)]
# Can split mid-function, loses context
```

Use **semantic code splitting**:

```python
# GOOD: Extract functions/classes with signatures only

# Original function:
def calculate_metrics(y_true, y_pred):
    """Calculate precision, recall, F1."""
    precision = precision_score(y_true, y_pred)
    recall = recall_score(y_true, y_pred)
    f1 = f1_score(y_true, y_pred)
    return {"precision": precision, "recall": recall, "f1": f1}

# Stored in vector DB (body removed):
def calculate_metrics(y_true, y_pred):
    """Calculate precision, recall, F1."""
    pass
```

**Why This Works:**

- LLM only needs to know **how to call** the function
- Signature + docstring = sufficient information
- Body is noise for most use cases
- Dramatically reduces token usage

**For Classes:**

```python
# Original class:
class DataProcessor:
    """Handles data preprocessing."""
    
    def __init__(self, config):
        self.config = config
        self.scaler = StandardScaler()
        self.encoder = LabelEncoder()
    
    def fit(self, X, y):
        # ... implementation ...
    
    def transform(self, X):
        # ... implementation ...

# Stored in vector DB:
class DataProcessor:
    """Handles data preprocessing."""
    
    def __init__(self, config):
        self.config = config
        self.scaler = StandardScaler()
        self.encoder = LabelEncoder()
    
    def fit(self, X, y):
        """Fit preprocessing transformers."""
        pass
    
    def transform(self, X):
        """Transform data using fitted transformers."""
        pass
```

**Prerequisites System:**

The planner now generates tasks with dependencies:

```json
{
  "tasks": [
    {
      "id": 1,
      "description": "Implement data loading function",
      "output_file": "data_loader.py",
      "prerequisites": []
    },
    {
      "id": 2,
      "description": "Implement model training",
      "output_file": "train.py",
      "prerequisites": [
        "data_loader.load_data()",
        "config.MODEL_TYPE",
        "config.BATCH_SIZE"
      ]
    }
  ]
}
```

**Retrieval Process:**

```python
# For Task 2:
for prereq in task.prerequisites:
    # Convert to embedding
    prereq_embedding = embed(prereq)
    
    # Query vector store
    similar_chunks = vector_store.similarity_search(
        prereq_embedding,
        k=3  # Top 3 matches
    )
    
    # Add to context
    context += "\n".join(similar_chunks)
```

**Results:**

✅ Stays within token limits
✅ Lower cost per generation
✅ Faster generation
✅ Still maintains necessary context
✅ Good code organization

**Remaining Problems:**

❌ **Outdated Knowledge**: GPT-4 trained on data up to 2021
- Missing newer libraries
- Old API patterns
- Deprecated approaches

❌ **Inflexible Plan**: Can't modify if initial plan is wrong

---

### Final Architecture (Iteration 3)

```
┌─────────────────────────────────────────────────────────┐
│  User: "Build image classifier for license plates"      │
└───────────────────┬─────────────────────────────────────┘
                    ↓
            ┌───────────────┐
            │ Planner Agent │
            │ (GPT-4)       │
            └───────┬───────┘
                    ↓
        ┌───────────────────────┐
        │ Plan with Prerequisites│
        │ (JSON validated)       │
        └───────┬───────────────┘
                ↓
    ┌───────────────────────────┐
    │ For Each Task in Plan:    │
    └───────┬───────────────────┘
            ↓
    ┌────────────────────────────────┐
    │ 1. Extract Prerequisites       │
    │ 2. Query Vector Store          │
    │ 3. Retrieve Relevant Chunks    │
    └──────────┬─────────────────────┘
               ↓
    ┌────────────────────────────────┐
    │ Build Context:                 │
    │ - Task description             │
    │ - Retrieved code chunks        │
    │ - Previous task results        │
    └──────────┬─────────────────────┘
               ↓
    ┌────────────────────────────────┐
    │ Specialized Agent:             │
    │ - CodeWriter / TestWriter /    │
    │   CodeRefactorer / etc.        │
    └──────────┬─────────────────────┘
               ↓
    ┌────────────────────────────────┐
    │ Generate Code → Save File      │
    │ Update Vector Store with New   │
    │ Functions/Classes              │
    └──────────┬─────────────────────┘
               ↓
    ┌────────────────────────────────┐
    │ Next Task or Finish            │
    └────────────────────────────────┘
```

---

## Approach B: Incremental Planning

### Philosophy

Developers don't plan every step upfront - they:

1. Start with rough idea
2. Implement
3. Discover issues
4. Adapt and iterate

**Can agents do the same?**

### Architecture

```
User Input → Initial Problem
                ↓
        ┌───────────────┐
        │ PrepAgent     │
        │ (Extract info) │
        └───────┬───────┘
                ↓
    ┌──────────────────────┐
    │ Domain: Computer Vision│
    │ Problem: Classification│
    │ Model: CNN            │
    │ Metrics: Accuracy, F1 │
    │ Dataset: license_plates│
    └──────────┬──────────────┘
               ↓
        ┌─────────────────┐
        │  ActionAgent    │
        │  (Decide next)   │
        └────────┬────────┘
                 ↓
    ┌────────────────────────────┐
    │ Action Choices:            │
    │ - CREATE <file>            │
    │ - EDIT <file>              │
    │ - DELETE <file>            │
    │ - PEEK <file>              │
    │ - RUN <file>               │
    │ - FINISH                   │
    └────────┬───────────────────┘
             ↓
    ┌───────────────────┐
    │ Execute Action    │
    │ Update State      │
    └────────┬──────────┘
             ↓
    ┌───────────────────┐
    │ Back to ActionAgent│
    │ (Loop)             │
    └────────────────────┘
```

### Actions Explained

**CREATE <filename>**

```python
Action: CREATE model.py
Result: Empty model.py created
State: {"files": ["model.py"]}
```

**EDIT <filename>**

```python
Action: EDIT model.py
Agent: Generates CNN architecture code
Result: model.py now contains model definition
State: {"files": ["model.py"], "model_defined": True}
```

**PEEK <filename>**

```python
Action: PEEK data.csv
Result: First 5 rows loaded into context
State: Context now includes data schema
```

**DELETE <filename>**

```python
Action: DELETE old_model.py
Result: File removed
State: {"files": ["model.py"]}
```

**RUN <filename>**

```python
Action: RUN train.py
Result: Error - missing dataset
State: {"last_error": "FileNotFoundError: data.csv"}
```

**FINISH**

```python
Action: FINISH
Agent: "License plate classifier implemented and tested"
```

### Secondary Agents

**BobTheBuilderAgent**: Manages file operations

```python
def bob_the_builder(action, filename, content=None):
    if action == "CREATE":
        Path(filename).touch()
    elif action == "EDIT":
        Path(filename).write_text(content)
    elif action == "DELETE":
        Path(filename).unlink()
```

**ScriptKiddieAgent**: Generates/modifies code

```python
def script_kiddie(task, context):
    prompt = f"""
    Task: {task}
    Context: {context}
    Generate Python code:
    """
    return llm.generate(prompt)
```

### Example Execution Flow

```
State 0: "Build sentiment analyzer"

ActionAgent: CREATE sentiment_model.py
State 1: sentiment_model.py exists (empty)

ActionAgent: EDIT sentiment_model.py
ScriptKiddieAgent: Generates BERT-based model
State 2: Model implemented

ActionAgent: PEEK train_data.csv
State 3: Context now has data schema

ActionAgent: CREATE train.py
State 4: train.py exists

ActionAgent: EDIT train.py
State 5: Training script implemented

ActionAgent: RUN train.py
State 6: Error - missing transformers library

ActionAgent: EDIT requirements.txt
State 7: requirements.txt created with transformers

ActionAgent: RUN train.py
State 8: Training successful

ActionAgent: CREATE test.py
State 9: test.py created

ActionAgent: EDIT test.py
State 10: Tests implemented

ActionAgent: RUN test.py
State 11: All tests pass

ActionAgent: FINISH
```

### Problems Encountered

❌ **Infinite Loops**: Agent can get stuck in CREATE → EDIT → DELETE cycles

❌ **Invalid Actions**: Sometimes invents new actions like "CONTACT_CLIENT"

```
Expected: CREATE, EDIT, DELETE, PEEK, RUN, FINISH
Actual:   CONTACT_CLIENT meeting.ics
```

❌ **Non-Determinism**: Same prompt can lead to different paths

❌ **Inefficient Exploration**: May take roundabout paths

❌ **State Explosion**: History grows large quickly

### Advantages

✅ **Highly Flexible**: Adapts to discoveries
✅ **Recovers from Errors**: Can fix mistakes
✅ **Mimics Human Workflow**: Natural development process
✅ **No Planning Failures**: No upfront plan to go wrong

### When to Use

- **Exploratory Projects**: Requirements unclear
- **Debugging**: Unknown issues
- **Research Code**: Experimental work
- **Small Projects**: Limited scope

---

## Context Management

### The Token Problem

**GPT-4**: 8k or 32k context window
**GPT-3.5**: 4k or 16k context window

**Typical Data Science Project:**

```
data_loader.py:        500 tokens
preprocessing.py:      800 tokens
model.py:            1,200 tokens
train.py:              600 tokens
evaluate.py:           400 tokens
utils.py:              300 tokens
----------------------
Total:               3,800 tokens (just code)

+ Task description:    200 tokens
+ Previous outputs:    500 tokens
+ System prompt:       150 tokens
----------------------
Grand Total:         4,650 tokens (approaching limits)
```

### Solution 1: File Truncation (Naive)

```python
MAX_CHARS = 10000
content = file.read(MAX_CHARS)
if len(file_content) > MAX_CHARS:
    content += "[... truncated ...]"
```

**Problems:**
- May cut mid-function
- Loses important context
- Arbitrary cutoff

### Solution 2: Vector Embeddings (Deepsense Solution)

**Step 1: Code Splitting**

```python
import ast

def split_code_semantically(file_content):
    tree = ast.parse(file_content)
    chunks = []
    
    for node in ast.walk(tree):
        if isinstance(node, ast.FunctionDef):
            # Extract function signature + docstring only
            signature = f"def {node.name}({args_to_string(node.args)}):"
            docstring = ast.get_docstring(node) or ""
            chunks.append(f"{signature}\n    \"\"\"{docstring}\"\"\"\n    pass")
        
        elif isinstance(node, ast.ClassDef):
            # Extract class with __init__ + method signatures
            class_chunk = f"class {node.name}:\n"
            class_chunk += f"    \"\"\"{ast.get_docstring(node)}\"\"\"\n"
            
            # Include full __init__
            for item in node.body:
                if isinstance(item, ast.FunctionDef) and item.name == "__init__":
                    class_chunk += ast.unparse(item)
                elif isinstance(item, ast.FunctionDef):
                    # Just signature for other methods
                    sig = f"    def {item.name}({args_to_string(item.args)}):"
                    doc = f'\n        """{ast.get_docstring(item)}"""'
                    class_chunk += sig + doc + "\n        pass\n"
            
            chunks.append(class_chunk)
    
    return chunks
```

**Step 2: Create Embeddings**

```python
from langchain.embeddings import OpenAIEmbeddings
from langchain.vectorstores import FAISS

embeddings = OpenAIEmbeddings()
chunks = split_code_semantically(code)

# Create vector store
vector_store = FAISS.from_texts(chunks, embeddings)
```

**Step 3: Retrieval**

```python
# When CodeWriter needs context for task
task = "Implement model training using the data loader"

# Extract prerequisites from task
prerequisites = ["load_data", "DataLoader", "preprocess"]

# Retrieve relevant chunks
relevant_chunks = []
for prereq in prerequisites:
    results = vector_store.similarity_search(prereq, k=3)
    relevant_chunks.extend(results)

# Build optimized context
context = "\n\n".join(set(relevant_chunks))  # Remove duplicates
```

**Token Savings:**

```
Before (full files):  3,800 tokens
After (embeddings):     800 tokens (5x reduction!)
```

### Solution 3: Hierarchical Context

```python
# Level 1: File tree only (minimal tokens)
context_l1 = """
project/
├── src/
│   ├── data_loader.py
│   ├── model.py
│   └── train.py
└── tests/
    └── test_model.py
"""

# Level 2: File summaries (medium tokens)
context_l2 = """
data_loader.py: Contains load_data(), preprocess() functions
model.py: Defines CNNModel class with train() and predict()
train.py: Main training loop, uses data_loader and model
"""

# Level 3: Relevant code chunks (targeted)
context_l3 = vector_store.retrieve(task_description)
```

**Use appropriate level based on task:**
- File creation → L1 (tree)
- High-level planning → L2 (summaries)
- Implementation → L3 (code chunks)

---

## Evaluation

### Benchmarks

7 Data Science projects:
1. License plate detection (CV)
2. Text summarizer (NLP)
3. Spotify genre classification (tabular)
4-7. (Not fully detailed in article)

### Agents Tested

1. **Their Approach A (Iteration 3)**: Upfront planning + vector embeddings
2. **Their Approach B**: Incremental planning
3. **MetaGPT**: Multi-agent collaboration
4. **GPT Engineer**: Iterative function calling

### Evaluation Criteria

**Quantitative:**
- ✅/❌ Does code execute?
- 💰 Cost to generate
- ⏱️ Time to complete

**Qualitative:**
- Effort to fix (hours needed)
- Solution completeness (0-10)
- Code quality (0-10)
- Adherence to best practices

### Results

**Winner: MetaGPT & Deepsense Approach A (Iteration 3)** (tied)

| Agent | Code Quality | Planning | Cost | Execution | Best For |
|-------|-------------|----------|------|-----------|----------|
| **MetaGPT** | 9/10 | Excellent | $$$ | Slow | Production |
| **Deepsense A v3** | 9/10 | Excellent | $$ | Medium | Data Science |
| **GPT Engineer** | 7/10 | Good | $ | Fast | Prototypes |
| **Deepsense B** | 6/10 | Flexible | $$$ | Varies | Exploration |

### Key Findings

✅ **Specialized Roles Work**: Separate CodeWriter, TestWriter, Refactorer better than single agent

✅ **Upfront Planning > Incremental** (for defined tasks): More efficient, predictable

✅ **Vector Embeddings Essential**: For codebases > 5 files

✅ **Few-Shot Prompting Critical**: For structured outputs (JSON)

✅ **GPT-4 >>> GPT-3.5**: For code generation (worth the cost)

❌ **All Solutions Have Limits**: None perfect for production

---

## Lessons Learned

### 1. GPT-4 is Essential

```
GPT-3.5: Adequate for simple tasks, frequent errors
GPT-4:   Much better code quality, worth 10x cost
```

**Recommendation**: Use GPT-4 for CodeWriter, GPT-3.5 for Planner/Refactorer

### 2. Specialized Roles > Generalization

**Bad:**

```python
agent = GenericAgent()
agent.generate("Write code and tests and refactor it")
```

**Good:**

```python
code = CodeWriter().generate(task)
tests = TestWriter().generate(code)
improved = CodeRefactorer().refactor(code)
```

**Why**: Each role has focused prompt, specialized instructions

### 3. Few-Shot Prompting for Structured Output

**Bad:**

```python
prompt = "Generate a plan in JSON format"
# Often returns invalid JSON
```

**Good:**

```python
prompt = """
Generate a plan in JSON format.

Example:
{
  "tasks": [
    {"id": 1, "description": "...", "output_file": "..."},
    {"id": 2, "description": "...", "output_file": "..."}
  ]
}

Your plan:
"""
# Much more reliable
```

### 4. Mix Models for Cost Efficiency

```python
# Expensive operations
CodeWriter: GPT-4

# Cheaper operations
Planner: GPT-3.5
Summarizer: GPT-3.5
Refactorer: GPT-3.5-turbo
```

### 5. Structured Answers > Plain Text

**Instead of:**

```python
response = llm.generate("List all files in the project")
# Returns: "There are 5 files: main.py, utils.py, ..."
# Hard to parse
```

**Do:**

```python
response = llm.generate("List all files in JSON array format")
# Returns: ["main.py", "utils.py", "config.py", "train.py", "test.py"]
# Easy to parse and use
```

### 6. Add Comprehensive Logging

```python
import logging

logger.info(f"Task {task_id}: {task_description}")
logger.debug(f"Context size: {len(context)} tokens")
logger.info(f"Generated {len(code)} lines of code")
logger.warning(f"Code execution failed: {error}")
logger.info(f"Cost so far: ${total_cost:.2f}")
```

**Essential for:**
- Debugging
- Cost tracking
- Performance optimization
- Understanding agent decisions

### 7. State-of-the-Art Knowledge Gap

**Problem**: Models trained on 2021 data

**Examples:**
- Missing: Transformers 4.x new APIs
- Missing: PyTorch 2.0 features
- Uses: Deprecated TensorFlow 1.x patterns

**Workarounds:**
- Provide library docs in context
- Add examples of modern APIs
- Manual post-generation updates

---

## Comparison to FreeCodeCamp Approach

| Aspect | Deepsense | FreeCodeCamp |
|--------|-----------|--------------|
| **Task Focus** | Data Science projects | General bug fixing |
| **Planning** | Upfront (Approach A) | None (iterative) |
| **Context** | Vector embeddings | Full file injection |
| **Agents** | Multiple specialized | Single agent |
| **Complexity** | High | Low |
| **Production Ready** | Closer | Educational only |
| **Cost Optimization** | Significant effort | Basic |
| **Best For** | Complete projects | Quick fixes |

---

## Next Steps

- See `04_COMPARATIVE_ANALYSIS.md` for detailed comparison of all approaches
- See `05_GOAGENTS_IMPLEMENTATION.md` for implementing these patterns with GoAgents
- See `06_BEST_PRACTICES.md` for production recommendations
