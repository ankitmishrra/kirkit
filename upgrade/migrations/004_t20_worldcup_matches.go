package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"

	"kirkit/sdk/cricapi"
)

// seriesIDForT20Matches is the series used to fetch match list from CricAPI series_info.
// Example: https://api.cricapi.com/v1/series_info?apikey=...&id=<series_id>
const seriesIDForT20Matches = "0cdf6736-ad9b-4e95-a647-5ee3a99c5510"

// Up004T20WorldCupMatches fetches match list from CricAPI series_info for the given series
// and inserts them into the game table. If CRICAPI_KEY is not set, the migration is a no-op.
func Up004T20WorldCupMatches(ctx context.Context, db *sql.DB) error {
	apiKey := os.Getenv("CRICAPI_KEY")
	if apiKey == "" {
		return nil // no-op when key not set (e.g. CI); matches can be filled later
	}

	client := cricapi.NewClient(apiKey, os.Getenv("CRICAPI_BASE"))

	stmt, err := db.PrepareContext(ctx,
		`INSERT INTO game (series_id, match_id, match_info) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE match_info = VALUES(match_info), updated_at = CURRENT_TIMESTAMP`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	const pageSize = 20
	for offset := 0; ; offset += pageSize {
		resp, err := client.SeriesInfo(ctx, seriesIDForT20Matches, offset)
		if err != nil {
			return err
		}
		if resp.Data == nil {
			break
		}
		// Ensure tournament exists for this series (so it appears in app and matches can be listed).
		if offset == 0 {
			seriesName := "T20 Series"
			if resp.Data.Info != nil && resp.Data.Info.Name != "" {
				seriesName = resp.Data.Info.Name
			}
			_, err = db.ExecContext(ctx,
				`INSERT INTO tournament (series_id, name, status) VALUES (?, ?, ?)
				 ON DUPLICATE KEY UPDATE name = VALUES(name), status = VALUES(status)`,
				seriesIDForT20Matches, seriesName, 1)
			if err != nil {
				return err
			}
		}
		list := resp.Data.MatchListOrMatches()
		if len(list) == 0 {
			break
		}
		for _, m := range list {
			infoJSON, err := json.Marshal(m)
			if err != nil {
				return err
			}
			_, err = stmt.ExecContext(ctx, seriesIDForT20Matches, m.ID, infoJSON)
			if err != nil {
				return err
			}
		}
		if len(list) < pageSize {
			break
		}
	}
	return nil
}
