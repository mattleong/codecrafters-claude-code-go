package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func agent_loop(client openai.Client, prompt []openai.ChatCompletionMessageParamUnion) string {
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
			},
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	messages := []openai.ChatCompletionMessageParamUnion{}
	messages = append(messages, prompt...)

	fmt.Fprintln(os.Stderr, "tool calls: ", len(resp.Choices[0].Message.ToolCalls))

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
		}

		if tool_call.Type == "function" && tool_call.Function.Name == "Write" {
			err = os.WriteFile(params["file_path"], []byte(params["content"]), 0644)
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
		}
	}

	return agent_loop(client, messages)
}

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))

	initial_prompt := []openai.ChatCompletionMessageParamUnion{
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	}

	result := agent_loop(client, initial_prompt)

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	fmt.Print(result)
}
