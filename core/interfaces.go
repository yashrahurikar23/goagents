// Package core provides fundamental types and interfaces for the GoAgent framework.
//
// This package contains the core abstractions that all other packages build upon,
// including interfaces for LLMs, tools, and agents, as well as common types like
// messages and responses.
//
// Example usage:
//
//	import "github.com/yashrahurikar/goagents/core"
//
//	// Implement the LLM interface
//	type MyLLM struct {}
//
//	func (m *MyLLM) Chat(ctx context.Context, messages []core.Message) (*core.Response, error) {
//	    // Implementation
//	}
package core

import "context"

// LLM is the interface for language model providers.
// All LLM integrations (OpenAI, Anthropic, Ollama, etc.) implement this interface.
type LLM interface {
	// Chat sends a conversation and receives a response.
	// The messages slice should contain the conversation history.
	Chat(ctx context.Context, messages []Message) (*Response, error)

	// Complete sends a single prompt and receives a completion.
	// This is a convenience wrapper around Chat for simple use cases.
	Complete(ctx context.Context, prompt string) (string, error)
}

// Tool represents something an agent can use to accomplish tasks.
// Tools can be functions, APIs, databases, search engines, etc.
type Tool interface {
	// Name returns the tool's unique identifier (e.g., "calculator", "web_search").
	Name() string

	// Description explains what the tool does.
	// This is shown to the LLM so it can decide when to use the tool.
	Description() string

	// Schema returns the tool's parameter schema.
	// This defines what arguments the tool accepts.
	Schema() *ToolSchema

	// Execute runs the tool with the given arguments.
	// The args map contains parameter name → value mappings.
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// Agent coordinates between an LLM and a set of tools to accomplish tasks.
// Agents implement different strategies (function calling, ReAct, workflows, etc.).
type Agent interface {
	// Run executes the agent with the given input and returns a response.
	// The agent decides which tools to use (if any) and how to use them.
	Run(ctx context.Context, input string) (*Response, error)

	// AddTool registers a new tool that the agent can use.
	AddTool(tool Tool) error

	// Reset clears any conversation history or state.
	Reset() error
}
