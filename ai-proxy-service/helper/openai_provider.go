package helper

import (
	"context"

	"github.com/blcvn/backend/services/ai-proxy-service/entities"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAIProvider struct {
	apiKey  string
	baseURL string
}

func NewOpenAIProvider(apiKey, baseURL string) *OpenAIProvider {
	return &OpenAIProvider{apiKey: apiKey, baseURL: baseURL}
}

func (p *OpenAIProvider) Complete(ctx context.Context, req *entities.CompletionRequest) (*entities.CompletionResponse, error) {
	opts := []openai.Option{openai.WithToken(p.apiKey)}
	if p.baseURL != "" {
		opts = append(opts, openai.WithBaseURL(p.baseURL))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	resp, err := llm.GenerateContent(ctx, u2l_messages(req.Messages),
		llms.WithModel(req.ModelID),
		llms.WithTemperature(float64(req.Temperature)),
		llms.WithMaxTokens(int(req.MaxTokens)),
	)
	if err != nil {
		return nil, err
	}

	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Content
	}

	// LangChainGo returns usage in GenerateContent response
	return &entities.CompletionResponse{
		Content: content,
		Usage: entities.Usage{
			PromptTokens:     0, // TODO: Extract from resp.Usage if available? only available in some providers
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}, nil
}

func (p *OpenAIProvider) StreamComplete(ctx context.Context, req *entities.CompletionRequest, callback func(*entities.StreamResponse) error) error {
	opts := []openai.Option{openai.WithToken(p.apiKey)}
	if p.baseURL != "" {
		opts = append(opts, openai.WithBaseURL(p.baseURL))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return err
	}

	_, err = llm.GenerateContent(ctx, u2l_messages(req.Messages),
		llms.WithModel(req.ModelID),
		llms.WithTemperature(float64(req.Temperature)),
		llms.WithMaxTokens(int(req.MaxTokens)),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			return callback(&entities.StreamResponse{
				Content: string(chunk),
			})
		}),
	)
	return err
}

func u2l_messages(msgs []entities.Message) []llms.MessageContent {
	res := make([]llms.MessageContent, len(msgs))
	for i, m := range msgs {
		role := llms.ChatMessageTypeGeneric
		switch m.Role {
		case entities.RoleSystem:
			role = llms.ChatMessageTypeSystem
		case entities.RoleUser:
			role = llms.ChatMessageTypeHuman
		case entities.RoleAssistant:
			role = llms.ChatMessageTypeAI
		}
		res[i] = llms.TextParts(role, m.Content)
	}
	return res
}
