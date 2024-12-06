package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"
)

type queryMsg struct {
	content string
	err     error
}

func (t *thread) sendQueryMsg(ctx context.Context, message string) tea.Cmd {
	return func() tea.Msg {
		content, err := t.query(ctx, message)
		return queryMsg{content, err}
	}
}

type thread struct {
	client         *openai.Client
	messageHistory *[]openai.ChatCompletionMessage
}

func NewThread() *thread {
	return &thread{
		client:         openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		messageHistory: &[]openai.ChatCompletionMessage{},
	}
}

func (t *thread) reset() {
	t.messageHistory = &[]openai.ChatCompletionMessage{}
}

func (t *thread) query(ctx context.Context, message string) (string, error) {
	// Add new user message
	*t.messageHistory = append(*t.messageHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	resp, err := t.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT4,
			Messages: *t.messageHistory,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %v", err)
	}

	content := resp.Choices[0].Message.Content

	*t.messageHistory = append(*t.messageHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	})

	return content, nil
}

func chat(ctx context.Context, messageHistory *[]openai.ChatCompletionMessage, message string) error {
	// Post message to OpenAI
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Add new user message
	*messageHistory = append(*messageHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: *messageHistory,
		},
	)

	if err != nil {
		return fmt.Errorf("ChatCompletion error: %v", err)
	}

	content := resp.Choices[0].Message.Content
	// 🤖 ⏩
	fmt.Println("🤖 " + content)
	*messageHistory = append(*messageHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	})

	return nil
}
