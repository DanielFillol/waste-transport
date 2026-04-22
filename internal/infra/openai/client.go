package openai

import (
	"context"
	"errors"

	"github.com/danielfillol/waste/internal/config"
	goopenai "github.com/sashabaranov/go-openai"
)

type Client struct {
	client *goopenai.Client
	model  string
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		client: goopenai.NewClient(cfg.OpenAIAPIKey),
		model:  cfg.OpenAIModel,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *Client) Chat(ctx context.Context, messages []Message, systemPrompt string) (string, error) {
	msgs := []goopenai.ChatCompletionMessage{
		{Role: goopenai.ChatMessageRoleSystem, Content: systemPrompt},
	}
	for _, m := range messages {
		msgs = append(msgs, goopenai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	resp, err := c.client.CreateChatCompletion(ctx, goopenai.ChatCompletionRequest{
		Model:    c.model,
		Messages: msgs,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty response from OpenAI")
	}
	return resp.Choices[0].Message.Content, nil
}
