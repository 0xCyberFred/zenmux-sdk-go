package openai

import (
	"net/http"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// NewClient creates a configured openai.Client for the given API key and base URL.
func NewClient(apiKey, baseURL string, httpClient *http.Client, maxRetries int) openai.Client {
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
