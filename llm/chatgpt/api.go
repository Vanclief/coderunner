package chatgpt

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/vanclief/ez"
)

type API struct {
	Model  string
	client *openai.Client
}

func NewAPI(apiKey, model string) (*API, error) {
	const op = "chatgpt.NewAPI"

	switch model {
	case "o1":
		model = openai.O1Preview
	case "o1-mini":
		model = openai.O1Mini
	case "4o":
		model = openai.GPT4o

	default:
		return nil, ez.New(op, ez.EINVALID, "Invalid model", nil)

	}

	client := openai.NewClient(apiKey)

	api := &API{
		Model:  model,
		client: client,
	}

	return api, nil
}

func (a *API) Prompt(prompt string) (string, error) {
	const op = "chatgpt.Prompt"

	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: a.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", ez.Wrap(op, err)
	} else if len(resp.Choices) == 0 {
		return "", ez.New(op, ez.ENOTFOUND, "No choices in response", nil)
	}

	return resp.Choices[0].Message.Content, nil
}
