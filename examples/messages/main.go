package main

import (
	"context"
	"fmt"
	"log"
	"os"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := zenmux.NewClient(os.Getenv("ZENMUX_API_KEY"))

	result, err := client.Messages.Create(context.Background(), anthropic.MessageNewParams{
		Model:     "anthropic/claude-sonnet-4-5",
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("What is 2+2? Answer in one word.")),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Role: %s\n", result.Role)
	for _, block := range result.Content {
		fmt.Println(block.Text)
	}
	fmt.Printf("Usage: input=%d output=%d\n", result.Usage.InputTokens, result.Usage.OutputTokens)
}
