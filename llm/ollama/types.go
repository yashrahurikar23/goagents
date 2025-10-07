package ollama

import "time"

// ChatRequest represents a request to the Ollama chat API
type ChatRequest struct {
	Model    string          `json:"model"`
	Messages []ChatMessage   `json:"messages"`
	Stream   bool            `json:"stream"`           // Must explicitly set to false
	Format   string          `json:"format,omitempty"` // json, etc.
	Options  *RequestOptions `json:"options,omitempty"`
	Tools    []Tool          `json:"tools,omitempty"`
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role      string     `json:"role"` // user, assistant, system, tool
	Content   string     `json:"content"`
	Images    []string   `json:"images,omitempty"` // base64 encoded images
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call in a chat response
type ToolCall struct {
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type"` // function
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function details of a tool call
type ToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Tool represents a tool that can be called by the model
type Tool struct {
	Type     string       `json:"type"` // function
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a tool function definition
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ChatResponse represents a response from the Ollama chat API
type ChatResponse struct {
	Model              string      `json:"model"`
	CreatedAt          time.Time   `json:"created_at"`
	Message            ChatMessage `json:"message"`
	Done               bool        `json:"done"`
	DoneReason         string      `json:"done_reason,omitempty"`
	TotalDuration      int64       `json:"total_duration,omitempty"`
	LoadDuration       int64       `json:"load_duration,omitempty"`
	PromptEvalCount    int         `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64       `json:"prompt_eval_duration,omitempty"`
	EvalCount          int         `json:"eval_count,omitempty"`
	EvalDuration       int64       `json:"eval_duration,omitempty"`
}

// GenerateRequest represents a request to the Ollama generate API
type GenerateRequest struct {
	Model   string          `json:"model"`
	Prompt  string          `json:"prompt"`
	Stream  bool            `json:"stream"` // Must explicitly set to false
	Format  string          `json:"format,omitempty"`
	Options *RequestOptions `json:"options,omitempty"`
	System  string          `json:"system,omitempty"`
	Context []int           `json:"context,omitempty"`
	Images  []string        `json:"images,omitempty"`
}

// GenerateResponse represents a response from the Ollama generate API
type GenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	DoneReason         string    `json:"done_reason,omitempty"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// RequestOptions represents optional parameters for requests
type RequestOptions struct {
	// Model parameters
	Temperature   *float64 `json:"temperature,omitempty"`
	TopP          *float64 `json:"top_p,omitempty"`
	TopK          *int     `json:"top_k,omitempty"`
	RepeatPenalty *float64 `json:"repeat_penalty,omitempty"`
	Seed          *int     `json:"seed,omitempty"`

	// Generation parameters
	NumPredict *int     `json:"num_predict,omitempty"`
	NumCtx     *int     `json:"num_ctx,omitempty"`
	Stop       []string `json:"stop,omitempty"`

	// Advanced parameters
	TfsZ             *float64 `json:"tfs_z,omitempty"`
	TypicalP         *float64 `json:"typical_p,omitempty"`
	RepeatLastN      *int     `json:"repeat_last_n,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	Mirostat         *int     `json:"mirostat,omitempty"`
	MirostatTau      *float64 `json:"mirostat_tau,omitempty"`
	MirostatEta      *float64 `json:"mirostat_eta,omitempty"`
}

// EmbeddingRequest represents a request to the Ollama embeddings API
type EmbeddingRequest struct {
	Model   string          `json:"model"`
	Prompt  string          `json:"prompt"`
	Options *RequestOptions `json:"options,omitempty"`
}

// EmbeddingResponse represents a response from the Ollama embeddings API
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// ListModelsResponse represents the response from listing models
type ListModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// ModelInfo represents information about an Ollama model
type ModelInfo struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
}

// ModelDetails represents detailed information about a model
type ModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// ErrorResponse represents an error response from Ollama
type ErrorResponse struct {
	Error string `json:"error"`
}
