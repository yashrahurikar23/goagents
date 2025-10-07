package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/yashrahurikar23/goagents/core"
)

// ReActAgent implements the ReAct (Reasoning + Acting) pattern.
// It interleaves thinking (reasoning) and acting (tool usage) steps,
// making the decision-making process transparent and debuggable.
//
// Unlike FunctionAgent which uses OpenAI's function calling API,
// ReActAgent works with ANY LLM through prompting. This makes it:
// - Compatible with Ollama, Anthropic, custom models
// - More transparent (shows reasoning steps)
// - Easier to debug (reasoning trace visible)
// - More flexible for research and experimentation
//
// The agent follows this loop:
//  1. Thought: Agent reasons about what to do next
//  2. Action: Agent decides to use a tool
//  3. Observation: Tool execution result
//  4. Repeat until final answer
//
// Example usage:
//
//	llm := ollama.New(ollama.WithModel("llama2"))
//	agent := NewReActAgent(llm)
//	agent.AddTool(calculator)
//
//	response, err := agent.Run(ctx, "What is 25 * 4?")
//	// Agent shows its reasoning:
//	// Thought: I need to multiply 25 by 4
//	// Action: calculator(operation=multiply, a=25, b=4)
//	// Observation: 100
//	// Thought: I have the answer
//	// Final Answer: 100
type ReActAgent struct {
	llm          core.LLM
	tools        map[string]core.Tool
	systemPrompt string
	maxIter      int
	trace        []ReActStep // Reasoning trace for debugging
}

// ReActStep represents one step in the reasoning trace.
type ReActStep struct {
	Iteration   int
	Thought     string
	Action      string
	ActionInput map[string]interface{}
	Observation string
}

// ReActAgentOption configures a ReActAgent.
type ReActAgentOption func(*ReActAgent)

// ReActWithSystemPrompt sets a custom system prompt.
func ReActWithSystemPrompt(prompt string) ReActAgentOption {
	return func(a *ReActAgent) {
		a.systemPrompt = prompt
	}
}

// ReActWithMaxIterations sets maximum reasoning iterations.
func ReActWithMaxIterations(max int) ReActAgentOption {
	return func(a *ReActAgent) {
		a.maxIter = max
	}
}

// NewReActAgent creates a new ReAct agent that works with any LLM.
func NewReActAgent(llm core.LLM, opts ...ReActAgentOption) *ReActAgent {
	agent := &ReActAgent{
		llm:          llm,
		tools:        make(map[string]core.Tool),
		systemPrompt: buildReActSystemPrompt(),
		maxIter:      10,
		trace:        make([]ReActStep, 0),
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// AddTool registers a tool that the agent can use.
func (a *ReActAgent) AddTool(tool core.Tool) error {
	if tool == nil {
		return &core.ErrInvalidArgument{
			Argument: "tool",
			Reason:   "cannot be nil",
		}
	}

	name := tool.Name()
	if name == "" {
		return &core.ErrInvalidArgument{
			Argument: "tool.Name()",
			Reason:   "cannot be empty",
		}
	}

	if _, exists := a.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	a.tools[name] = tool
	return nil
}

// Run executes the agent with ReAct reasoning loop.
func (a *ReActAgent) Run(ctx context.Context, input string) (*core.Response, error) {
	// Reset trace for this run
	a.trace = make([]ReActStep, 0)

	// Build initial prompt with tools and input
	prompt := a.buildPrompt(input)

	conversationHistory := prompt

	// ReAct reasoning loop
	for iteration := 0; iteration < a.maxIter; iteration++ {
		step := ReActStep{
			Iteration: iteration + 1,
		}

		// Get LLM response
		response, err := a.llm.Complete(ctx, conversationHistory)
		if err != nil {
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// Parse the response for Thought, Action, or Final Answer
		thought, action, actionInput, finalAnswer := a.parseResponse(response)

		step.Thought = thought

		// Check if we have a final answer
		if finalAnswer != "" {
			a.trace = append(a.trace, step)
			return &core.Response{
				Content: finalAnswer,
				Meta: map[string]interface{}{
					"iterations": iteration + 1,
					"trace":      a.trace,
				},
			}, nil
		}

		// Check if we have an action
		if action == "" {
			// No action and no final answer - prompt for next step
			conversationHistory += "\n" + response + "\n"
			continue
		}

		step.Action = action
		step.ActionInput = actionInput

		// Execute the action (tool)
		observation, err := a.executeAction(ctx, action, actionInput)
		if err != nil {
			observation = fmt.Sprintf("Error: %v", err)
		}

		step.Observation = observation
		a.trace = append(a.trace, step)

		// Add observation to conversation
		conversationHistory += fmt.Sprintf("\n%s\nObservation: %s\n", response, observation)
	}

	// Max iterations reached
	return nil, fmt.Errorf("max iterations (%d) reached without final answer", a.maxIter)
}

// Reset clears the reasoning trace.
func (a *ReActAgent) Reset() error {
	a.trace = make([]ReActStep, 0)
	return nil
}

// GetTrace returns the reasoning trace from the last run.
func (a *ReActAgent) GetTrace() []ReActStep {
	return a.trace
}

// buildPrompt creates the initial prompt with system instructions and tools.
func (a *ReActAgent) buildPrompt(input string) string {
	var sb strings.Builder

	// System prompt
	sb.WriteString(a.systemPrompt)
	sb.WriteString("\n\n")

	// Available tools
	if len(a.tools) > 0 {
		sb.WriteString("Available tools:\n")
		for _, tool := range a.tools {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description()))

			// Add parameter information
			schema := tool.Schema()
			if schema != nil && len(schema.Parameters) > 0 {
				sb.WriteString("  Parameters:\n")
				for _, param := range schema.Parameters {
					required := ""
					if param.Required {
						required = " (required)"
					}
					sb.WriteString(fmt.Sprintf("  - %s (%s)%s: %s\n",
						param.Name, param.Type, required, param.Description))
				}
			}
		}
		sb.WriteString("\n")
	}

	// User question
	sb.WriteString(fmt.Sprintf("Question: %s\n", input))
	sb.WriteString("Let's approach this step-by-step:\n")

	return sb.String()
}

// parseResponse extracts Thought, Action, ActionInput, and Final Answer from LLM response.
func (a *ReActAgent) parseResponse(response string) (thought, action string, actionInput map[string]interface{}, finalAnswer string) {
	// Look for "Thought:" or "Think:"
	thoughtRegex := regexp.MustCompile(`(?i)(?:Thought|Think):\s*(.+?)(?:\n|$)`)
	if matches := thoughtRegex.FindStringSubmatch(response); len(matches) > 1 {
		thought = strings.TrimSpace(matches[1])
	}

	// Look for "Final Answer:"
	finalRegex := regexp.MustCompile(`(?i)Final Answer:\s*(.+?)(?:\n\n|$)`)
	if matches := finalRegex.FindStringSubmatch(response); len(matches) > 1 {
		finalAnswer = strings.TrimSpace(matches[1])
		return
	}

	// Look for "Action:" followed by tool call
	// Format: Action: tool_name(param1=value1, param2=value2)
	actionRegex := regexp.MustCompile(`(?i)Action:\s*(\w+)\s*\(([^)]*)\)`)
	if matches := actionRegex.FindStringSubmatch(response); len(matches) > 2 {
		action = matches[1]

		// Parse parameters
		actionInput = make(map[string]interface{})
		paramsStr := matches[2]

		if paramsStr != "" {
			// Split by comma
			params := strings.Split(paramsStr, ",")
			for _, param := range params {
				// Split by = to get key-value
				parts := strings.SplitN(strings.TrimSpace(param), "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
					actionInput[key] = value
				}
			}
		}
	}

	return
}

// executeAction executes a tool with given parameters.
func (a *ReActAgent) executeAction(ctx context.Context, action string, input map[string]interface{}) (string, error) {
	tool, exists := a.tools[action]
	if !exists {
		return "", fmt.Errorf("tool '%s' not found", action)
	}

	result, err := tool.Execute(ctx, input)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", result), nil
}

// buildReActSystemPrompt creates the default system prompt for ReAct.
func buildReActSystemPrompt() string {
	return `You are a helpful AI assistant that solves problems step-by-step using the ReAct framework.

Follow this format exactly:

Thought: [Your reasoning about what to do next]
Action: tool_name(param1=value1, param2=value2)
Observation: [You will see the result here]

After seeing the observation, continue:

Thought: [Your reasoning about the observation]
Action: [Next action, or Final Answer if done]
...

When you have the final answer, respond with:
Thought: [Final reasoning]
Final Answer: [Your final answer to the user's question]

Important rules:
1. Always start with a Thought
2. Use Action to call tools when needed
3. Wait for Observation before continuing
4. Only provide Final Answer when you're confident
5. Be concise and clear in your reasoning`
}
