package service

import (
	"context"
	"os"

	"github.com/openai/openai-go/v2"
)

type AiService struct {
	client *openai.Client
}

func NewAiService(client *openai.Client) *AiService {
	return &AiService{client: client}
}

func (a *AiService) Generate(ctx context.Context, systemPrompt string, prompt string) (string, error) {
	model := openai.ChatModel(os.Getenv("OPENAI_CHAT_MODEL"))
	if model == "" {
		model = openai.ChatModelGPT5Nano
	}
	chatCompletion, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(prompt),
		},
		Model: model,
	})
	if err != nil {
		return "", err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
