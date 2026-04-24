package platform

import (
	"net/http"

	"github.com/0xCyberFred/zenmux-sdk-go/internal/httpclient"
)

// Client provides access to the ZenMux Platform API. It uses a separate
// Management API Key (not the provider API key) and communicates via HTTP GET
// requests.
type Client struct {
	http *httpclient.Client
}

// NewClient creates a new Platform API client. If httpClient is nil a default
// client with a 30-second timeout is used.
func NewClient(baseURL, managementKey string, httpClient *http.Client) *Client {
	return &Client{http: httpclient.New(baseURL, managementKey, httpClient)}
}
