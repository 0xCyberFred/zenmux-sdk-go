package platform

import "context"

// GetSubscription returns the current subscription details including plan,
// quotas, and pricing information.
func (c *Client) GetSubscription(ctx context.Context) (*SubscriptionDetail, error) {
	var resp apiResponse[SubscriptionDetail]
	if err := c.http.Get(ctx, "/subscription/detail", nil, &resp); err != nil {
		return nil, err
	}
	if err := resp.validate(); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
