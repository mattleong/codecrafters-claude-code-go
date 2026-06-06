package client

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var ctx context.Context
var client openai.Client

func InitClient() error {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseURL := os.Getenv("OPENROUTER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		return fmt.Errorf("no API key found")
	}

	ctx = context.Background()
	client = openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseURL))

	return nil
}

func MakeRequest(prompt []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	return client.Chat.Completions.New(ctx,
		openai.ChatCompletionNewParams{
			Model:    "anthropic/claude-haiku-4.5",
			Messages: prompt,
			Tools: []openai.ChatCompletionToolUnionParam{
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name:        "Read",
					Description: openai.String("Read and return the contents of a file"),
					Parameters: openai.FunctionParameters{
						"type": "object",
						"properties": map[string]any{
							"file_path": map[string]any{
								"type":        "string",
								"description": "The path to the file to read",
							},
						},
						"required": []string{"file_path"},
					},
				}),
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name:        "Write",
					Description: openai.String("Write content to a file"),
					Parameters: openai.FunctionParameters{
						"type": "object",
						"properties": map[string]any{
							"file_path": map[string]any{
								"type":        "string",
								"description": "The path of the file to write to",
							},
							"content": map[string]any{
								"type":        "string",
								"description": "The content to write to the file",
							},
						},
					},
				}),
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name:        "Bash",
					Description: openai.String("Execute a shell command"),
					Parameters: openai.FunctionParameters{
						"type": "object",
						"required": []string{
							"command",
						},
						"properties": map[string]any{
							"command": map[string]any{
								"type":        "string",
								"description": "The command to execute",
							},
						},
					},
				}),
			},
		},
	)
}
