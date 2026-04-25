package zenmux

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go/v3"
	"google.golang.org/genai"

	"github.com/0xCyberFred/zenmux-sdk-go/platform"
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
	// Models provides unified model listing across all providers.
	Models *ModelService
	// Platform provides access to the ZenMux Platform management API.
	Platform *platform.Client

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

	hc := cfg.effectiveHTTPClient()

	oc := openaiprovider.NewClient(cfg.apiKey, cfg.baseURL(ProviderOpenAI), hc, cfg.maxRetries)
	ac := anthropicprovider.NewClient(cfg.apiKey, cfg.baseURL(ProviderAnthropic), hc, cfg.maxRetries)
	gc, _ := googleprovider.NewClient(context.Background(), cfg.apiKey, cfg.baseURL(ProviderGoogle), hc)

	var gemini *GeminiService
	if gc != nil {
		gemini = newGeminiService(gc, cfg)
	}

	var pc *platform.Client
	if cfg.managementKey != "" {
		pc = platform.NewClient(cfg.platformBaseURL(), cfg.managementKey, hc)
	}

	return &Client{
		cfg:             cfg,
		Chat:            newChatService(oc),
		Responses:       newResponseService(oc),
		Embeddings:      newEmbeddingService(oc),
		Messages:        newMessageService(ac),
		Gemini:          gemini,
		Models:          newModelService(cfg),
		Platform:        pc,
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
