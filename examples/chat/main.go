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

	result, err := client.Chat.Create(context.Background(), openai.ChatCompletionNewParams{
		Model: "openai/gpt-4.1",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say hello in three languages."),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Choices[0].Message.Content)
}
