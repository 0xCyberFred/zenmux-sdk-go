package zenmux

import (
	"github.com/openai/openai-go/v3"

	openaiprovider "github.com/0xCyberFred/zenmux-sdk-go/provider/openai"
)

// Client is the top-level ZenMux SDK client. It exposes service objects that
// map to provider APIs.
type Client struct {
	cfg *config

	// Chat provides access to chat completion endpoints.
	Chat *ChatService

	openaiClient openai.Client
}

// NewClient creates a new ZenMux client configured with the given API key and
// optional settings.
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

// OpenAI returns the underlying openai-go client for direct access to
// provider-specific functionality.
func (c *Client) OpenAI() openai.Client {
	return c.openaiClient
}
