package main

import (
	"flag"
	"fmt"
	"github.com/codecrafters-io/claude-code-starter-go/internal/agent"
	c "github.com/codecrafters-io/claude-code-starter-go/internal/client"
	"github.com/openai/openai-go/v3"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	client := c.GetClient()

	initial_prompt := []openai.ChatCompletionMessageParamUnion{
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	}

	result := agent.AgentLoop(client, initial_prompt)
	fmt.Print(result)
}
