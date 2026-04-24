package anthropic

import (
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// NewClient creates a configured anthropic.Client for the given API key and base URL.
func NewClient(apiKey, baseURL string, httpClient *http.Client, maxRetries int) anthropic.Client {
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
