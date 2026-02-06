// Package main runs DB migrations (Go-based, version-managed).
package main

import (
	"context"
	"log"
	"os"

	"kirkit/dl/mysql"
	"kirkit/upgrade"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "kirkit:kirkit_secret@tcp(localhost:3306)/kirkit?parseTime=true"
	}

	db, err := mysql.NewDB(dsn)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := upgrade.Runner(ctx, db); err != nil {
		log.Fatalf("upgrade: %v", err)
	}
	log.Printf("upgrade complete (current version from version table)")
}
