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
	if u, ok := c.baseURLs["platform"]; ok {
		return u
	}
	return defaultPlatformBaseURL
}

func (c *config) effectiveHTTPClient() *http.Client {
	if c.httpClient != nil {
		return c.httpClient
	}
	return &http.Client{Timeout: c.timeout}
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
