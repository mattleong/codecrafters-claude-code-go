package agent

import (
	"fmt"

	"github.com/openai/openai-go/v3"
)

var maxSteps = 100

func (c *Agent) Loop(prompt []openai.ChatCompletionMessageParamUnion) (string, error) {
	messages := []openai.ChatCompletionMessageParamUnion{}
	messages = append(messages, prompt...)

	for range maxSteps {
		resp, err := c.SendChatCompletion(messages)
		if err != nil {
			return "", fmt.Errorf("error getting response back from client: %w", err)
		}

		if len(resp.Choices[0].Message.ToolCalls) == 0 {
			return resp.Choices[0].Message.Content, nil
		}

		toolCallParams := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(resp.Choices[0].Message.ToolCalls))
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			toolCallParams = append(toolCallParams, tc.ToParam())
		}

		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Role:      "assistant",
				ToolCalls: toolCallParams,
			},
		})

		messages, err = ExecuteTool(resp.Choices[0].Message.ToolCalls, messages)
		if err != nil {
			return "", fmt.Errorf("error executing tool call: %w", err)
		}
	}

	return "", fmt.Errorf("agent exceeded max steps: %d", maxSteps)
}
