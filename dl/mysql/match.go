package mysql

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"kirkit/dl"
)

type matchRepo struct {
	db *sql.DB
}

// NewMatchRepo returns a MySQL implementation of MatchRepo.
func NewMatchRepo(db *sql.DB) dl.MatchRepo {
	return &matchRepo{db: db}
}

func (r *matchRepo) Create(ctx context.Context, seriesID, matchID string, matchInfo []byte) (int64, error) {
	q := Builder.Insert("game").Columns("series_id", "match_id", "match_info").Values(seriesID, matchID, matchInfo)
	res, err := q.RunWith(r.db).ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *matchRepo) Upsert(ctx context.Context, seriesID, matchID string, matchInfo []byte) (int64, error) {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO game (series_id, match_id, match_info) VALUES (?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE match_info = VALUES(match_info), updated_at = CURRENT_TIMESTAMP",
		seriesID, matchID, matchInfo)
	if err != nil {
		return 0, err
	}
	q := Builder.Select("id").From("game").Where(sq.Eq{"series_id": seriesID, "match_id": matchID})
	var id int64
	err = q.RunWith(r.db).QueryRowContext(ctx).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *matchRepo) UpdateScorecard(ctx context.Context, seriesID, matchID string, scorecard []byte) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE game SET scorecard = ?, updated_at = CURRENT_TIMESTAMP WHERE series_id = ? AND match_id = ?",
		scorecard, seriesID, matchID)
	return err
}

func (r *matchRepo) GetByID(ctx context.Context, id int) (*dl.MatchRow, error) {
	q := Builder.Select("id", "series_id", "match_id", "match_info", "scorecard").From("game").Where(sq.Eq{"id": id})
	var row dl.MatchRow
	var scorecard []byte
	err := q.RunWith(r.db).QueryRowContext(ctx).Scan(&row.ID, &row.SeriesID, &row.MatchID, &row.MatchInfo, &scorecard)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	row.Scorecard = scorecard
	return &row, nil
}

func (r *matchRepo) GetByMatchID(ctx context.Context, seriesID, matchID string) (*dl.MatchRow, error) {
	q := Builder.Select("id", "series_id", "match_id", "match_info", "scorecard").From("game").
		Where(sq.Eq{"series_id": seriesID, "match_id": matchID})
	var row dl.MatchRow
	var scorecard []byte
	err := q.RunWith(r.db).QueryRowContext(ctx).Scan(&row.ID, &row.SeriesID, &row.MatchID, &row.MatchInfo, &scorecard)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	row.Scorecard = scorecard
	return &row, nil
}

func (r *matchRepo) ListBySeriesID(ctx context.Context, seriesID string) ([]dl.MatchRow, error) {
	q := Builder.Select("id", "series_id", "match_id", "match_info", "scorecard").From("game").
		Where(sq.Eq{"series_id": seriesID}).OrderBy("id")
	rows, err := q.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dl.MatchRow
	for rows.Next() {
		var row dl.MatchRow
		var scorecard []byte
		if err := rows.Scan(&row.ID, &row.SeriesID, &row.MatchID, &row.MatchInfo, &scorecard); err != nil {
			return nil, err
		}
		row.Scorecard = scorecard
		list = append(list, row)
	}
	return list, rows.Err()
}

func (r *matchRepo) Delete(ctx context.Context, id int) error {
	q := Builder.Delete("game").Where(sq.Eq{"id": id})
	_, err := q.RunWith(r.db).ExecContext(ctx)
	return err
}
