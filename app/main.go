package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/codecrafters-io/claude-code-starter-go/internal/agent"
	c "github.com/codecrafters-io/claude-code-starter-go/internal/client"
	"github.com/openai/openai-go/v3"
	"log"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		log.Fatalf("Prompt must not be empty")
	}

	client, err := c.NewClient()
	if err != nil {
		log.Fatalf("Unable to initialize client: %s", err)
	}

	initialPrompt := []openai.ChatCompletionMessageParamUnion{
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	}

	ctx := context.Background()

	result, err := agent.AgentLoop(ctx, client, initialPrompt)
	if err != nil {
		log.Fatalf("Could not complete agent loop: %s", err)
	}

	fmt.Print(result)
}
