package migrations

import (
	"context"
	"database/sql"
)

// Up001InitialSchema creates the version table and full schema (tournament, fantasy_league, squad, game, snapshot, leaderboard).
// DDL uses constant strings only; no user input (injection-safe).
func Up001InitialSchema(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS version (
			id INT PRIMARY KEY DEFAULT 1,
			version INT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CONSTRAINT single_row CHECK (id = 1)
		)`,
		`INSERT INTO version (id, version) VALUES (1, 0) ON DUPLICATE KEY UPDATE id = id`,
		`CREATE TABLE IF NOT EXISTS tournament (
			id INT AUTO_INCREMENT PRIMARY KEY,
			series_id VARCHAR(64) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			status TINYINT NOT NULL DEFAULT 1 COMMENT '1=ongoing, 2=done',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS fantasy_league (
			id INT AUTO_INCREMENT PRIMARY KEY,
			series_id VARCHAR(64) NOT NULL,
			team_name VARCHAR(255) NOT NULL,
			team_owner VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_series_team (series_id, team_name)
		)`,
		`CREATE TABLE IF NOT EXISTS squad (
			id INT AUTO_INCREMENT PRIMARY KEY,
			team_id INT NOT NULL,
			player_name VARCHAR(255) NOT NULL,
			is_captain TINYINT(1) NOT NULL DEFAULT 0,
			is_vice_captain TINYINT(1) NOT NULL DEFAULT 0,
			player_type VARCHAR(64) DEFAULT NULL COMMENT 'e.g. batsman, bowler, allrounder',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (team_id) REFERENCES fantasy_league(id) ON DELETE CASCADE,
			UNIQUE KEY uk_team_player (team_id, player_name)
		)`,
		`CREATE TABLE IF NOT EXISTS game (
			id INT AUTO_INCREMENT PRIMARY KEY,
			series_id VARCHAR(64) NOT NULL,
			match_id VARCHAR(64) NOT NULL,
			match_info JSON NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_series_match (series_id, match_id)
		)`,
		`CREATE TABLE IF NOT EXISTS snapshot (
			id INT AUTO_INCREMENT PRIMARY KEY,
			snapshot_date DATE NOT NULL,
			tournament_id INT NOT NULL,
			leaderboard_json JSON NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tournament_id) REFERENCES tournament(id) ON DELETE CASCADE,
			UNIQUE KEY uk_date_tournament (snapshot_date, tournament_id)
		)`,
		"CREATE TABLE IF NOT EXISTS leaderboard (" +
			"id INT AUTO_INCREMENT PRIMARY KEY, " +
			"tournament_id INT NOT NULL, " +
			"team_id INT NOT NULL, " +
			"points INT NOT NULL DEFAULT 0, " +
			"`rank` INT NOT NULL DEFAULT 0, " +
			"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, " +
			"FOREIGN KEY (tournament_id) REFERENCES tournament(id) ON DELETE CASCADE, " +
			"FOREIGN KEY (team_id) REFERENCES fantasy_league(id) ON DELETE CASCADE, " +
			"UNIQUE KEY uk_tournament_team (tournament_id, team_id)" +
			")",
		`CREATE INDEX idx_tournament_series ON tournament(series_id)`,
		`CREATE INDEX idx_fantasy_series ON fantasy_league(series_id)`,
		`CREATE INDEX idx_game_series ON game(series_id)`,
		`CREATE INDEX idx_leaderboard_tournament ON leaderboard(tournament_id)`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
