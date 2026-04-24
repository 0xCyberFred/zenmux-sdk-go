package platform

import "context"

// GetPAYGBalance returns the current pay-as-you-go credit balance.
func (c *Client) GetPAYGBalance(ctx context.Context) (*PAYGBalance, error) {
	var resp apiResponse[PAYGBalance]
	if err := c.http.Get(ctx, "/balance", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
