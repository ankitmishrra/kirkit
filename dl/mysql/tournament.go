package mysql

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"kirkit/dl"
)

type tournamentRepo struct {
	db *sql.DB
}

// NewTournamentRepo returns a MySQL implementation of TournamentRepo.
func NewTournamentRepo(db *sql.DB) dl.TournamentRepo {
	return &tournamentRepo{db: db}
}

func (r *tournamentRepo) Create(ctx context.Context, seriesID, name string, status int) (int64, error) {
	q := Builder.Insert("tournament").Columns("series_id", "name", "status").Values(seriesID, name, status)
	res, err := q.RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *tournamentRepo) GetByID(ctx context.Context, id int) (*dl.TournamentRow, error) {
	q := Builder.Select("id", "series_id", "name", "status").From("tournament").Where(sq.Eq{"id": id})
	var row dl.TournamentRow
	err := q.RunWith(r.db).QueryRowContext(ctx).Scan(&row.ID, &row.SeriesID, &row.Name, &row.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *tournamentRepo) GetBySeriesID(ctx context.Context, seriesID string) (*dl.TournamentRow, error) {
	q := Builder.Select("id", "series_id", "name", "status").From("tournament").Where(sq.Eq{"series_id": seriesID})
	var row dl.TournamentRow
	err := q.RunWith(r.db).QueryRowContext(ctx).Scan(&row.ID, &row.SeriesID, &row.Name, &row.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *tournamentRepo) List(ctx context.Context) ([]dl.TournamentRow, error) {
	q := Builder.Select("id", "series_id", "name", "status").From("tournament").OrderBy("id")
	rows, err := q.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dl.TournamentRow
	for rows.Next() {
		var row dl.TournamentRow
		if err := rows.Scan(&row.ID, &row.SeriesID, &row.Name, &row.Status); err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (r *tournamentRepo) Update(ctx context.Context, id int, name *string, status *int) error {
	if name == nil && status == nil {
		return nil
	}
	u := Builder.Update("tournament").Where(sq.Eq{"id": id})
	if name != nil {
		u = u.Set("name", *name)
	}
	if status != nil {
		u = u.Set("status", *status)
	}
	_, err := u.RunWith(r.db).ExecContext(ctx)
	return err
}

func (r *tournamentRepo) Delete(ctx context.Context, id int) error {
	q := Builder.Delete("tournament").Where(sq.Eq{"id": id})
	_, err := q.RunWith(r.db).ExecContext(ctx)
	return err
}
