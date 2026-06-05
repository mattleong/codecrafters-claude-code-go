package agent

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openai/openai-go/v3"
)

func ReadTool(toolCall openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	filePath, ok := params["file_path"]
	if !ok || filePath == "" {
		return nil, fmt.Errorf("missing file_path")
	}

	content, err := os.ReadFile(filePath)
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
	filePath, ok := params["file_path"]
	if !ok || filePath == "" {
		return nil, fmt.Errorf("missing file_path")
	}

	content, ok := params["content"]
	if !ok {
		return nil, fmt.Errorf("missing content")
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write file: %w", err)
	}

	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role:       "tool",
			ToolCallID: toolCall.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(content),
			},
		},
	})
	return messages, nil
}

func BashTool(toolCall openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	command, ok := params["command"]
	if !ok {
		return nil, fmt.Errorf("missing command")
	}

	commandParts := strings.Fields(command)
	out, err := exec.Command(commandParts[0], commandParts[1:]...).CombinedOutput()
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
