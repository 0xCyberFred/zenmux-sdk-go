package main

import (
	"context"
	"fmt"
	"log"
	"os"

	zenmux "github.com/0xCyberFred/zenmux-sdk-go"
)

func main() {
	client := zenmux.NewClient(os.Getenv("ZENMUX_API_KEY"))

	for _, provider := range []zenmux.Provider{zenmux.ProviderOpenAI, zenmux.ProviderAnthropic, zenmux.ProviderGoogle} {
		fmt.Printf("=== %s ===\n", provider)

		result, err := client.Models.List(context.Background(), provider)
		if err != nil {
			log.Printf("  error: %v", err)
			continue
		}

		for _, m := range result.Models {
			fmt.Printf("  %-40s context=%d reasoning=%v\n", m.ID, m.ContextLength, m.Reasoning)
		}
		fmt.Println()
	}
}
