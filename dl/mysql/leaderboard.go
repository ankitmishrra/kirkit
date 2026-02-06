package mysql

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"kirkit/dl"
)

type leaderboardRepo struct {
	db *sql.DB
}

// NewLeaderboardRepo returns a MySQL implementation of LeaderboardRepo.
func NewLeaderboardRepo(db *sql.DB) dl.LeaderboardRepo {
	return &leaderboardRepo{db: db}
}

func (r *leaderboardRepo) Upsert(ctx context.Context, tournamentID, teamID, points, rank int) (int64, error) {
	// INSERT ... ON DUPLICATE KEY UPDATE; parameterized Exec.
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO leaderboard (tournament_id, team_id, points, `rank`) VALUES (?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE points = VALUES(points), `rank` = VALUES(`rank`), updated_at = CURRENT_TIMESTAMP",
		tournamentID, teamID, points, rank)
	if err != nil {
		return 0, err
	}
	q := Builder.Select("id").From("leaderboard").Where(sq.Eq{"tournament_id": tournamentID, "team_id": teamID})
	var id int64
	err = q.RunWith(r.db).QueryRowContext(ctx).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *leaderboardRepo) GetByTournamentID(ctx context.Context, tournamentID int) ([]dl.LeaderboardRow, error) {
	q := Builder.Select("id", "tournament_id", "team_id", "points", "`rank`").From("leaderboard").
		Where(sq.Eq{"tournament_id": tournamentID}).OrderBy("`rank`", "points DESC")
	rows, err := q.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dl.LeaderboardRow
	for rows.Next() {
		var row dl.LeaderboardRow
		if err := rows.Scan(&row.ID, &row.TournamentID, &row.TeamID, &row.Points, &row.Rank); err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (r *leaderboardRepo) DeleteByTournamentID(ctx context.Context, tournamentID int) error {
	q := Builder.Delete("leaderboard").Where(sq.Eq{"tournament_id": tournamentID})
	_, err := q.RunWith(r.db).ExecContext(ctx)
	return err
}
