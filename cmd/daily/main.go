// Package main runs the daily leaderboard sync: for each live tournament (status=ongoing),
// finds matches that have no scorecard yet, fetches scorecard from CricAPI match_scorecard,
// updates game.scorecard, computes and persists the leaderboard, then emails the leaderboard.
//
// Run once per day (e.g. cron): CRICAPI_KEY=xxx go run ./cmd/daily
// Or: make daily (loads .env)
//
// Email: set SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM; EMAIL_TO defaults to ankitmishra7991@gmail.com.
package main

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"kirkit/bl"
	"kirkit/dl"
	"kirkit/dl/mysql"
	"kirkit/sdk/cricapi"
	"kirkit/spec/params"
)

const (
	ongoingStatus = 1 // tournament status: 1=ongoing (live), 2=done
	defaultEmailTo = "ankitmishra7991@gmail.com"
)

func main() {
	cfg := params.SvcParam{
		MySQLDSN:   getEnv("MYSQL_DSN", "kirkit:kirkit_secret@tcp(localhost:3306)/kirkit?parseTime=true"),
		CricAPIKey: getEnv("CRICAPI_KEY", ""),
		CricAPIBase: getEnv("CRICAPI_BASE", "https://api.cricapi.com/v1"),
	}

	if cfg.CricAPIKey == "" {
		log.Fatalf("CRICAPI_KEY required for daily sync")
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
	client := cricapi.NewClient(cfg.CricAPIKey, cfg.CricAPIBase)

	ctx := context.Background()
	allTournaments, err := tr.List(ctx)
	if err != nil {
		log.Fatalf("list tournaments: %v", err)
	}

	var liveTournaments []dl.TournamentRow
	for _, t := range allTournaments {
		if t.Status == ongoingStatus {
			liveTournaments = append(liveTournaments, t)
		}
	}

	for _, t := range liveTournaments {
		games, err := mr.ListBySeriesID(ctx, t.SeriesID)
		if err != nil {
			log.Printf("series %s list games: %v", t.SeriesID, err)
			continue
		}
		updated := 0
		for _, g := range games {
			if len(g.Scorecard) > 0 {
				continue
			}
			scorecardData, err := client.MatchScorecard(ctx, g.MatchID)
			if err != nil {
				log.Printf("match %s fetch scorecard: %v", g.MatchID, err)
				continue
			}
			if err := svc.UpdateMatchScorecard(ctx, t.SeriesID, g.MatchID, scorecardData); err != nil {
				log.Printf("match %s update scorecard: %v", g.MatchID, err)
				continue
			}
			entries, err := svc.ComputeLeaderboardForMatch(ctx, t.ID, g.MatchID)
			if err != nil {
				log.Printf("match %s compute leaderboard: %v", g.MatchID, err)
				continue
			}
			if err := svc.UpdateLeaderboard(ctx, t.ID, entries); err != nil {
				log.Printf("match %s update leaderboard: %v", g.MatchID, err)
				continue
			}
			updated++
			log.Printf("tournament %d (%s) match %s: scorecard updated, leaderboard refreshed", t.ID, t.Name, g.MatchID)
		}
		if updated > 0 {
			log.Printf("tournament %d (%s): %d match(es) processed", t.ID, t.Name, updated)
		}
	}

	// Build and send email with leaderboard for each live tournament
	body := buildLeaderboardEmail(ctx, svc, liveTournaments)
	emailTo := getEnv("EMAIL_TO", defaultEmailTo)
	if err := sendEmail(body, emailTo); err != nil {
		log.Printf("send email: %v", err)
	} else {
		log.Printf("leaderboard email sent to %s", emailTo)
	}
	log.Println("daily sync done")
}

func buildLeaderboardEmail(ctx context.Context, svc bl.Service, tournaments []dl.TournamentRow) string {
	var b strings.Builder
	b.WriteString("Kirkit – Daily Leaderboard\n\n")
	for _, t := range tournaments {
		entries, err := svc.GetLeaderboard(ctx, t.ID)
		if err != nil {
			b.WriteString(fmt.Sprintf("--- %s ---\n(error: %v)\n\n", t.Name, err))
			continue
		}
		b.WriteString(fmt.Sprintf("--- %s ---\n", t.Name))
		for _, e := range entries {
			teamLabel := e.TeamName
			if teamLabel == "" {
				teamLabel = fmt.Sprintf("Team %d", e.TeamID)
			}
			if e.TeamOwner != "" {
				teamLabel += " (" + e.TeamOwner + ")"
			}
			b.WriteString(fmt.Sprintf("  %d. %s – %d pts\n", e.Rank, teamLabel, e.Points))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func sendEmail(body, to string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	if host == "" || port == "" || user == "" || pass == "" || from == "" {
		return fmt.Errorf("email skipped: set SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM to send")
	}
	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)
	subject := "Subject: Kirkit Daily Leaderboard\n"
	mime := "MIME-Version: 1.0\nContent-Type: text/plain; charset=utf-8\n\n"
	msg := []byte(subject + mime + body)
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
