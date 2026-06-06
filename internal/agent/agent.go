package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codecrafters-io/claude-code-starter-go/internal/client"
	"github.com/openai/openai-go/v3"
)

func AgentLoop(ctx context.Context, prompt []openai.ChatCompletionMessageParamUnion) (string, error) {
	resp, err := client.MakeRequest(ctx, prompt)

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

	return AgentLoop(ctx, messages)
}
