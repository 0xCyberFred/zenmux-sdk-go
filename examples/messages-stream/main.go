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

	stream := client.Messages.CreateStream(context.Background(), anthropic.MessageNewParams{
		Model:     "anthropic/claude-sonnet-4-5",
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Tell me a haiku about coding.")),
		},
	})

	for stream.Next() {
		event := stream.Current()
		switch v := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			if v.Delta.Type == "text_delta" {
				fmt.Print(v.Delta.Text)
			}
		}
	}
	fmt.Println()

	if err := stream.Err(); err != nil {
		log.Fatal(err)
	}
	stream.Close()
}
