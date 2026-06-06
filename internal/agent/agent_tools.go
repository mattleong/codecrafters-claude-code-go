package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openai/openai-go/v3"
)

func readTool(toolCall openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
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
			ToolCallID: toolCall.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(string(content)),
			},
		},
	})

	return messages, nil
}

func writeTool(toolCall openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
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
			ToolCallID: toolCall.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(content),
			},
		},
	})
	return messages, nil
}

func bashTool(toolCall openai.ChatCompletionMessageToolCallUnion, params map[string]string, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
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
			ToolCallID: toolCall.ID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(string(out)),
			},
		},
	})
	return messages, nil
}

func GetToolDefinitionParams() []openai.ChatCompletionToolUnionParam {
	return []openai.ChatCompletionToolUnionParam{
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
	}
}

func GetToolCallParams(toolCalls []openai.ChatCompletionMessageToolCallUnion) []openai.ChatCompletionMessageToolCallUnionParam {
	toolCallParams := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(toolCalls))
	for _, tc := range toolCalls {
		toolCallParams = append(toolCallParams, tc.ToParam())
	}

	return toolCallParams
}

func ExecuteToolCalls(toolCalls []openai.ChatCompletionMessageToolCallUnion, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	for _, toolCall := range toolCalls {
		argsJSON := toolCall.Function.Arguments
		var params map[string]string
		err := json.Unmarshal([]byte(argsJSON), &params)
		if err != nil {
			return nil, fmt.Errorf("could not parse response params %w", err)
		}

		switch toolCall.Function.Name {
		case "Read":
			messages, err = readTool(toolCall, params, messages)
		case "Write":
			messages, err = writeTool(toolCall, params, messages)
		case "Bash":
			messages, err = bashTool(toolCall, params, messages)
		default:
			return nil, fmt.Errorf("unsupported tool call")
		}

		if err != nil {
			return nil, err
		}
	}

	return messages, nil
}
