package zenmux_test

import (
	"fmt"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
)

func ExampleNewClient() {
	client := zenmux.NewClient("sk-your-zenmux-key",
		zenmux.WithManagementKey("sk-mgmt-your-key"),
	)

	_ = client.Chat       // OpenAI Chat Completions
	_ = client.Responses  // OpenAI Responses
	_ = client.Embeddings // OpenAI Embeddings
	_ = client.Messages   // Anthropic Messages
	_ = client.Gemini     // Google Gemini
	_ = client.Models     // Unified model listing
	_ = client.Platform   // Platform management API

	_ = client.OpenAI()    // *openai.Client escape hatch
	_ = client.Anthropic() // *anthropic.Client escape hatch
	_ = client.Google()    // *genai.Client escape hatch

	fmt.Println("client created")
	// Output: client created
}
