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
				Tools:    GetToolDefinitionParams(),
			},
		)
		if err != nil {
			return "", fmt.Errorf("error getting response back from client: %w", err)
		}

		toolCalls := resp.Choices[0].Message.ToolCalls

		if len(toolCalls) == 0 {
			return resp.Choices[0].Message.Content, nil
		}

		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Role:      "assistant",
				ToolCalls: GetToolCallParams(toolCalls),
			},
		})

		messages, err = ExecuteToolCalls(toolCalls, messages)
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
