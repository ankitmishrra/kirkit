package mysql

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"kirkit/dl"
)

type fantasyLeagueRepo struct {
	db *sql.DB
}

// NewFantasyLeagueRepo returns a MySQL implementation of FantasyLeagueRepo.
func NewFantasyLeagueRepo(db *sql.DB) dl.FantasyLeagueRepo {
	return &fantasyLeagueRepo{db: db}
}

func (r *fantasyLeagueRepo) Create(ctx context.Context, seriesID, teamName, teamOwner string) (int64, error) {
	q := Builder.Insert("fantasy_league").Columns("series_id", "team_name", "team_owner").Values(seriesID, teamName, teamOwner)
	res, err := q.RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *fantasyLeagueRepo) GetByID(ctx context.Context, id int) (*dl.FantasyLeagueRow, error) {
	q := Builder.Select("id", "series_id", "team_name", "team_owner").From("fantasy_league").Where(sq.Eq{"id": id})
	var row dl.FantasyLeagueRow
	err := q.RunWith(r.db).QueryRowContext(ctx).Scan(&row.ID, &row.SeriesID, &row.TeamName, &row.TeamOwner)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *fantasyLeagueRepo) ListBySeriesID(ctx context.Context, seriesID string) ([]dl.FantasyLeagueRow, error) {
	q := Builder.Select("id", "series_id", "team_name", "team_owner").From("fantasy_league").Where(sq.Eq{"series_id": seriesID}).OrderBy("id")
	rows, err := q.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dl.FantasyLeagueRow
	for rows.Next() {
		var row dl.FantasyLeagueRow
		if err := rows.Scan(&row.ID, &row.SeriesID, &row.TeamName, &row.TeamOwner); err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (r *fantasyLeagueRepo) Update(ctx context.Context, id int, teamName, teamOwner *string) error {
	if teamName == nil && teamOwner == nil {
		return nil
	}
	u := Builder.Update("fantasy_league").Where(sq.Eq{"id": id})
	if teamName != nil {
		u = u.Set("team_name", *teamName)
	}
	if teamOwner != nil {
		u = u.Set("team_owner", *teamOwner)
	}
	_, err := u.RunWith(r.db).ExecContext(ctx)
	return err
}

func (r *fantasyLeagueRepo) Delete(ctx context.Context, id int) error {
	q := Builder.Delete("fantasy_league").Where(sq.Eq{"id": id})
	_, err := q.RunWith(r.db).ExecContext(ctx)
	return err
}
