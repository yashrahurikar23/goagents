package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/yashrahurikar23/goagents/core"
	"github.com/yashrahurikar23/goagents/tests/mocks"
)

// TestReActAgent_NewAgent tests ReAct agent creation.
func TestReActAgent_NewAgent(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	if agent == nil {
		t.Fatal("NewReActAgent() returned nil")
	}

	if agent.llm == nil {
		t.Error("agent.llm is nil")
	}

	if agent.tools == nil {
		t.Error("agent.tools is nil")
	}

	if agent.systemPrompt == "" {
		t.Error("agent.systemPrompt is empty")
	}

	if agent.maxIter != 10 {
		t.Errorf("agent.maxIter = %d, want 10", agent.maxIter)
	}
}

// TestReActAgent_WithOptions tests agent configuration.
func TestReActAgent_WithOptions(t *testing.T) {
	llm := mocks.NewMockLLM()
	customPrompt := "Custom ReAct prompt"
	maxIter := 5

	agent := NewReActAgent(
		llm,
		ReActWithSystemPrompt(customPrompt),
		ReActWithMaxIterations(maxIter),
	)

	if agent.systemPrompt != customPrompt {
		t.Errorf("systemPrompt = %q, want %q", agent.systemPrompt, customPrompt)
	}

	if agent.maxIter != maxIter {
		t.Errorf("maxIter = %d, want %d", agent.maxIter, maxIter)
	}
}

// TestReActAgent_AddTool tests tool registration.
func TestReActAgent_AddTool(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	tool := mocks.NewMockTool("calculator", "Performs calculations")

	err := agent.AddTool(tool)
	if err != nil {
		t.Fatalf("AddTool() error = %v", err)
	}

	if len(agent.tools) != 1 {
		t.Errorf("len(tools) = %d, want 1", len(agent.tools))
	}
}

// TestReActAgent_AddTool_Nil tests nil tool handling.
func TestReActAgent_AddTool_Nil(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	err := agent.AddTool(nil)
	if err == nil {
		t.Fatal("AddTool(nil) expected error, got nil")
	}

	var invalidArgErr *core.ErrInvalidArgument
	if !errors.As(err, &invalidArgErr) {
		t.Errorf("error type = %T, want *core.ErrInvalidArgument", err)
	}
}

// TestReActAgent_AddTool_EmptyName tests empty name handling.
func TestReActAgent_AddTool_EmptyName(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	tool := mocks.NewMockTool("", "Tool with empty name")

	err := agent.AddTool(tool)
	if err == nil {
		t.Fatal("AddTool() expected error for empty name, got nil")
	}
}

// TestReActAgent_AddTool_Duplicate tests duplicate tool handling.
func TestReActAgent_AddTool_Duplicate(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	tool1 := mocks.NewMockTool("calculator", "First")
	tool2 := mocks.NewMockTool("calculator", "Second")

	agent.AddTool(tool1)
	err := agent.AddTool(tool2)

	if err == nil {
		t.Fatal("AddTool() expected error for duplicate, got nil")
	}
}

// TestReActAgent_ParseResponse tests response parsing.
func TestReActAgent_ParseResponse(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	tests := []struct {
		name        string
		response    string
		wantThought string
		wantAction  string
		wantFinal   string
	}{
		{
			name:        "thought only",
			response:    "Thought: I need to calculate something",
			wantThought: "I need to calculate something",
			wantAction:  "",
			wantFinal:   "",
		},
		{
			name:        "thought and action",
			response:    "Thought: Let me calculate\nAction: calculator(operation=add, a=5, b=3)",
			wantThought: "Let me calculate",
			wantAction:  "calculator",
			wantFinal:   "",
		},
		{
			name:        "final answer",
			response:    "Thought: I have the answer\nFinal Answer: The result is 8",
			wantThought: "I have the answer",
			wantAction:  "",
			wantFinal:   "The result is 8",
		},
		{
			name:        "case insensitive",
			response:    "THOUGHT: Testing case\nACTION: test()",
			wantThought: "Testing case",
			wantAction:  "test",
			wantFinal:   "",
		},
		{
			name:        "think variant",
			response:    "Think: Alternative keyword",
			wantThought: "Alternative keyword",
			wantAction:  "",
			wantFinal:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thought, action, _, finalAnswer := agent.parseResponse(tt.response)

			if thought != tt.wantThought {
				t.Errorf("thought = %q, want %q", thought, tt.wantThought)
			}

			if action != tt.wantAction {
				t.Errorf("action = %q, want %q", action, tt.wantAction)
			}

			if finalAnswer != tt.wantFinal {
				t.Errorf("finalAnswer = %q, want %q", finalAnswer, tt.wantFinal)
			}
		})
	}
}

// TestReActAgent_ParseResponse_ActionInput tests action parameter parsing.
func TestReActAgent_ParseResponse_ActionInput(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	tests := []struct {
		name      string
		response  string
		wantInput map[string]interface{}
	}{
		{
			name:     "single param",
			response: "Action: tool(param=value)",
			wantInput: map[string]interface{}{
				"param": "value",
			},
		},
		{
			name:     "multiple params",
			response: "Action: calculator(operation=add, a=5, b=3)",
			wantInput: map[string]interface{}{
				"operation": "add",
				"a":         "5",
				"b":         "3",
			},
		},
		{
			name:     "quoted values",
			response: `Action: search(query="hello world")`,
			wantInput: map[string]interface{}{
				"query": "hello world",
			},
		},
		{
			name:      "no params",
			response:  "Action: tool()",
			wantInput: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, input, _ := agent.parseResponse(tt.response)

			if len(input) != len(tt.wantInput) {
				t.Errorf("input length = %d, want %d", len(input), len(tt.wantInput))
			}

			for key, wantValue := range tt.wantInput {
				gotValue, exists := input[key]
				if !exists {
					t.Errorf("input missing key %q", key)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("input[%q] = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

// TestReActAgent_ExecuteAction tests tool execution.
func TestReActAgent_ExecuteAction(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	tool := mocks.NewMockTool("calculator", "Does math")
	tool.ExecuteFunc = func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "42", nil
	}
	agent.AddTool(tool)

	ctx := context.Background()
	result, err := agent.executeAction(ctx, "calculator", map[string]interface{}{
		"operation": "add",
		"a":         "10",
		"b":         "32",
	})

	if err != nil {
		t.Fatalf("executeAction() error = %v", err)
	}

	if result != "42" {
		t.Errorf("result = %q, want %q", result, "42")
	}
}

// TestReActAgent_ExecuteAction_ToolNotFound tests missing tool.
func TestReActAgent_ExecuteAction_ToolNotFound(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	ctx := context.Background()
	_, err := agent.executeAction(ctx, "nonexistent", map[string]interface{}{})

	if err == nil {
		t.Fatal("executeAction() expected error for missing tool, got nil")
	}

	expectedMsg := "tool 'nonexistent' not found"
	if err.Error() != expectedMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedMsg)
	}
}

// TestReActAgent_ExecuteAction_ToolError tests tool execution error.
func TestReActAgent_ExecuteAction_ToolError(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	tool := mocks.NewMockTool("calculator", "Does math")
	tool.ExecuteFunc = func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return nil, fmt.Errorf("division by zero")
	}
	agent.AddTool(tool)

	ctx := context.Background()
	_, err := agent.executeAction(ctx, "calculator", map[string]interface{}{})

	if err == nil {
		t.Fatal("executeAction() expected error from tool, got nil")
	}

	if err.Error() != "division by zero" {
		t.Errorf("error message = %q, want %q", err.Error(), "division by zero")
	}
}

// TestReActAgent_Run_SimpleFinalAnswer tests direct final answer.
func TestReActAgent_Run_SimpleFinalAnswer(t *testing.T) {
	llm := mocks.NewMockLLM()
	llm.WithCompleteResponse("Thought: This is simple\nFinal Answer: 42")

	agent := NewReActAgent(llm)

	ctx := context.Background()
	response, err := agent.Run(ctx, "What is the answer?")

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if response.Content != "42" {
		t.Errorf("response.Content = %q, want %q", response.Content, "42")
	}

	// Check trace
	trace := agent.GetTrace()
	if len(trace) != 1 {
		t.Errorf("trace length = %d, want 1", len(trace))
	}

	if trace[0].Thought != "This is simple" {
		t.Errorf("trace[0].Thought = %q, want %q", trace[0].Thought, "This is simple")
	}
}

// TestReActAgent_Run_WithToolCall tests reasoning with tool execution.
func TestReActAgent_Run_WithToolCall(t *testing.T) {
	llm := mocks.NewMockLLM()

	// Configure sequential responses
	responses := []string{
		"Thought: I need to calculate\nAction: calculator(operation=multiply, a=25, b=4)",
		"Thought: I have the result\nFinal Answer: 100",
	}
	responseIndex := 0
	llm.CompleteFunc = func(ctx context.Context, prompt string) (string, error) {
		if responseIndex >= len(responses) {
			return responses[len(responses)-1], nil
		}
		resp := responses[responseIndex]
		responseIndex++
		return resp, nil
	}

	agent := NewReActAgent(llm)

	// Add tool
	tool := mocks.NewMockTool("calculator", "Does math")
	tool.ExecuteFunc = func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "100", nil
	}
	agent.AddTool(tool)

	ctx := context.Background()
	response, err := agent.Run(ctx, "What is 25 * 4?")

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if response.Content != "100" {
		t.Errorf("response.Content = %q, want %q", response.Content, "100")
	}

	// Check trace
	trace := agent.GetTrace()
	if len(trace) != 2 {
		t.Errorf("trace length = %d, want 2", len(trace))
	}

	// First step should have tool call
	if trace[0].Action != "calculator" {
		t.Errorf("trace[0].Action = %q, want %q", trace[0].Action, "calculator")
	}

	if trace[0].Observation != "100" {
		t.Errorf("trace[0].Observation = %q, want %q", trace[0].Observation, "100")
	}

	// Second step should have final answer
	if trace[1].Thought != "I have the result" {
		t.Errorf("trace[1].Thought = %q, want %q", trace[1].Thought, "I have the result")
	}

	// Check tool was called
	if tool.CallCount() != 1 {
		t.Errorf("tool.CallCount() = %d, want 1", tool.CallCount())
	}
}

// TestReActAgent_Run_MaxIterations tests iteration limit.
func TestReActAgent_Run_MaxIterations(t *testing.T) {
	llm := mocks.NewMockLLM()

	// Always return thought without final answer (infinite loop)
	llm.WithCompleteResponse("Thought: Still thinking...")

	agent := NewReActAgent(llm, ReActWithMaxIterations(3))

	ctx := context.Background()
	_, err := agent.Run(ctx, "Question")

	if err == nil {
		t.Fatal("Run() expected error for max iterations, got nil")
	}

	expectedMsg := "max iterations (3) reached without final answer"
	if err.Error() != expectedMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedMsg)
	}
}

// TestReActAgent_Reset tests trace reset.
func TestReActAgent_Reset(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	// Add some trace entries
	agent.trace = []ReActStep{
		{Iteration: 1, Thought: "Test"},
		{Iteration: 2, Thought: "Test2"},
	}

	err := agent.Reset()
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	if len(agent.trace) != 0 {
		t.Errorf("trace length after reset = %d, want 0", len(agent.trace))
	}
}

// TestReActAgent_GetTrace tests trace retrieval.
func TestReActAgent_GetTrace(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	// Add trace entries
	agent.trace = []ReActStep{
		{Iteration: 1, Thought: "First"},
		{Iteration: 2, Thought: "Second"},
	}

	trace := agent.GetTrace()

	if len(trace) != 2 {
		t.Errorf("trace length = %d, want 2", len(trace))
	}

	if trace[0].Thought != "First" {
		t.Errorf("trace[0].Thought = %q, want %q", trace[0].Thought, "First")
	}
}

// TestReActAgent_BuildPrompt tests prompt generation.
func TestReActAgent_BuildPrompt(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewReActAgent(llm)

	// Add a tool
	tool := mocks.NewMockTool("calculator", "Performs arithmetic operations")
	agent.AddTool(tool)

	prompt := agent.buildPrompt("What is 2 + 2?")

	// Check prompt contains key elements
	if !strings.Contains(prompt, "calculator") {
		t.Error("prompt should contain tool name")
	}

	if !strings.Contains(prompt, "Performs arithmetic operations") {
		t.Error("prompt should contain tool description")
	}

	if !strings.Contains(prompt, "What is 2 + 2?") {
		t.Error("prompt should contain user question")
	}

	if !strings.Contains(prompt, "step-by-step") {
		t.Error("prompt should contain reasoning instruction")
	}
}
