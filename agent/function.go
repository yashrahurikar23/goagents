// Package agent provides different agent implementations for orchestrating
// LLMs and tools to accomplish complex tasks.
//
// This package includes various agent types:
//   - FunctionAgent: Uses LLM function calling for tool execution
//   - ReActAgent: Implements reasoning and acting pattern
//   - ConversationalAgent: Maintains conversation history and context
//   - Multi-agent coordinators for complex workflows
package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yashrahurikar23/goagents/core"
	"github.com/yashrahurikar23/goagents/llm/openai"
)

// FunctionAgent uses OpenAI's function calling capability to execute tools.
// It automatically converts tool calls into function executions and handles
// multi-turn conversations with the LLM.
//
// Example usage:
//
//	client := openai.New(openai.WithAPIKey("sk-..."))
//	agent := agent.NewFunctionAgent(client)
//
//	calc := tools.NewCalculator()
//	agent.AddTool(calc)
//
//	response, err := agent.Run(ctx, "What is 25 * 4?")
type FunctionAgent struct {
	llm          core.LLM
	tools        map[string]core.Tool
	messages     []core.Message
	systemPrompt string
	maxIter      int
}

// FunctionAgentOption configures a FunctionAgent.
type FunctionAgentOption func(*FunctionAgent)

// WithSystemPrompt sets a custom system prompt for the agent.
func WithSystemPrompt(prompt string) FunctionAgentOption {
	return func(a *FunctionAgent) {
		a.systemPrompt = prompt
	}
}

// WithMaxIterations sets the maximum number of tool execution iterations.
// This prevents infinite loops in case the LLM keeps calling tools.
func WithMaxIterations(max int) FunctionAgentOption {
	return func(a *FunctionAgent) {
		a.maxIter = max
	}
}

// NewFunctionAgent creates a new function calling agent with the given LLM.
func NewFunctionAgent(llm core.LLM, opts ...FunctionAgentOption) *FunctionAgent {
	agent := &FunctionAgent{
		llm:          llm,
		tools:        make(map[string]core.Tool),
		messages:     make([]core.Message, 0),
		systemPrompt: "You are a helpful assistant with access to tools. Use them when needed to accomplish tasks.",
		maxIter:      5, // Default max iterations
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// AddTool registers a tool that the agent can use.
func (a *FunctionAgent) AddTool(tool core.Tool) error {
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

// Run executes the agent with the given input and returns a response.
// The agent will:
//  1. Send the user message to the LLM with available tools
//  2. If LLM requests tool calls, execute them
//  3. Send tool results back to LLM
//  4. Repeat until LLM provides final answer (or max iterations reached)
func (a *FunctionAgent) Run(ctx context.Context, input string) (*core.Response, error) {
	// Initialize messages if this is the first run
	if len(a.messages) == 0 && a.systemPrompt != "" {
		a.messages = append(a.messages, core.SystemMessage(a.systemPrompt))
	}

	// Add user message
	a.messages = append(a.messages, core.UserMessage(input))

	// Check if LLM supports OpenAI interface (for function calling)
	openaiClient, ok := a.llm.(*openai.Client)
	if !ok {
		return nil, fmt.Errorf("FunctionAgent requires an LLM that supports function calling (OpenAI)")
	}

	// Convert tools to OpenAI function format
	functions := a.convertToolsToFunctions()

	// Main execution loop
	for iter := 0; iter < a.maxIter; iter++ {
		// Create chat completion request with functions
		req := openai.ChatCompletionRequest{
			Model:    "gpt-3.5-turbo", // Default model
			Messages: a.convertToOpenAIMessages(a.messages),
		}

		// Add functions if we have tools
		if len(functions) > 0 {
			req.Tools = functions
			req.ToolChoice = "auto"
		}

		// Call LLM
		resp, err := openaiClient.CreateChatCompletion(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// Extract the response message
		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("no response from LLM")
		}

		choice := resp.Choices[0]

		// Convert content to string (handle interface{})
		contentStr := ""
		if choice.Message.Content != nil {
			contentStr = fmt.Sprintf("%v", choice.Message.Content)
		}

		// Add assistant's message to history
		assistantMsg := core.Message{
			Role:    "assistant",
			Content: contentStr,
		}

		// Check if there are tool calls
		if len(choice.Message.ToolCalls) > 0 {
			// Execute tool calls
			toolResults, err := a.executeToolCalls(ctx, choice.Message.ToolCalls)
			if err != nil {
				return nil, fmt.Errorf("tool execution failed: %w", err)
			}

			// Add tool calls to assistant message
			assistantMsg.ToolCalls = toolResults

			// Add assistant message with tool calls
			a.messages = append(a.messages, assistantMsg)

			// Add tool results as separate messages
			for _, result := range toolResults {
				toolMsg := core.Message{
					Role:       "tool",
					Content:    fmt.Sprintf("%v", result.Result),
					Name:       result.Name,
					ToolCallID: result.ID,
				}
				a.messages = append(a.messages, toolMsg)
			}

			// Continue loop to send tool results back to LLM
			continue
		}

		// No tool calls - this is the final response
		a.messages = append(a.messages, assistantMsg)

		// Return response
		return &core.Response{
			Content: fmt.Sprintf("%v", choice.Message.Content),
			Meta: map[string]interface{}{
				"model":             resp.Model,
				"finish":            choice.FinishReason,
				"iterations":        iter + 1,
				"prompt_tokens":     resp.Usage.PromptTokens,
				"completion_tokens": resp.Usage.CompletionTokens,
				"total_tokens":      resp.Usage.TotalTokens,
			},
		}, nil
	}

	return nil, fmt.Errorf("max iterations (%d) reached without final answer", a.maxIter)
}

// Reset clears the conversation history.
func (a *FunctionAgent) Reset() error {
	a.messages = make([]core.Message, 0)
	if a.systemPrompt != "" {
		a.messages = append(a.messages, core.SystemMessage(a.systemPrompt))
	}
	return nil
}

// GetMessages returns the current conversation history.
func (a *FunctionAgent) GetMessages() []core.Message {
	return a.messages
}

// executeToolCalls executes the tool calls requested by the LLM.
func (a *FunctionAgent) executeToolCalls(ctx context.Context, toolCalls []openai.ToolCall) ([]core.ToolCall, error) {
	results := make([]core.ToolCall, 0, len(toolCalls))

	for _, tc := range toolCalls {
		// Find the tool
		tool, exists := a.tools[tc.Function.Name]
		if !exists {
			// Tool not found - return error result
			results = append(results, core.ToolCall{
				ID:     tc.ID,
				Name:   tc.Function.Name,
				Result: fmt.Sprintf("Error: tool '%s' not found", tc.Function.Name),
			})
			continue
		}

		// Parse arguments
		var args map[string]interface{}
		if tc.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				results = append(results, core.ToolCall{
					ID:     tc.ID,
					Name:   tc.Function.Name,
					Result: fmt.Sprintf("Error: invalid arguments: %v", err),
				})
				continue
			}
		} else {
			args = make(map[string]interface{})
		}

		// Execute tool
		result, err := tool.Execute(ctx, args)
		if err != nil {
			results = append(results, core.ToolCall{
				ID:     tc.ID,
				Name:   tc.Function.Name,
				Result: fmt.Sprintf("Error: %v", err),
			})
			continue
		}

		// Format result as string
		resultStr := fmt.Sprintf("%v", result)

		results = append(results, core.ToolCall{
			ID:     tc.ID,
			Name:   tc.Function.Name,
			Args:   args,
			Result: resultStr,
		})
	}

	return results, nil
}

// convertToolsToFunctions converts core.Tool to OpenAI function format.
func (a *FunctionAgent) convertToolsToFunctions() []openai.Tool {
	if len(a.tools) == 0 {
		return nil
	}

	functions := make([]openai.Tool, 0, len(a.tools))
	for _, tool := range a.tools {
		schema := tool.Schema()
		if schema == nil {
			continue
		}

		// Build parameters
		params := map[string]interface{}{
			"type":       "object",
			"properties": make(map[string]interface{}),
		}

		required := make([]string, 0)

		// Add each parameter
		for _, param := range schema.Parameters {
			paramDef := map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
			}

			// Add enum if present
			if len(param.Enum) > 0 {
				paramDef["enum"] = param.Enum
			}

			params["properties"].(map[string]interface{})[param.Name] = paramDef

			if param.Required {
				required = append(required, param.Name)
			}
		}

		if len(required) > 0 {
			params["required"] = required
		}

		functions = append(functions, openai.Tool{
			Type: "function",
			Function: &openai.Function{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  params,
			},
		})
	}

	return functions
}

// convertToOpenAIMessages converts core.Message to openai.ChatMessage.
func (a *FunctionAgent) convertToOpenAIMessages(messages []core.Message) []openai.ChatMessage {
	result := make([]openai.ChatMessage, 0, len(messages))

	for _, msg := range messages {
		openaiMsg := openai.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}

		// Add tool call ID for tool messages
		if msg.Role == "tool" {
			openaiMsg.ToolCallID = msg.ToolCallID
		}

		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			openaiMsg.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				// Marshal args back to JSON string
				argsJSON, _ := json.Marshal(tc.Args)
				openaiMsg.ToolCalls[i] = openai.ToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: &openai.FunctionCall{
						Name:      tc.Name,
						Arguments: string(argsJSON),
					},
				}
			}
		}

		result = append(result, openaiMsg)
	}

	return result
}
