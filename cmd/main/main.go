package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"kirkit/bl"
	"kirkit/dl/mysql"
	"kirkit/endpoints"
	"kirkit/spec/params"
	"kirkit/transport"
)

func main() {
	cfg := params.SvcParam{
		MySQLDSN:    getEnv("MYSQL_DSN", "kirkit:kirkit_secret@tcp(localhost:3306)/kirkit?parseTime=true"),
		CricAPIKey:  getEnv("CRICAPI_KEY", ""),
		CricAPIBase: getEnv("CRICAPI_BASE", "https://api.cricapi.com/v1"),
		ServerPort:  getEnv("PORT", "8080"),
	}

	db, err := mysql.NewDB(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	defer db.Close()

	tr := mysql.NewTournamentRepo(db)
	flr := mysql.NewFantasyLeagueRepo(db)
	sr := mysql.NewSquadRepo(db)
	mr := mysql.NewMatchRepo(db)
	snapr := mysql.NewSnapshotRepo(db)
	lbr := mysql.NewLeaderboardRepo(db)

	svc := bl.NewBL(tr, flr, sr, mr, snapr, lbr)
	eps := endpoints.New(svc)
	handler := transport.NewHTTPHandler(eps)
	// CORS for web app (e.g. React on port 3000 calling API on 8080)
	handler = corsHandler(handler)

	srv := &http.Server{Addr: ":" + cfg.ServerPort, Handler: handler}
	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
