package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/yashrahurikar/goagents/core"
	"github.com/yashrahurikar/goagents/tests/mocks"
)

// TestFunctionAgent_NewAgent tests agent creation.
func TestFunctionAgent_NewAgent(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	if agent == nil {
		t.Fatal("NewFunctionAgent() returned nil")
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

	if agent.maxIter != 5 {
		t.Errorf("agent.maxIter = %d, want 5", agent.maxIter)
	}
}

// TestFunctionAgent_WithOptions tests agent configuration options.
func TestFunctionAgent_WithOptions(t *testing.T) {
	llm := mocks.NewMockLLM()
	customPrompt := "You are a math tutor"
	maxIter := 10

	agent := NewFunctionAgent(
		llm,
		WithSystemPrompt(customPrompt),
		WithMaxIterations(maxIter),
	)

	if agent.systemPrompt != customPrompt {
		t.Errorf("systemPrompt = %q, want %q", agent.systemPrompt, customPrompt)
	}

	if agent.maxIter != maxIter {
		t.Errorf("maxIter = %d, want %d", agent.maxIter, maxIter)
	}
}

// TestFunctionAgent_AddTool tests tool registration.
func TestFunctionAgent_AddTool(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	tool := mocks.NewMockTool("calculator", "Performs calculations")

	err := agent.AddTool(tool)
	if err != nil {
		t.Fatalf("AddTool() error = %v", err)
	}

	if len(agent.tools) != 1 {
		t.Errorf("len(agent.tools) = %d, want 1", len(agent.tools))
	}

	if _, exists := agent.tools["calculator"]; !exists {
		t.Error("tool 'calculator' not found in agent.tools")
	}
}

// TestFunctionAgent_AddTool_Nil tests adding a nil tool.
func TestFunctionAgent_AddTool_Nil(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	err := agent.AddTool(nil)
	if err == nil {
		t.Fatal("AddTool(nil) expected error, got nil")
	}

	var invalidArgErr *core.ErrInvalidArgument
	if !errors.As(err, &invalidArgErr) {
		t.Errorf("error type = %T, want *core.ErrInvalidArgument", err)
	}
}

// TestFunctionAgent_AddTool_EmptyName tests adding a tool with empty name.
func TestFunctionAgent_AddTool_EmptyName(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	tool := mocks.NewMockTool("", "Empty name tool")

	err := agent.AddTool(tool)
	if err == nil {
		t.Fatal("AddTool() expected error for empty name, got nil")
	}

	var invalidArgErr *core.ErrInvalidArgument
	if !errors.As(err, &invalidArgErr) {
		t.Errorf("error type = %T, want *core.ErrInvalidArgument", err)
	}
}

// TestFunctionAgent_AddTool_Duplicate tests adding a duplicate tool.
func TestFunctionAgent_AddTool_Duplicate(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	tool1 := mocks.NewMockTool("calculator", "First calculator")
	tool2 := mocks.NewMockTool("calculator", "Second calculator")

	err := agent.AddTool(tool1)
	if err != nil {
		t.Fatalf("AddTool(tool1) error = %v", err)
	}

	err = agent.AddTool(tool2)
	if err == nil {
		t.Fatal("AddTool(tool2) expected error for duplicate, got nil")
	}
}

// TestFunctionAgent_Reset tests resetting conversation history.
func TestFunctionAgent_Reset(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	// Add some messages
	agent.messages = append(agent.messages, core.UserMessage("Hello"))
	agent.messages = append(agent.messages, core.AssistantMessage("Hi there"))

	if len(agent.messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(agent.messages))
	}

	err := agent.Reset()
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Should have system prompt
	if len(agent.messages) != 1 {
		t.Errorf("len(messages) after reset = %d, want 1 (system prompt)", len(agent.messages))
	}

	if agent.messages[0].Role != "system" {
		t.Errorf("first message role = %q, want \"system\"", agent.messages[0].Role)
	}
}

// TestFunctionAgent_GetMessages tests retrieving conversation history.
func TestFunctionAgent_GetMessages(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	// Add some messages
	agent.messages = append(agent.messages, core.UserMessage("Test 1"))
	agent.messages = append(agent.messages, core.AssistantMessage("Response 1"))

	messages := agent.GetMessages()

	if len(messages) != 2 {
		t.Errorf("len(GetMessages()) = %d, want 2", len(messages))
	}

	if messages[0].Content != "Test 1" {
		t.Errorf("messages[0].Content = %q, want \"Test 1\"", messages[0].Content)
	}
}

// TestFunctionAgent_Run_RequiresOpenAI tests that Run() requires OpenAI client.
func TestFunctionAgent_Run_RequiresOpenAI(t *testing.T) {
	// Use generic mock LLM (not OpenAI)
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	ctx := context.Background()
	_, err := agent.Run(ctx, "Hello")

	if err == nil {
		t.Fatal("Run() expected error for non-OpenAI LLM, got nil")
	}

	expectedMsg := "FunctionAgent requires an LLM that supports function calling (OpenAI)"
	if err.Error() != expectedMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedMsg)
	}
}

// TestFunctionAgent_ConvertToolsToFunctions tests basic tool to function conversion.
func TestFunctionAgent_ConvertToolsToFunctions(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	// Add a simple tool
	tool := mocks.NewMockTool("calculator", "Performs arithmetic")
	agent.AddTool(tool)

	functions := agent.convertToolsToFunctions()

	if len(functions) != 1 {
		t.Fatalf("len(functions) = %d, want 1", len(functions))
	}

	fn := functions[0]
	if fn.Type != "function" {
		t.Errorf("function.Type = %q, want \"function\"", fn.Type)
	}

	if fn.Function == nil {
		t.Fatal("function.Function is nil")
	}

	if fn.Function.Name != "calculator" {
		t.Errorf("function.Name = %q, want \"calculator\"", fn.Function.Name)
	}

	if fn.Function.Description != "Performs arithmetic" {
		t.Errorf("function.Description = %q, want \"Performs arithmetic\"", fn.Function.Description)
	}
}

// TestFunctionAgent_ConvertToOpenAIMessages tests message conversion.
func TestFunctionAgent_ConvertToOpenAIMessages(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewFunctionAgent(llm)

	messages := []core.Message{
		core.SystemMessage("System prompt"),
		core.UserMessage("Hello"),
		core.AssistantMessage("Hi"),
		{
			Role:       "tool",
			Content:    "42",
			Name:       "calculator",
			ToolCallID: "call_123",
		},
	}

	openaiMsgs := agent.convertToOpenAIMessages(messages)

	if len(openaiMsgs) != 4 {
		t.Fatalf("len(openaiMsgs) = %d, want 4", len(openaiMsgs))
	}

	// Check system message
	if openaiMsgs[0].Role != "system" {
		t.Errorf("openaiMsgs[0].Role = %q, want \"system\"", openaiMsgs[0].Role)
	}

	// Check tool message
	if openaiMsgs[3].Role != "tool" {
		t.Errorf("openaiMsgs[3].Role = %q, want \"tool\"", openaiMsgs[3].Role)
	}

	if openaiMsgs[3].ToolCallID != "call_123" {
		t.Errorf("openaiMsgs[3].ToolCallID = %q, want \"call_123\"", openaiMsgs[3].ToolCallID)
	}
}
