package providers

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// CompletionRequest represents a request to an LLM provider
type CompletionRequest struct {
	Prompt        string
	Temperature   float64
	MaxTokens     int
	TopP          float64
	StopSequences []string
	SystemPrompt  string
}

// CompletionResponse represents a response from an LLM provider
type CompletionResponse struct {
	Content          string
	TokensUsed       int
	PromptTokens     int
	CompletionTokens int
	FinishReason     string
	Cost             float64
	ModelUsed        string
}

// ProviderInfo contains metadata about a provider
type ProviderInfo struct {
	Name    string
	Type    string
	BaseURL string
	Models  []string
}

// LLMProvider defines the interface that all LLM providers must implement
// This wraps LangChainGo's LLM interface with our custom logic
type LLMProvider interface {
	// Complete sends a completion request and returns the response
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// GenerateContent uses LangChainGo's standardized interface
	GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error)

	// HealthCheck verifies the provider is accessible
	HealthCheck(ctx context.Context) error

	// GetProviderInfo returns metadata about the provider
	GetProviderInfo() ProviderInfo
}
