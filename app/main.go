package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/codecrafters-io/claude-code-starter-go/internal/agent"
	"github.com/openai/openai-go/v3"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		log.Fatalf("Prompt must not be empty")
	}

	agentClient, err := agent.NewAgent("anthropic/claude-haiku-4.5")
	if err != nil {
		log.Fatalf("Unable to initialize client: %s", err)
	}

	result, err := agentClient.Loop([]openai.ChatCompletionMessageParamUnion{
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Could not complete agent loop: %s", err)
	}

	fmt.Print(result)
}
