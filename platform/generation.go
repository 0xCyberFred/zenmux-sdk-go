package platform

import (
	"context"
	"net/url"
)

// GetGeneration returns details about a single generation request identified
// by id. Unlike other endpoints the response is NOT wrapped in apiResponse.
func (c *Client) GetGeneration(ctx context.Context, id string) (*Generation, error) {
	q := url.Values{}
	q.Set("id", id)
	var resp Generation
	if err := c.http.Get(ctx, "/generation", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
