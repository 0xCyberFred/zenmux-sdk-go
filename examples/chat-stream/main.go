package main

import (
	"context"
	"fmt"
	"log"
	"os"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
	"github.com/openai/openai-go/v3"
)

func main() {
	client := zenmux.NewClient(os.Getenv("ZENMUX_API_KEY"))

	stream := client.Chat.CreateStream(context.Background(), openai.ChatCompletionNewParams{
		Model: "openai/gpt-4.1",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Tell me a short joke."),
		},
	})

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}
	fmt.Println()

	if err := stream.Err(); err != nil {
		log.Fatal(err)
	}
	stream.Close()
}
