package agent

import (
	"fmt"

	"github.com/openai/openai-go/v3"
)

var maxSteps = 100

func (c *Agent) Loop(messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	messages = append(messages, messages...)

	for range maxSteps {
		resp, err := c.client.Chat.Completions.New(c.ctx,
			openai.ChatCompletionNewParams{
				Model:    c.model,
				Messages: messages,
				Tools:    GetToolParams(),
			},
		)
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

func (c *Agent) StartLoop(prompt string) (string, error) {
	return c.Loop([]openai.ChatCompletionMessageParamUnion{
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	})
}
