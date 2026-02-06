package entities

import (
	"context"
)

type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

type Message struct {
	Role    MessageRole
	Content string
}

type CompletionRequest struct {
	ModelID       string
	Messages      []Message
	Temperature   float32
	MaxTokens     int32
	StopSequences []string
	// Credentials injected by usecase
	APIKey  string
	BaseURL string
}

type CompletionResponse struct {
	Content      string
	Usage        Usage
	FinishReason string
}

type Usage struct {
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
}

type StreamResponse struct {
	Content      string
	Usage        *Usage
	FinishReason string
}

type LLMProvider interface {
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	StreamComplete(ctx context.Context, req *CompletionRequest, callback func(*StreamResponse) error) error
}
