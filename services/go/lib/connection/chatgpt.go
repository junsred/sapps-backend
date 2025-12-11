package connection

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type ChatGPT struct {
	client openai.Client
}

func NewChatGPT() (*ChatGPT, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	log.Println("CHATGPT CONNECTED")
	return &ChatGPT{
		client: client,
	}, nil
}

func (c *ChatGPT) GenerateCompletion(ctx context.Context, model shared.ChatModel, systemPrompt string, prompt string, responseFormat openai.ChatCompletionNewParamsResponseFormatUnion, temperature float64) (string, openai.CompletionUsage, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
				URL: "",
			})}),
			openai.UserMessage(prompt),
		},
		Model:               model,
		MaxCompletionTokens: openai.Int(5000),
		ResponseFormat:      responseFormat,
		Temperature:         openai.Float(temperature),
	})
	if err != nil {
		return "", openai.CompletionUsage{}, err
	}

	if len(resp.Choices) == 0 {
		return "", openai.CompletionUsage{}, fmt.Errorf("empty response")
	}

	return resp.Choices[0].Message.Content, resp.Usage, nil
}

func (c *ChatGPT) GenerateCompletionWithHistory(ctx context.Context, model shared.ChatModel, params []openai.ChatCompletionMessageParamUnion, responseFormat openai.ChatCompletionNewParamsResponseFormatUnion, temperature float64) (string, openai.CompletionUsage, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:            params,
		Model:               model,
		MaxCompletionTokens: openai.Int(5000),
		ResponseFormat:      responseFormat,
		Temperature:         openai.Float(temperature),
	})
	if err != nil {
		return "", openai.CompletionUsage{}, err
	}

	if len(resp.Choices) == 0 {
		return "", openai.CompletionUsage{}, fmt.Errorf("empty response")
	}

	return resp.Choices[0].Message.Content, resp.Usage, nil
}

func (c *ChatGPT) GenerateCompletionWithImage(ctx context.Context, model shared.ChatModel, systemPrompt string, url string, responseFormat openai.ChatCompletionNewParamsResponseFormatUnion, temperature float64) (string, openai.CompletionUsage, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
				URL: url,
			})}),
		},
		Model:               model,
		MaxCompletionTokens: openai.Int(5000),
		ResponseFormat:      responseFormat,
		Temperature:         openai.Float(temperature),
	})
	if err != nil {
		return "", openai.CompletionUsage{}, err
	}

	if len(resp.Choices) == 0 {
		return "", openai.CompletionUsage{}, fmt.Errorf("empty response")
	}

	return resp.Choices[0].Message.Content, resp.Usage, nil
}
