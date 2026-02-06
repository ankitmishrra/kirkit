package migrations

import (
	"context"
	"database/sql"
	"strings"
)

// Up003RenameMatchToGame renames table `match` to game (MySQL reserved keyword fix).
// No-op if `match` does not exist (e.g. fresh install where 001 already creates game).
func Up003RenameMatchToGame(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "RENAME TABLE `match` TO game")
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "Unknown table") {
		return nil
	}
	return err
}
