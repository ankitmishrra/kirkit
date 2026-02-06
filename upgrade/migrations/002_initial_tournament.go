package migrations

import (
	"context"
	"database/sql"
)

const defaultT20WorldCupSeriesID = "c4ca5cd5-e25c-4d83-bb77-2d193d93475a"

// Up002InitialTournament inserts the default Men's T20 World Cup tournament (parameterized Exec).
func Up002InitialTournament(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO tournament (series_id, name, status) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE name = VALUES(name), status = VALUES(status)`,
		defaultT20WorldCupSeriesID, "Mens T20 World Cup", 1)
	return err
}
