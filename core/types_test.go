package core

import (
	"testing"
	"time"
)

// TestNewMessage tests the NewMessage constructor
func TestNewMessage(t *testing.T) {
	msg := NewMessage("user", "Hello, world!")

	if msg.Role != "user" {
		t.Errorf("expected role 'user', got %q", msg.Role)
	}

	if msg.Content != "Hello, world!" {
		t.Errorf("expected content 'Hello, world!', got %q", msg.Content)
	}

	if msg.Meta == nil {
		t.Error("expected Meta map to be initialized")
	}

	if msg.Name != "" {
		t.Errorf("expected Name to be empty, got %q", msg.Name)
	}
}

// TestSystemMessage tests the SystemMessage helper
func TestSystemMessage(t *testing.T) {
	msg := SystemMessage("You are a helpful assistant")

	if msg.Role != "system" {
		t.Errorf("expected role 'system', got %q", msg.Role)
	}

	if msg.Content != "You are a helpful assistant" {
		t.Errorf("expected system content, got %q", msg.Content)
	}

	if msg.Meta == nil {
		t.Error("expected Meta map to be initialized")
	}
}

// TestUserMessage tests the UserMessage helper
func TestUserMessage(t *testing.T) {
	msg := UserMessage("What is the weather?")

	if msg.Role != "user" {
		t.Errorf("expected role 'user', got %q", msg.Role)
	}

	if msg.Content != "What is the weather?" {
		t.Errorf("expected user content, got %q", msg.Content)
	}

	if msg.Meta == nil {
		t.Error("expected Meta map to be initialized")
	}
}

// TestAssistantMessage tests the AssistantMessage helper
func TestAssistantMessage(t *testing.T) {
	msg := AssistantMessage("The weather is sunny")

	if msg.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", msg.Role)
	}

	if msg.Content != "The weather is sunny" {
		t.Errorf("expected assistant content, got %q", msg.Content)
	}

	if msg.Meta == nil {
		t.Error("expected Meta map to be initialized")
	}
}

// TestMessage_WithName tests setting the Name field
func TestMessage_WithName(t *testing.T) {
	msg := UserMessage("Hello")
	msg.Name = "Alice"

	if msg.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", msg.Name)
	}
}

// TestMessage_WithMeta tests setting metadata
func TestMessage_WithMeta(t *testing.T) {
	msg := UserMessage("Hello")
	msg.Meta["custom_field"] = "custom_value"
	msg.Meta["count"] = 42

	if val, ok := msg.Meta["custom_field"].(string); !ok || val != "custom_value" {
		t.Errorf("expected custom_field='custom_value', got %v", msg.Meta["custom_field"])
	}

	if val, ok := msg.Meta["count"].(int); !ok || val != 42 {
		t.Errorf("expected count=42, got %v", msg.Meta["count"])
	}
}

// TestResponse_Empty tests an empty Response
func TestResponse_Empty(t *testing.T) {
	resp := Response{}

	if resp.Content != "" {
		t.Errorf("expected empty content, got %q", resp.Content)
	}

	if resp.ToolCalls != nil {
		t.Errorf("expected nil ToolCalls, got %v", resp.ToolCalls)
	}

	if resp.Meta != nil {
		t.Errorf("expected nil Meta, got %v", resp.Meta)
	}
}

// TestResponse_WithContent tests a Response with content
func TestResponse_WithContent(t *testing.T) {
	resp := Response{
		Content: "This is a response",
	}

	if resp.Content != "This is a response" {
		t.Errorf("expected 'This is a response', got %q", resp.Content)
	}
}

// TestResponse_WithToolCalls tests a Response with tool calls
func TestResponse_WithToolCalls(t *testing.T) {
	resp := Response{
		Content: "",
		ToolCalls: []ToolCall{
			{
				ID:   "call_123",
				Name: "calculator",
				Args: map[string]interface{}{
					"operation": "add",
					"a":         5,
					"b":         3,
				},
			},
		},
	}

	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}

	call := resp.ToolCalls[0]
	if call.ID != "call_123" {
		t.Errorf("expected ID 'call_123', got %q", call.ID)
	}

	if call.Name != "calculator" {
		t.Errorf("expected Name 'calculator', got %q", call.Name)
	}

	if len(call.Args) != 3 {
		t.Errorf("expected 3 args, got %d", len(call.Args))
	}
}

// TestResponse_WithMeta tests a Response with metadata
func TestResponse_WithMeta(t *testing.T) {
	resp := Response{
		Content: "Response",
		Meta: map[string]interface{}{
			"tokens_used": 150,
			"model":       "gpt-4",
			"latency_ms":  234.5,
		},
	}

	if tokens, ok := resp.Meta["tokens_used"].(int); !ok || tokens != 150 {
		t.Errorf("expected tokens_used=150, got %v", resp.Meta["tokens_used"])
	}

	if model, ok := resp.Meta["model"].(string); !ok || model != "gpt-4" {
		t.Errorf("expected model='gpt-4', got %v", resp.Meta["model"])
	}

	if latency, ok := resp.Meta["latency_ms"].(float64); !ok || latency != 234.5 {
		t.Errorf("expected latency_ms=234.5, got %v", resp.Meta["latency_ms"])
	}
}

// TestToolCall_Complete tests a complete ToolCall
func TestToolCall_Complete(t *testing.T) {
	call := ToolCall{
		ID:   "call_456",
		Name: "search",
		Args: map[string]interface{}{
			"query": "golang testing",
		},
		Result:   "Found 42 results",
		Error:    nil,
		Duration: 150 * time.Millisecond,
	}

	if call.ID != "call_456" {
		t.Errorf("expected ID 'call_456', got %q", call.ID)
	}

	if call.Name != "search" {
		t.Errorf("expected Name 'search', got %q", call.Name)
	}

	if call.Result.(string) != "Found 42 results" {
		t.Errorf("expected result 'Found 42 results', got %v", call.Result)
	}

	if call.Error != nil {
		t.Errorf("expected no error, got %v", call.Error)
	}

	if call.Duration != 150*time.Millisecond {
		t.Errorf("expected duration 150ms, got %v", call.Duration)
	}
}

// TestToolCall_WithError tests a ToolCall with an error
func TestToolCall_WithError(t *testing.T) {
	call := ToolCall{
		ID:    "call_789",
		Name:  "calculator",
		Args:  map[string]interface{}{"operation": "divide", "a": 10, "b": 0},
		Error: &ErrToolExecution{ToolName: "calculator", Err: nil},
	}

	if call.Error == nil {
		t.Error("expected error, got nil")
	}

	if call.Result != nil {
		t.Errorf("expected nil result on error, got %v", call.Result)
	}
}

// TestToolSchema tests ToolSchema structure
func TestToolSchema(t *testing.T) {
	schema := ToolSchema{
		Name:        "calculator",
		Description: "Performs basic arithmetic operations",
		Parameters: []Parameter{
			{
				Name:        "operation",
				Type:        "string",
				Description: "The operation to perform: add, subtract, multiply, divide",
				Required:    true,
				Enum:        []interface{}{"add", "subtract", "multiply", "divide"},
			},
			{
				Name:        "a",
				Type:        "number",
				Description: "First number",
				Required:    true,
			},
			{
				Name:        "b",
				Type:        "number",
				Description: "Second number",
				Required:    true,
			},
		},
	}

	if schema.Name != "calculator" {
		t.Errorf("expected name 'calculator', got %q", schema.Name)
	}

	if schema.Description != "Performs basic arithmetic operations" {
		t.Errorf("unexpected description: %q", schema.Description)
	}

	if len(schema.Parameters) != 3 {
		t.Fatalf("expected 3 parameters, got %d", len(schema.Parameters))
	}

	// Check first parameter (operation)
	param := schema.Parameters[0]
	if param.Name != "operation" {
		t.Errorf("expected param name 'operation', got %q", param.Name)
	}

	if param.Type != "string" {
		t.Errorf("expected param type 'string', got %q", param.Type)
	}

	if !param.Required {
		t.Error("expected operation to be required")
	}

	if len(param.Enum) != 4 {
		t.Errorf("expected 4 enum values, got %d", len(param.Enum))
	}
}

// TestParameter_Required tests required parameters
func TestParameter_Required(t *testing.T) {
	param := Parameter{
		Name:        "api_key",
		Type:        "string",
		Description: "API key for authentication",
		Required:    true,
	}

	if !param.Required {
		t.Error("expected parameter to be required")
	}

	if param.Default != nil {
		t.Errorf("expected no default for required param, got %v", param.Default)
	}
}

// TestParameter_Optional tests optional parameters with defaults
func TestParameter_Optional(t *testing.T) {
	param := Parameter{
		Name:        "timeout",
		Type:        "number",
		Description: "Request timeout in seconds",
		Required:    false,
		Default:     30,
	}

	if param.Required {
		t.Error("expected parameter to be optional")
	}

	if param.Default.(int) != 30 {
		t.Errorf("expected default 30, got %v", param.Default)
	}
}

// TestParameter_WithEnum tests parameter with enum values
func TestParameter_WithEnum(t *testing.T) {
	param := Parameter{
		Name:        "model",
		Type:        "string",
		Description: "LLM model to use",
		Required:    false,
		Enum:        []interface{}{"gpt-4", "gpt-3.5-turbo", "claude-3"},
		Default:     "gpt-4",
	}

	if len(param.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(param.Enum))
	}

	// Check enum contains expected values
	hasGPT4 := false
	for _, val := range param.Enum {
		if val.(string) == "gpt-4" {
			hasGPT4 = true
			break
		}
	}

	if !hasGPT4 {
		t.Error("expected enum to contain 'gpt-4'")
	}

	if param.Default.(string) != "gpt-4" {
		t.Errorf("expected default 'gpt-4', got %v", param.Default)
	}
}

// TestParameter_Types tests different parameter types
func TestParameter_Types(t *testing.T) {
	tests := []struct {
		name     string
		param    Parameter
		typeName string
	}{
		{
			name: "string parameter",
			param: Parameter{
				Name: "text",
				Type: "string",
			},
			typeName: "string",
		},
		{
			name: "number parameter",
			param: Parameter{
				Name: "count",
				Type: "number",
			},
			typeName: "number",
		},
		{
			name: "boolean parameter",
			param: Parameter{
				Name: "enabled",
				Type: "boolean",
			},
			typeName: "boolean",
		},
		{
			name: "object parameter",
			param: Parameter{
				Name: "config",
				Type: "object",
			},
			typeName: "object",
		},
		{
			name: "array parameter",
			param: Parameter{
				Name: "items",
				Type: "array",
			},
			typeName: "array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.param.Type != tt.typeName {
				t.Errorf("expected type %q, got %q", tt.typeName, tt.param.Type)
			}
		})
	}
}

// TestMessage_EmptyContent tests message with empty content
func TestMessage_EmptyContent(t *testing.T) {
	msg := UserMessage("")

	if msg.Content != "" {
		t.Errorf("expected empty content, got %q", msg.Content)
	}

	if msg.Role != "user" {
		t.Errorf("expected role 'user' even with empty content, got %q", msg.Role)
	}
}

// TestResponse_MultipleToolCalls tests response with multiple tool calls
func TestResponse_MultipleToolCalls(t *testing.T) {
	resp := Response{
		Content: "I'll search and calculate for you",
		ToolCalls: []ToolCall{
			{
				ID:   "call_1",
				Name: "search",
				Args: map[string]interface{}{"query": "weather"},
			},
			{
				ID:   "call_2",
				Name: "calculator",
				Args: map[string]interface{}{"operation": "add", "a": 5, "b": 3},
			},
		},
	}

	if len(resp.ToolCalls) != 2 {
		t.Fatalf("expected 2 tool calls, got %d", len(resp.ToolCalls))
	}

	if resp.ToolCalls[0].Name != "search" {
		t.Errorf("expected first call to be 'search', got %q", resp.ToolCalls[0].Name)
	}

	if resp.ToolCalls[1].Name != "calculator" {
		t.Errorf("expected second call to be 'calculator', got %q", resp.ToolCalls[1].Name)
	}
}

// Benchmark tests

func BenchmarkNewMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewMessage("user", "Hello")
	}
}

func BenchmarkUserMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UserMessage("Hello")
	}
}

func BenchmarkSystemMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SystemMessage("You are a helpful assistant")
	}
}

func BenchmarkAssistantMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = AssistantMessage("Response")
	}
}
