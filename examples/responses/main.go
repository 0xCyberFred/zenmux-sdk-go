package main

import (
	"context"
	"fmt"
	"log"
	"os"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

func main() {
	client := zenmux.NewClient(os.Getenv("ZENMUX_API_KEY"))

	result, err := client.Responses.Create(context.Background(), responses.ResponseNewParams{
		Model: shared.ResponsesModel("openai/gpt-4.1"),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("What is the capital of France?"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Status: %s\n", result.Status)
	for _, item := range result.Output {
		fmt.Printf("Output: %+v\n", item)
	}
}
