package local

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"

	"github.com/blcvn/backend/services/ai-proxy-service/providers"
)

// OllamaProvider implements the LLMProvider interface for local LLMs via Ollama
type OllamaProvider struct {
	llm     llms.Model
	baseURL string
	modelID string
}

// NewOllamaProvider creates a new Ollama provider for local LLMs using LangChainGo
func NewOllamaProvider(baseURL, modelID string) (*OllamaProvider, error) {
	llm, err := ollama.New(
		ollama.WithServerURL(baseURL),
		ollama.WithModel(modelID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama LLM: %w", err)
	}

	return &OllamaProvider{
		llm:     llm,
		baseURL: baseURL,
		modelID: modelID,
	}, nil
}

// Complete sends a completion request to Ollama using LangChainGo
func (o *OllamaProvider) Complete(ctx context.Context, req *providers.CompletionRequest) (*providers.CompletionResponse, error) {
	// Build messages
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: req.Prompt},
			},
		},
	}

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		messages = append([]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: req.SystemPrompt},
				},
			},
		}, messages...)
	}

	// Build options
	options := []llms.CallOption{
		llms.WithTemperature(req.Temperature),
		llms.WithMaxTokens(req.MaxTokens),
		llms.WithTopP(req.TopP),
	}

	if len(req.StopSequences) > 0 {
		options = append(options, llms.WithStopWords(req.StopSequences))
	}

	// Call LLM
	response, err := o.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract text from response
	var text string
	var stopReason string
	if len(response.Choices) > 0 {
		text = response.Choices[0].Content
		stopReason = response.Choices[0].StopReason
	}

	// Estimate token counts
	promptTokens := estimateTokens(req.Prompt + req.SystemPrompt)
	completionTokens := estimateTokens(text)

	return &providers.CompletionResponse{
		Content:          text,
		TokensUsed:       promptTokens + completionTokens,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		FinishReason:     stopReason,
		Cost:             0.0, // Local LLMs are free
		ModelUsed:        o.modelID,
	}, nil
}

// GenerateContent uses LangChainGo's standardized interface directly
func (o *OllamaProvider) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return o.llm.GenerateContent(ctx, messages, options...)
}

// HealthCheck verifies Ollama is accessible
func (o *OllamaProvider) HealthCheck(ctx context.Context) error {
	_, err := o.Complete(ctx, &providers.CompletionRequest{
		Prompt:    "Hi",
		MaxTokens: 10,
	})
	return err
}

// GetProviderInfo returns metadata about the Ollama provider
func (o *OllamaProvider) GetProviderInfo() providers.ProviderInfo {
	return providers.ProviderInfo{
		Name:    "Ollama (Local LLM)",
		Type:    "ollama",
		BaseURL: o.baseURL,
		Models:  []string{"llama2", "mistral", "codellama", "phi"},
	}
}

// estimateTokens provides a rough token count estimate
func estimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters
	return len(strings.TrimSpace(text)) / 4
}
