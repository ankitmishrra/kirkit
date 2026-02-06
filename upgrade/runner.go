package upgrade

import (
	"context"
	"database/sql"
	"fmt"

	"kirkit/dl/mysql"
	"kirkit/upgrade/migrations"
)

// Runner runs Go migrations in order. Version management:
// - A single row in the `version` table holds the current schema version (integer).
// - Only migrations with Version > current are run.
// - After each migration, the version table is updated to that migration's Version.
// - Migrations are ordered by Version (1, 2, 3, ...).
func Runner(ctx context.Context, db *sql.DB) error {
	if err := ensureVersionTable(ctx, db); err != nil {
		return fmt.Errorf("ensure version table: %w", err)
	}

	versionRepo := mysql.NewVersionRepo(db)
	current, err := versionRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	for _, m := range migrations.All {
		if m.Version <= current {
			continue
		}
		if err := m.Up(ctx, db); err != nil {
			return fmt.Errorf("migration %d %s: %w", m.Version, m.Name, err)
		}
		if err := versionRepo.Set(ctx, m.Version); err != nil {
			return fmt.Errorf("set version to %d: %w", m.Version, err)
		}
		current = m.Version
	}

	return nil
}

// ensureVersionTable creates the version table and inserts row (id=1, version=0) if not present.
// DDL uses constant strings only (no user input).
func ensureVersionTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS version (
		id INT PRIMARY KEY DEFAULT 1,
		version INT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		CONSTRAINT single_row CHECK (id = 1)
	)`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `INSERT INTO version (id, version) VALUES (1, 0) ON DUPLICATE KEY UPDATE id = id`)
	return err
}
