package service

import (
	"context"

	"github.com/openai/openai-go"
)

type AiService struct {
	client *openai.Client
}

func NewAiService(client *openai.Client) *AiService {
	return &AiService{client: client}
}

func (a *AiService) Generate(ctx context.Context, systemPrompt string, prompt string) (string, error) {
	chatCompletion, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4_1Mini,
	})
	if err != nil {
		return "", err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
