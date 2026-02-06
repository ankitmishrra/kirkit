package mysql

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
)

// NewDB opens a MySQL connection. Caller must call db.Close().
func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// Builder is Squirrel statement builder for MySQL (? placeholders). Use for all
// parameterized queries so input is never concatenated into SQL (injection-proof).
var Builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)
