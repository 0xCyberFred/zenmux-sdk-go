# ZenMux Go SDK Design Spec

## Overview

A Go SDK for [ZenMux AI](https://zenmux.ai) that provides a **unified abstraction layer** over three AI provider APIs (OpenAI, Anthropic, Google Vertex AI) plus ZenMux's own Platform Management API. Users interact primarily through a single `zenmux.Client`, with native SDK clients exposed as escape hatches for advanced use cases.

**Module path:** `github.com/0xCyberFred/zenmux-sdk-go`
**Go version:** 1.23+ (required by `google.golang.org/genai` for `iter.Seq2`)
**License:** MIT

## API Scope

| Provider | Capabilities | ZenMux Endpoint |
|----------|-------------|-----------------|
| OpenAI | Chat Completions, Responses, Embeddings, List Models | `https://zenmux.ai/api/v1` |
| Anthropic | Messages, List Models | `https://zenmux.ai/api/anthropic` |
| Google | Gemini GenerateContent, Imagen, Video, List Models | `https://zenmux.ai/api/vertex-ai` |
| Platform | Flow Rate, PAYG Balance, Subscription, Generation, Statistics | `https://zenmux.ai/api/v1/management` |

## External Dependencies

| SDK | Import Path | Purpose |
|-----|-------------|---------|
| openai-go v3 | `github.com/openai/openai-go/v3` | OpenAI Chat Completions, Responses, Embeddings |
| anthropic-sdk-go | `github.com/anthropics/anthropic-sdk-go` | Anthropic Messages |
| Google Gen AI | `google.golang.org/genai` | Google Gemini, Imagen, Video |

## Architecture

### Directory Structure

```
zenmux-sdk-go/
├── client.go                  // Client construction and configuration
├── option.go                  // Functional options (WithAPIKey, WithTimeout, etc.)
├── chat.go                    // ChatService (OpenAI Chat Completions)
├── response.go                // ResponseService (OpenAI Responses)
├── embedding.go               // EmbeddingService (OpenAI Embeddings)
├── message.go                 // MessageService (Anthropic Messages)
├── gemini.go                  // GeminiService (Google GenerateContent/Imagen/Video)
├── model.go                   // ModelService (unified List Models across 3 providers)
├── types.go                   // Shared types (Provider enum, Model, Pricing, etc.)
├── error.go                   // Unified error type
│
├── provider/
│   ├── openai/
│   │   ├── adapter.go         // openai-go adapter
│   │   └── client.go          // Native client construction
│   ├── anthropic/
│   │   ├── adapter.go         // anthropic-sdk-go adapter
│   │   └── client.go          // Native client construction
│   └── google/
│       ├── adapter.go         // genai adapter
│       └── client.go          // Native client construction
│
├── platform/
│   ├── client.go              // Platform HTTP client
│   ├── flow_rate.go           // GET /management/flow_rate
│   ├── balance.go             // GET /management/payg/balance
│   ├── subscription.go        // GET /management/subscription/detail
│   ├── generation.go          // GET /management/generation
│   └── statistics.go          // Timeseries, Leaderboard, MarketShare
│
└── internal/
    └── httpclient/
        └── client.go          // Shared HTTP utilities for Platform API
```

### Core Client

```go
package zenmux

type Client struct {
    cfg        *config

    // Unified abstraction services
    Chat       *ChatService
    Responses  *ResponseService
    Embeddings *EmbeddingService
    Messages   *MessageService
    Gemini     *GeminiService
    Models     *ModelService
    Platform   *platform.Client
}

func NewClient(apiKey string, opts ...Option) *Client
```

`NewClient` takes a single ZenMux API key shared across all three providers. Each provider adapter is internally constructed with the appropriate BaseURL and auth configuration.

### Configuration Options

```go
type Option func(*config)

func WithBaseURL(provider Provider, url string) Option
func WithHTTPClient(client *http.Client) Option
func WithMaxRetries(n int) Option
func WithTimeout(d time.Duration) Option
func WithManagementKey(key string) Option
```

Default BaseURLs:

| Provider | Default BaseURL |
|----------|----------------|
| OpenAI | `https://zenmux.ai/api/v1` |
| Anthropic | `https://zenmux.ai/api/anthropic` |
| Google | `https://zenmux.ai/api/vertex-ai` |
| Platform | `https://zenmux.ai/api/v1/management` |

### Native Client Escape Hatches

```go
func (c *Client) OpenAI() *openai.Client
func (c *Client) Anthropic() *anthropic.Client
func (c *Client) Google() *genai.Client
```

Returns pre-configured native SDK clients pointing at ZenMux endpoints.

## Unified API Surface

### Design Principle

Each service **reuses the native SDK's parameter and response types** directly. ZenMux's provider APIs are compatible with the official APIs, so wrapping them in custom types would add no value and increase maintenance burden.

The only exception is `ModelService`, where three providers return different response formats that need normalization into a unified `Model` type.

### ChatService (OpenAI Chat Completions)

```go
type ChatService struct { ... }

func (s *ChatService) Create(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
func (s *ChatService) CreateStream(ctx context.Context, params openai.ChatCompletionNewParams) *ChatCompletionStream
```

`ChatCompletionStream`:

```go
type ChatCompletionStream struct { ... }

func (s *ChatCompletionStream) Next() bool
func (s *ChatCompletionStream) Current() openai.ChatCompletionChunk
func (s *ChatCompletionStream) Err() error
func (s *ChatCompletionStream) Close() error
```

### ResponseService (OpenAI Responses)

```go
type ResponseService struct { ... }

func (s *ResponseService) Create(ctx context.Context, params openai.ResponseNewParams) (*openai.Response, error)
func (s *ResponseService) CreateStream(ctx context.Context, params openai.ResponseNewParams) *ResponseStream
```

`ResponseStream` follows the same `Next()/Current()/Err()/Close()` pattern.

### EmbeddingService (OpenAI Embeddings)

```go
type EmbeddingService struct { ... }

func (s *EmbeddingService) Create(ctx context.Context, params openai.EmbeddingNewParams) (*openai.CreateEmbeddingResponse, error)
```

Synchronous only, no streaming.

### MessageService (Anthropic Messages)

```go
type MessageService struct { ... }

func (s *MessageService) Create(ctx context.Context, params anthropic.MessageNewParams) (*anthropic.Message, error)
func (s *MessageService) CreateStream(ctx context.Context, params anthropic.MessageNewParams) *MessageStream
```

`MessageStream`:

```go
type MessageStream struct { ... }

func (s *MessageStream) Next() bool
func (s *MessageStream) Current() anthropic.MessageStreamEvent
func (s *MessageStream) Err() error
func (s *MessageStream) Close() error
```

### GeminiService (Google Gemini / Imagen / Video)

```go
type GeminiService struct { ... }

// Text generation
func (s *GeminiService) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
func (s *GeminiService) GenerateContentStream(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) iter.Seq2[*genai.GenerateContentResponse, error]

// Image generation
func (s *GeminiService) GenerateImages(ctx context.Context, model string, prompt string, config *genai.GenerateImagesConfig) (*genai.GenerateImagesResponse, error)

// Video generation
func (s *GeminiService) GenerateVideos(ctx context.Context, model string, prompt string, config *genai.GenerateVideosConfig) (*genai.GenerateVideosOperation, error)
```

Streaming uses Go 1.23 `iter.Seq2`, consistent with the `genai` SDK's native pattern.

### ModelService (Unified List Models)

```go
type ModelService struct { ... }

type Provider string

const (
    ProviderOpenAI    Provider = "openai"
    ProviderAnthropic Provider = "anthropic"
    ProviderGoogle    Provider = "google"
)

func (s *ModelService) List(ctx context.Context, provider Provider) (*ModelList, error)
```

Normalized response type:

```go
type ModelList struct {
    Models []Model
}

type Model struct {
    ID               string
    DisplayName      string
    Provider         Provider
    InputModalities  []string
    OutputModalities []string
    ContextLength    int
    Reasoning        bool
    Pricings         map[string][]Pricing
}

type Pricing struct {
    Value      float64
    Unit       string  // "perMTokens", "perCount", "perSecond"
    Currency   string  // "USD"
    Conditions *PricingConditions
}

type PricingConditions struct {
    PromptTokens     *TokenRange
    CompletionTokens *TokenRange
}

type TokenRange struct {
    Gte *float64 // unit: kTokens
    Lte *float64
    Gt  *float64
    Lt  *float64
}
```

This is the only service with custom ZenMux types, because the three providers return different response schemas that must be normalized.

## Platform API

Uses `Authorization: Bearer <MANAGEMENT_API_KEY>` authentication. Requires a separate Management API Key passed via `WithManagementKey()`.

```go
package platform

type Client struct { ... }

// Account & Billing
func (c *Client) GetFlowRate(ctx context.Context) (*FlowRate, error)
func (c *Client) GetPAYGBalance(ctx context.Context) (*PAYGBalance, error)
func (c *Client) GetSubscription(ctx context.Context) (*SubscriptionDetail, error)

// Generation
func (c *Client) GetGeneration(ctx context.Context, id string) (*Generation, error)

// Statistics
func (c *Client) GetTimeseries(ctx context.Context, params TimeseriesParams) (*Timeseries, error)
func (c *Client) GetLeaderboard(ctx context.Context, params LeaderboardParams) (*Leaderboard, error)
func (c *Client) GetMarketShare(ctx context.Context, params MarketShareParams) (*MarketShare, error)
```

### Platform Types

```go
type FlowRate struct {
    Currency            string  // "usd"
    BaseUSDPerFlow      float64
    EffectiveUSDPerFlow float64
}

type PAYGBalance struct {
    Currency     string
    TotalCredits float64
    TopUpCredits float64
    BonusCredits float64
}

type SubscriptionDetail struct {
    Plan                Plan
    Currency            string
    BaseUSDPerFlow      float64
    EffectiveUSDPerFlow float64
    AccountStatus       string // "healthy", "monitored", "abusive", "suspended", "banned"
    Quota5Hour          Quota
    Quota7Day           Quota
    QuotaMonthly        QuotaMonthly
}

type Plan struct {
    Tier      string  // "free", "pro", "max", "ultra"
    AmountUSD float64
    Interval  string  // "month"
    ExpiresAt string  // ISO 8601
}

type Quota struct {
    UsagePercentage float64
    ResetsAt        *string // nil if window hasn't started
    MaxFlows        float64
    UsedFlows       float64
    RemainingFlows  float64
    UsedValueUSD    float64
    MaxValueUSD     float64
}

type QuotaMonthly struct {
    MaxFlows    float64
    MaxValueUSD float64
}

type Generation struct {
    API              string // "chat.completions", "responses", "messages", "generateContent"
    GenerationID     string
    Model            string
    CreateAt         string
    GenerationTime   int // ms
    Latency          int // ms, TTFT
    NativeTokens     TokenUsage
    Streamed         bool
    FinishReason     string
    Usage            float64
    RatingResponses  *RatingResponses // nil for subscription keys
    RequestRetryTimes int
    FinalRetry       bool
}

type TokenUsage struct {
    CompletionTokens int
    PromptTokens     int
    TotalTokens      int
    ReasoningTokens  int
    CachedTokens     int
}

type RatingResponses struct {
    BillAmount     float64
    DiscountAmount float64
    OriginAmount   float64
    PriceVersion   string
    RatingDetails  []RatingDetail
}

type RatingDetail struct {
    BillAmount     float64
    DiscountAmount float64
    FeeItemCode    string // "prompt", "completion"
    OriginAmount   float64
    Rate           float64
}

// Statistics query params
type TimeseriesParams struct {
    Metric      string // "tokens" | "cost"
    BucketWidth string // "1d" | "1w"
    StartingAt  string // YYYY-MM-DD, optional
    EndingAt    string // YYYY-MM-DD, optional
    Limit       int    // default 10, max 50
}

type LeaderboardParams struct {
    Metric     string // "tokens" | "cost"
    StartingAt string // YYYY-MM-DD, optional
    EndingAt   string // YYYY-MM-DD, optional
    Limit      int    // default 10, max 50
}

type MarketShareParams struct {
    Metric      string // "tokens" | "cost"
    BucketWidth string // "1d" | "1w"
    StartingAt  string // YYYY-MM-DD, optional
    EndingAt    string // YYYY-MM-DD, optional
    Limit       int    // default 10, max 50
}

type Timeseries struct {
    Metric       string
    BucketWidth  string
    StartingAt   string
    EndingAt     string
    TotalBuckets int
    Series       []TimeseriesBucket
}

type TimeseriesBucket struct {
    Period string // YYYYMMDD or YYYYWW
    Date   string // YYYY-MM-DD
    Models []ModelMetric
}

type ModelMetric struct {
    Model string
    Label string
    Value float64
}

type Leaderboard struct {
    Metric     string
    StartingAt string
    EndingAt   string
    Entries    []LeaderboardEntry
}

type LeaderboardEntry struct {
    Rank        int
    Model       string
    Label       string
    Author      string
    AuthorLabel string
    Value       float64
}

type MarketShare struct {
    Metric       string
    BucketWidth  string
    StartingAt   string
    EndingAt     string
    TotalBuckets int
    Series       []MarketShareBucket
}

type MarketShareBucket struct {
    Period  string
    Date    string
    Authors []AuthorMetric
}

type AuthorMetric struct {
    Author string
    Label  string
    Value  float64
}
```

## Error Handling

```go
package zenmux

type Error struct {
    Provider   Provider
    StatusCode int
    Code       string
    Message    string
    Err        error // original SDK error, supports errors.Is/As unwrapping
}

func (e *Error) Error() string
func (e *Error) Unwrap() error

func IsRateLimitError(err error) bool // 422
func IsAuthError(err error) bool      // 401/403
func IsNotFoundError(err error) bool  // 404
```

Each provider adapter catches the native SDK's error type and wraps it in `*zenmux.Error`, preserving the original error for unwrapping via `errors.As`.

## Provider Adapter Details

### OpenAI Adapter

Straightforward — `openai-go` natively supports `option.WithBaseURL()` and `option.WithAPIKey()`:

```go
client := openai.NewClient(
    option.WithAPIKey(apiKey),
    option.WithBaseURL("https://zenmux.ai/api/v1"),
)
```

ZenMux uses `Authorization: Bearer <key>`, which matches `openai-go`'s default auth format.

### Anthropic Adapter

Similarly straightforward — `anthropic-sdk-go` supports the same options:

```go
client := anthropic.NewClient(
    option.WithAPIKey(apiKey),
    option.WithBaseURL("https://zenmux.ai/api/anthropic"),
)
```

ZenMux uses `x-api-key` header for the Anthropic endpoint, which matches `anthropic-sdk-go`'s default behavior.

### Google Adapter

Requires special handling due to two mismatches:

1. **Auth format:** ZenMux uses `Authorization: Bearer <key>`, but `genai` SDK uses `x-goog-api-key` (Gemini API backend) or Google ADC (Vertex AI backend).
2. **Base URL isolation:** `genai.SetDefaultBaseURLs()` is a **package-level** function that affects all `genai.Client` instances in the process. As a library, we must not mutate global state.

**Recommended approach:** Use `BackendGeminiAPI` with `APIKey` and a custom `http.RoundTripper` that rewrites the base URL and auth header per-request. This provides full isolation without global side effects:

```go
type zenMuxTransport struct {
    apiKey  string
    baseURL string
    base    http.RoundTripper
}

func (t *zenMuxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    req.Header.Set("Authorization", "Bearer "+t.apiKey)
    req.Header.Del("x-goog-api-key")
    // Rewrite host to ZenMux endpoint
    u, _ := url.Parse(t.baseURL)
    req.URL.Scheme = u.Scheme
    req.URL.Host = u.Host
    req.Host = u.Host
    return t.base.RoundTrip(req)
}
```

**Alternative approach (if `genai` supports per-client config):** During implementation, investigate whether `ClientConfig.HTTPOptions` or similar per-client mechanisms can set base URL without global mutation. If available, prefer that over transport rewriting.

## Authentication Summary

| Endpoint | Header | Key Type |
|----------|--------|----------|
| OpenAI (`/api/v1`) | `Authorization: Bearer <key>` | ZenMux API Key |
| Anthropic (`/api/anthropic`) | `x-api-key: <key>` | ZenMux API Key |
| Google (`/api/vertex-ai`) | `Authorization: Bearer <key>` | ZenMux API Key |
| Platform (`/api/v1/management`) | `Authorization: Bearer <key>` | Management API Key |

## Testing Strategy

- **Unit tests:** Type conversion logic in each provider adapter, error wrapping, Model normalization
- **Integration tests:** `httptest.Server` mocking ZenMux endpoints to verify request format (headers, body, URL paths)
- **Example tests:** `example_test.go` as living documentation showing usage of each API

## Usage Example

```go
package main

import (
    "context"
    "fmt"

    "github.com/0xCyberFred/zenmux-sdk-go"
    "github.com/anthropics/anthropic-sdk-go"
    "github.com/openai/openai-go"
    "google.golang.org/genai"
)

func main() {
    ctx := context.Background()

    client := zenmux.NewClient("sk-your-zenmux-key",
        zenmux.WithManagementKey("sk-mgmt-your-key"),
    )

    // OpenAI Chat Completions
    chat, _ := client.Chat.Create(ctx, openai.ChatCompletionNewParams{
        Model: "openai/gpt-4.1",
        Messages: []openai.ChatCompletionMessageParamUnion{
            openai.UserMessage("Hello"),
        },
    })
    fmt.Println(chat.Choices[0].Message.Content)

    // OpenAI Chat Completions (streaming)
    stream := client.Chat.CreateStream(ctx, openai.ChatCompletionNewParams{
        Model: "openai/gpt-4.1",
        Messages: []openai.ChatCompletionMessageParamUnion{
            openai.UserMessage("Tell me a story"),
        },
    })
    for stream.Next() {
        fmt.Print(stream.Current().Choices[0].Delta.Content)
    }
    stream.Close()

    // Anthropic Messages
    msg, _ := client.Messages.Create(ctx, anthropic.MessageNewParams{
        Model:     "anthropic/claude-sonnet-4-5",
        MaxTokens: 1024,
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock("Hi")),
        },
    })
    fmt.Println(msg.Content[0].Text)

    // Google Gemini
    resp, _ := client.Gemini.GenerateContent(ctx, "google/gemini-2.5-pro",
        []*genai.Content{
            {Parts: []*genai.Part{genai.NewPartFromText("Hello")}},
        }, nil)
    fmt.Println(resp.Text())

    // Model Listing
    models, _ := client.Models.List(ctx, zenmux.ProviderOpenAI)
    for _, m := range models.Models {
        fmt.Printf("%s (%s)\n", m.ID, m.DisplayName)
    }

    // Platform API
    balance, _ := client.Platform.GetPAYGBalance(ctx)
    fmt.Printf("Balance: $%.2f\n", balance.TotalCredits)

    // Escape hatch: native clients
    _ = client.OpenAI()    // *openai.Client
    _ = client.Anthropic() // *anthropic.Client
    _ = client.Google()    // *genai.Client
}
```
