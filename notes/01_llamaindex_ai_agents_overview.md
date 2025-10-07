# LlamaIndex AI Agents Framework Overview

**Research Date:** October 7, 2025  
**Purpose:** Understanding LlamaIndex multi-agent systems and workflows for building a Go-based alternative

---

## Table of Contents
1. [Core Concepts](#core-concepts)
2. [Architecture Components](#architecture-components)
3. [Multi-Agent Patterns](#multi-agent-patterns)
4. [API Reference](#api-reference)
5. [Key Takeaways](#key-takeaways)

---

## Core Concepts

### What is LlamaIndex?
LlamaIndex is a Python framework for building LLM-powered agents over your data. It provides tools for:
- Data ingestion and indexing
- Building context-augmented LLM applications
- Creating complex agentic workflows
- Enabling natural language access to data

### Key Philosophy
- **Event-driven architecture**: Workflows respond to events
- **Modular design**: Composable agents with specific roles
- **State management**: Shared context across agent interactions
- **Hand-off pattern**: Agents explicitly transfer control

---

## Architecture Components

### 1. FunctionAgent

The basic building block for specialized agents.

```python
from llama_index.core.agent.workflow import FunctionAgent

agent = FunctionAgent(
    name="ResearchAgent",                    # Unique identifier
    description="Search and record notes",   # Agent's purpose
    system_prompt="You are a researcher...", # Behavioral guidelines
    llm=llm,                                 # OpenAI/LLM instance
    tools=[search_web, record_notes],        # Available tools
    can_handoff_to=["WriteAgent"],          # Agents to hand off to
    initial_state={"notes": {}}             # Optional initial state
)

# Run the agent
response = await agent.run(
    user_msg="Research AI trends",
    ctx=ctx,                                 # Optional Context
    memory=memory                            # Optional Memory
)
```

**Key Properties:**
- `name`: Unique identifier for the agent
- `description`: What the agent does
- `system_prompt`: Instructions guiding agent behavior
- `llm`: Language model instance
- `tools`: List of functions the agent can call
- `can_handoff_to`: List of agent names this agent can transfer to
- `initial_state`: Default state values

### 2. AgentWorkflow

Orchestrates multiple agents with hand-off patterns.

```python
from llama_index.core.agent.workflow import AgentWorkflow

workflow = AgentWorkflow(
    agents=[research_agent, write_agent, review_agent],
    root_agent="ResearchAgent",              # Starting agent
    initial_state={                          # Shared state
        "research_notes": {},
        "report_content": "Not written yet.",
        "review": "Review required."
    },
    output_cls=MyPydanticModel                # Optional structured output
)

# Execute workflow
response = await workflow.run(
    user_msg="Create a report on AI",
    ctx=ctx
)

# Access results
print(response.structured_response)
print(response.get_pydantic_model(MyModel))
```

**Features:**
- Automatic agent hand-offs based on `can_handoff_to`
- Shared state management across all agents
- Type-safe structured outputs with Pydantic
- Event streaming for real-time updates

### 3. Workflow (Base Class)

Custom workflow with full control over execution flow.

```python
from llama_index.core.workflow import Workflow, step, Context, Event

class CustomWorkflow(Workflow):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.llm = OpenAI(model="gpt-4o")
    
    @step
    async def my_step(self, ctx: Context, ev: StartEvent) -> StopEvent:
        # Read state
        state = await ctx.store.get("key", default=None)
        
        # Modify state atomically
        async with ctx.store.edit_state() as ctx_state:
            ctx_state["state"]["key"] = value
        
        # Stream events to UI
        ctx.write_event_to_stream(StreamEvent(delta="Processing..."))
        
        # Send events to other steps
        ctx.send_event(MyCustomEvent(data="..."))
        
        return StopEvent(result="completed")

# Run workflow
workflow = CustomWorkflow(timeout=120)
response = await workflow.run(start_event=StartEvent(query="..."))
```

**Key Methods:**
- `@step`: Decorator to define workflow steps
- `ctx.store.get/set`: State management
- `ctx.write_event_to_stream`: UI updates
- `ctx.send_event`: Inter-step communication
- `ctx.wait_for_event`: Human-in-the-loop

### 4. Context

Manages state and communication across workflow execution.

```python
from llama_index.core.workflow import Context, JsonSerializer

# Create context
ctx = Context(agent_or_workflow)

# State operations
state = await ctx.store.get("key", default=None)
await ctx.store.set("key", value)

# Atomic state updates
async with ctx.store.edit_state() as ctx_state:
    ctx_state["state"]["key"] = value

# Serialization for persistence
ctx_dict = ctx.to_dict(serializer=JsonSerializer())
restored_ctx = Context.from_dict(
    agent_or_workflow,
    ctx_dict,
    serializer=JsonSerializer()
)

# Event communication
ctx.send_event(MyEvent(...))
ctx.write_event_to_stream(StreamEvent(delta="..."))

# Human-in-the-loop
response = await ctx.wait_for_event(
    HumanResponseEvent,
    waiter_id="confirmation",
    waiter_event=InputRequiredEvent(prefix="Continue?"),
    requirements={"user_name": "User"}
)
```

### 5. Tools with Context

Tools can access and modify workflow state.

```python
from llama_index.core.workflow import Context

async def search_web(query: str) -> str:
    """Search the web for information."""
    client = AsyncTavilyClient(api_key="...")
    return str(await client.search(query))

async def record_notes(ctx: Context, notes: str, title: str) -> str:
    """Record research notes in workflow state."""
    async with ctx.store.edit_state() as ctx_state:
        if "research_notes" not in ctx_state["state"]:
            ctx_state["state"]["research_notes"] = {}
        ctx_state["state"]["research_notes"][title] = notes
    return "Notes recorded."

async def write_report(ctx: Context, content: str) -> str:
    """Write report to workflow state."""
    async with ctx.store.edit_state() as ctx_state:
        ctx_state["state"]["report_content"] = content
    return "Report written."
```

---

## Multi-Agent Patterns

### Pattern 1: Simple Hand-off Pattern

Agents explicitly transfer control to each other.

```python
# Research Agent
research_agent = FunctionAgent(
    name="ResearchAgent",
    tools=[search_web, record_notes],
    can_handoff_to=["WriteAgent"],
    system_prompt="Research and hand off to WriteAgent when ready"
)

# Write Agent
write_agent = FunctionAgent(
    name="WriteAgent",
    tools=[write_report],
    can_handoff_to=["ReviewAgent"],
    system_prompt="Write report and get feedback from ReviewAgent"
)

# Review Agent
review_agent = FunctionAgent(
    name="ReviewAgent",
    tools=[review_report],
    can_handoff_to=["WriteAgent"],
    system_prompt="Review and send back to WriteAgent if changes needed"
)

# Orchestrate
workflow = AgentWorkflow(
    agents=[research_agent, write_agent, review_agent],
    root_agent="ResearchAgent"
)
```

**Flow:**
1. ResearchAgent searches and records notes
2. When satisfied, hands off to WriteAgent
3. WriteAgent creates report
4. ReviewAgent provides feedback
5. If changes needed, hands back to WriteAgent
6. Loop until approved

### Pattern 2: Planner/Executor Pattern

A central planner generates steps for specialized agents.

```python
class PlanStep(BaseModel):
    agent_name: str
    agent_input: str

class Plan(BaseModel):
    steps: list[PlanStep]

class PlannerWorkflow(Workflow):
    llm: OpenAI = OpenAI(model="gpt-4o")
    agents: dict[str, FunctionAgent] = {
        "ResearchAgent": research_agent,
        "WriteAgent": write_agent,
        "ReviewAgent": review_agent
    }
    
    @step
    async def plan(self, ctx: Context, ev: InputEvent) -> ExecuteEvent:
        # Generate plan using LLM
        state = await ctx.store.get("state")
        
        # Create prompt with available agents
        agents_description = format_agents(self.agents)
        prompt = f"State: {state}\nAgents: {agents_description}\nTask: {ev.user_msg}"
        
        # Get plan from LLM
        response = await self.llm.achat([ChatMessage(content=prompt)])
        plan = parse_plan_from_response(response)
        
        return ExecuteEvent(plan=plan)
    
    @step
    async def execute(self, ctx: Context, ev: ExecuteEvent) -> InputEvent:
        # Execute each step with appropriate agent
        for step in ev.plan.steps:
            agent = self.agents[step.agent_name]
            await agent.run(step.agent_input, ctx=ctx)
        
        # Check if more planning needed
        return InputEvent(chat_history=updated_history)
```

**Flow:**
1. Planner analyzes task and current state
2. Generates plan with steps for each agent
3. Executor runs each step sequentially
4. After execution, checks if more planning needed
5. Repeat until task complete

### Pattern 3: Parallel Execution

Multiple agents work concurrently on different sub-tasks.

```python
class ParallelWorkflow(Workflow):
    @step
    async def dispatch(self, ctx: Context, ev: StartEvent) -> CollectEvent:
        # Split task into parallel sub-tasks
        subtasks = split_task(ev.task)
        
        # Dispatch to multiple agents
        for subtask in subtasks:
            ctx.send_event(AgentTaskEvent(
                agent_name=subtask.agent,
                input=subtask.query
            ))
        
        await ctx.store.set("expected_results", len(subtasks))
        return None  # Wait for collection
    
    @step
    async def collect(self, ctx: Context, ev: AgentResultEvent) -> StopEvent | None:
        # Collect results from agents
        results = await ctx.store.get("results", default=[])
        results.append(ev.result)
        await ctx.store.set("results", results)
        
        expected = await ctx.store.get("expected_results")
        
        if len(results) >= expected:
            # All results collected, aggregate
            final_result = aggregate_results(results)
            return StopEvent(result=final_result)
        
        return None  # Wait for more results
```

---

## API Reference

### Event Streaming

Monitor agent execution in real-time.

```python
from llama_index.core.agent.workflow import (
    AgentStream,
    AgentInput,
    AgentOutput,
    ToolCall,
    ToolCallResult
)

handler = agent.run(user_msg="query")

async for event in handler.stream_events():
    if isinstance(event, AgentStream):
        print(event.delta, end="", flush=True)
        # event.response - current full response
        # event.raw - raw LLM API response
        # event.current_agent_name - active agent
    
    elif isinstance(event, ToolCall):
        print(f"Calling tool: {event.tool_name}")
        print(f"Arguments: {event.tool_kwargs}")
    
    elif isinstance(event, ToolCallResult):
        print(f"Tool result: {event.tool_output}")
    
    elif isinstance(event, AgentInput):
        print(f"Agent input: {event.input}")
    
    elif isinstance(event, AgentOutput):
        print(f"Agent output: {event.response}")

# Get final result
result = await handler
```

### Human-in-the-Loop

Pause execution to get human input.

```python
from llama_index.core.workflow import (
    InputRequiredEvent,
    HumanResponseEvent
)

# In tool definition
async def risky_operation(ctx: Context) -> str:
    """Operation requiring human confirmation."""
    response = await ctx.wait_for_event(
        HumanResponseEvent,
        waiter_id="confirmation",
        waiter_event=InputRequiredEvent(
            prefix="Are you sure? (yes/no): ",
            user_name="User"
        ),
        requirements={"user_name": "User"}
    )
    
    if response.response.lower() == "yes":
        return "Operation completed"
    return "Operation cancelled"

# In main execution
handler = agent.run(user_msg="Do risky operation")

async for event in handler.stream_events():
    if isinstance(event, InputRequiredEvent):
        user_input = input(event.prefix)
        handler.ctx.send_event(
            HumanResponseEvent(
                response=user_input,
                user_name=event.user_name
            )
        )

result = await handler
```

### Structured Outputs

Ensure agent responses match a schema.

```python
from pydantic import BaseModel, Field

class WeatherReport(BaseModel):
    location: str = Field(description="City name")
    temperature: float = Field(description="Temperature in Celsius")
    conditions: str = Field(description="Weather conditions")
    forecast: list[str] = Field(description="3-day forecast")

workflow = AgentWorkflow(
    agents=[weather_agent],
    root_agent="WeatherAgent",
    output_cls=WeatherReport  # Enforce schema
)

response = await workflow.run("What's the weather in Tokyo?")

# Type-safe access
report: WeatherReport = response.get_pydantic_model(WeatherReport)
print(f"Temperature: {report.temperature}°C")
print(f"Conditions: {report.conditions}")
```

---

## Key Takeaways

### Strengths of LlamaIndex Approach

1. **Clear Separation of Concerns**
   - Each agent has a single responsibility
   - Tools are modular and reusable
   - State management is centralized

2. **Flexible Orchestration**
   - Simple hand-offs for linear workflows
   - Complex planning for dynamic tasks
   - Parallel execution for independent sub-tasks

3. **Production-Ready Features**
   - State serialization/deserialization
   - Event streaming for UI updates
   - Human-in-the-loop support
   - Structured output validation

4. **Type Safety (with Pydantic)**
   - Schema validation
   - Auto-generated documentation
   - IDE autocomplete

5. **Excellent Developer Experience**
   - Intuitive API design
   - Clear event model
   - Good error handling
   - Comprehensive examples

### Design Principles to Adopt

1. **Event-Driven Architecture**
   - Everything is an event
   - Steps process events and emit new ones
   - Clean separation between logic and flow

2. **Explicit Hand-offs**
   - Agents declare who they can hand off to
   - No implicit routing
   - Clear workflow visualization

3. **Context-Aware Tools**
   - Tools can access workflow state
   - State updates are atomic
   - No global variables

4. **Composability**
   - Agents are composable units
   - Workflows can be nested
   - Tools are reusable across agents

5. **Streaming by Default**
   - Real-time feedback to users
   - Progress visibility
   - Cancellation support

### Implementation Challenges for Go

1. **Async/Await**
   - Python's `async/await` is elegant
   - Go: Use goroutines + channels
   - Need good channel management patterns

2. **Dynamic Typing**
   - Python allows flexible tool definitions
   - Go: Need interface design for extensibility
   - Generics can help (Go 1.18+)

3. **Decorators**
   - Python's `@step` decorator is clean
   - Go: Struct methods or function registration
   - May need code generation

4. **Context Management**
   - Python's context managers (`async with`)
   - Go: `defer` and explicit error handling
   - Need careful resource management

### Opportunities for Go

1. **Better Performance**
   - Native concurrency with goroutines
   - No GIL limitations
   - Lower memory usage

2. **Type Safety from Start**
   - Compile-time error checking
   - Better IDE support
   - Easier refactoring

3. **Simpler Deployment**
   - Single binary
   - No dependency conflicts
   - Easy containerization

4. **Better Concurrency**
   - Channels for agent communication
   - Select for event multiplexing
   - Built-in timeout support

---

## Next Steps

1. **Design Go API** - Create idiomatic Go interfaces
2. **Prototype Core** - Build FunctionAgent equivalent
3. **Event System** - Implement event-driven workflow
4. **State Management** - Context and state store
5. **Tool System** - Flexible tool registration
6. **Examples** - Port LlamaIndex examples to Go

---

## References

- [LlamaIndex Documentation](https://docs.llamaindex.ai/)
- [Multi-Agent Guide](https://docs.llamaindex.ai/en/stable/understanding/agent/multi_agent/)
- [Workflow Documentation](https://docs.llamaindex.ai/en/stable/understanding/workflows/)
- [GitHub Repository](https://github.com/run-llama/llama_index)
