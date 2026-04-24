package platform

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0xCyberFred/zenmux-sdk-go/internal/httpclient"
)

// newTestClient creates a Platform client that points at the given test server.
func newTestClient(url string) *Client {
	return NewClient(url, "test-mgmt-key", nil)
}

func TestGetFlowRate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/flow_rate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-mgmt-key" {
			t.Errorf("unexpected auth header: %s", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("unexpected accept header: %s", got)
		}

		resp := apiResponse[FlowRate]{
			Success: true,
			Data: FlowRate{
				Currency:            "USD",
				BaseUSDPerFlow:      0.01,
				EffectiveUSDPerFlow: 0.008,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetFlowRate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", result.Currency)
	}
	if result.BaseUSDPerFlow != 0.01 {
		t.Errorf("expected base rate 0.01, got %f", result.BaseUSDPerFlow)
	}
	if result.EffectiveUSDPerFlow != 0.008 {
		t.Errorf("expected effective rate 0.008, got %f", result.EffectiveUSDPerFlow)
	}
}

func TestGetPAYGBalance(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/balance" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := apiResponse[PAYGBalance]{
			Success: true,
			Data: PAYGBalance{
				Currency:     "USD",
				TotalCredits: 150.50,
				TopUpCredits: 100.00,
				BonusCredits: 50.50,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetPAYGBalance(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", result.Currency)
	}
	if result.TotalCredits != 150.50 {
		t.Errorf("expected total credits 150.50, got %f", result.TotalCredits)
	}
	if result.TopUpCredits != 100.00 {
		t.Errorf("expected top-up credits 100.00, got %f", result.TopUpCredits)
	}
	if result.BonusCredits != 50.50 {
		t.Errorf("expected bonus credits 50.50, got %f", result.BonusCredits)
	}
}

func TestGetSubscription(t *testing.T) {
	resetTime := "2026-04-25T00:00:00Z"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscription/detail" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := apiResponse[SubscriptionDetail]{
			Success: true,
			Data: SubscriptionDetail{
				Plan: Plan{
					Tier:      "pro",
					AmountUSD: 49.99,
					Interval:  "monthly",
					ExpiresAt: "2026-05-24T00:00:00Z",
				},
				Currency:            "USD",
				BaseUSDPerFlow:      0.01,
				EffectiveUSDPerFlow: 0.008,
				AccountStatus:       "active",
				Quota5Hour: Quota{
					UsagePercentage: 25.0,
					ResetsAt:        &resetTime,
					MaxFlows:        1000,
					UsedFlows:       250,
					RemainingFlows:  750,
					UsedValueUSD:    2.50,
					MaxValueUSD:     10.00,
				},
				Quota7Day: Quota{
					UsagePercentage: 10.0,
					ResetsAt:        nil,
					MaxFlows:        5000,
					UsedFlows:       500,
					RemainingFlows:  4500,
					UsedValueUSD:    5.00,
					MaxValueUSD:     50.00,
				},
				QuotaMonthly: QuotaMonthly{
					MaxFlows:    20000,
					MaxValueUSD: 200.00,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetSubscription(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify nested Plan struct
	if result.Plan.Tier != "pro" {
		t.Errorf("expected plan tier pro, got %s", result.Plan.Tier)
	}
	if result.Plan.AmountUSD != 49.99 {
		t.Errorf("expected plan amount 49.99, got %f", result.Plan.AmountUSD)
	}
	if result.Plan.Interval != "monthly" {
		t.Errorf("expected plan interval monthly, got %s", result.Plan.Interval)
	}

	// Verify account status
	if result.AccountStatus != "active" {
		t.Errorf("expected account status active, got %s", result.AccountStatus)
	}

	// Verify nested Quota5Hour with non-nil ResetsAt
	if result.Quota5Hour.UsagePercentage != 25.0 {
		t.Errorf("expected 5h usage 25.0, got %f", result.Quota5Hour.UsagePercentage)
	}
	if result.Quota5Hour.ResetsAt == nil {
		t.Fatal("expected Quota5Hour.ResetsAt to be non-nil")
	}
	if *result.Quota5Hour.ResetsAt != resetTime {
		t.Errorf("expected resets_at %s, got %s", resetTime, *result.Quota5Hour.ResetsAt)
	}
	if result.Quota5Hour.RemainingFlows != 750 {
		t.Errorf("expected remaining flows 750, got %f", result.Quota5Hour.RemainingFlows)
	}

	// Verify Quota7Day with nil ResetsAt
	if result.Quota7Day.ResetsAt != nil {
		t.Errorf("expected Quota7Day.ResetsAt to be nil, got %s", *result.Quota7Day.ResetsAt)
	}

	// Verify QuotaMonthly
	if result.QuotaMonthly.MaxFlows != 20000 {
		t.Errorf("expected monthly max flows 20000, got %f", result.QuotaMonthly.MaxFlows)
	}
	if result.QuotaMonthly.MaxValueUSD != 200.00 {
		t.Errorf("expected monthly max value 200.00, got %f", result.QuotaMonthly.MaxValueUSD)
	}
}

func TestGetGeneration(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/generation" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameter
		if got := r.URL.Query().Get("id"); got != "gen-abc-123" {
			t.Errorf("expected query param id=gen-abc-123, got %s", got)
		}

		// Generation response is NOT wrapped in apiResponse
		resp := Generation{
			API:            "chat/completions",
			GenerationID:   "gen-abc-123",
			Model:          "openai/gpt-4.1",
			CreateAt:       "2026-04-24T10:00:00Z",
			GenerationTime: 1500,
			Latency:        200,
			NativeTokens: TokenUsage{
				CompletionTokens: 100,
				PromptTokens:     50,
				TotalTokens:      150,
				CompletionTokensDetails: CompletionTokensDetails{
					ReasoningTokens: 20,
				},
				PromptTokensDetails: PromptTokensDetails{
					CachedTokens: 10,
				},
			},
			Streamed:     true,
			FinishReason: "stop",
			Usage:        0.5,
			RatingResponses: &RatingResponses{
				BillAmount:     0.003,
				DiscountAmount: 0.001,
				OriginAmount:   0.004,
				PriceVersion:   "v2",
				RatingDetails: []RatingDetail{
					{
						BillAmount:     0.002,
						DiscountAmount: 0.001,
						FeeItemCode:    "input",
						OriginAmount:   0.003,
						Rate:           0.00001,
					},
				},
			},
			RequestRetryTimes: 0,
			FinalRetry:        false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetGeneration(context.Background(), "gen-abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GenerationID != "gen-abc-123" {
		t.Errorf("expected generation ID gen-abc-123, got %s", result.GenerationID)
	}
	if result.Model != "openai/gpt-4.1" {
		t.Errorf("expected model openai/gpt-4.1, got %s", result.Model)
	}
	if !result.Streamed {
		t.Error("expected streamed to be true")
	}

	// Verify nested TokenUsage
	if result.NativeTokens.TotalTokens != 150 {
		t.Errorf("expected total tokens 150, got %d", result.NativeTokens.TotalTokens)
	}
	if result.NativeTokens.CompletionTokensDetails.ReasoningTokens != 20 {
		t.Errorf("expected reasoning tokens 20, got %d", result.NativeTokens.CompletionTokensDetails.ReasoningTokens)
	}
	if result.NativeTokens.PromptTokensDetails.CachedTokens != 10 {
		t.Errorf("expected cached tokens 10, got %d", result.NativeTokens.PromptTokensDetails.CachedTokens)
	}

	// Verify RatingResponses
	if result.RatingResponses == nil {
		t.Fatal("expected RatingResponses to be non-nil")
	}
	if result.RatingResponses.BillAmount != 0.003 {
		t.Errorf("expected bill amount 0.003, got %f", result.RatingResponses.BillAmount)
	}
	if len(result.RatingResponses.RatingDetails) != 1 {
		t.Fatalf("expected 1 rating detail, got %d", len(result.RatingResponses.RatingDetails))
	}
	if result.RatingResponses.RatingDetails[0].FeeItemCode != "input" {
		t.Errorf("expected fee item code input, got %s", result.RatingResponses.RatingDetails[0].FeeItemCode)
	}
}

func TestGetTimeseries(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/statistics/timeseries" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify all query params
		q := r.URL.Query()
		if got := q.Get("metric"); got != "tokens" {
			t.Errorf("expected metric=tokens, got %s", got)
		}
		if got := q.Get("bucket_width"); got != "1d" {
			t.Errorf("expected bucket_width=1d, got %s", got)
		}
		if got := q.Get("starting_at"); got != "2026-04-01" {
			t.Errorf("expected starting_at=2026-04-01, got %s", got)
		}
		if got := q.Get("ending_at"); got != "2026-04-24" {
			t.Errorf("expected ending_at=2026-04-24, got %s", got)
		}
		if got := q.Get("limit"); got != "10" {
			t.Errorf("expected limit=10, got %s", got)
		}

		resp := apiResponse[Timeseries]{
			Success: true,
			Data: Timeseries{
				Metric:       "tokens",
				BucketWidth:  "1d",
				StartingAt:   "2026-04-01",
				EndingAt:     "2026-04-24",
				TotalBuckets: 1,
				Series: []TimeseriesBucket{
					{
						Period: "2026-04-01",
						Date:   "2026-04-01T00:00:00Z",
						Models: []ModelMetric{
							{Model: "gpt-4.1", Label: "GPT-4.1", Value: 50000},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetTimeseries(context.Background(), TimeseriesParams{
		Metric:      "tokens",
		BucketWidth: "1d",
		StartingAt:  "2026-04-01",
		EndingAt:    "2026-04-24",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Metric != "tokens" {
		t.Errorf("expected metric tokens, got %s", result.Metric)
	}
	if result.TotalBuckets != 1 {
		t.Errorf("expected 1 bucket, got %d", result.TotalBuckets)
	}
	if len(result.Series) != 1 {
		t.Fatalf("expected 1 series entry, got %d", len(result.Series))
	}
	if len(result.Series[0].Models) != 1 {
		t.Fatalf("expected 1 model metric, got %d", len(result.Series[0].Models))
	}
	if result.Series[0].Models[0].Value != 50000 {
		t.Errorf("expected value 50000, got %f", result.Series[0].Models[0].Value)
	}
}

func TestGetLeaderboard(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/statistics/leaderboard" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		q := r.URL.Query()
		if got := q.Get("metric"); got != "requests" {
			t.Errorf("expected metric=requests, got %s", got)
		}

		resp := apiResponse[Leaderboard]{
			Success: true,
			Data: Leaderboard{
				Metric:     "requests",
				StartingAt: "2026-04-01",
				EndingAt:   "2026-04-24",
				Entries: []LeaderboardEntry{
					{
						Rank:        1,
						Model:       "gpt-4.1",
						Label:       "GPT-4.1",
						Author:      "openai",
						AuthorLabel: "OpenAI",
						Value:       10000,
					},
					{
						Rank:        2,
						Model:       "claude-opus-4",
						Label:       "Claude Opus 4",
						Author:      "anthropic",
						AuthorLabel: "Anthropic",
						Value:       8000,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetLeaderboard(context.Background(), LeaderboardParams{
		Metric:     "requests",
		StartingAt: "2026-04-01",
		EndingAt:   "2026-04-24",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	if result.Entries[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", result.Entries[0].Rank)
	}
	if result.Entries[0].Model != "gpt-4.1" {
		t.Errorf("expected model gpt-4.1, got %s", result.Entries[0].Model)
	}
	if result.Entries[1].Author != "anthropic" {
		t.Errorf("expected author anthropic, got %s", result.Entries[1].Author)
	}
}

func TestGetMarketShare(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/statistics/market_share" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := apiResponse[MarketShare]{
			Success: true,
			Data: MarketShare{
				Metric:       "tokens",
				BucketWidth:  "1w",
				StartingAt:   "2026-04-01",
				EndingAt:     "2026-04-24",
				TotalBuckets: 1,
				Series: []MarketShareBucket{
					{
						Period: "2026-04-01",
						Date:   "2026-04-01T00:00:00Z",
						Authors: []AuthorMetric{
							{Author: "openai", Label: "OpenAI", Value: 0.60},
							{Author: "anthropic", Label: "Anthropic", Value: 0.40},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetMarketShare(context.Background(), MarketShareParams{
		Metric:      "tokens",
		BucketWidth: "1w",
		StartingAt:  "2026-04-01",
		EndingAt:    "2026-04-24",
		Limit:       5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Series) != 1 {
		t.Fatalf("expected 1 series entry, got %d", len(result.Series))
	}
	if len(result.Series[0].Authors) != 2 {
		t.Fatalf("expected 2 authors, got %d", len(result.Series[0].Authors))
	}
	if result.Series[0].Authors[0].Value != 0.60 {
		t.Errorf("expected value 0.60, got %f", result.Series[0].Authors[0].Value)
	}
}

func TestHTTPError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid management key"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetFlowRate(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}

	var httpErr *httpclient.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *httpclient.HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", httpErr.StatusCode)
	}
	if httpErr.Body != `{"error":"invalid management key"}` {
		t.Errorf("unexpected error body: %s", httpErr.Body)
	}
}

func TestHTTPErrorServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetPAYGBalance(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	var httpErr *httpclient.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *httpclient.HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", httpErr.StatusCode)
	}
}

func TestGetTimeseriesOptionalParams(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("metric"); got != "cost" {
			t.Errorf("expected metric=cost, got %s", got)
		}
		// Optional params should be absent
		if q.Has("bucket_width") {
			t.Error("expected bucket_width to be absent")
		}
		if q.Has("starting_at") {
			t.Error("expected starting_at to be absent")
		}
		if q.Has("ending_at") {
			t.Error("expected ending_at to be absent")
		}
		if q.Has("limit") {
			t.Error("expected limit to be absent")
		}

		resp := apiResponse[Timeseries]{
			Success: true,
			Data:    Timeseries{Metric: "cost"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetTimeseries(context.Background(), TimeseriesParams{
		Metric: "cost",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Metric != "cost" {
		t.Errorf("expected metric cost, got %s", result.Metric)
	}
}

func TestGetGenerationNilRatingResponses(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Generation{
			GenerationID:    "gen-no-rating",
			Model:           "test-model",
			RatingResponses: nil,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(server.URL)
	result, err := client.GetGeneration(context.Background(), "gen-no-rating")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RatingResponses != nil {
		t.Error("expected RatingResponses to be nil")
	}
}
