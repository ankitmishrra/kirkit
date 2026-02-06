package mysql

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"kirkit/dl"
)

type snapshotRepo struct {
	db *sql.DB
}

// NewSnapshotRepo returns a MySQL implementation of SnapshotRepo.
func NewSnapshotRepo(db *sql.DB) dl.SnapshotRepo {
	return &snapshotRepo{db: db}
}

func (r *snapshotRepo) Create(ctx context.Context, snapshotDate string, tournamentID int, leaderboardJSON []byte) (int64, error) {
	// INSERT ... ON DUPLICATE KEY UPDATE; use parameterized Exec.
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO snapshot (snapshot_date, tournament_id, leaderboard_json) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE leaderboard_json = VALUES(leaderboard_json)`,
		snapshotDate, tournamentID, leaderboardJSON)
	if err != nil {
		return 0, err
	}
	q := Builder.Select("id").From("snapshot").Where(sq.Eq{"snapshot_date": snapshotDate, "tournament_id": tournamentID})
	var id int64
	err = q.RunWith(r.db).QueryRowContext(ctx).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *snapshotRepo) GetByDateAndTournament(ctx context.Context, snapshotDate string, tournamentID int) (*dl.SnapshotRow, error) {
	q := Builder.Select("id", "snapshot_date", "tournament_id", "leaderboard_json").From("snapshot").
		Where(sq.Eq{"snapshot_date": snapshotDate, "tournament_id": tournamentID})
	var row dl.SnapshotRow
	err := q.RunWith(r.db).QueryRowContext(ctx).Scan(&row.ID, &row.SnapshotDate, &row.TournamentID, &row.LeaderboardJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *snapshotRepo) ListByTournament(ctx context.Context, tournamentID int) ([]dl.SnapshotRow, error) {
	q := Builder.Select("id", "snapshot_date", "tournament_id", "leaderboard_json").From("snapshot").
		Where(sq.Eq{"tournament_id": tournamentID}).OrderBy("snapshot_date DESC")
	rows, err := q.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dl.SnapshotRow
	for rows.Next() {
		var row dl.SnapshotRow
		if err := rows.Scan(&row.ID, &row.SnapshotDate, &row.TournamentID, &row.LeaderboardJSON); err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}
