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

	for resp, err := range client.Gemini.GenerateContentStream(context.Background(), "google/gemini-2.5-pro",
		[]*genai.Content{
			{Parts: []*genai.Part{genai.NewPartFromText("Write a short poem about Go programming.")}},
		}, nil) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(resp.Text())
	}
	fmt.Println()
}
