package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/openai/openai-go/v3"
	"log"
	"os"
)

func AgentLoop(client openai.Client, prompt []openai.ChatCompletionMessageParamUnion) string {
	resp, err := client.Chat.Completions.New(context.Background(),
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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	messages := []openai.ChatCompletionMessageParamUnion{}
	messages = append(messages, prompt...)

	if len(resp.Choices[0].Message.ToolCalls) == 0 {
		return resp.Choices[0].Message.Content
	}

	var toolCalls []openai.ChatCompletionMessageToolCallUnionParam
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
		var tool_call = resp.Choices[0].Message.ToolCalls[i]
		argsJSON := tool_call.Function.Arguments
		var params map[string]string
		err := json.Unmarshal([]byte(argsJSON), &params)
		if err != nil {
			log.Fatalf("Failed to parse response: %s", err)
		}

		if tool_call.Type == "function" && tool_call.Function.Name == "Read" {
			messages = ReadTool(tool_call, params, messages)
		}

		if tool_call.Type == "function" && tool_call.Function.Name == "Write" {
			messages = WriteTool(tool_call, params, messages)
		}

		if tool_call.Type == "function" && tool_call.Function.Name == "Bash" {
			messages = BashTool(tool_call, params, messages)
		}
	}

	return AgentLoop(client, messages)
}
