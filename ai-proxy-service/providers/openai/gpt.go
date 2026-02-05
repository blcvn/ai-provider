package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/blcvn/backend/services/ai-proxy-service/providers"
)

// GPTProvider implements the LLMProvider interface for OpenAI GPT using LangChainGo
type GPTProvider struct {
	llm     llms.Model
	apiKey  string
	modelID string
}

// NewGPTProvider creates a new OpenAI GPT provider using LangChainGo
func NewGPTProvider(apiKey, modelID string) (*GPTProvider, error) {
	llm, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithModel(modelID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI LLM: %w", err)
	}

	return &GPTProvider{
		llm:     llm,
		apiKey:  apiKey,
		modelID: modelID,
	}, nil
}

// Complete sends a completion request to OpenAI API using LangChainGo
func (g *GPTProvider) Complete(ctx context.Context, req *providers.CompletionRequest) (*providers.CompletionResponse, error) {
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
	response, err := g.llm.GenerateContent(ctx, messages, options...)
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

	// Estimate token counts (rough approximation: 1 token ≈ 4 characters)
	promptTokens := estimateTokens(req.Prompt + req.SystemPrompt)
	completionTokens := estimateTokens(text)

	// Calculate cost (GPT-4 pricing)
	// $10 per 1M input tokens, $30 per 1M output tokens
	cost := (float64(promptTokens) / 1_000_000.0 * 10.0) + (float64(completionTokens) / 1_000_000.0 * 30.0)

	return &providers.CompletionResponse{
		Content:          text,
		TokensUsed:       promptTokens + completionTokens,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		FinishReason:     stopReason,
		Cost:             cost,
		ModelUsed:        g.modelID,
	}, nil
}

// GenerateContent uses LangChainGo's standardized interface directly
func (g *GPTProvider) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return g.llm.GenerateContent(ctx, messages, options...)
}

// HealthCheck verifies the OpenAI API is accessible
func (g *GPTProvider) HealthCheck(ctx context.Context) error {
	_, err := g.Complete(ctx, &providers.CompletionRequest{
		Prompt:    "Hi",
		MaxTokens: 10,
	})
	return err
}

// GetProviderInfo returns metadata about the GPT provider
func (g *GPTProvider) GetProviderInfo() providers.ProviderInfo {
	return providers.ProviderInfo{
		Name:    "OpenAI GPT",
		Type:    "openai",
		BaseURL: "https://api.openai.com",
		Models:  []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo", "gpt-4o"},
	}
}

// estimateTokens provides a rough token count estimate
// More accurate counting would require tiktoken or similar
func estimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters
	return len(strings.TrimSpace(text)) / 4
}
