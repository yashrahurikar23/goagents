package core

import (
	"errors"
	"testing"
	"time"
)

func TestStreamChunk(t *testing.T) {
	chunk := StreamChunk{
		Content:      "Hello world",
		Delta:        "world",
		Index:        1,
		FinishReason: "",
		Metadata:     map[string]interface{}{"model": "gpt-4"},
		Timestamp:    time.Now(),
	}

	if chunk.Content != "Hello world" {
		t.Errorf("Expected Content 'Hello world', got '%s'", chunk.Content)
	}

	if chunk.Delta != "world" {
		t.Errorf("Expected Delta 'world', got '%s'", chunk.Delta)
	}

	if chunk.Index != 1 {
		t.Errorf("Expected Index 1, got %d", chunk.Index)
	}

	if chunk.FinishReason != "" {
		t.Errorf("Expected empty FinishReason, got '%s'", chunk.FinishReason)
	}

	if model, ok := chunk.Metadata["model"]; !ok || model != "gpt-4" {
		t.Errorf("Expected model metadata 'gpt-4'")
	}
}

func TestStreamChunkWithError(t *testing.T) {
	testErr := errors.New("stream error")
	chunk := StreamChunk{
		Error:     testErr,
		Timestamp: time.Now(),
	}

	if chunk.Error == nil {
		t.Error("Expected Error to be set")
	}

	if chunk.Error.Error() != "stream error" {
		t.Errorf("Expected error message 'stream error', got '%s'", chunk.Error.Error())
	}
}

func TestStreamEvent(t *testing.T) {
	event := StreamEvent{
		Type:      EventTypeToken,
		Content:   "test",
		Data:      map[string]interface{}{"index": 0},
		Error:     nil,
		Timestamp: time.Now(),
	}

	if event.Type != EventTypeToken {
		t.Errorf("Expected Type 'token', got '%s'", event.Type)
	}

	if event.Content != "test" {
		t.Errorf("Expected Content 'test', got '%s'", event.Content)
	}

	if index, ok := event.Data["index"]; !ok || index != 0 {
		t.Errorf("Expected index data 0")
	}
}

func TestStreamEventWithError(t *testing.T) {
	testErr := errors.New("execution error")
	event := StreamEvent{
		Type:      EventTypeError,
		Content:   "execution error",
		Error:     testErr,
		Timestamp: time.Now(),
	}

	if event.Type != EventTypeError {
		t.Errorf("Expected Type 'error', got '%s'", event.Type)
	}

	if event.Error == nil {
		t.Error("Expected Error to be set")
	}
}

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		constant string
		expected string
	}{
		{EventTypeToken, "token"},
		{EventTypeThought, "thought"},
		{EventTypeToolStart, "tool_start"},
		{EventTypeToolEnd, "tool_end"},
		{EventTypeAnswer, "answer"},
		{EventTypeComplete, "complete"},
		{EventTypeError, "error"},
	}

	for _, tt := range tests {
		if tt.constant != tt.expected {
			t.Errorf("Expected constant '%s', got '%s'", tt.expected, tt.constant)
		}
	}
}

func TestNewStreamChunk(t *testing.T) {
	chunk := NewStreamChunk("hello", 0)

	if chunk.Delta != "hello" {
		t.Errorf("Expected Delta 'hello', got '%s'", chunk.Delta)
	}

	if chunk.Index != 0 {
		t.Errorf("Expected Index 0, got %d", chunk.Index)
	}

	if chunk.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}

	if chunk.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

func TestNewStreamEvent(t *testing.T) {
	event := NewStreamEvent(EventTypeToken, "test content")

	if event.Type != EventTypeToken {
		t.Errorf("Expected Type 'token', got '%s'", event.Type)
	}

	if event.Content != "test content" {
		t.Errorf("Expected Content 'test content', got '%s'", event.Content)
	}

	if event.Data == nil {
		t.Error("Expected Data to be initialized")
	}

	if event.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

func TestNewStreamEventWithData(t *testing.T) {
	data := map[string]interface{}{
		"tool":  "calculator",
		"input": "2+2",
	}

	event := NewStreamEventWithData(EventTypeToolStart, "calculator", data)

	if event.Type != EventTypeToolStart {
		t.Errorf("Expected Type 'tool_start', got '%s'", event.Type)
	}

	if event.Content != "calculator" {
		t.Errorf("Expected Content 'calculator', got '%s'", event.Content)
	}

	if event.Data["tool"] != "calculator" {
		t.Error("Expected tool data to be set")
	}

	if event.Data["input"] != "2+2" {
		t.Error("Expected input data to be set")
	}
}

func TestNewErrorEvent(t *testing.T) {
	testErr := errors.New("test error")
	event := NewErrorEvent(testErr)

	if event.Type != EventTypeError {
		t.Errorf("Expected Type 'error', got '%s'", event.Type)
	}

	if event.Content != "test error" {
		t.Errorf("Expected Content 'test error', got '%s'", event.Content)
	}

	if event.Error == nil {
		t.Error("Expected Error to be set")
	}

	if event.Error.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", event.Error.Error())
	}

	if event.Data == nil {
		t.Error("Expected Data to be initialized")
	}
}

func TestStreamChunkFinishReasons(t *testing.T) {
	finishReasons := []string{"", "stop", "length", "tool_calls", "content_filter"}

	for _, reason := range finishReasons {
		chunk := StreamChunk{
			FinishReason: reason,
		}

		if chunk.FinishReason != reason {
			t.Errorf("Expected FinishReason '%s', got '%s'", reason, chunk.FinishReason)
		}
	}
}

func TestStreamEventTypes(t *testing.T) {
	eventTypes := []string{
		EventTypeToken,
		EventTypeThought,
		EventTypeToolStart,
		EventTypeToolEnd,
		EventTypeAnswer,
		EventTypeComplete,
		EventTypeError,
	}

	for _, eventType := range eventTypes {
		event := NewStreamEvent(eventType, "test")

		if event.Type != eventType {
			t.Errorf("Expected Type '%s', got '%s'", eventType, event.Type)
		}
	}
}

func TestStreamChunkMetadata(t *testing.T) {
	chunk := NewStreamChunk("test", 0)

	// Add metadata
	chunk.Metadata["model"] = "gpt-4"
	chunk.Metadata["tokens"] = 100
	chunk.Metadata["temperature"] = 0.7

	if chunk.Metadata["model"] != "gpt-4" {
		t.Error("Failed to set model metadata")
	}

	if chunk.Metadata["tokens"] != 100 {
		t.Error("Failed to set tokens metadata")
	}

	if chunk.Metadata["temperature"] != 0.7 {
		t.Error("Failed to set temperature metadata")
	}
}

func TestStreamEventData(t *testing.T) {
	event := NewStreamEvent(EventTypeToolStart, "calculator")

	// Add data
	event.Data["input"] = map[string]interface{}{
		"operation": "add",
		"a":         2,
		"b":         2,
	}

	input, ok := event.Data["input"].(map[string]interface{})
	if !ok {
		t.Fatal("Failed to get input data")
	}

	if input["operation"] != "add" {
		t.Error("Failed to set operation in input data")
	}

	if input["a"] != 2 || input["b"] != 2 {
		t.Error("Failed to set operands in input data")
	}
}

func TestTimestampPresence(t *testing.T) {
	// Test chunk timestamp
	chunk := NewStreamChunk("test", 0)
	if chunk.Timestamp.IsZero() {
		t.Error("StreamChunk Timestamp should not be zero")
	}

	// Test event timestamp
	event := NewStreamEvent(EventTypeToken, "test")
	if event.Timestamp.IsZero() {
		t.Error("StreamEvent Timestamp should not be zero")
	}

	// Test error event timestamp
	errorEvent := NewErrorEvent(errors.New("test"))
	if errorEvent.Timestamp.IsZero() {
		t.Error("Error event Timestamp should not be zero")
	}
}

func TestStreamChunkContentAccumulation(t *testing.T) {
	// Simulate accumulating content from deltas
	var content string
	chunks := []StreamChunk{
		{Delta: "Hello", Index: 0},
		{Delta: " ", Index: 1},
		{Delta: "world", Index: 2},
		{Delta: "!", Index: 3},
	}

	for _, chunk := range chunks {
		content += chunk.Delta
	}

	if content != "Hello world!" {
		t.Errorf("Expected accumulated content 'Hello world!', got '%s'", content)
	}
}
