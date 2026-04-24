// Generate images using OpenAI GPT-Image-2 via ZenMux's Vertex AI protocol.
//
// Demonstrates both sync and async (goroutine) patterns, since the genai SDK
// only provides a synchronous GenerateImages method.
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
	"sync"
	"time"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
	"google.golang.org/genai"
)

func main() {
	client := zenmux.NewClient(os.Getenv("ZENMUX_API_KEY"))

	// --- Sync with timeout ---
	fmt.Println("=== Sync (with 2-minute timeout) ===")
	syncGenerate(client)

	// --- Async via goroutine ---
	fmt.Println("\n=== Async (goroutine + channel) ===")
	asyncGenerate(client)

	// --- Multiple concurrent requests ---
	fmt.Println("\n=== Concurrent (3 images in parallel) ===")
	concurrentGenerate(client)
}

// syncGenerate is the simplest approach: block until done, with a timeout
// to prevent hanging forever.
func syncGenerate(client *zenmux.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	resp, err := client.Gemini.GenerateImages(ctx,
		"openai/gpt-image-2",
		"A cat wearing sunglasses sitting on a surfboard, digital art",
		&genai.GenerateImagesConfig{
			NumberOfImages: 1,
			OutputMIMEType: "image/png",
		},
	)
	if err != nil {
		log.Printf("sync error: %v", err)
		return
	}
	saveImages(resp, "sync")
}

// ImageResult holds the result of an async image generation.
type ImageResult struct {
	Response *genai.GenerateImagesResponse
	Err      error
}

// asyncGenerate fires the request in a goroutine and returns immediately.
// The caller can do other work while waiting on the channel.
func asyncGenerate(client *zenmux.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ch := make(chan ImageResult, 1)

	go func() {
		resp, err := client.Gemini.GenerateImages(ctx,
			"openai/gpt-image-2",
			"A robot painting a landscape, watercolor style",
			&genai.GenerateImagesConfig{
				NumberOfImages: 1,
				OutputMIMEType: "image/png",
			},
		)
		ch <- ImageResult{Response: resp, Err: err}
	}()

	// Do other work here while image generates...
	fmt.Println("  request sent, doing other work...")

	// Wait for result
	result := <-ch
	if result.Err != nil {
		log.Printf("async error: %v", result.Err)
		return
	}
	saveImages(result.Response, "async")
}

// concurrentGenerate launches multiple image requests in parallel.
func concurrentGenerate(client *zenmux.Client) {
	prompts := []string{
		"A futuristic city at sunset, cyberpunk",
		"A cozy cabin in snowy mountains, oil painting",
		"An underwater coral reef with tropical fish, photorealistic",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	for i, prompt := range prompts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("  [%d] generating: %s\n", i, prompt[:40])

			resp, err := client.Gemini.GenerateImages(ctx,
				"openai/gpt-image-2",
				prompt,
				&genai.GenerateImagesConfig{
					NumberOfImages: 1,
					OutputMIMEType: "image/png",
				},
			)
			if err != nil {
				log.Printf("  [%d] error: %v", i, err)
				return
			}
			saveImages(resp, fmt.Sprintf("concurrent_%d", i))
		}()
	}
	wg.Wait()
}

func saveImages(resp *genai.GenerateImagesResponse, prefix string) {
	for i, img := range resp.GeneratedImages {
		filename := fmt.Sprintf("%s_%d.png", prefix, i)
		if err := os.WriteFile(filename, img.Image.ImageBytes, 0644); err != nil {
			log.Printf("failed to save %s: %v", filename, err)
			continue
		}
		fmt.Printf("  Saved %s (%d bytes)\n", filename, len(img.Image.ImageBytes))
	}
}
