package zenmux

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go/v3"
	"google.golang.org/genai"

	anthropicprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/anthropic"
	googleprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/google"
	openaiprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/openai"
)

// Client is the top-level ZenMux SDK client. It exposes service objects that
// map to provider APIs.
type Client struct {
	cfg *config

	// Chat provides access to chat completion endpoints.
	Chat *ChatService
	// Responses provides access to the Responses API endpoints.
	Responses *ResponseService
	// Embeddings provides access to the embeddings API endpoints.
	Embeddings *EmbeddingService
	// Messages provides access to the Anthropic Messages API endpoints.
	Messages *MessageService
	// Gemini provides access to the Google Gemini API endpoints.
	Gemini *GeminiService

	openaiClient    openai.Client
	anthropicClient anthropic.Client
	googleClient    *genai.Client
}

// NewClient creates a new ZenMux client configured with the given API key and
// optional settings.
func NewClient(apiKey string, opts ...Option) *Client {
	cfg := defaultConfig(apiKey)
	for _, opt := range opts {
		opt(cfg)
	}

	oc := openaiprovider.NewClient(cfg.apiKey, cfg.baseURL(ProviderOpenAI), cfg.httpClient, cfg.maxRetries)
	ac := anthropicprovider.NewClient(cfg.apiKey, cfg.baseURL(ProviderAnthropic), cfg.httpClient, cfg.maxRetries)
	gc, _ := googleprovider.NewClient(context.Background(), cfg.apiKey, cfg.baseURL(ProviderGoogle), cfg.httpClient)

	return &Client{
		cfg:             cfg,
		Chat:            newChatService(oc),
		Responses:       newResponseService(oc),
		Embeddings:      newEmbeddingService(oc),
		Messages:        newMessageService(ac),
		Gemini:          newGeminiService(gc),
		openaiClient:    oc,
		anthropicClient: ac,
		googleClient:    gc,
	}
}

// OpenAI returns the underlying openai-go client for direct access to
// provider-specific functionality.
func (c *Client) OpenAI() openai.Client {
	return c.openaiClient
}

// Anthropic returns the underlying anthropic-sdk-go client for direct access
// to provider-specific functionality.
func (c *Client) Anthropic() anthropic.Client {
	return c.anthropicClient
}

// Google returns the underlying genai client for direct access to
// provider-specific functionality.
func (c *Client) Google() *genai.Client {
	return c.googleClient
}
