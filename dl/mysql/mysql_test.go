package mysql

import (
	"context"
	"os"
	"testing"
)

// TestComponent_CRUD requires MYSQL_DSN (e.g. kirkit:kirkit_secret@tcp(localhost:3306)/kirkit?parseTime=true).
// Run with: docker-compose up -d mysql && go test -v -run TestComponent ./dl/mysql/
func TestComponent_CRUD(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN not set (component test)")
	}
	db, err := NewDB(dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	tr := NewTournamentRepo(db)
	id, err := tr.Create(ctx, "test-series-id", "Test Tournament", 1)
	if err != nil {
		t.Fatalf("create tournament: %v", err)
	}
	row, err := tr.GetByID(ctx, int(id))
	if err != nil || row == nil {
		t.Fatalf("get tournament: %v", err)
	}
	if row.Name != "Test Tournament" || row.SeriesID != "test-series-id" {
		t.Errorf("unexpected row: %+v", row)
	}
	// Cleanup
	_ = tr.Delete(ctx, int(id))
}
