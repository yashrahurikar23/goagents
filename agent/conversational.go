package agent

import (
	"context"
	"fmt"

	"github.com/yashrahurikar23/goagents/core"
)

// ConversationalAgent maintains conversation history and provides
// memory management for multi-turn conversations.
//
// This agent is designed for chatbots and assistants that need to:
// - Remember previous conversation turns
// - Maintain context over long conversations
// - Manage token limits through memory strategies
// - Support streaming responses
//
// Memory Strategies:
// - Windowing: Keep last N messages
// - Summarization: Compress old messages into summaries
// - Selective: Keep important messages (system, tool results, etc.)
//
// Example usage:
//
//	llm := openai.New(openai.WithAPIKey("sk-..."))
//	agent := NewConversationalAgent(llm)
//
//	// First turn
//	resp1, _ := agent.Run(ctx, "Hi, I'm Alice")
//	// "Hello Alice! How can I help you today?"
//
//	// Second turn (remembers Alice)
//	resp2, _ := agent.Run(ctx, "What's my name?")
//	// "Your name is Alice."
type ConversationalAgent struct {
	llm             core.LLM
	messages        []core.Message
	systemPrompt    string
	memoryStrategy  MemoryStrategy
	maxMessages     int
	summarizationLM core.LLM // Optional separate LLM for summarization
}

// MemoryStrategy defines how conversation history is managed.
type MemoryStrategy string

const (
	// MemoryStrategyWindow keeps the last N messages.
	MemoryStrategyWindow MemoryStrategy = "window"

	// MemoryStrategySummarize compresses old messages into summaries.
	MemoryStrategySummarize MemoryStrategy = "summarize"

	// MemoryStrategySelective keeps important messages and summarizes the rest.
	MemoryStrategySelective MemoryStrategy = "selective"

	// MemoryStrategyAll keeps all messages (no limit).
	MemoryStrategyAll MemoryStrategy = "all"
)

// ConversationalAgentOption configures a ConversationalAgent.
type ConversationalAgentOption func(*ConversationalAgent)

// ConvWithSystemPrompt sets the system prompt.
func ConvWithSystemPrompt(prompt string) ConversationalAgentOption {
	return func(a *ConversationalAgent) {
		a.systemPrompt = prompt
	}
}

// ConvWithMemoryStrategy sets the memory management strategy.
func ConvWithMemoryStrategy(strategy MemoryStrategy) ConversationalAgentOption {
	return func(a *ConversationalAgent) {
		a.memoryStrategy = strategy
	}
}

// ConvWithMaxMessages sets the maximum number of messages to keep (for windowing).
func ConvWithMaxMessages(max int) ConversationalAgentOption {
	return func(a *ConversationalAgent) {
		a.maxMessages = max
	}
}

// ConvWithSummarizationLLM sets a separate LLM for summarization (cheaper model).
func ConvWithSummarizationLLM(llm core.LLM) ConversationalAgentOption {
	return func(a *ConversationalAgent) {
		a.summarizationLM = llm
	}
}

// NewConversationalAgent creates a conversational agent with memory management.
func NewConversationalAgent(llm core.LLM, opts ...ConversationalAgentOption) *ConversationalAgent {
	agent := &ConversationalAgent{
		llm:            llm,
		messages:       make([]core.Message, 0),
		systemPrompt:   "You are a helpful assistant.",
		memoryStrategy: MemoryStrategyWindow,
		maxMessages:    20, // Default: keep last 20 messages (10 turns)
	}

	for _, opt := range opts {
		opt(agent)
	}

	// Initialize with system prompt
	if agent.systemPrompt != "" {
		agent.messages = append(agent.messages, core.SystemMessage(agent.systemPrompt))
	}

	return agent
}

// Run processes a user message and returns a response.
func (a *ConversationalAgent) Run(ctx context.Context, input string) (*core.Response, error) {
	// Add user message
	a.messages = append(a.messages, core.UserMessage(input))

	// Apply memory management before calling LLM
	if err := a.applyMemoryStrategy(ctx); err != nil {
		return nil, fmt.Errorf("memory management failed: %w", err)
	}

	// Call LLM with conversation history
	response, err := a.llm.Chat(ctx, a.messages)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Add assistant response to history
	a.messages = append(a.messages, core.AssistantMessage(response.Content))

	return response, nil
}

// Chat is an alias for Run to match conversational patterns.
func (a *ConversationalAgent) Chat(ctx context.Context, message string) (*core.Response, error) {
	return a.Run(ctx, message)
}

// Reset clears the conversation history.
func (a *ConversationalAgent) Reset() error {
	a.messages = make([]core.Message, 0)
	if a.systemPrompt != "" {
		a.messages = append(a.messages, core.SystemMessage(a.systemPrompt))
	}
	return nil
}

// GetMessages returns the conversation history.
func (a *ConversationalAgent) GetMessages() []core.Message {
	return a.messages
}

// GetMessageCount returns the number of messages in history.
func (a *ConversationalAgent) GetMessageCount() int {
	return len(a.messages)
}

// AddTool is required by the Agent interface but not implemented.
// ConversationalAgent focuses on conversation, not tool execution.
func (a *ConversationalAgent) AddTool(tool core.Tool) error {
	return fmt.Errorf("ConversationalAgent does not support tools")
}

// applyMemoryStrategy manages conversation history based on strategy.
func (a *ConversationalAgent) applyMemoryStrategy(ctx context.Context) error {
	switch a.memoryStrategy {
	case MemoryStrategyAll:
		// Keep everything, no management needed
		return nil

	case MemoryStrategyWindow:
		return a.applyWindowStrategy()

	case MemoryStrategySummarize:
		return a.applySummarizeStrategy(ctx)

	case MemoryStrategySelective:
		return a.applySelectiveStrategy(ctx)

	default:
		return fmt.Errorf("unknown memory strategy: %s", a.memoryStrategy)
	}
}

// applyWindowStrategy keeps only the last N messages.
func (a *ConversationalAgent) applyWindowStrategy() error {
	if len(a.messages) <= a.maxMessages {
		return nil
	}

	// Always keep system prompt (first message)
	systemMsg := a.messages[0]

	// Keep last N messages
	startIdx := len(a.messages) - a.maxMessages + 1 // +1 for system message
	if startIdx < 1 {
		startIdx = 1
	}

	a.messages = append([]core.Message{systemMsg}, a.messages[startIdx:]...)
	return nil
}

// applySummarizeStrategy summarizes old messages and keeps recent ones.
func (a *ConversationalAgent) applySummarizeStrategy(ctx context.Context) error {
	// If we're under the limit, no need to summarize
	if len(a.messages) <= a.maxMessages {
		return nil
	}

	// Keep system message and recent messages
	systemMsg := a.messages[0]

	// Messages to summarize (middle portion)
	summarizeEndIdx := len(a.messages) - (a.maxMessages / 2)
	if summarizeEndIdx <= 1 {
		return nil // Nothing to summarize
	}

	messagesToSummarize := a.messages[1:summarizeEndIdx]

	// Use summarization LLM or main LLM
	summarizerLLM := a.summarizationLM
	if summarizerLLM == nil {
		summarizerLLM = a.llm
	}

	// Create summarization prompt
	conversationText := a.formatMessagesForSummarization(messagesToSummarize)
	summaryPrompt := fmt.Sprintf(
		"Summarize the following conversation concisely, preserving key facts and context:\n\n%s\n\nSummary:",
		conversationText,
	)

	// Get summary
	summary, err := summarizerLLM.Complete(ctx, summaryPrompt)
	if err != nil {
		// If summarization fails, fall back to window strategy
		return a.applyWindowStrategy()
	}

	// Rebuild messages: system + summary + recent messages
	summaryMsg := core.Message{
		Role:    "system",
		Content: fmt.Sprintf("Previous conversation summary: %s", summary),
	}

	a.messages = append(
		[]core.Message{systemMsg, summaryMsg},
		a.messages[summarizeEndIdx:]...,
	)

	return nil
}

// applySelectiveStrategy keeps important messages and summarizes less important ones.
func (a *ConversationalAgent) applySelectiveStrategy(ctx context.Context) error {
	// If we're under the limit, no need to manage
	if len(a.messages) <= a.maxMessages {
		return nil
	}

	// Keep system messages, recent messages, and messages with metadata
	systemMsg := a.messages[0]

	// Identify important messages
	importantMsgs := make([]core.Message, 0)
	regularMsgs := make([]core.Message, 0)

	recentStartIdx := len(a.messages) - (a.maxMessages / 2)

	for i, msg := range a.messages[1:] { // Skip system message
		actualIdx := i + 1

		// Keep recent messages
		if actualIdx >= recentStartIdx {
			regularMsgs = append(regularMsgs, msg)
			continue
		}

		// Keep messages with metadata (tool calls, etc.)
		if len(msg.Meta) > 0 || msg.Role == "tool" {
			importantMsgs = append(importantMsgs, msg)
		} else {
			regularMsgs = append(regularMsgs, msg)
		}
	}

	// If we still have too many, summarize the regular ones
	if len(importantMsgs)+len(regularMsgs) > a.maxMessages-1 {
		// Use summarization LLM or main LLM
		summarizerLLM := a.summarizationLM
		if summarizerLLM == nil {
			summarizerLLM = a.llm
		}

		conversationText := a.formatMessagesForSummarization(regularMsgs[:len(regularMsgs)/2])
		summaryPrompt := fmt.Sprintf(
			"Summarize the following conversation concisely:\n\n%s\n\nSummary:",
			conversationText,
		)

		summary, err := summarizerLLM.Complete(ctx, summaryPrompt)
		if err != nil {
			return a.applyWindowStrategy()
		}

		summaryMsg := core.Message{
			Role:    "system",
			Content: fmt.Sprintf("Conversation summary: %s", summary),
		}

		a.messages = append(
			[]core.Message{systemMsg, summaryMsg},
			append(importantMsgs, regularMsgs[len(regularMsgs)/2:]...)...,
		)
	}

	return nil
}

// formatMessagesForSummarization formats messages into text for summarization.
func (a *ConversationalAgent) formatMessagesForSummarization(messages []core.Message) string {
	var text string
	for _, msg := range messages {
		text += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}
	return text
}

// SetSystemPrompt updates the system prompt and resets conversation.
func (a *ConversationalAgent) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
	a.Reset()
}

// GetSystemPrompt returns the current system prompt.
func (a *ConversationalAgent) GetSystemPrompt() string {
	return a.systemPrompt
}

// ExportConversation exports the conversation history as a formatted string.
func (a *ConversationalAgent) ExportConversation() string {
	return a.formatMessagesForSummarization(a.messages)
}
