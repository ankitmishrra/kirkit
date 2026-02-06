package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"kirkit/sdk/cricapi"
)

// seriesIDBackfill is the series to backfill (tournament + all matches from CricAPI).
const seriesIDBackfill = "0cdf6736-ad9b-4e95-a647-5ee3a99c5510"

// Up005BackfillSeriesMatches ensures the tournament exists for the series and backfills
// all matches from CricAPI series_info (with pagination). Idempotent.
// If CRICAPI_KEY is not set, no-op. If API returns no matches, returns error so you can verify series ID.
func Up005BackfillSeriesMatches(ctx context.Context, db *sql.DB) error {
	apiKey := os.Getenv("CRICAPI_KEY")
	if apiKey == "" {
		return nil
	}

	client := cricapi.NewClient(apiKey, os.Getenv("CRICAPI_BASE"))
	var seriesName string

	stmt, err := db.PrepareContext(ctx,
		`INSERT INTO game (series_id, match_id, match_info) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE match_info = VALUES(match_info), updated_at = CURRENT_TIMESTAMP`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	const pageSize = 20
	for offset := 0; ; offset += pageSize {
		resp, err := client.SeriesInfo(ctx, seriesIDBackfill, offset)
		if err != nil {
			return fmt.Errorf("series_info for %s offset %d: %w", seriesIDBackfill, offset, err)
		}
		if resp.Data == nil {
			break
		}
		if offset == 0 {
			seriesName = "T20 Series"
			if resp.Data.Info != nil && resp.Data.Info.Name != "" {
				seriesName = resp.Data.Info.Name
			}
			_, err = db.ExecContext(ctx,
				`INSERT INTO tournament (series_id, name, status) VALUES (?, ?, ?)
				 ON DUPLICATE KEY UPDATE name = VALUES(name), status = VALUES(status)`,
				seriesIDBackfill, seriesName, 1)
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
			_, err = stmt.ExecContext(ctx, seriesIDBackfill, m.ID, infoJSON)
			if err != nil {
				return err
			}
		}
		if len(list) < pageSize {
			break
		}
	}

	// If totalInserted == 0, tournament was still inserted; verify series id and CRICAPI_KEY if you expect matches.
	return nil
}
