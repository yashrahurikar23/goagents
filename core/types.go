package core

import "time"

// Message represents a single message in a conversation.
type Message struct {
	// Role identifies who sent the message: "user", "assistant", "system", or "tool"
	Role string

	// Content is the text content of the message
	Content string

	// Name optionally identifies the speaker (for multi-agent scenarios)
	Name string

	// ToolCallID is set for tool result messages, linking to the tool call
	ToolCallID string

	// ToolCalls contains tool calls made by the assistant (for function calling)
	ToolCalls []ToolCall

	// Meta contains additional metadata (model-specific fields, etc.)
	Meta map[string]interface{}
}

// Response represents a response from an LLM or agent.
type Response struct {
	// Content is the main text response
	Content string

	// ToolCalls contains any tools that were invoked
	ToolCalls []ToolCall

	// Meta contains metadata about the response
	// Common fields: "tokens_used", "model", "latency_ms", "cost"
	Meta map[string]interface{}
}

// ToolCall represents a single invocation of a tool.
type ToolCall struct {
	// ID is a unique identifier for this call
	ID string

	// Name is the tool that was called
	Name string

	// Args contains the arguments passed to the tool
	Args map[string]interface{}

	// Result contains the tool's return value (set after execution)
	Result interface{}

	// Error contains any error that occurred during execution
	Error error

	// Duration is how long the tool took to execute
	Duration time.Duration
}

// ToolSchema defines a tool's interface.
type ToolSchema struct {
	// Name is the tool's unique identifier
	Name string

	// Description explains what the tool does
	Description string

	// Parameters defines the tool's inputs
	Parameters []Parameter
}

// Parameter defines a single parameter for a tool.
type Parameter struct {
	// Name is the parameter name
	Name string

	// Type is the parameter type: "string", "number", "boolean", "object", "array"
	Type string

	// Description explains what this parameter is for
	Description string

	// Required indicates if this parameter must be provided
	Required bool

	// Enum optionally restricts values to a specific set
	Enum []interface{}

	// Default is the default value if not provided
	Default interface{}
}

// NewMessage creates a new message with the given role and content.
func NewMessage(role, content string) Message {
	return Message{
		Role:    role,
		Content: content,
		Meta:    make(map[string]interface{}),
	}
}

// SystemMessage creates a system message.
func SystemMessage(content string) Message {
	return NewMessage("system", content)
}

// UserMessage creates a user message.
func UserMessage(content string) Message {
	return NewMessage("user", content)
}

// AssistantMessage creates an assistant message.
func AssistantMessage(content string) Message {
	return NewMessage("assistant", content)
}
