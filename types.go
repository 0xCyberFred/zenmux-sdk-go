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
