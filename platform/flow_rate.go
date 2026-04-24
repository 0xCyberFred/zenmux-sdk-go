package platform

import "context"

// GetFlowRate returns the current flow rate pricing.
func (c *Client) GetFlowRate(ctx context.Context) (*FlowRate, error) {
	var resp apiResponse[FlowRate]
	if err := c.http.Get(ctx, "/flow_rate", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
