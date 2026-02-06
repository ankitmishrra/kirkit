package mysql

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"kirkit/dl"
)

type versionRepo struct {
	db *sql.DB
}

// NewVersionRepo returns a MySQL implementation of VersionRepo.
func NewVersionRepo(db *sql.DB) dl.VersionRepo {
	return &versionRepo{db: db}
}

func (r *versionRepo) Get(ctx context.Context) (int, error) {
	q := Builder.Select("version").From("version").Where(sq.Eq{"id": 1})
	var v int
	err := q.RunWith(r.db).QueryRowContext(ctx).Scan(&v)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (r *versionRepo) Set(ctx context.Context, version int) error {
	q := Builder.Update("version").Set("version", version).Where(sq.Eq{"id": 1})
	_, err := q.RunWith(r.db).ExecContext(ctx)
	return err
}
