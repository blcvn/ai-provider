package anthropic

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"

	"github.com/blcvn/backend/services/ai-proxy-service/providers"
)

// ClaudeProvider implements the LLMProvider interface for Anthropic Claude using LangChainGo
type ClaudeProvider struct {
	llm     llms.Model
	apiKey  string
	modelID string
}

// NewClaudeProvider creates a new Anthropic Claude provider using LangChainGo
func NewClaudeProvider(apiKey, modelID string) (*ClaudeProvider, error) {
	llm, err := anthropic.New(
		anthropic.WithToken(apiKey),
		anthropic.WithModel(modelID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic LLM: %w", err)
	}

	return &ClaudeProvider{
		llm:     llm,
		apiKey:  apiKey,
		modelID: modelID,
	}, nil
}

// Complete sends a completion request to Claude API using LangChainGo
func (c *ClaudeProvider) Complete(ctx context.Context, req *providers.CompletionRequest) (*providers.CompletionResponse, error) {
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
	response, err := c.llm.GenerateContent(ctx, messages, options...)
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

	// Calculate cost (Claude 3.5 Sonnet pricing)
	// $3 per 1M input tokens, $15 per 1M output tokens
	cost := (float64(promptTokens) / 1_000_000.0 * 3.0) + (float64(completionTokens) / 1_000_000.0 * 15.0)

	return &providers.CompletionResponse{
		Content:          text,
		TokensUsed:       promptTokens + completionTokens,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		FinishReason:     stopReason,
		Cost:             cost,
		ModelUsed:        c.modelID,
	}, nil
}

// GenerateContent uses LangChainGo's standardized interface directly
func (c *ClaudeProvider) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return c.llm.GenerateContent(ctx, messages, options...)
}

// HealthCheck verifies the Claude API is accessible
func (c *ClaudeProvider) HealthCheck(ctx context.Context) error {
	_, err := c.Complete(ctx, &providers.CompletionRequest{
		Prompt:    "Hi",
		MaxTokens: 10,
	})
	return err
}

// GetProviderInfo returns metadata about the Claude provider
func (c *ClaudeProvider) GetProviderInfo() providers.ProviderInfo {
	return providers.ProviderInfo{
		Name:    "Anthropic Claude",
		Type:    "anthropic",
		BaseURL: "https://api.anthropic.com",
		Models:  []string{"claude-3-5-sonnet-20241022", "claude-3-opus-20240229", "claude-3-haiku-20240307"},
	}
}

// estimateTokens provides a rough token count estimate
// More accurate counting would require tiktoken or similar
func estimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters
	// This is a simplification; actual tokenization is more complex
	return len(strings.TrimSpace(text)) / 4
}
