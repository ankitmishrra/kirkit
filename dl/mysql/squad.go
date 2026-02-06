package mysql

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"kirkit/dl"
)

type squadRepo struct {
	db *sql.DB
}

// NewSquadRepo returns a MySQL implementation of SquadRepo.
func NewSquadRepo(db *sql.DB) dl.SquadRepo {
	return &squadRepo{db: db}
}

func (r *squadRepo) Create(ctx context.Context, teamID int, playerName string, isCaptain, isViceCaptain bool, playerType string) (int64, error) {
	ic, iv := 0, 0
	if isCaptain {
		ic = 1
	}
	if isViceCaptain {
		iv = 1
	}
	cols := []string{"team_id", "player_name", "is_captain", "is_vice_captain"}
	vals := []interface{}{teamID, playerName, ic, iv}
	if playerType != "" {
		cols = append(cols, "player_type")
		vals = append(vals, playerType)
	}
	q := Builder.Insert("squad").Columns(cols...).Values(vals...)
	res, err := q.RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *squadRepo) GetByTeamID(ctx context.Context, teamID int) ([]dl.SquadRow, error) {
	q := Builder.Select("id", "team_id", "player_name", "is_captain", "is_vice_captain", "player_type").
		From("squad").Where(sq.Eq{"team_id": teamID}).OrderBy("id")
	rows, err := q.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dl.SquadRow
	for rows.Next() {
		var row dl.SquadRow
		var ic, iv int
		if err := rows.Scan(&row.ID, &row.TeamID, &row.PlayerName, &ic, &iv, &row.PlayerType); err != nil {
			return nil, err
		}
		row.IsCaptain = ic == 1
		row.IsViceCaptain = iv == 1
		list = append(list, row)
	}
	return list, rows.Err()
}

func (r *squadRepo) DeleteByTeamID(ctx context.Context, teamID int) error {
	q := Builder.Delete("squad").Where(sq.Eq{"team_id": teamID})
	_, err := q.RunWith(r.db).ExecContext(ctx)
	return err
}

func (r *squadRepo) DeleteByTeamIDAndPlayer(ctx context.Context, teamID int, playerName string) error {
	q := Builder.Delete("squad").Where(sq.Eq{"team_id": teamID, "player_name": playerName})
	_, err := q.RunWith(r.db).ExecContext(ctx)
	return err
}
