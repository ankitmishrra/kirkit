package migrations

import (
	"context"
	"database/sql"
)

// Up006AddScorecardColumn adds a nullable scorecard JSON column to game.
// Scorecard is populated after the match (e.g. from CricAPI match_scorecard) for leaderboard computation.
func Up006AddScorecardColumn(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `ALTER TABLE game ADD COLUMN scorecard JSON NULL AFTER match_info`)
	return err
}
