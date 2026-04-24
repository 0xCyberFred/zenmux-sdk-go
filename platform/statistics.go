package platform

import (
	"context"
	"net/url"
	"strconv"
)

// GetTimeseries returns time-bucketed metrics for the account.
func (c *Client) GetTimeseries(ctx context.Context, params TimeseriesParams) (*Timeseries, error) {
	q := url.Values{}
	q.Set("metric", params.Metric)
	if params.BucketWidth != "" {
		q.Set("bucket_width", params.BucketWidth)
	}
	if params.StartingAt != "" {
		q.Set("starting_at", params.StartingAt)
	}
	if params.EndingAt != "" {
		q.Set("ending_at", params.EndingAt)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}

	var resp apiResponse[Timeseries]
	if err := c.http.Get(ctx, "/statistics/timeseries", q, &resp); err != nil {
		return nil, err
	}
	if err := resp.validate(); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetLeaderboard returns a ranked list of models by the given metric.
func (c *Client) GetLeaderboard(ctx context.Context, params LeaderboardParams) (*Leaderboard, error) {
	q := url.Values{}
	q.Set("metric", params.Metric)
	if params.StartingAt != "" {
		q.Set("starting_at", params.StartingAt)
	}
	if params.EndingAt != "" {
		q.Set("ending_at", params.EndingAt)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}

	var resp apiResponse[Leaderboard]
	if err := c.http.Get(ctx, "/statistics/leaderboard", q, &resp); err != nil {
		return nil, err
	}
	if err := resp.validate(); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetMarketShare returns time-bucketed market share data by author.
func (c *Client) GetMarketShare(ctx context.Context, params MarketShareParams) (*MarketShare, error) {
	q := url.Values{}
	q.Set("metric", params.Metric)
	if params.BucketWidth != "" {
		q.Set("bucket_width", params.BucketWidth)
	}
	if params.StartingAt != "" {
		q.Set("starting_at", params.StartingAt)
	}
	if params.EndingAt != "" {
		q.Set("ending_at", params.EndingAt)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}

	var resp apiResponse[MarketShare]
	if err := c.http.Get(ctx, "/statistics/market_share", q, &resp); err != nil {
		return nil, err
	}
	if err := resp.validate(); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
