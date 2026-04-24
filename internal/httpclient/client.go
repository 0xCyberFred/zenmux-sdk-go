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

// Client is a lightweight HTTP client that handles authentication, JSON
// parsing, and error handling for simple GET-based APIs.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a new Client. If httpClient is nil a default client with a
// 30-second timeout is used.
func New(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{baseURL: baseURL, apiKey: apiKey, httpClient: httpClient}
}

// Get performs an authenticated GET request, appending query to path and
// JSON-decoding the response body into out. Non-2xx responses are returned as
// *HTTPError.
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
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	return json.Unmarshal(body, out)
}

// HTTPError represents a non-success HTTP response.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}
