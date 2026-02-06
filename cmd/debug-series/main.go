// Package main prints CricAPI series_info response for a given series ID (debug).
// Usage: CRICAPI_KEY=yourkey go run ./cmd/debug-series [series_id]
// Default series_id: 0cdf6736-ad9b-4e95-a647-5ee3a99c5510
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"kirkit/sdk/cricapi"
)

func main() {
	apiKey := os.Getenv("CRICAPI_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "set CRICAPI_KEY")
		os.Exit(1)
	}
	seriesID := "0cdf6736-ad9b-4e95-a647-5ee3a99c5510"
	if len(os.Args) > 1 {
		seriesID = os.Args[1]
	}
	client := cricapi.NewClient(apiKey, os.Getenv("CRICAPI_BASE"))
	ctx := context.Background()
	resp, err := client.SeriesInfo(ctx, seriesID, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SeriesInfo: %v\n", err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	matches := 0
	if resp.Data != nil {
		matches = len(resp.Data.MatchListOrMatches())
	}
	fmt.Fprintf(os.Stderr, "status=%s matches=%d (use offset 20,40,... for more)\n", resp.Status, matches)
}
