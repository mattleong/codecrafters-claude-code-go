package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/openai/openai-go/v3"
)

func AgentLoop(ctx context.Context, client openai.Client, prompt []openai.ChatCompletionMessageParamUnion) (string, error) {
	resp, err := client.Chat.Completions.New(ctx,
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

	if err != nil {
		return "", fmt.Errorf("error getting response back from client: %w", err)
	}

	messages := []openai.ChatCompletionMessageParamUnion{}
	messages = append(messages, prompt...)

	if len(resp.Choices[0].Message.ToolCalls) == 0 {
		return resp.Choices[0].Message.Content, nil
	}

	toolCalls := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(resp.Choices[0].Message.ToolCalls))
	for _, tc := range resp.Choices[0].Message.ToolCalls {
		toolCalls = append(toolCalls, tc.ToParam())
	}

	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			Role:      "assistant",
			ToolCalls: toolCalls,
		},
	})

	for i := range resp.Choices[0].Message.ToolCalls {
		var toolCall = resp.Choices[0].Message.ToolCalls[i]
		argsJSON := toolCall.Function.Arguments
		var params map[string]string
		err := json.Unmarshal([]byte(argsJSON), &params)
		if err != nil {
			return "", fmt.Errorf("could not parse response params %w", err)
		}

		switch toolCall.Function.Name {
		case "Read":
			messages, err = ReadTool(toolCall, params, messages)
		case "Write":
			messages, err = WriteTool(toolCall, params, messages)
		case "Bash":
			messages, err = BashTool(toolCall, params, messages)
		default:
			return "", fmt.Errorf("unsupported tool call")
		}

		if err != nil {
			return "", err
		}
	}

	return AgentLoop(ctx, client, messages)
}
