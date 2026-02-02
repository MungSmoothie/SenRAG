package rag

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type LLM struct {
	client *openai.Client
	cfg    interface{}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewLLM(baseURL, apiKey, model string) *LLM {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	return &LLM{
		client: openai.NewClientWithConfig(config),
		cfg: map[string]interface{}{
			"model": model,
		},
	}
}

func (l *LLM) Chat(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, error) {
	msgs := make([]openai.ChatCompletionMessage, 0, len(messages)+1)

	if systemPrompt != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}

	for _, m := range messages {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	resp, err := l.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    l.cfg.(map[string]interface{})["model"].(string),
		Messages: msgs,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (l *LLM) StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string) (<-chan string, error) {
	msgs := make([]openai.ChatCompletionMessage, 0, len(messages)+1)

	if systemPrompt != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}

	for _, m := range messages {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	stream, err := l.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    l.cfg.(map[string]interface{})["model"].(string),
		Messages: msgs,
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan string)
	go func() {
		defer close(ch)
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if err != nil {
				break
			}
			if len(resp.Choices) > 0 {
				ch <- resp.Choices[0].Delta.Content
			}
		}
	}()
	return ch, nil
}
