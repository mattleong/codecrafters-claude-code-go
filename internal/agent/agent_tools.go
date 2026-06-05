package agent

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openai/openai-go/v3"
)

func ReadTool(toolCall openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	content, err := os.ReadFile(params["file_path"])
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role:       "tool",
			ToolCallID: toolCall.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(string(content)),
			},
		},
	})

	return messages, nil
}

func WriteTool(toolCall openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	err := os.WriteFile(params["file_path"], []byte(params["content"]), 0644)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role:       "tool",
			ToolCallID: toolCall.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(params["content"]),
			},
		},
	})
	return messages, nil
}

func BashTool(toolCall openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	command_parts := strings.Fields(params["command"])
	out, err := exec.Command(command_parts[0], command_parts[1:]...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not run shell command: %w", err)
	}

	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role:       "tool",
			ToolCallID: toolCall.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(string(out)),
			},
		},
	})
	return messages, nil
}
