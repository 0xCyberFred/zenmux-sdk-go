// Generate images using OpenAI GPT-Image-2 via ZenMux's Vertex AI protocol.
//
// ZenMux routes all image generation through the Vertex AI protocol,
// even for non-Google models like openai/gpt-image-2. The SDK translates
// Vertex AI parameters to the provider's native format internally.
//
// Usage:
//
//	ZENMUX_API_KEY=sk-xxx go run ./examples/image-generate/
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

	resp, err := client.Gemini.GenerateImages(context.Background(),
		"openai/gpt-image-2",
		"A cat wearing sunglasses sitting on a surfboard, digital art",
		&genai.GenerateImagesConfig{
			NumberOfImages: 1,
			OutputMIMEType: "image/png",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	for i, img := range resp.GeneratedImages {
		filename := fmt.Sprintf("output_%d.png", i)
		if err := os.WriteFile(filename, img.Image.ImageBytes, 0644); err != nil {
			log.Fatalf("failed to save %s: %v", filename, err)
		}
		fmt.Printf("Saved %s (%d bytes)\n", filename, len(img.Image.ImageBytes))
	}
}
