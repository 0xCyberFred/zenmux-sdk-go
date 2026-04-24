package platform

import "context"

// GetSubscription returns the current subscription details including plan,
// quotas, and pricing information.
func (c *Client) GetSubscription(ctx context.Context) (*SubscriptionDetail, error) {
	var resp apiResponse[SubscriptionDetail]
	if err := c.http.Get(ctx, "/subscription", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
