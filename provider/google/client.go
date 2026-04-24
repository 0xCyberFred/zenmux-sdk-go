package google

import (
	"context"
	"net/http"

	"google.golang.org/genai"
)

// NewClient creates a genai.Client configured with ZenMux-compatible auth
// transport and base URL.
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
