// Generate a video via ZenMux's Vertex AI protocol.
//
// Unlike GenerateImages (synchronous), GenerateVideos is a long-running
// operation. The flow is:
//  1. Call GenerateVideos → returns an Operation (not the video)
//  2. Poll Gemini.GetVideosOperation until operation.Done == true
//  3. Read the video bytes from operation.Response
//
// Note: this uses client.Gemini.GetVideosOperation (a ZenMux SDK method),
// not gc.Operations.GetVideosOperation from the genai library, because
// ZenMux speaks Vertex protocol on a non-default API version.
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

	fmt.Println("Submitting video generation request...")
	op, err := client.Gemini.GenerateVideos(ctx,
		"bytedance/doubao-seedance-2.0",
		"A golden retriever running on a beach at sunset, slow motion, cinematic",
		nil,
		&genai.GenerateVideosConfig{
			AspectRatio: "16:9",
		},
	)
	if err != nil {
		log.Fatalf("GenerateVideos failed: %v", err)
	}
	fmt.Printf("Operation started: %s\n", op.Name)

	for !op.Done {
		fmt.Print(".")
		time.Sleep(10 * time.Second)
		op, err = client.Gemini.GetVideosOperation(ctx, op)
		if err != nil {
			log.Fatalf("poll failed: %v", err)
		}
	}
	fmt.Println(" done!")

	if op.Error != nil {
		log.Fatalf("video generation failed: %v", op.Error)
	}
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
