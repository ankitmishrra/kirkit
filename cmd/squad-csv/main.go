// One-off: fetch T20 World Cup squad from CricAPI series_squad and print CSV (name, nationality).
// Usage: from repo root, load .env then run:
//
//	go run ./cmd/squad-csv
//
// Or: CRICAPI_KEY=yourkey go run ./cmd/squad-csv
//
// Writes to stdout. Redirect to file: go run ./cmd/squad-csv > t20_worldcup_squad.csv
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"kirkit/sdk/cricapi"
)

const t20WorldCupSeriesID = "0cdf6736-ad9b-4e95-a647-5ee3a99c5510"

func main() {
	apiKey := os.Getenv("CRICAPI_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "set CRICAPI_KEY (e.g. in .env or export)")
		os.Exit(1)
	}
	client := cricapi.NewClient(apiKey, os.Getenv("CRICAPI_BASE"))
	ctx := context.Background()
	resp, err := client.SeriesSquad(ctx, t20WorldCupSeriesID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SeriesSquad: %v\n", err)
		os.Exit(1)
	}
	if resp.Data == nil || len(resp.Data.Squad) == 0 {
		fmt.Fprintln(os.Stderr, "no squad data in response (check API response shape)")
		os.Exit(1)
	}
	w := csv.NewWriter(os.Stdout)
	_ = w.Write([]string{"name", "nationality", "team"})
	for _, team := range resp.Data.Squad {
		teamName := team.Name
		if teamName == "" {
			teamName = team.Country
		}
		for _, p := range team.Players {
			nat := p.NationalityOrCountry(teamName)
			name := strings.TrimSpace(p.Name)
			if name == "" {
				continue
			}
			_ = w.Write([]string{name, nat, teamName})
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
