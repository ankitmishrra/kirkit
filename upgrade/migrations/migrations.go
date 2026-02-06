package migrations

import (
	"context"
	"database/sql"
)

// Migration represents a single versioned migration.
type Migration struct {
	Version int
	Name    string
	Up      func(ctx context.Context, db *sql.DB) error
}

// All migrations in order. Version must be strictly increasing.
var All = []Migration{
	{Version: 1, Name: "initial_schema", Up: Up001InitialSchema},
	{Version: 2, Name: "initial_tournament", Up: Up002InitialTournament},
	{Version: 3, Name: "rename_match_to_game", Up: Up003RenameMatchToGame},
	{Version: 4, Name: "t20_worldcup_matches", Up: Up004T20WorldCupMatches},
	{Version: 5, Name: "backfill_series_matches", Up: Up005BackfillSeriesMatches},
	{Version: 6, Name: "add_scorecard_column", Up: Up006AddScorecardColumn},
}
