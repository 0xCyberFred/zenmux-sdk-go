// Generate a video using Google Veo via ZenMux's Vertex AI protocol.
//
// Unlike GenerateImages (synchronous), GenerateVideos is a long-running
// operation. The flow is:
//   1. Call GenerateVideos → returns an Operation (not the video)
//   2. Poll GetVideosOperation until operation.Done == true
//   3. Read the video bytes from operation.Response
//
// Usage:
//
//	ZENMUX_API_KEY=sk-xxx go run ./examples/gemini-video/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
	"google.golang.org/genai"
)

func main() {
	client := zenmux.NewClient(os.Getenv("ZENMUX_API_KEY"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Step 1: Submit the video generation request.
	fmt.Println("Submitting video generation request...")
	op, err := client.Gemini.GenerateVideos(ctx,
		"google/veo-3.0-generate-preview",
		"A golden retriever running on a beach at sunset, slow motion, cinematic",
		nil, // no reference image
		&genai.GenerateVideosConfig{
			AspectRatio: "16:9",
		},
	)
	if err != nil {
		log.Fatalf("GenerateVideos failed: %v", err)
	}
	fmt.Printf("Operation started: %s\n", op.Name)

	// Step 2: Poll until the operation completes.
	gc := client.Google()
	for !op.Done {
		fmt.Print(".")
		time.Sleep(10 * time.Second)

		op, err = gc.Operations.GetVideosOperation(ctx, op, nil)
		if err != nil {
			log.Fatalf("poll failed: %v", err)
		}
	}
	fmt.Println(" done!")

	// Step 3: Check for errors.
	if op.Error != nil {
		log.Fatalf("video generation failed: %v", op.Error)
	}

	// Step 4: Save the generated videos.
	if op.Response == nil || len(op.Response.GeneratedVideos) == 0 {
		log.Fatal("no videos generated")
	}

	for i, v := range op.Response.GeneratedVideos {
		filename := fmt.Sprintf("output_%d.mp4", i)
		if err := os.WriteFile(filename, v.Video.VideoBytes, 0644); err != nil {
			log.Printf("failed to save %s: %v", filename, err)
			continue
		}
		fmt.Printf("Saved %s (%d bytes)\n", filename, len(v.Video.VideoBytes))
	}
}
