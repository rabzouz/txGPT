package ai

import (
	"context"
	"fmt"
	"io"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// Config customizes OpenAI chat completion requests.
type Config struct {
	Model        string
	Temperature  float32
	MaxTokens    int
	BaseURL      string
	SystemPrompt string
	Stream       bool
}

// DefaultConfig returns conservative defaults for the CLI.
func DefaultConfig() Config {
	return Config{
		Model:        openai.GPT4oMini,
		Temperature:  0.7,
		MaxTokens:    0,
		SystemPrompt: "You are txGPT, a concise technical assistant for code, Linux and authorized security testing. Refuse requests that facilitate unauthorized harm.",
		Stream:       false,
	}
}

// Ask sends a prompt with optional chat history.
func Ask(prompt string, history []openai.ChatCompletionMessage, cfg Config) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	clientCfg := openai.DefaultConfig(apiKey)
	if cfg.BaseURL != "" {
		clientCfg.BaseURL = cfg.BaseURL
	}
	client := openai.NewClientWithConfig(clientCfg)

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: cfg.SystemPrompt},
	}
	messages = append(messages, history...)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	req := openai.ChatCompletionRequest{
		Model:       cfg.Model,
		Messages:    messages,
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
	}

	if cfg.Stream {
		return streamAnswer(client, req)
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return resp.Choices[0].Message.Content, nil
}

func streamAnswer(client *openai.Client, req openai.ChatCompletionRequest) (string, error) {
	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var full string
	for {
		part, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if len(part.Choices) == 0 {
			continue
		}
		txt := part.Choices[0].Delta.Content
		fmt.Print(txt)
		full += txt
	}
	fmt.Println()
	return full, nil
}
