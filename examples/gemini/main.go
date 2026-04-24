package main

import (
	"context"
	"fmt"
	"log"
	"os"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
	"google.golang.org/genai"
)

func main() {
	client := zenmux.NewClient(os.Getenv("ZENMUX_API_KEY"))

	result, err := client.Gemini.GenerateContent(context.Background(), "google/gemini-2.5-pro",
		[]*genai.Content{
			{Parts: []*genai.Part{genai.NewPartFromText("What is the meaning of life? Answer briefly.")}},
		}, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Text())
}
