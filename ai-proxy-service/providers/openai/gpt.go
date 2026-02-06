package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/blcvn/backend/services/ai-proxy-service/entities"
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
	// Initial placeholder initialization
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
func (g *GPTProvider) Complete(ctx context.Context, req *entities.CompletionRequest) (*entities.CompletionResponse, error) {
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
	clientOpts := []openai.Option{
		openai.WithToken(req.APIKey),
		openai.WithModel(req.ModelID),
	}
	if req.BaseURL != "" {
		clientOpts = append(clientOpts, openai.WithBaseURL(req.BaseURL))
	}

	ll, err := openai.New(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI LLM: %w", err)
	}

	// Build options
	callOpts := []llms.CallOption{
		llms.WithTemperature(float64(req.Temperature)),
		llms.WithMaxTokens(int(req.MaxTokens)),
		// llms.WithTopP(req.TopP), // Removed TopP as it was missing in struct or not critical
	}

	if len(req.StopSequences) > 0 {
		callOpts = append(callOpts, llms.WithStopWords(req.StopSequences))
	}

	// Call LLM
	response, err := ll.GenerateContent(ctx, messages, callOpts...)
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

// GenerateContent uses LangChainGo's standardized interface directly
func (g *GPTProvider) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return g.llm.GenerateContent(ctx, messages, options...)
}

// StreamComplete implements streaming completion
func (g *GPTProvider) StreamComplete(ctx context.Context, req *entities.CompletionRequest, callback func(*entities.StreamResponse) error) error {
	// Dynamic client creation using injected credentials
	clientOpts := []openai.Option{
		openai.WithToken(req.APIKey),
		openai.WithModel(req.ModelID),
	}
	if req.BaseURL != "" {
		clientOpts = append(clientOpts, openai.WithBaseURL(req.BaseURL))
	}

	ll, err := openai.New(clientOpts...)
	if err != nil {
		return fmt.Errorf("failed to create OpenAI LLM: %w", err)
	}

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

	_, err = ll.GenerateContent(ctx, messages, callOpts...)
	return err
}

// HealthCheck verifies the OpenAI API is accessible
func (g *GPTProvider) HealthCheck(ctx context.Context) error {
	_, err := g.Complete(ctx, &entities.CompletionRequest{
		Messages:  []entities.Message{{Role: entities.RoleUser, Content: "Hi"}},
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
