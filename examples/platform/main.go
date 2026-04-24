package main

import (
	"context"
	"fmt"
	"log"
	"os"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
	"github.com/0xCyberFred/zenmux-sdk-go/platform"
)

func main() {
	client := zenmux.NewClient(
		os.Getenv("ZENMUX_API_KEY"),
		zenmux.WithManagementKey(os.Getenv("ZENMUX_MANAGEMENT_KEY")),
	)

	if client.Platform == nil {
		log.Fatal("ZENMUX_MANAGEMENT_KEY is required for Platform API")
	}

	ctx := context.Background()

	// Flow Rate
	rate, err := client.Platform.GetFlowRate(ctx)
	if err != nil {
		log.Printf("FlowRate error: %v", err)
	} else {
		fmt.Printf("Flow Rate: $%.5f/flow (effective: $%.5f)\n", rate.BaseUSDPerFlow, rate.EffectiveUSDPerFlow)
	}

	// PAYG Balance
	balance, err := client.Platform.GetPAYGBalance(ctx)
	if err != nil {
		log.Printf("Balance error: %v", err)
	} else {
		fmt.Printf("Balance: $%.2f (top-up: $%.2f, bonus: $%.2f)\n",
			balance.TotalCredits, balance.TopUpCredits, balance.BonusCredits)
	}

	// Subscription
	sub, err := client.Platform.GetSubscription(ctx)
	if err != nil {
		log.Printf("Subscription error: %v", err)
	} else {
		fmt.Printf("Plan: %s ($%.2f/%s) Status: %s\n",
			sub.Plan.Tier, sub.Plan.AmountUSD, sub.Plan.Interval, sub.AccountStatus)
		fmt.Printf("5h quota: %.0f/%.0f flows (%.1f%%)\n",
			sub.Quota5Hour.UsedFlows, sub.Quota5Hour.MaxFlows, sub.Quota5Hour.UsagePercentage*100)
	}

	// Statistics Leaderboard
	lb, err := client.Platform.GetLeaderboard(ctx, platform.LeaderboardParams{
		Metric: "tokens",
		Limit:  5,
	})
	if err != nil {
		log.Printf("Leaderboard error: %v", err)
	} else {
		fmt.Printf("\nTop %d models by tokens:\n", len(lb.Entries))
		for _, e := range lb.Entries {
			fmt.Printf("  #%d %-35s %.0f tokens\n", e.Rank, e.Label, e.Value)
		}
	}
}
