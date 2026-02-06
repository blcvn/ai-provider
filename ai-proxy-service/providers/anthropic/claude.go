package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/blcvn/backend/services/ai-proxy-service/entities"
	"github.com/blcvn/backend/services/ai-proxy-service/providers"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

// ClaudeProvider implements the LLMProvider interface for Anthropic Claude using LangChainGo
type ClaudeProvider struct {
}

// NewClaudeProvider creates a new Anthropic Claude provider using LangChainGo
func NewClaudeProvider() (*ClaudeProvider, error) {
	return &ClaudeProvider{}, nil
}

// Complete sends a completion request to Claude API using LangChainGo
func (c *ClaudeProvider) Complete(ctx context.Context, req *entities.CompletionRequest) (*entities.CompletionResponse, error) {
	// Debug Log
	reqJSON, _ := json.Marshal(req)
	log.Printf("Anthropic Complete Request: %s", string(reqJSON))

	// Build messages
	// entities.Message -> llms.MessageContent
	messages := make([]llms.MessageContent, len(req.Messages))
	for i, m := range req.Messages {
		role := llms.ChatMessageTypeGeneric
		switch m.Role {
		case entities.RoleSystem:
			role = llms.ChatMessageTypeSystem
		case entities.RoleUser:
			role = llms.ChatMessageTypeHuman
		case entities.RoleAssistant:
			role = llms.ChatMessageTypeAI
		}
		messages[i] = llms.TextParts(role, m.Content)
	}

	// Dynamic client creation using injected credentials
	opts := []anthropic.Option{
		anthropic.WithToken(req.APIKey),
		anthropic.WithModel(req.ModelID),
	}
	if req.BaseURL != "" {
		opts = append(opts, anthropic.WithBaseURL(req.BaseURL))
	}

	ll, err := anthropic.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic LLM: %w", err)
	}

	// Build options
	callOpts := []llms.CallOption{
		llms.WithTemperature(float64(req.Temperature)),
		llms.WithMaxTokens(16384),
	}

	if len(req.StopSequences) > 0 {
		callOpts = append(callOpts, llms.WithStopWords(req.StopSequences))
	}

	// Call LLM
	response, err := ll.GenerateContent(ctx, messages, callOpts...)
	if err != nil {
		log.Printf("Anthropic GenerateContent Error: %v", err)
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
	var promptText string
	for _, m := range req.Messages {
		promptText += m.Content + " "
	}
	promptTokens := int32(estimateTokens(promptText))
	completionTokens := int32(estimateTokens(text))

	return &entities.CompletionResponse{
		Content:      text,
		FinishReason: stopReason,
		Usage: entities.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}, nil
}

// StreamComplete implements streaming completion
func (c *ClaudeProvider) StreamComplete(ctx context.Context, req *entities.CompletionRequest, callback func(*entities.StreamResponse) error) error {
	// Build messages
	messages := make([]llms.MessageContent, len(req.Messages))
	for i, m := range req.Messages {
		role := llms.ChatMessageTypeGeneric
		switch m.Role {
		case entities.RoleSystem:
			role = llms.ChatMessageTypeSystem
		case entities.RoleUser:
			role = llms.ChatMessageTypeHuman
		case entities.RoleAssistant:
			role = llms.ChatMessageTypeAI
		}
		messages[i] = llms.TextParts(role, m.Content)
	}

	// Dynamic client creation using injected credentials
	opts := []anthropic.Option{
		anthropic.WithToken(req.APIKey),
		anthropic.WithModel(req.ModelID),
	}
	if req.BaseURL != "" {
		opts = append(opts, anthropic.WithBaseURL(req.BaseURL))
	}

	ll, err := anthropic.New(opts...)
	if err != nil {
		return fmt.Errorf("failed to create Anthropic LLM: %w", err)
	}

	// Build options
	callOpts := []llms.CallOption{
		llms.WithTemperature(float64(req.Temperature)),
		llms.WithMaxTokens(int(req.MaxTokens)),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			return callback(&entities.StreamResponse{Content: string(chunk)})
		}),
	}

	if len(req.StopSequences) > 0 {
		callOpts = append(callOpts, llms.WithStopWords(req.StopSequences))
	}

	// Call LLM
	_, err = ll.GenerateContent(ctx, messages, callOpts...)
	return err
}

// // GenerateContent uses LangChainGo's standardized interface directly
// func (c *ClaudeProvider) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
// 	return c.llm.GenerateContent(ctx, messages, options...)
// }

// HealthCheck verifies the Claude API is accessible
func (c *ClaudeProvider) HealthCheck(ctx context.Context) error {
	_, err := c.Complete(ctx, &entities.CompletionRequest{
		Messages: []entities.Message{
			{Role: entities.RoleUser, Content: "Hi"},
		},
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
