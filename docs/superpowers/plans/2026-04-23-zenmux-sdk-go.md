# ZenMux Go SDK Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go SDK that provides a unified abstraction over OpenAI, Anthropic, and Google Vertex AI APIs proxied through ZenMux, plus ZenMux's Platform Management API.

**Architecture:** A single `zenmux.Client` entry point with per-provider services (`Chat`, `Responses`, `Embeddings`, `Messages`, `Gemini`, `Models`, `Platform`). Each service delegates to pre-configured official SDK clients. A unified `Error` type wraps all provider errors. Stream wrappers hide native SDK stream types while preserving the `Next()/Current()/Err()/Close()` iteration pattern.

**Tech Stack:** Go 1.24+, openai-go v3, anthropic-sdk-go, google.golang.org/genai

**Spec:** `docs/superpowers/specs/2026-04-23-zenmux-sdk-go-design.md`

---

## File Map

| File | Responsibility |
|------|---------------|
| `client.go` | `Client` struct, `NewClient()`, escape hatch methods |
| `option.go` | `config` struct, `Option` type, all `With*` functions |
| `types.go` | `Provider` enum |
| `error.go` | `Error` type, `Unwrap()`, helper predicates |
| `chat.go` | `ChatService`, `ChatCompletionStream` |
| `response.go` | `ResponseService`, `ResponseStream` |
| `embedding.go` | `EmbeddingService` |
| `message.go` | `MessageService`, `MessageStream` |
| `gemini.go` | `GeminiService` |
| `model.go` | `ModelService`, unified `Model`/`ModelList` types |
| `provider/openai/client.go` | Construct `openai.Client` with ZenMux config |
| `provider/openai/models.go` | Normalize OpenAI model list → unified `Model` |
| `provider/anthropic/client.go` | Construct `anthropic.Client` with ZenMux config |
| `provider/anthropic/models.go` | Normalize Anthropic model list → unified `Model` |
| `provider/google/client.go` | Construct `genai.Client` with custom transport |
| `provider/google/transport.go` | `zenMuxTransport` for auth header rewriting |
| `provider/google/models.go` | Normalize Google model list → unified `Model` |
| `platform/client.go` | Platform HTTP client construction |
| `platform/types.go` | All Platform API request/response types |
| `platform/flow_rate.go` | `GetFlowRate()` |
| `platform/balance.go` | `GetPAYGBalance()` |
| `platform/subscription.go` | `GetSubscription()` |
| `platform/generation.go` | `GetGeneration()` |
| `platform/statistics.go` | `GetTimeseries()`, `GetLeaderboard()`, `GetMarketShare()` |
| `internal/httpclient/client.go` | Shared HTTP helpers for Platform API |

---

### Task 1: Project Scaffolding & Core Foundation

**Files:**
- Create: `go.mod`
- Create: `types.go`
- Create: `option.go`
- Create: `error.go`
- Test: `error_test.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/fred/Workspace/0xCyberFred/zenmux-sdk-go
go mod init github.com/0xCyberFred/zenmux-sdk-go
```

- [ ] **Step 2: Install dependencies**

```bash
go get github.com/openai/openai-go/v3@latest
go get github.com/anthropics/anthropic-sdk-go@latest
go get google.golang.org/genai@latest
```

- [ ] **Step 3: Create directory structure**

```bash
mkdir -p provider/openai provider/anthropic provider/google platform internal/httpclient
```

- [ ] **Step 4: Write `types.go`**

```go
package zenmux

type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
)

const (
	defaultOpenAIBaseURL    = "https://zenmux.ai/api/v1"
	defaultAnthropicBaseURL = "https://zenmux.ai/api/anthropic"
	defaultGoogleBaseURL    = "https://zenmux.ai/api/vertex-ai"
	defaultPlatformBaseURL  = "https://zenmux.ai/api/v1/management"
)
```

- [ ] **Step 5: Write `option.go`**

```go
package zenmux

import (
	"net/http"
	"time"
)

type config struct {
	apiKey        string
	managementKey string
	httpClient    *http.Client
	maxRetries    int
	timeout       time.Duration
	baseURLs      map[Provider]string
}

func defaultConfig(apiKey string) *config {
	return &config{
		apiKey:     apiKey,
		maxRetries: 2,
		timeout:    30 * time.Second,
		baseURLs: map[Provider]string{
			ProviderOpenAI:    defaultOpenAIBaseURL,
			ProviderAnthropic: defaultAnthropicBaseURL,
			ProviderGoogle:    defaultGoogleBaseURL,
		},
	}
}

func (c *config) baseURL(p Provider) string {
	if u, ok := c.baseURLs[p]; ok {
		return u
	}
	return ""
}

func (c *config) platformBaseURL() string {
	return defaultPlatformBaseURL
}

type Option func(*config)

func WithBaseURL(provider Provider, url string) Option {
	return func(c *config) {
		c.baseURLs[provider] = url
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		c.httpClient = client
	}
}

func WithMaxRetries(n int) Option {
	return func(c *config) {
		c.maxRetries = n
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.timeout = d
	}
}

func WithManagementKey(key string) Option {
	return func(c *config) {
		c.managementKey = key
	}
}
```

- [ ] **Step 6: Write `error.go`**

```go
package zenmux

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Provider   Provider
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("zenmux %s error (HTTP %d, %s): %s", e.Provider, e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("zenmux %s error (HTTP %d): %s", e.Provider, e.StatusCode, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func IsRateLimitError(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.StatusCode == http.StatusTooManyRequests || e.StatusCode == 422
	}
	return false
}

func IsAuthError(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
	}
	return false
}

func IsNotFoundError(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.StatusCode == http.StatusNotFound
	}
	return false
}
```

- [ ] **Step 7: Write `error_test.go`**

```go
package zenmux

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorMessage(t *testing.T) {
	e := &Error{
		Provider:   ProviderOpenAI,
		StatusCode: 429,
		Code:       "rate_limit_exceeded",
		Message:    "too many requests",
	}
	want := "zenmux openai error (HTTP 429, rate_limit_exceeded): too many requests"
	if got := e.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrorMessageWithoutCode(t *testing.T) {
	e := &Error{
		Provider:   ProviderAnthropic,
		StatusCode: 500,
		Message:    "internal server error",
	}
	want := "zenmux anthropic error (HTTP 500): internal server error"
	if got := e.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrorUnwrap(t *testing.T) {
	inner := errors.New("connection refused")
	e := &Error{Provider: ProviderOpenAI, StatusCode: 0, Message: "connection failed", Err: inner}
	if !errors.Is(e, inner) {
		t.Error("Unwrap should expose inner error")
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		want   bool
	}{
		{"429", &Error{StatusCode: 429}, true},
		{"422", &Error{StatusCode: 422}, true},
		{"500", &Error{StatusCode: 500}, false},
		{"non-zenmux", errors.New("other"), false},
		{"wrapped 429", fmt.Errorf("wrap: %w", &Error{StatusCode: 429}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimitError(tt.err); got != tt.want {
				t.Errorf("IsRateLimitError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"401", &Error{StatusCode: 401}, true},
		{"403", &Error{StatusCode: 403}, true},
		{"200", &Error{StatusCode: 200}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthError(tt.err); got != tt.want {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	e := &Error{StatusCode: 404}
	if !IsNotFoundError(e) {
		t.Error("expected true for 404")
	}
	if IsNotFoundError(&Error{StatusCode: 200}) {
		t.Error("expected false for 200")
	}
}
```

- [ ] **Step 8: Run tests**

```bash
go test ./... -v
```

Expected: All tests PASS.

- [ ] **Step 9: Commit**

```bash
git add go.mod go.sum types.go option.go error.go error_test.go
git commit -m "feat: add project scaffolding with core types, options, and error handling"
```

---

### Task 2: OpenAI Provider & ChatService

**Files:**
- Create: `provider/openai/client.go`
- Create: `chat.go`
- Test: `chat_test.go`

- [ ] **Step 1: Write `provider/openai/client.go`**

```go
package openai

import (
	"net/http"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func NewClient(apiKey, baseURL string, httpClient *http.Client, maxRetries int) *openai.Client {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
		option.WithMaxRetries(maxRetries),
	}
	if httpClient != nil {
		opts = append(opts, option.WithHTTPClient(httpClient))
	}
	return openai.NewClient(opts...)
}
```

- [ ] **Step 2: Write the failing test `chat_test.go`**

```go
package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestChatServiceCreate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		resp := map[string]any{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"model":   "openai/gpt-4.1",
			"choices": []map[string]any{
				{
					"index":         0,
					"finish_reason": "stop",
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello!",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     5,
				"completion_tokens": 2,
				"total_tokens":      7,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderOpenAI, server.URL))
	result, err := client.Chat.Create(context.Background(), openai.ChatCompletionNewParams{
		Model: "openai/gpt-4.1",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hi"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "chatcmpl-123" {
		t.Errorf("unexpected ID: %s", result.ID)
	}
}

func TestChatServiceCreateError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "invalid api key",
				"type":    "authentication_error",
			},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("bad-key", WithBaseURL(ProviderOpenAI, server.URL))
	_, err := client.Chat.Create(context.Background(), openai.ChatCompletionNewParams{
		Model: "openai/gpt-4.1",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hi"),
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthError(err) {
		t.Errorf("expected auth error, got: %v", err)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./... -run TestChatService -v
```

Expected: FAIL — `ChatService` not defined.

- [ ] **Step 4: Write `chat.go`**

```go
package zenmux

import (
	"context"

	"github.com/openai/openai-go/v3"
)

type ChatService struct {
	client *openai.Client
}

func newChatService(client *openai.Client) *ChatService {
	return &ChatService{client: client}
}

func (s *ChatService) Create(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	result, err := s.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, wrapOpenAIError(err)
	}
	return result, nil
}

type ChatCompletionStream struct {
	stream openaiStream[openai.ChatCompletionChunk]
}

func (s *ChatService) CreateStream(ctx context.Context, params openai.ChatCompletionNewParams) *ChatCompletionStream {
	stream := s.client.Chat.Completions.NewStreaming(ctx, params)
	return &ChatCompletionStream{stream: stream}
}

func (s *ChatCompletionStream) Next() bool {
	return s.stream.Next()
}

func (s *ChatCompletionStream) Current() openai.ChatCompletionChunk {
	return s.stream.Current()
}

func (s *ChatCompletionStream) Err() error {
	if err := s.stream.Err(); err != nil {
		return wrapOpenAIError(err)
	}
	return nil
}

func (s *ChatCompletionStream) Close() error {
	return s.stream.Close()
}

type openaiStream[T any] interface {
	Next() bool
	Current() T
	Err() error
	Close() error
}
```

- [ ] **Step 5: Add `wrapOpenAIError` to `error.go`**

Append to `error.go`:

```go
func wrapOpenAIError(err error) error {
	if err == nil {
		return nil
	}
	e := &Error{
		Provider: ProviderOpenAI,
		Message:  err.Error(),
		Err:      err,
	}
	// Extract status code from openai error if possible
	type statusCoder interface{ StatusCode() int }
	var sc statusCoder
	if errors.As(err, &sc) {
		e.StatusCode = sc.StatusCode()
	}
	return e
}
```

- [ ] **Step 6: Write minimal `client.go` to make tests pass**

```go
package zenmux

import (
	"github.com/openai/openai-go/v3"
	openaiprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/openai"
)

type Client struct {
	cfg *config

	Chat *ChatService

	openaiClient *openai.Client
}

func NewClient(apiKey string, opts ...Option) *Client {
	cfg := defaultConfig(apiKey)
	for _, opt := range opts {
		opt(cfg)
	}

	oc := openaiprovider.NewClient(cfg.apiKey, cfg.baseURL(ProviderOpenAI), cfg.httpClient, cfg.maxRetries)

	return &Client{
		cfg:          cfg,
		Chat:         newChatService(oc),
		openaiClient: oc,
	}
}

func (c *Client) OpenAI() *openai.Client {
	return c.openaiClient
}
```

- [ ] **Step 7: Run tests**

```bash
go test ./... -v
```

Expected: All tests PASS.

- [ ] **Step 8: Commit**

```bash
git add provider/openai/client.go chat.go chat_test.go client.go error.go
git commit -m "feat: add OpenAI provider client and ChatService with streaming support"
```

---

### Task 3: ResponseService & EmbeddingService

**Files:**
- Create: `response.go`
- Create: `embedding.go`
- Test: `response_test.go`
- Test: `embedding_test.go`
- Modify: `client.go`

- [ ] **Step 1: Write `response_test.go`**

```go
package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go/v3/responses"
)

func TestResponseServiceCreate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := map[string]any{
			"id":     "resp-123",
			"object": "response",
			"model":  "openai/gpt-4.1",
			"output": []map[string]any{},
			"status": "completed",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderOpenAI, server.URL))
	result, err := client.Responses.Create(context.Background(), responses.ResponseNewParams{
		Model: "openai/gpt-4.1",
		Input: responses.ResponseNewParamsInputUnion{
			OfString: "Hello",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "resp-123" {
		t.Errorf("unexpected ID: %s", result.ID)
	}
}
```

- [ ] **Step 2: Write `embedding_test.go`**

```go
package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestEmbeddingServiceCreate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := map[string]any{
			"object": "list",
			"data": []map[string]any{
				{
					"object":    "embedding",
					"embedding": []float64{0.1, 0.2, 0.3},
					"index":     0,
				},
			},
			"model": "openai/text-embedding-3-small",
			"usage": map[string]any{
				"prompt_tokens": 5,
				"total_tokens":  5,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderOpenAI, server.URL))
	result, err := client.Embeddings.Create(context.Background(), openai.EmbeddingNewParams{
		Model: "openai/text-embedding-3-small",
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: "Hello world",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 embedding, got %d", len(result.Data))
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
go test ./... -run "TestResponseService|TestEmbeddingService" -v
```

Expected: FAIL — services not defined.

- [ ] **Step 4: Write `response.go`**

```go
package zenmux

import (
	"context"

	"github.com/openai/openai-go/v3/responses"
)

type ResponseService struct {
	client interface {
		New(ctx context.Context, params responses.ResponseNewParams, opts ...interface{}) (*responses.Response, error)
		NewStreaming(ctx context.Context, params responses.ResponseNewParams, opts ...interface{}) openaiStream[responses.ResponseStreamEventUnion]
	}
	raw *openai.Client
}

func newResponseService(client *openai.Client) *ResponseService {
	return &ResponseService{raw: client}
}

func (s *ResponseService) Create(ctx context.Context, params responses.ResponseNewParams) (*responses.Response, error) {
	result, err := s.raw.Responses.New(ctx, params)
	if err != nil {
		return nil, wrapOpenAIError(err)
	}
	return result, nil
}

type ResponseStream struct {
	stream openaiStream[responses.ResponseStreamEventUnion]
}

func (s *ResponseService) CreateStream(ctx context.Context, params responses.ResponseNewParams) *ResponseStream {
	stream := s.raw.Responses.NewStreaming(ctx, params)
	return &ResponseStream{stream: stream}
}

func (s *ResponseStream) Next() bool        { return s.stream.Next() }
func (s *ResponseStream) Current() responses.ResponseStreamEventUnion { return s.stream.Current() }
func (s *ResponseStream) Err() error {
	if err := s.stream.Err(); err != nil {
		return wrapOpenAIError(err)
	}
	return nil
}
func (s *ResponseStream) Close() error { return s.stream.Close() }
```

Note: The `ResponseService` struct above uses an interface for testing flexibility. During implementation, adjust the field to hold the concrete OpenAI client directly — the exact approach depends on the openai-go v3 API surface. The simplest working version:

```go
package zenmux

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

type ResponseService struct {
	client *openai.Client
}

func newResponseService(client *openai.Client) *ResponseService {
	return &ResponseService{client: client}
}

func (s *ResponseService) Create(ctx context.Context, params responses.ResponseNewParams) (*responses.Response, error) {
	result, err := s.client.Responses.New(ctx, params)
	if err != nil {
		return nil, wrapOpenAIError(err)
	}
	return result, nil
}

type ResponseStream struct {
	stream openaiStream[responses.ResponseStreamEventUnion]
}

func (s *ResponseService) CreateStream(ctx context.Context, params responses.ResponseNewParams) *ResponseStream {
	stream := s.client.Responses.NewStreaming(ctx, params)
	return &ResponseStream{stream: stream}
}

func (s *ResponseStream) Next() bool                                  { return s.stream.Next() }
func (s *ResponseStream) Current() responses.ResponseStreamEventUnion { return s.stream.Current() }
func (s *ResponseStream) Err() error {
	if err := s.stream.Err(); err != nil {
		return wrapOpenAIError(err)
	}
	return nil
}
func (s *ResponseStream) Close() error { return s.stream.Close() }
```

- [ ] **Step 5: Write `embedding.go`**

```go
package zenmux

import (
	"context"

	"github.com/openai/openai-go/v3"
)

type EmbeddingService struct {
	client *openai.Client
}

func newEmbeddingService(client *openai.Client) *EmbeddingService {
	return &EmbeddingService{client: client}
}

func (s *EmbeddingService) Create(ctx context.Context, params openai.EmbeddingNewParams) (*openai.CreateEmbeddingResponse, error) {
	result, err := s.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, wrapOpenAIError(err)
	}
	return result, nil
}
```

- [ ] **Step 6: Update `client.go` — add Responses and Embeddings fields**

Add to `Client` struct:

```go
Responses  *ResponseService
Embeddings *EmbeddingService
```

Add to `NewClient`:

```go
Responses:  newResponseService(oc),
Embeddings: newEmbeddingService(oc),
```

- [ ] **Step 7: Run tests**

```bash
go test ./... -v
```

Expected: All tests PASS.

- [ ] **Step 8: Commit**

```bash
git add response.go response_test.go embedding.go embedding_test.go client.go
git commit -m "feat: add ResponseService and EmbeddingService"
```

---

### Task 4: Anthropic Provider & MessageService

**Files:**
- Create: `provider/anthropic/client.go`
- Create: `message.go`
- Test: `message_test.go`
- Modify: `client.go`
- Modify: `error.go`

- [ ] **Step 1: Write `provider/anthropic/client.go`**

```go
package anthropic

import (
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func NewClient(apiKey, baseURL string, httpClient *http.Client, maxRetries int) *anthropic.Client {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
		option.WithMaxRetries(maxRetries),
	}
	if httpClient != nil {
		opts = append(opts, option.WithHTTPClient(httpClient))
	}
	return anthropic.NewClient(opts...)
}
```

- [ ] **Step 2: Write `message_test.go`**

```go
package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

func TestMessageServiceCreate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("unexpected x-api-key header: %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("missing anthropic-version header")
		}

		resp := map[string]any{
			"id":    "msg-123",
			"type":  "message",
			"role":  "assistant",
			"model": "anthropic/claude-sonnet-4-5",
			"content": []map[string]any{
				{"type": "text", "text": "Hello!"},
			},
			"stop_reason": "end_turn",
			"usage": map[string]any{
				"input_tokens":  10,
				"output_tokens": 5,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderAnthropic, server.URL))
	result, err := client.Messages.Create(context.Background(), anthropic.MessageNewParams{
		Model:     "anthropic/claude-sonnet-4-5",
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Hi")),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "msg-123" {
		t.Errorf("unexpected ID: %s", result.ID)
	}
}

func TestMessageServiceCreateError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "rate_limit_error",
				"message": "rate limited",
			},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderAnthropic, server.URL))
	_, err := client.Messages.Create(context.Background(), anthropic.MessageNewParams{
		Model:     "anthropic/claude-sonnet-4-5",
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Hi")),
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsRateLimitError(err) {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
go test ./... -run TestMessageService -v
```

Expected: FAIL — `MessageService` not defined.

- [ ] **Step 4: Write `message.go`**

```go
package zenmux

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
)

type MessageService struct {
	client *anthropic.Client
}

func newMessageService(client *anthropic.Client) *MessageService {
	return &MessageService{client: client}
}

func (s *MessageService) Create(ctx context.Context, params anthropic.MessageNewParams) (*anthropic.Message, error) {
	result, err := s.client.Messages.New(ctx, params)
	if err != nil {
		return nil, wrapAnthropicError(err)
	}
	return result, nil
}

type MessageStream struct {
	stream anthropicStream
}

func (s *MessageService) CreateStream(ctx context.Context, params anthropic.MessageNewParams) *MessageStream {
	stream := s.client.Messages.NewStreaming(ctx, params)
	return &MessageStream{stream: stream}
}

func (s *MessageStream) Next() bool                             { return s.stream.Next() }
func (s *MessageStream) Current() anthropic.MessageStreamEventUnion { return s.stream.Current() }
func (s *MessageStream) Err() error {
	if err := s.stream.Err(); err != nil {
		return wrapAnthropicError(err)
	}
	return nil
}
func (s *MessageStream) Close() error { return s.stream.Close() }

type anthropicStream interface {
	Next() bool
	Current() anthropic.MessageStreamEventUnion
	Err() error
	Close() error
}
```

- [ ] **Step 5: Add `wrapAnthropicError` to `error.go`**

```go
func wrapAnthropicError(err error) error {
	if err == nil {
		return nil
	}
	e := &Error{
		Provider: ProviderAnthropic,
		Message:  err.Error(),
		Err:      err,
	}
	type statusCoder interface{ StatusCode() int }
	var sc statusCoder
	if errors.As(err, &sc) {
		e.StatusCode = sc.StatusCode()
	}
	return e
}
```

- [ ] **Step 6: Update `client.go` — add Anthropic client and MessageService**

Add imports:

```go
anthropicprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/anthropic"
nativeanthropic "github.com/anthropics/anthropic-sdk-go"
```

Add to `Client` struct:

```go
Messages        *MessageService
anthropicClient *nativeanthropic.Client
```

Add to `NewClient`:

```go
ac := anthropicprovider.NewClient(cfg.apiKey, cfg.baseURL(ProviderAnthropic), cfg.httpClient, cfg.maxRetries)
```

Set fields:

```go
Messages:        newMessageService(ac),
anthropicClient: ac,
```

Add escape hatch:

```go
func (c *Client) Anthropic() *nativeanthropic.Client {
	return c.anthropicClient
}
```

- [ ] **Step 7: Run tests**

```bash
go test ./... -v
```

Expected: All tests PASS.

- [ ] **Step 8: Commit**

```bash
git add provider/anthropic/client.go message.go message_test.go client.go error.go
git commit -m "feat: add Anthropic provider client and MessageService with streaming support"
```

---

### Task 5: Google Provider & GeminiService

**Files:**
- Create: `provider/google/client.go`
- Create: `provider/google/transport.go`
- Create: `gemini.go`
- Test: `gemini_test.go`
- Modify: `client.go`
- Modify: `error.go`

- [ ] **Step 1: Write `provider/google/transport.go`**

```go
package google

import (
	"net/http"
)

type zenMuxTransport struct {
	apiKey string
	base   http.RoundTripper
}

func newTransport(apiKey string, base http.RoundTripper) *zenMuxTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &zenMuxTransport{apiKey: apiKey, base: base}
}

func (t *zenMuxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.Header.Set("Authorization", "Bearer "+t.apiKey)
	r.Header.Del("x-goog-api-key")
	return t.base.RoundTrip(r)
}
```

- [ ] **Step 2: Write `provider/google/client.go`**

```go
package google

import (
	"context"
	"net/http"

	"google.golang.org/genai"
)

func NewClient(ctx context.Context, apiKey, baseURL string, httpClient *http.Client) (*genai.Client, error) {
	var baseTransport http.RoundTripper
	if httpClient != nil {
		baseTransport = httpClient.Transport
	}

	transport := newTransport(apiKey, baseTransport)

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			BaseURL: baseURL,
		},
		HTTPClient: &http.Client{Transport: transport},
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
```

- [ ] **Step 3: Write `gemini_test.go`**

```go
package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/genai"
)

func TestGeminiServiceGenerateContent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("x-goog-api-key") != "" {
			t.Error("x-goog-api-key should have been removed")
		}

		resp := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]any{
							{"text": "Hello from Gemini!"},
						},
						"role": "model",
					},
					"finishReason": "STOP",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderGoogle, server.URL))
	result, err := client.Gemini.GenerateContent(context.Background(), "google/gemini-2.5-pro",
		[]*genai.Content{
			{Parts: []*genai.Part{genai.NewPartFromText("Hi")}},
		}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Text()
	if text != "Hello from Gemini!" {
		t.Errorf("unexpected text: %s", text)
	}
}
```

Note: This test may need adjustment depending on how the `genai` SDK constructs request URLs internally. The test verifies auth headers and basic response parsing. During implementation, adjust the test handler to match the actual request path that `genai` sends.

- [ ] **Step 4: Run tests to verify they fail**

```bash
go test ./... -run TestGeminiService -v
```

Expected: FAIL — `GeminiService` not defined.

- [ ] **Step 5: Write `gemini.go`**

```go
package zenmux

import (
	"context"
	"iter"

	"google.golang.org/genai"
)

type GeminiService struct {
	client *genai.Client
}

func newGeminiService(client *genai.Client) *GeminiService {
	return &GeminiService{client: client}
}

func (s *GeminiService) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	result, err := s.client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, wrapGoogleError(err)
	}
	return result, nil
}

func (s *GeminiService) GenerateContentStream(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) iter.Seq2[*genai.GenerateContentResponse, error] {
	return func(yield func(*genai.GenerateContentResponse, error) bool) {
		for resp, err := range s.client.Models.GenerateContentStream(ctx, model, contents, config) {
			if err != nil {
				yield(nil, wrapGoogleError(err))
				return
			}
			if !yield(resp, nil) {
				return
			}
		}
	}
}

func (s *GeminiService) GenerateImages(ctx context.Context, model string, prompt string, config *genai.GenerateImagesConfig) (*genai.GenerateImagesResponse, error) {
	result, err := s.client.Models.GenerateImages(ctx, model, prompt, config)
	if err != nil {
		return nil, wrapGoogleError(err)
	}
	return result, nil
}

func (s *GeminiService) GenerateVideos(ctx context.Context, model string, prompt string, image *genai.Image, config *genai.GenerateVideosConfig) (*genai.GenerateVideosOperation, error) {
	result, err := s.client.Models.GenerateVideos(ctx, model, prompt, image, config)
	if err != nil {
		return nil, wrapGoogleError(err)
	}
	return result, nil
}
```

- [ ] **Step 6: Add `wrapGoogleError` to `error.go`**

```go
func wrapGoogleError(err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Provider: ProviderGoogle,
		Message:  err.Error(),
		Err:      err,
	}
}
```

- [ ] **Step 7: Update `client.go` — add Google client and GeminiService**

Add imports:

```go
googleprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/google"
"google.golang.org/genai"
```

Add to `Client` struct:

```go
Gemini       *GeminiService
googleClient *genai.Client
```

Update `NewClient` — note that `genai.NewClient` requires `context.Context`. Add `context.Background()` for initialization:

```go
gc, err := googleprovider.NewClient(context.Background(), cfg.apiKey, cfg.baseURL(ProviderGoogle), cfg.httpClient)
if err != nil {
	// Google client creation failure is not fatal — Gemini service will be nil
	gc = nil
}
```

Set fields:

```go
Gemini:       newGeminiService(gc),
googleClient: gc,
```

Add escape hatch:

```go
func (c *Client) Google() *genai.Client {
	return c.googleClient
}
```

- [ ] **Step 8: Run tests**

```bash
go test ./... -v
```

Expected: All tests PASS.

- [ ] **Step 9: Commit**

```bash
git add provider/google/ gemini.go gemini_test.go client.go error.go
git commit -m "feat: add Google provider with custom transport and GeminiService"
```

---

### Task 6: ModelService (Unified)

**Files:**
- Create: `model.go`
- Create: `provider/openai/models.go`
- Create: `provider/anthropic/models.go`
- Create: `provider/google/models.go`
- Test: `model_test.go`
- Modify: `client.go`

- [ ] **Step 1: Write `model_test.go`**

```go
package zenmux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModelServiceListOpenAI(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := map[string]any{
			"object": "list",
			"data": []map[string]any{
				{
					"id":                "openai/gpt-4.1",
					"object":            "model",
					"display_name":      "GPT-4.1",
					"created":           1700000000,
					"owned_by":          "openai",
					"input_modalities":  []string{"text", "image"},
					"output_modalities": []string{"text"},
					"capabilities":      map[string]any{"reasoning": true},
					"context_length":    128000,
					"pricings":          map[string]any{},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(ProviderOpenAI, server.URL))
	result, err := client.Models.List(context.Background(), ProviderOpenAI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(result.Models))
	}
	m := result.Models[0]
	if m.ID != "openai/gpt-4.1" {
		t.Errorf("unexpected ID: %s", m.ID)
	}
	if m.DisplayName != "GPT-4.1" {
		t.Errorf("unexpected display name: %s", m.DisplayName)
	}
	if m.Provider != ProviderOpenAI {
		t.Errorf("unexpected provider: %s", m.Provider)
	}
	if m.ContextLength != 128000 {
		t.Errorf("unexpected context length: %d", m.ContextLength)
	}
	if !m.Reasoning {
		t.Error("expected reasoning=true")
	}
}

func TestModelServiceListInvalidProvider(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.Models.List(context.Background(), Provider("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./... -run TestModelService -v
```

Expected: FAIL.

- [ ] **Step 3: Add unified Model types to `model.go`**

```go
package zenmux

import (
	"context"
	"fmt"
)

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
	Unit       string
	Currency   string
	Conditions *PricingConditions
}

type PricingConditions struct {
	PromptTokens     *TokenRange
	CompletionTokens *TokenRange
}

type TokenRange struct {
	Gte *float64
	Lte *float64
	Gt  *float64
	Lt  *float64
}

type ModelService struct {
	openaiClient    *openai.Client
	anthropicClient *anthropic.Client
	googleClient    *genai.Client
}

func newModelService(oc *openai.Client, ac *anthropic.Client, gc *genai.Client) *ModelService {
	return &ModelService{
		openaiClient:    oc,
		anthropicClient: ac,
		googleClient:    gc,
	}
}

func (s *ModelService) List(ctx context.Context, provider Provider) (*ModelList, error) {
	switch provider {
	case ProviderOpenAI:
		return s.listOpenAI(ctx)
	case ProviderAnthropic:
		return s.listAnthropic(ctx)
	case ProviderGoogle:
		return s.listGoogle(ctx)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
```

- [ ] **Step 4: Write `provider/openai/models.go`**

This file provides a helper function that the ModelService calls to fetch and normalize OpenAI models. Since the ZenMux OpenAI List Models endpoint returns a custom JSON format (not the standard openai-go pagination type), we use a raw HTTP request via the openai client:

```go
package openai

import (
	"github.com/0xCyberFred/zenmux-sdk-go/internal/modelconv"
)

// NormalizeModel converts a raw OpenAI model JSON map to a unified Model.
// The actual implementation depends on the response format from ZenMux.
// See model.go for how this is called.
```

Note: Because the ZenMux model list response includes ZenMux-specific fields (`display_name`, `input_modalities`, `capabilities`, `pricings`) that differ from the standard openai-go `Model` type, the `ModelService.listOpenAI()` method should fetch models using the native SDK's `List` method, then normalize the results. The exact implementation depends on how much of ZenMux's custom fields the openai-go `Model` type captures. If the SDK type doesn't include them, fall back to a raw HTTP request via the Platform API.

Implement `listOpenAI` directly in `model.go`:

```go
func (s *ModelService) listOpenAI(ctx context.Context) (*ModelList, error) {
	page, err := s.openaiClient.Models.List(ctx)
	if err != nil {
		return nil, wrapOpenAIError(err)
	}

	var models []Model
	for _, m := range page.Data {
		models = append(models, Model{
			ID:          m.ID,
			DisplayName: m.ID, // fallback if display_name not in SDK type
			Provider:    ProviderOpenAI,
		})
	}
	return &ModelList{Models: models}, nil
}
```

Note: During implementation, inspect the actual fields available on the openai-go `Model` struct. ZenMux's response includes `display_name`, `input_modalities`, `output_modalities`, `capabilities`, `context_length`, and `pricings`. If the openai-go Model type uses `JSON()` or raw JSON access, extract these fields. Otherwise, make a raw HTTP GET to `/models` and unmarshal manually. Adjust this implementation accordingly.

- [ ] **Step 5: Implement `listAnthropic` and `listGoogle` in `model.go`**

```go
func (s *ModelService) listAnthropic(ctx context.Context) (*ModelList, error) {
	page, err := s.anthropicClient.Models.List(ctx, anthropic.ModelListParams{})
	if err != nil {
		return nil, wrapAnthropicError(err)
	}

	var models []Model
	for _, m := range page.Data {
		models = append(models, Model{
			ID:          m.ID,
			DisplayName: m.DisplayName,
			Provider:    ProviderAnthropic,
		})
	}
	return &ModelList{Models: models}, nil
}

func (s *ModelService) listGoogle(ctx context.Context) (*ModelList, error) {
	if s.googleClient == nil {
		return nil, fmt.Errorf("google client not initialized")
	}
	page, err := s.googleClient.Models.List(ctx, nil)
	if err != nil {
		return nil, wrapGoogleError(err)
	}

	var models []Model
	for _, m := range page.Items {
		models = append(models, Model{
			ID:          m.Name,
			DisplayName: m.DisplayName,
			Provider:    ProviderGoogle,
		})
	}
	return &ModelList{Models: models}, nil
}
```

Note: Field names on the native SDK Model types (`page.Data`, `m.ID`, `m.DisplayName`, `page.Items`, `m.Name`) need to be verified against the actual SDK types during implementation. The normalization logic follows the same pattern regardless of exact field names.

- [ ] **Step 6: Update `client.go` — add ModelService**

Add to `Client` struct:

```go
Models *ModelService
```

Add to `NewClient`:

```go
Models: newModelService(oc, ac, gc),
```

- [ ] **Step 7: Run tests**

```bash
go test ./... -v
```

Expected: All tests PASS. Note: The OpenAI model test may need adjustment depending on how the openai-go SDK handles the `/models` response from ZenMux's custom format vs standard OpenAI format. Adjust the test handler response to match.

- [ ] **Step 8: Commit**

```bash
git add model.go model_test.go provider/openai/models.go provider/anthropic/models.go provider/google/models.go client.go
git commit -m "feat: add ModelService with unified model listing across all providers"
```

---

### Task 7: Platform API

**Files:**
- Create: `internal/httpclient/client.go`
- Create: `platform/types.go`
- Create: `platform/client.go`
- Create: `platform/flow_rate.go`
- Create: `platform/balance.go`
- Create: `platform/subscription.go`
- Create: `platform/generation.go`
- Create: `platform/statistics.go`
- Test: `platform/client_test.go`

- [ ] **Step 1: Write `internal/httpclient/client.go`**

```go
package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) error {
	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", path, err)
	}
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
	}

	return json.Unmarshal(body, out)
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}
```

- [ ] **Step 2: Write `platform/types.go`**

```go
package platform

type apiResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type FlowRate struct {
	Currency            string  `json:"currency"`
	BaseUSDPerFlow      float64 `json:"base_usd_per_flow"`
	EffectiveUSDPerFlow float64 `json:"effective_usd_per_flow"`
}

type PAYGBalance struct {
	Currency     string  `json:"currency"`
	TotalCredits float64 `json:"total_credits"`
	TopUpCredits float64 `json:"top_up_credits"`
	BonusCredits float64 `json:"bonus_credits"`
}

type SubscriptionDetail struct {
	Plan                Plan         `json:"plan"`
	Currency            string       `json:"currency"`
	BaseUSDPerFlow      float64      `json:"base_usd_per_flow"`
	EffectiveUSDPerFlow float64      `json:"effective_usd_per_flow"`
	AccountStatus       string       `json:"account_status"`
	Quota5Hour          Quota        `json:"quota_5_hour"`
	Quota7Day           Quota        `json:"quota_7_day"`
	QuotaMonthly        QuotaMonthly `json:"quota_monthly"`
}

type Plan struct {
	Tier      string  `json:"tier"`
	AmountUSD float64 `json:"amount_usd"`
	Interval  string  `json:"interval"`
	ExpiresAt string  `json:"expires_at"`
}

type Quota struct {
	UsagePercentage float64 `json:"usage_percentage"`
	ResetsAt        *string `json:"resets_at"`
	MaxFlows        float64 `json:"max_flows"`
	UsedFlows       float64 `json:"used_flows"`
	RemainingFlows  float64 `json:"remaining_flows"`
	UsedValueUSD    float64 `json:"used_value_usd"`
	MaxValueUSD     float64 `json:"max_value_usd"`
}

type QuotaMonthly struct {
	MaxFlows    float64 `json:"max_flows"`
	MaxValueUSD float64 `json:"max_value_usd"`
}

type Generation struct {
	API               string           `json:"api"`
	GenerationID      string           `json:"generationId"`
	Model             string           `json:"model"`
	CreateAt          string           `json:"createAt"`
	GenerationTime    int              `json:"generationTime"`
	Latency           int              `json:"latency"`
	NativeTokens      TokenUsage       `json:"nativeTokens"`
	Streamed          bool             `json:"streamed"`
	FinishReason      string           `json:"finishReason"`
	Usage             float64          `json:"usage"`
	RatingResponses   *RatingResponses `json:"ratingResponses"`
	RequestRetryTimes int              `json:"requestRetryTimes"`
	FinalRetry        bool             `json:"finalRetry"`
}

type TokenUsage struct {
	CompletionTokens        int                     `json:"completion_tokens"`
	PromptTokens            int                     `json:"prompt_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
	PromptTokensDetails     PromptTokensDetails     `json:"prompt_tokens_details"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type RatingResponses struct {
	BillAmount     float64        `json:"billAmount"`
	DiscountAmount float64        `json:"discountAmount"`
	OriginAmount   float64        `json:"originAmount"`
	PriceVersion   string         `json:"priceVersion"`
	RatingDetails  []RatingDetail `json:"ratingDetails"`
}

type RatingDetail struct {
	BillAmount     float64 `json:"billAmount"`
	DiscountAmount float64 `json:"discountAmount"`
	FeeItemCode    string  `json:"feeItemCode"`
	OriginAmount   float64 `json:"originAmount"`
	Rate           float64 `json:"rate"`
}

type TimeseriesParams struct {
	Metric      string
	BucketWidth string
	StartingAt  string
	EndingAt    string
	Limit       int
}

type Timeseries struct {
	Metric       string             `json:"metric"`
	BucketWidth  string             `json:"bucket_width"`
	StartingAt   string             `json:"starting_at"`
	EndingAt     string             `json:"ending_at"`
	TotalBuckets int                `json:"total_buckets"`
	Series       []TimeseriesBucket `json:"series"`
}

type TimeseriesBucket struct {
	Period string        `json:"period"`
	Date   string        `json:"date"`
	Models []ModelMetric `json:"models"`
}

type ModelMetric struct {
	Model string  `json:"model"`
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type LeaderboardParams struct {
	Metric     string
	StartingAt string
	EndingAt   string
	Limit      int
}

type Leaderboard struct {
	Metric     string             `json:"metric"`
	StartingAt string             `json:"starting_at"`
	EndingAt   string             `json:"ending_at"`
	Entries    []LeaderboardEntry `json:"entries"`
}

type LeaderboardEntry struct {
	Rank        int     `json:"rank"`
	Model       string  `json:"model"`
	Label       string  `json:"label"`
	Author      string  `json:"author"`
	AuthorLabel string  `json:"author_label"`
	Value       float64 `json:"value"`
}

type MarketShareParams struct {
	Metric      string
	BucketWidth string
	StartingAt  string
	EndingAt    string
	Limit       int
}

type MarketShare struct {
	Metric       string              `json:"metric"`
	BucketWidth  string              `json:"bucket_width"`
	StartingAt   string              `json:"starting_at"`
	EndingAt     string              `json:"ending_at"`
	TotalBuckets int                 `json:"total_buckets"`
	Series       []MarketShareBucket `json:"series"`
}

type MarketShareBucket struct {
	Period  string         `json:"period"`
	Date    string         `json:"date"`
	Authors []AuthorMetric `json:"authors"`
}

type AuthorMetric struct {
	Author string  `json:"author"`
	Label  string  `json:"label"`
	Value  float64 `json:"value"`
}
```

- [ ] **Step 3: Write `platform/client.go`**

```go
package platform

import (
	"net/http"

	"github.com/0xCyberFred/zenmux-sdk-go/internal/httpclient"
)

type Client struct {
	http *httpclient.Client
}

func NewClient(baseURL, managementKey string, httpClient *http.Client) *Client {
	return &Client{
		http: httpclient.New(baseURL, managementKey, httpClient),
	}
}
```

- [ ] **Step 4: Write `platform/flow_rate.go`**

```go
package platform

import "context"

func (c *Client) GetFlowRate(ctx context.Context) (*FlowRate, error) {
	var resp apiResponse[FlowRate]
	if err := c.http.Get(ctx, "/flow_rate", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
```

- [ ] **Step 5: Write `platform/balance.go`**

```go
package platform

import "context"

func (c *Client) GetPAYGBalance(ctx context.Context) (*PAYGBalance, error) {
	var resp apiResponse[PAYGBalance]
	if err := c.http.Get(ctx, "/payg/balance", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
```

- [ ] **Step 6: Write `platform/subscription.go`**

```go
package platform

import "context"

func (c *Client) GetSubscription(ctx context.Context) (*SubscriptionDetail, error) {
	var resp apiResponse[SubscriptionDetail]
	if err := c.http.Get(ctx, "/subscription/detail", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
```

- [ ] **Step 7: Write `platform/generation.go`**

```go
package platform

import (
	"context"
	"net/url"
)

func (c *Client) GetGeneration(ctx context.Context, id string) (*Generation, error) {
	q := url.Values{}
	q.Set("id", id)
	var resp Generation
	if err := c.http.Get(ctx, "/generation", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 8: Write `platform/statistics.go`**

```go
package platform

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) GetTimeseries(ctx context.Context, params TimeseriesParams) (*Timeseries, error) {
	q := url.Values{}
	q.Set("metric", params.Metric)
	q.Set("bucket_width", params.BucketWidth)
	if params.StartingAt != "" {
		q.Set("starting_at", params.StartingAt)
	}
	if params.EndingAt != "" {
		q.Set("ending_at", params.EndingAt)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}

	var resp apiResponse[Timeseries]
	if err := c.http.Get(ctx, "/statistics/timeseries", q, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetLeaderboard(ctx context.Context, params LeaderboardParams) (*Leaderboard, error) {
	q := url.Values{}
	q.Set("metric", params.Metric)
	if params.StartingAt != "" {
		q.Set("starting_at", params.StartingAt)
	}
	if params.EndingAt != "" {
		q.Set("ending_at", params.EndingAt)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}

	var resp apiResponse[Leaderboard]
	if err := c.http.Get(ctx, "/statistics/leaderboard", q, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetMarketShare(ctx context.Context, params MarketShareParams) (*MarketShare, error) {
	q := url.Values{}
	q.Set("metric", params.Metric)
	q.Set("bucket_width", params.BucketWidth)
	if params.StartingAt != "" {
		q.Set("starting_at", params.StartingAt)
	}
	if params.EndingAt != "" {
		q.Set("ending_at", params.EndingAt)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}

	var resp apiResponse[MarketShare]
	if err := c.http.Get(ctx, "/statistics/market_share", q, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
```

- [ ] **Step 9: Write `platform/client_test.go`**

```go
package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFlowRate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/flow_rate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer mgmt-key" {
			t.Errorf("unexpected auth: %s", r.Header.Get("Authorization"))
		}

		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"currency":              "usd",
				"base_usd_per_flow":     0.03283,
				"effective_usd_per_flow": 0.03283,
			},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient(server.URL, "mgmt-key", nil)
	result, err := client.GetFlowRate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Currency != "usd" {
		t.Errorf("unexpected currency: %s", result.Currency)
	}
	if result.BaseUSDPerFlow != 0.03283 {
		t.Errorf("unexpected base rate: %f", result.BaseUSDPerFlow)
	}
}

func TestGetPAYGBalance(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payg/balance" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"currency":      "usd",
				"total_credits": 482.74,
				"top_up_credits": 35.00,
				"bonus_credits": 447.74,
			},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient(server.URL, "mgmt-key", nil)
	result, err := client.GetPAYGBalance(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCredits != 482.74 {
		t.Errorf("unexpected total: %f", result.TotalCredits)
	}
}

func TestGetSubscription(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"plan": map[string]any{
					"tier":       "pro",
					"amount_usd": 29.99,
					"interval":   "month",
					"expires_at": "2026-05-23T00:00:00Z",
				},
				"currency":               "usd",
				"base_usd_per_flow":      0.03283,
				"effective_usd_per_flow": 0.03283,
				"account_status":         "healthy",
				"quota_5_hour":           map[string]any{"max_flows": 1000, "used_flows": 50, "remaining_flows": 950, "usage_percentage": 0.05, "used_value_usd": 1.64, "max_value_usd": 32.83},
				"quota_7_day":            map[string]any{"max_flows": 5000, "used_flows": 200, "remaining_flows": 4800, "usage_percentage": 0.04, "used_value_usd": 6.57, "max_value_usd": 164.15},
				"quota_monthly":          map[string]any{"max_flows": 20000, "max_value_usd": 656.60},
			},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient(server.URL, "mgmt-key", nil)
	result, err := client.GetSubscription(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Plan.Tier != "pro" {
		t.Errorf("unexpected tier: %s", result.Plan.Tier)
	}
	if result.AccountStatus != "healthy" {
		t.Errorf("unexpected status: %s", result.AccountStatus)
	}
}

func TestGetGeneration(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "gen-123" {
			t.Errorf("unexpected id param: %s", r.URL.Query().Get("id"))
		}

		json.NewEncoder(w).Encode(map[string]any{
			"api":            "chat.completions",
			"generationId":   "gen-123",
			"model":          "openai/gpt-4.1",
			"createAt":       "2026-04-23T10:00:00Z",
			"generationTime": 1500,
			"latency":        200,
			"nativeTokens": map[string]any{
				"completion_tokens": 100,
				"prompt_tokens":     50,
				"total_tokens":      150,
			},
			"streamed":     true,
			"finishReason": "stop",
			"usage":        0.005,
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient(server.URL, "mgmt-key", nil)
	result, err := client.GetGeneration(context.Background(), "gen-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GenerationID != "gen-123" {
		t.Errorf("unexpected generation ID: %s", result.GenerationID)
	}
	if result.GenerationTime != 1500 {
		t.Errorf("unexpected generation time: %d", result.GenerationTime)
	}
}

func TestGetTimeseries(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("metric") != "tokens" {
			t.Errorf("unexpected metric: %s", r.URL.Query().Get("metric"))
		}
		if r.URL.Query().Get("bucket_width") != "1d" {
			t.Errorf("unexpected bucket_width: %s", r.URL.Query().Get("bucket_width"))
		}

		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"metric":        "tokens",
				"bucket_width":  "1d",
				"starting_at":   "2026-04-01",
				"ending_at":     "2026-04-23",
				"total_buckets": 1,
				"series": []map[string]any{
					{
						"period": "20260423",
						"date":   "2026-04-23",
						"models": []map[string]any{
							{"model": "openai/gpt-4.1", "label": "GPT-4.1", "value": 50000},
						},
					},
				},
			},
		})
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient(server.URL, "mgmt-key", nil)
	result, err := client.GetTimeseries(context.Background(), TimeseriesParams{
		Metric:      "tokens",
		BucketWidth: "1d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Series) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(result.Series))
	}
	if result.Series[0].Models[0].Value != 50000 {
		t.Errorf("unexpected value: %f", result.Series[0].Models[0].Value)
	}
}

func TestHTTPError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewClient(server.URL, "bad-key", nil)
	_, err := client.GetFlowRate(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
```

- [ ] **Step 10: Run tests**

```bash
go test ./... -v
```

Expected: All tests PASS.

- [ ] **Step 11: Update `client.go` — wire Platform client**

Add to `NewClient`:

```go
var pc *platform.Client
if cfg.managementKey != "" {
	pc = platform.NewClient(cfg.platformBaseURL(), cfg.managementKey, cfg.httpClient)
}
```

Set field:

```go
Platform: pc,
```

- [ ] **Step 12: Commit**

```bash
git add internal/httpclient/ platform/ client.go
git commit -m "feat: add Platform API client with all management endpoints"
```

---

### Task 8: Core Client Integration

**Files:**
- Modify: `client.go` (finalize)
- Test: `client_test.go`

- [ ] **Step 1: Review and finalize `client.go`**

Ensure the final `client.go` has all services wired up:

```go
package zenmux

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/anthropics/anthropic-sdk-go"
	"google.golang.org/genai"

	openaiprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/openai"
	anthropicprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/anthropic"
	googleprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/google"
	"github.com/0xCyberFred/zenmux-sdk-go/platform"
)

type Client struct {
	cfg *config

	Chat       *ChatService
	Responses  *ResponseService
	Embeddings *EmbeddingService
	Messages   *MessageService
	Gemini     *GeminiService
	Models     *ModelService
	Platform   *platform.Client

	openaiClient    *openai.Client
	anthropicClient *anthropic.Client
	googleClient    *genai.Client
}

func NewClient(apiKey string, opts ...Option) *Client {
	cfg := defaultConfig(apiKey)
	for _, opt := range opts {
		opt(cfg)
	}

	oc := openaiprovider.NewClient(cfg.apiKey, cfg.baseURL(ProviderOpenAI), cfg.httpClient, cfg.maxRetries)
	ac := anthropicprovider.NewClient(cfg.apiKey, cfg.baseURL(ProviderAnthropic), cfg.httpClient, cfg.maxRetries)

	var gc *genai.Client
	gc, _ = googleprovider.NewClient(context.Background(), cfg.apiKey, cfg.baseURL(ProviderGoogle), cfg.httpClient)

	var pc *platform.Client
	if cfg.managementKey != "" {
		pc = platform.NewClient(cfg.platformBaseURL(), cfg.managementKey, cfg.httpClient)
	}

	return &Client{
		cfg:             cfg,
		Chat:            newChatService(oc),
		Responses:       newResponseService(oc),
		Embeddings:      newEmbeddingService(oc),
		Messages:        newMessageService(ac),
		Gemini:          newGeminiService(gc),
		Models:          newModelService(oc, ac, gc),
		Platform:        pc,
		openaiClient:    oc,
		anthropicClient: ac,
		googleClient:    gc,
	}
}

func (c *Client) OpenAI() *openai.Client       { return c.openaiClient }
func (c *Client) Anthropic() *anthropic.Client  { return c.anthropicClient }
func (c *Client) Google() *genai.Client         { return c.googleClient }
```

- [ ] **Step 2: Write `client_test.go`**

```go
package zenmux

import (
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	client := NewClient("test-key")

	if client.Chat == nil {
		t.Error("Chat service should not be nil")
	}
	if client.Responses == nil {
		t.Error("Responses service should not be nil")
	}
	if client.Embeddings == nil {
		t.Error("Embeddings service should not be nil")
	}
	if client.Messages == nil {
		t.Error("Messages service should not be nil")
	}
	if client.Gemini == nil {
		t.Error("Gemini service should not be nil")
	}
	if client.Models == nil {
		t.Error("Models service should not be nil")
	}
	if client.Platform != nil {
		t.Error("Platform should be nil without management key")
	}
}

func TestNewClientWithManagementKey(t *testing.T) {
	client := NewClient("test-key", WithManagementKey("mgmt-key"))
	if client.Platform == nil {
		t.Error("Platform should not be nil with management key")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClient("test-key",
		WithMaxRetries(5),
		WithTimeout(60*time.Second),
		WithBaseURL(ProviderOpenAI, "https://custom.example.com"),
	)
	if client.cfg.maxRetries != 5 {
		t.Errorf("expected maxRetries=5, got %d", client.cfg.maxRetries)
	}
	if client.cfg.timeout != 60*time.Second {
		t.Errorf("expected timeout=60s, got %v", client.cfg.timeout)
	}
	if client.cfg.baseURLs[ProviderOpenAI] != "https://custom.example.com" {
		t.Errorf("unexpected base URL: %s", client.cfg.baseURLs[ProviderOpenAI])
	}
}

func TestEscapeHatches(t *testing.T) {
	client := NewClient("test-key")
	if client.OpenAI() == nil {
		t.Error("OpenAI() should return non-nil client")
	}
	if client.Anthropic() == nil {
		t.Error("Anthropic() should return non-nil client")
	}
	// Google client may be nil if initialization fails in test env
}
```

- [ ] **Step 3: Run all tests**

```bash
go test ./... -v
```

Expected: All tests PASS.

- [ ] **Step 4: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: finalize Client with all services wired and escape hatches"
```

---

### Task 9: Example Tests & Cleanup

**Files:**
- Create: `example_test.go`
- Modify: `go.mod` (tidy)

- [ ] **Step 1: Write `example_test.go`**

```go
package zenmux_test

import (
	"context"
	"fmt"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
	"github.com/openai/openai-go/v3"
)

func ExampleNewClient() {
	client := zenmux.NewClient("sk-your-zenmux-key",
		zenmux.WithManagementKey("sk-mgmt-your-key"),
	)

	// All services are available
	_ = client.Chat       // OpenAI Chat Completions
	_ = client.Responses  // OpenAI Responses
	_ = client.Embeddings // OpenAI Embeddings
	_ = client.Messages   // Anthropic Messages
	_ = client.Gemini     // Google Gemini
	_ = client.Models     // Unified model listing
	_ = client.Platform   // Platform management API

	// Native client escape hatches
	_ = client.OpenAI()    // *openai.Client
	_ = client.Anthropic() // *anthropic.Client
	_ = client.Google()    // *genai.Client

	fmt.Println("client created")
	// Output: client created
}

func ExampleChatService_Create() {
	client := zenmux.NewClient("sk-your-key")

	result, err := client.Chat.Create(context.Background(), openai.ChatCompletionNewParams{
		Model: "openai/gpt-4.1",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hello"),
		},
	})
	if err != nil {
		// Handle error — use zenmux.IsRateLimitError, IsAuthError, etc.
		panic(err)
	}

	_ = result // *openai.ChatCompletion
}
```

Note: `ExampleChatService_Create` will panic since there's no real server. It serves as documentation, not a runnable test. If needed, wrap in `// Output:` or mark with `//nolint` to prevent test runner from executing the API call. Alternatively, only include examples that don't make network calls (like `ExampleNewClient`).

- [ ] **Step 2: Run `go mod tidy`**

```bash
go mod tidy
```

- [ ] **Step 3: Run all tests one final time**

```bash
go test ./... -v -count=1
```

Expected: All tests PASS.

- [ ] **Step 4: Commit**

```bash
git add example_test.go go.mod go.sum
git commit -m "feat: add example tests and tidy module dependencies"
```

---

## Implementation Notes

**Type name verification:** The openai-go, anthropic-sdk-go, and genai SDKs evolve. During implementation, verify exact type names, method signatures, and field names against the installed SDK versions. The code in this plan is based on API research as of 2026-04-23. Adjust as needed — the architecture and patterns remain the same regardless of minor type name differences.

**Google auth adaptation:** The `zenMuxTransport` approach (Task 5) is the safest path. If during implementation you discover that `genai.ClientConfig.HTTPOptions.Headers` can set `Authorization: Bearer` without needing the custom transport, simplify accordingly.

**Model listing normalization:** The ZenMux model list endpoints return custom fields (`display_name`, `pricings`, `capabilities`, etc.) that may not map cleanly to the native SDK's `Model` types. If the native SDK types don't capture these fields, use raw HTTP requests (via `internal/httpclient`) to fetch and unmarshal the model list directly. This is a known area where implementation may diverge from the plan.

**`TokenUsage` nested JSON:** The `nativeTokens` field in the Generation response has nested paths like `completion_tokens_details.reasoning_tokens`. Go's `json` package doesn't support dot-notation in struct tags. During implementation, use nested structs:

```go
type TokenUsage struct {
	CompletionTokens        int                      `json:"completion_tokens"`
	PromptTokens            int                      `json:"prompt_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	CompletionTokensDetails CompletionTokensDetails  `json:"completion_tokens_details"`
	PromptTokensDetails     PromptTokensDetails      `json:"prompt_tokens_details"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}
```
