package platform

// apiResponse is the standard wrapper for most Platform API responses.
type apiResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

// FlowRate holds pricing information for a single flow unit.
type FlowRate struct {
	Currency            string  `json:"currency"`
	BaseUSDPerFlow      float64 `json:"base_usd_per_flow"`
	EffectiveUSDPerFlow float64 `json:"effective_usd_per_flow"`
}

// PAYGBalance holds the pay-as-you-go credit balance breakdown.
type PAYGBalance struct {
	Currency     string  `json:"currency"`
	TotalCredits float64 `json:"total_credits"`
	TopUpCredits float64 `json:"top_up_credits"`
	BonusCredits float64 `json:"bonus_credits"`
}

// SubscriptionDetail holds the full subscription state including plan, quotas,
// and pricing.
type SubscriptionDetail struct {
	Plan                Plan         `json:"plan"`
	Currency            string       `json:"currency"`
	BaseUSDPerFlow      float64      `json:"base_usd_per_flow"`
	EffectiveUSDPerFlow float64      `json:"effective_usd_per_flow"`
	AccountStatus       string       `json:"account_status"`
	Quota5Hour          Quota        `json:"quota_5_hour"`
	Quota7Day           Quota        `json:"quota_7_day"`
	QuotaMonthly        QuotaMonthly `json:"quota_monthly"`
}

// Plan describes the subscription tier and billing interval.
type Plan struct {
	Tier      string  `json:"tier"`
	AmountUSD float64 `json:"amount_usd"`
	Interval  string  `json:"interval"`
	ExpiresAt string  `json:"expires_at"`
}

// Quota holds usage information for a rolling time window.
type Quota struct {
	UsagePercentage float64 `json:"usage_percentage"`
	ResetsAt        *string `json:"resets_at"`
	MaxFlows        float64 `json:"max_flows"`
	UsedFlows       float64 `json:"used_flows"`
	RemainingFlows  float64 `json:"remaining_flows"`
	UsedValueUSD    float64 `json:"used_value_usd"`
	MaxValueUSD     float64 `json:"max_value_usd"`
}

// QuotaMonthly holds the monthly quota limits.
type QuotaMonthly struct {
	MaxFlows    float64 `json:"max_flows"`
	MaxValueUSD float64 `json:"max_value_usd"`
}

// Generation holds details about a single API generation request.
type Generation struct {
	API               string           `json:"api"`
	GenerationID      string           `json:"generationId"`
	Model             string           `json:"model"`
	CreateAt          string           `json:"createAt"`
	GenerationTime    int              `json:"generationTime"`
	Latency           int              `json:"latency"`
	NativeTokens      TokenUsage       `json:"nativeTokens"`
	Streamed          bool             `json:"streamed"`
	FinishReason      string           `json:"finishReason"`
	Usage             float64          `json:"usage"`
	RatingResponses   *RatingResponses `json:"ratingResponses"`
	RequestRetryTimes int              `json:"requestRetryTimes"`
	FinalRetry        bool             `json:"finalRetry"`
}

// TokenUsage contains prompt and completion token counts with detailed
// breakdowns.
type TokenUsage struct {
	CompletionTokens        int                     `json:"completion_tokens"`
	PromptTokens            int                     `json:"prompt_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
	PromptTokensDetails     PromptTokensDetails     `json:"prompt_tokens_details"`
}

// CompletionTokensDetails breaks down completion tokens.
type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// PromptTokensDetails breaks down prompt tokens.
type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

// RatingResponses holds billing and discount information for a generation.
type RatingResponses struct {
	BillAmount     float64        `json:"billAmount"`
	DiscountAmount float64        `json:"discountAmount"`
	OriginAmount   float64        `json:"originAmount"`
	PriceVersion   string         `json:"priceVersion"`
	RatingDetails  []RatingDetail `json:"ratingDetails"`
}

// RatingDetail holds a single line-item in the billing breakdown.
type RatingDetail struct {
	BillAmount     float64 `json:"billAmount"`
	DiscountAmount float64 `json:"discountAmount"`
	FeeItemCode    string  `json:"feeItemCode"`
	OriginAmount   float64 `json:"originAmount"`
	Rate           float64 `json:"rate"`
}

// TimeseriesParams holds query parameters for the GetTimeseries endpoint.
type TimeseriesParams struct {
	Metric      string
	BucketWidth string
	StartingAt  string
	EndingAt    string
	Limit       int
}

// Timeseries is the response from the timeseries statistics endpoint.
type Timeseries struct {
	Metric       string             `json:"metric"`
	BucketWidth  string             `json:"bucket_width"`
	StartingAt   string             `json:"starting_at"`
	EndingAt     string             `json:"ending_at"`
	TotalBuckets int                `json:"total_buckets"`
	Series       []TimeseriesBucket `json:"series"`
}

// TimeseriesBucket is a single bucket in a timeseries.
type TimeseriesBucket struct {
	Period string        `json:"period"`
	Date   string        `json:"date"`
	Models []ModelMetric `json:"models"`
}

// ModelMetric holds a metric value for a single model within a timeseries
// bucket.
type ModelMetric struct {
	Model string  `json:"model"`
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// LeaderboardParams holds query parameters for the GetLeaderboard endpoint.
type LeaderboardParams struct {
	Metric     string
	StartingAt string
	EndingAt   string
	Limit      int
}

// Leaderboard is the response from the leaderboard statistics endpoint.
type Leaderboard struct {
	Metric     string             `json:"metric"`
	StartingAt string             `json:"starting_at"`
	EndingAt   string             `json:"ending_at"`
	Entries    []LeaderboardEntry `json:"entries"`
}

// LeaderboardEntry is a single entry in the leaderboard ranking.
type LeaderboardEntry struct {
	Rank        int     `json:"rank"`
	Model       string  `json:"model"`
	Label       string  `json:"label"`
	Author      string  `json:"author"`
	AuthorLabel string  `json:"author_label"`
	Value       float64 `json:"value"`
}

// MarketShareParams holds query parameters for the GetMarketShare endpoint.
type MarketShareParams struct {
	Metric      string
	BucketWidth string
	StartingAt  string
	EndingAt    string
	Limit       int
}

// MarketShare is the response from the market share statistics endpoint.
type MarketShare struct {
	Metric       string              `json:"metric"`
	BucketWidth  string              `json:"bucket_width"`
	StartingAt   string              `json:"starting_at"`
	EndingAt     string              `json:"ending_at"`
	TotalBuckets int                 `json:"total_buckets"`
	Series       []MarketShareBucket `json:"series"`
}

// MarketShareBucket is a single bucket in a market share timeseries.
type MarketShareBucket struct {
	Period  string         `json:"period"`
	Date    string         `json:"date"`
	Authors []AuthorMetric `json:"authors"`
}

// AuthorMetric holds a metric value for a single author within a market share
// bucket.
type AuthorMetric struct {
	Author string  `json:"author"`
	Label  string  `json:"label"`
	Value  float64 `json:"value"`
}
