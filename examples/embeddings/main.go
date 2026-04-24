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

	result, err := client.Embeddings.Create(context.Background(), openai.EmbeddingNewParams{
		Model: openai.EmbeddingModelTextEmbedding3Small,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String("Hello world"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Model: %s\n", result.Model)
	fmt.Printf("Dimensions: %d\n", len(result.Data[0].Embedding))
	fmt.Printf("Usage: %d tokens\n", result.Usage.TotalTokens)
}
