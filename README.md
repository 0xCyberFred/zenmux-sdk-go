# zenmux-sdk-go

Go SDK for [ZenMux AI](https://zenmux.ai) — a unified AI gateway that routes requests to OpenAI, Anthropic, and Google Vertex AI through a single API key.

## Features

- **OpenAI** — Chat Completions, Responses API, Embeddings (streaming supported)
- **Anthropic** — Messages (streaming supported)
- **Google** — Gemini GenerateContent, Imagen, Video (streaming supported)
- **Model Listing** — unified model query across all three providers
- **Platform API** — account billing, subscription, usage statistics
- **Native Escape Hatches** — access underlying `openai.Client`, `anthropic.Client`, `genai.Client` directly

## Requirements

- Go 1.24+

## Install

```bash
go get github.com/0xCyberFred/zenmux-sdk-go@latest
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"

    zenmux "github.com/0xCyberFred/zenmux-sdk-go"
    "github.com/openai/openai-go/v3"
)

func main() {
    client := zenmux.NewClient(os.Getenv("ZENMUX_API_KEY"))

    result, _ := client.Chat.Create(context.Background(), openai.ChatCompletionNewParams{
        Model: "openai/gpt-4.1",
        Messages: []openai.ChatCompletionMessageParamUnion{
            openai.UserMessage("Hello!"),
        },
    })
    fmt.Println(result.Choices[0].Message.Content)
}
```

## Usage

### OpenAI Chat Completions

```go
// Non-streaming
result, err := client.Chat.Create(ctx, openai.ChatCompletionNewParams{
    Model:    "openai/gpt-4.1",
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("Hello"),
    },
})

// Streaming
stream := client.Chat.CreateStream(ctx, openai.ChatCompletionNewParams{
    Model:    "openai/gpt-4.1",
    Messages: []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("Hello"),
    },
})
for stream.Next() {
    chunk := stream.Current()
    fmt.Print(chunk.Choices[0].Delta.Content)
}
if err := stream.Err(); err != nil { /* handle */ }
stream.Close()
```

### Anthropic Messages

```go
msg, err := client.Messages.Create(ctx, anthropic.MessageNewParams{
    Model:     "anthropic/claude-sonnet-4-5",
    MaxTokens: 1024,
    Messages: []anthropic.MessageParam{
        anthropic.NewUserMessage(anthropic.NewTextBlock("Hello")),
    },
})
fmt.Println(msg.Content[0].Text)
```

### Google Gemini

```go
resp, err := client.Gemini.GenerateContent(ctx, "google/gemini-2.5-pro",
    []*genai.Content{
        {Parts: []*genai.Part{genai.NewPartFromText("Hello")}},
    }, nil)
fmt.Println(resp.Text())

// Streaming (Go 1.23+ iter.Seq2)
for resp, err := range client.Gemini.GenerateContentStream(ctx, "google/gemini-2.5-pro",
    []*genai.Content{
        {Parts: []*genai.Part{genai.NewPartFromText("Hello")}},
    }, nil) {
    if err != nil { /* handle */ }
    fmt.Print(resp.Text())
}
```

### Model Listing

```go
models, err := client.Models.List(ctx, zenmux.ProviderOpenAI)
for _, m := range models.Models {
    fmt.Printf("%s  context=%d  reasoning=%v\n", m.ID, m.ContextLength, m.Reasoning)
}
```

### Platform API

Requires a separate Management API Key.

```go
client := zenmux.NewClient(os.Getenv("ZENMUX_API_KEY"),
    zenmux.WithManagementKey(os.Getenv("ZENMUX_MANAGEMENT_KEY")),
)

balance, _ := client.Platform.GetPAYGBalance(ctx)
fmt.Printf("Balance: $%.2f\n", balance.TotalCredits)

sub, _ := client.Platform.GetSubscription(ctx)
fmt.Printf("Plan: %s  Status: %s\n", sub.Plan.Tier, sub.AccountStatus)
```

Available methods: `GetFlowRate`, `GetPAYGBalance`, `GetSubscription`, `GetGeneration`, `GetTimeseries`, `GetLeaderboard`, `GetMarketShare`.

### Native Client Escape Hatches

For advanced use cases, access the underlying SDK clients directly:

```go
openaiClient    := client.OpenAI()    // openai.Client
anthropicClient := client.Anthropic() // anthropic.Client
googleClient    := client.Google()    // *genai.Client
```

## Configuration

```go
client := zenmux.NewClient("sk-your-key",
    zenmux.WithManagementKey("sk-mgmt-key"),        // Platform API
    zenmux.WithTimeout(60 * time.Second),            // request timeout
    zenmux.WithMaxRetries(3),                        // retry count
    zenmux.WithHTTPClient(customHTTPClient),          // custom http.Client
    zenmux.WithBaseURL(zenmux.ProviderOpenAI, "..."), // override endpoint
)
```

## Error Handling

All provider errors are wrapped in `*zenmux.Error` with convenience helpers:

```go
_, err := client.Chat.Create(ctx, params)
if zenmux.IsRateLimitError(err) { /* back off */ }
if zenmux.IsAuthError(err)      { /* check key */ }
if zenmux.IsNotFoundError(err)  { /* check model */ }

// Unwrap to original SDK error
var zenErr *zenmux.Error
if errors.As(err, &zenErr) {
    fmt.Println(zenErr.Provider, zenErr.StatusCode)
}
```

## Examples

Runnable examples for every API surface are in the [`examples/`](examples/) directory:

```bash
export ZENMUX_API_KEY=sk-your-key
go run ./examples/chat/
go run ./examples/chat-stream/
go run ./examples/messages/
go run ./examples/gemini/
go run ./examples/models/
# ...
```

## License

MIT
