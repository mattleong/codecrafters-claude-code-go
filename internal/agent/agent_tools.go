package agent

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/openai/openai-go/v3"
)

func ReadTool(tool_call openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) []openai.ChatCompletionMessageParamUnion {
	content, err := os.ReadFile(params["file_path"])
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}

	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role:       "tool",
			ToolCallID: tool_call.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(string(content)),
			},
		},
	})

	return messages
}

func WriteTool(tool_call openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) []openai.ChatCompletionMessageParamUnion {
	err := os.WriteFile(params["file_path"], []byte(params["content"]), 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %s", params["file_path"])
	}

	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role:       "tool",
			ToolCallID: tool_call.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(params["content"]),
			},
		},
	})
	return messages
}

func BashTool(tool_call openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) []openai.ChatCompletionMessageParamUnion {
	command_parts := strings.Split(params["command"], " ")
	out, err := exec.Command(command_parts[0], command_parts[1:]...).CombinedOutput()
	if err != nil {
		log.Fatalf("Unable to run shell command: %s", err)
	}

	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role:       "tool",
			ToolCallID: tool_call.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(string(out)),
			},
		},
	})
	return messages
}
