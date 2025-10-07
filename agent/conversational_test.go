package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/yashrahurikar23/goagents/core"
	"github.com/yashrahurikar23/goagents/tests/mocks"
)

// TestConversationalAgent_NewAgent tests agent creation.
func TestConversationalAgent_NewAgent(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewConversationalAgent(llm)

	if agent == nil {
		t.Fatal("NewConversationalAgent() returned nil")
	}

	if agent.llm == nil {
		t.Error("agent.llm is nil")
	}

	if agent.memoryStrategy != MemoryStrategyWindow {
		t.Errorf("memoryStrategy = %v, want %v", agent.memoryStrategy, MemoryStrategyWindow)
	}

	if agent.maxMessages != 20 {
		t.Errorf("maxMessages = %d, want 20", agent.maxMessages)
	}

	// Should have system message
	if len(agent.messages) != 1 {
		t.Errorf("len(messages) = %d, want 1 (system message)", len(agent.messages))
	}
}

// TestConversationalAgent_WithOptions tests configuration options.
func TestConversationalAgent_WithOptions(t *testing.T) {
	llm := mocks.NewMockLLM()
	customPrompt := "Custom chatbot prompt"

	agent := NewConversationalAgent(
		llm,
		ConvWithSystemPrompt(customPrompt),
		ConvWithMemoryStrategy(MemoryStrategySummarize),
		ConvWithMaxMessages(10),
	)

	if agent.systemPrompt != customPrompt {
		t.Errorf("systemPrompt = %q, want %q", agent.systemPrompt, customPrompt)
	}

	if agent.memoryStrategy != MemoryStrategySummarize {
		t.Errorf("memoryStrategy = %v, want %v", agent.memoryStrategy, MemoryStrategySummarize)
	}

	if agent.maxMessages != 10 {
		t.Errorf("maxMessages = %d, want 10", agent.maxMessages)
	}
}

// TestConversationalAgent_Run tests basic conversation.
func TestConversationalAgent_Run(t *testing.T) {
	llm := mocks.NewMockLLM()
	llm.WithChatResponse("Hello! How can I help you?", nil)

	agent := NewConversationalAgent(llm)

	ctx := context.Background()
	response, err := agent.Run(ctx, "Hi")

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if response.Content != "Hello! How can I help you?" {
		t.Errorf("response.Content = %q, want greeting", response.Content)
	}

	// Check message history
	messages := agent.GetMessages()
	if len(messages) != 3 { // system + user + assistant
		t.Errorf("len(messages) = %d, want 3", len(messages))
	}

	if messages[1].Role != "user" || messages[1].Content != "Hi" {
		t.Error("user message not recorded correctly")
	}

	if messages[2].Role != "assistant" {
		t.Error("assistant message not recorded correctly")
	}
}

// TestConversationalAgent_Chat tests Chat alias.
func TestConversationalAgent_Chat(t *testing.T) {
	llm := mocks.NewMockLLM()
	llm.WithChatResponse("Test response", nil)

	agent := NewConversationalAgent(llm)

	ctx := context.Background()
	response, err := agent.Chat(ctx, "Test input")

	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if response.Content != "Test response" {
		t.Errorf("response.Content = %q, want %q", response.Content, "Test response")
	}
}

// TestConversationalAgent_MultiTurn tests multi-turn conversation.
func TestConversationalAgent_MultiTurn(t *testing.T) {
	llm := mocks.NewMockLLM()
	
	// Configure sequential responses
	responses := []string{"Hello!", "Nice to meet you!", "I'm doing well, thanks!"}
	responseIndex := 0
	llm.ChatFunc = func(ctx context.Context, messages []core.Message) (*core.Response, error) {
		if responseIndex >= len(responses) {
			return &core.Response{Content: responses[len(responses)-1]}, nil
		}
		resp := &core.Response{Content: responses[responseIndex]}
		responseIndex++
		return resp, nil
	}

	agent := NewConversationalAgent(llm)
	ctx := context.Background()

	// Turn 1
	resp1, _ := agent.Run(ctx, "Hi")
	if resp1.Content != "Hello!" {
		t.Errorf("turn 1: got %q, want %q", resp1.Content, "Hello!")
	}

	// Turn 2
	resp2, _ := agent.Run(ctx, "I'm Alice")
	if resp2.Content != "Nice to meet you!" {
		t.Errorf("turn 2: got %q, want %q", resp2.Content, "Nice to meet you!")
	}

	// Turn 3
	resp3, _ := agent.Run(ctx, "How are you?")
	if resp3.Content != "I'm doing well, thanks!" {
		t.Errorf("turn 3: got %q, want %q", resp3.Content, "I'm doing well, thanks!")
	}

	// Check conversation history
	messages := agent.GetMessages()
	expectedCount := 1 + (3 * 2) // system + 3 turns (user + assistant each)
	if len(messages) != expectedCount {
		t.Errorf("len(messages) = %d, want %d", len(messages), expectedCount)
	}
}

// TestConversationalAgent_Reset tests conversation reset.
func TestConversationalAgent_Reset(t *testing.T) {
	llm := mocks.NewMockLLM()
	llm.WithChatResponse("Response", nil)

	agent := NewConversationalAgent(llm)
	ctx := context.Background()

	// Add some messages
	agent.Run(ctx, "Test 1")
	agent.Run(ctx, "Test 2")

	if len(agent.messages) <= 1 {
		t.Fatal("agent should have multiple messages")
	}

	// Reset
	err := agent.Reset()
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Should only have system message
	if len(agent.messages) != 1 {
		t.Errorf("len(messages) after reset = %d, want 1", len(agent.messages))
	}

	if agent.messages[0].Role != "system" {
		t.Error("should have system message after reset")
	}
}

// TestConversationalAgent_GetMessages tests message retrieval.
func TestConversationalAgent_GetMessages(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewConversationalAgent(llm)

	messages := agent.GetMessages()

	if len(messages) != 1 { // system message
		t.Errorf("len(messages) = %d, want 1", len(messages))
	}
}

// TestConversationalAgent_GetMessageCount tests message count.
func TestConversationalAgent_GetMessageCount(t *testing.T) {
	llm := mocks.NewMockLLM()
	llm.WithChatResponse("Response", nil)

	agent := NewConversationalAgent(llm)
	ctx := context.Background()

	initialCount := agent.GetMessageCount()
	if initialCount != 1 { // system message
		t.Errorf("initial count = %d, want 1", initialCount)
	}

	agent.Run(ctx, "Test")

	afterCount := agent.GetMessageCount()
	if afterCount != 3 { // system + user + assistant
		t.Errorf("after count = %d, want 3", afterCount)
	}
}

// TestConversationalAgent_WindowStrategy tests memory windowing.
func TestConversationalAgent_WindowStrategy(t *testing.T) {
	llm := mocks.NewMockLLM()
	llm.WithChatResponse("Response", nil)

	agent := NewConversationalAgent(
		llm,
		ConvWithMemoryStrategy(MemoryStrategyWindow),
		ConvWithMaxMessages(5), // system + 4 messages (2 turns)
	)

	ctx := context.Background()

	// Add 5 turns (10 messages + 1 system = 11 total)
	for i := 0; i < 5; i++ {
		agent.Run(ctx, "Message")
	}

	// Should trigger windowing
	messageCount := agent.GetMessageCount()
	if messageCount > 6 { // system + 5 kept messages
		t.Errorf("messageCount = %d, want <= 6 (windowing should apply)", messageCount)
	}

	// Should still have system message
	messages := agent.GetMessages()
	if messages[0].Role != "system" {
		t.Error("system message should be preserved")
	}
}

// TestConversationalAgent_AllStrategy tests keeping all messages.
func TestConversationalAgent_AllStrategy(t *testing.T) {
	llm := mocks.NewMockLLM()
	llm.WithChatResponse("Response", nil)

	agent := NewConversationalAgent(
		llm,
		ConvWithMemoryStrategy(MemoryStrategyAll),
		ConvWithMaxMessages(5),
	)

	ctx := context.Background()

	// Add many turns
	for i := 0; i < 10; i++ {
		agent.Run(ctx, "Message")
	}

	// Should keep all messages (no limit with "all" strategy)
	messageCount := agent.GetMessageCount()
	expected := 1 + (10 * 2) // system + 10 turns
	if messageCount != expected {
		t.Errorf("messageCount = %d, want %d (should keep all)", messageCount, expected)
	}
}

// TestConversationalAgent_SetSystemPrompt tests system prompt update.
func TestConversationalAgent_SetSystemPrompt(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewConversationalAgent(llm)

	newPrompt := "New system prompt"
	agent.SetSystemPrompt(newPrompt)

	if agent.GetSystemPrompt() != newPrompt {
		t.Errorf("systemPrompt = %q, want %q", agent.GetSystemPrompt(), newPrompt)
	}

	// Should reset conversation
	if len(agent.messages) != 1 {
		t.Errorf("len(messages) = %d, want 1 (should reset)", len(agent.messages))
	}

	if agent.messages[0].Content != newPrompt {
		t.Error("system message should have new prompt")
	}
}

// TestConversationalAgent_ExportConversation tests conversation export.
func TestConversationalAgent_ExportConversation(t *testing.T) {
	llm := mocks.NewMockLLM()
	llm.WithChatResponse("Hello there!", nil)

	agent := NewConversationalAgent(llm)
	ctx := context.Background()

	agent.Run(ctx, "Hi")

	exported := agent.ExportConversation()

	if exported == "" {
		t.Error("exported conversation should not be empty")
	}

	// Should contain both user and assistant messages
	if !strings.Contains(exported, "user:") || !strings.Contains(exported, "assistant:") {
		t.Error("exported conversation should contain user and assistant messages")
	}
}

// TestConversationalAgent_AddTool tests that tools are not supported.
func TestConversationalAgent_AddTool(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewConversationalAgent(llm)

	tool := mocks.NewMockTool("test", "test tool")
	err := agent.AddTool(tool)

	if err == nil {
		t.Fatal("AddTool() expected error, got nil")
	}

	expectedMsg := "ConversationalAgent does not support tools"
	if err.Error() != expectedMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedMsg)
	}
}

// TestConversationalAgent_SummarizeStrategy tests summarization memory management.
func TestConversationalAgent_SummarizeStrategy(t *testing.T) {
	llm := mocks.NewMockLLM()
	llm.WithChatResponse("Response", nil)
	
	// Configure complete function for summarization
	llm.CompleteFunc = func(ctx context.Context, prompt string) (string, error) {
		// Return a summary when asked
		if strings.Contains(prompt, "Summarize") {
			return "Summary of previous conversation", nil
		}
		return "Response", nil
	}

	agent := NewConversationalAgent(
		llm,
		ConvWithMemoryStrategy(MemoryStrategySummarize),
		ConvWithMaxMessages(6),
	)

	ctx := context.Background()

	// Add many messages to trigger summarization
	for i := 0; i < 10; i++ {
		agent.Run(ctx, "Message")
	}

	// Check if summarization happened
	messages := agent.GetMessages()
	hasSummary := false
	for _, msg := range messages {
		if strings.Contains(msg.Content, "summary") || strings.Contains(msg.Content, "Summary") {
			hasSummary = true
			break
		}
	}

	if !hasSummary {
		t.Error("expected to find summary in messages after exceeding max")
	}
}

// TestConversationalAgent_ApplyWindowStrategy tests window strategy directly.
func TestConversationalAgent_ApplyWindowStrategy(t *testing.T) {
	llm := mocks.NewMockLLM()
	agent := NewConversationalAgent(
		llm,
		ConvWithMaxMessages(5),
	)

	// Manually add many messages
	for i := 0; i < 10; i++ {
		agent.messages = append(agent.messages,
			core.UserMessage("User message"),
			core.AssistantMessage("Assistant message"),
		)
	}

	initialCount := len(agent.messages)
	if initialCount <= 5 {
		t.Fatalf("test setup failed: need more than 5 messages, got %d", initialCount)
	}

	// Apply windowing
	err := agent.applyWindowStrategy()
	if err != nil {
		t.Fatalf("applyWindowStrategy() error = %v", err)
	}

	// Should be reduced
	afterCount := len(agent.messages)
	if afterCount > 6 { // system + 5
		t.Errorf("after windowing: count = %d, want <= 6", afterCount)
	}

	// System message should be preserved
	if agent.messages[0].Role != "system" {
		t.Error("system message should be first after windowing")
	}
}
