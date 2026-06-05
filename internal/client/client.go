package client

import (
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func NewClient() (openai.Client, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseURL := os.Getenv("OPENROUTER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		return openai.Client{}, fmt.Errorf("no API key found")
	}

	return openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseURL)), nil
}
