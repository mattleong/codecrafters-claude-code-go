package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type Agent struct {
	client openai.Client
	ctx    context.Context
	model  shared.ChatModel
}

func NewAgent(model shared.ChatModel) (*Agent, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseURL := os.Getenv("OPENROUTER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		return &Agent{}, fmt.Errorf("no API key found")
	}

	return &Agent{
		client: openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseURL)),
		ctx:    context.Background(),
		model:  model,
	}, nil
}
