package bl

import (
	"context"
	"encoding/json"
	"fmt"

	"kirkit/spec/response"
)

// CricapiMatchData represents the "data" part of cricapi match_info response.
// Score is team-level; player-level may come from scorecard if available.
type CricapiMatchData struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Score     []CricapiInningScore `json:"score,omitempty"`
	Teams     []string      `json:"teams,omitempty"`
	MatchEnded bool         `json:"matchEnded"`
	// Optional: if API returns scorecard with player stats
	Scorecard interface{}   `json:"scorecard,omitempty"`
}

type CricapiInningScore struct {
	R      int    `json:"r"`
	W      int    `json:"w"`
	O      float64 `json:"o"`
	Inning string `json:"inning"`
}

// ComputeLeaderboardForMatch computes new leaderboard entries after applying points from the given match.
// It loads current leaderboard, adds points from this match for each team's squad, and returns sorted entries.
func (b *BL) ComputeLeaderboardForMatch(ctx context.Context, tournamentID int, matchID string) ([]response.LeaderboardEntry, error) {
	tour, err := b.TournamentRepo.GetByID(ctx, tournamentID)
	if err != nil || tour == nil {
		return nil, fmt.Errorf("tournament not found: %w", err)
	}
	mrow, err := b.MatchRepo.GetByMatchID(ctx, tour.SeriesID, matchID)
	if err != nil || mrow == nil {
		return nil, fmt.Errorf("match not found: %w", err)
	}
	var matchData CricapiMatchData
	if err := json.Unmarshal(mrow.MatchInfo, &matchData); err != nil {
		return nil, fmt.Errorf("parse match_info: %w", err)
	}
	// Prefer dedicated scorecard column (updated after match); fall back to match_info.scorecard
	if len(mrow.Scorecard) > 0 {
		var sc interface{}
		if err := json.Unmarshal(mrow.Scorecard, &sc); err == nil {
			matchData.Scorecard = sc
		}
	}
	playerPoints := b.extractPlayerPointsFromMatch(matchData)

	teams, err := b.FantasyLeagueRepo.ListBySeriesID(ctx, tour.SeriesID)
	if err != nil {
		return nil, err
	}
	currentLB, _ := b.LeaderboardRepo.GetByTournamentID(ctx, tournamentID)
	pointsByTeamID := make(map[int]int)
	for _, row := range currentLB {
		pointsByTeamID[row.TeamID] = row.Points
	}
	for _, t := range teams {
		squad, err := b.SquadRepo.GetByTeamID(ctx, t.ID)
		if err != nil {
			continue
		}
		matchPoints := 0
		for _, s := range squad {
			pts, ok := playerPoints[s.PlayerName]
			if !ok {
				continue
			}
			mult := 1
			if s.IsCaptain {
				mult = b.PointConfig.CaptainMult
			} else if s.IsViceCaptain {
				// 1.5x: (pts * 3) / 2
				matchPoints += (pts * b.PointConfig.ViceMultNum) / b.PointConfig.ViceMultDen
				continue
			}
			matchPoints += pts * mult
		}
		pointsByTeamID[t.ID] += matchPoints
	}

	// Build entries and sort by points desc
	type pair struct {
		teamID int
		points int
	}
	var pairs []pair
	for teamID, points := range pointsByTeamID {
		pairs = append(pairs, pair{teamID, points})
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].points > pairs[i].points {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	out := make([]response.LeaderboardEntry, len(pairs))
	for i, p := range pairs {
		rank := i + 1
		fl, _ := b.FantasyLeagueRepo.GetByID(ctx, p.teamID)
		teamName, teamOwner := "", ""
		if fl != nil {
			teamName, teamOwner = fl.TeamName, fl.TeamOwner
		}
		out[i] = response.LeaderboardEntry{
			TournamentID: tournamentID,
			TeamID:       p.teamID,
			TeamName:     teamName,
			TeamOwner:    teamOwner,
			Points:       p.points,
			Rank:         rank,
		}
	}
	return out, nil
}

// playerStats holds per-player stats from scorecard for full point calculation (PointCalculation.xlsx).
type playerStats struct {
	// Batting
	runs, balls, fours, sixes int
	out                       bool // dismissed (not "not out")
	// Bowling
	wickets, maidens       int
	overs                  float64
	runsConceded           int
	lbwBowledCount         int // wickets that were lbw or bowled
	dotBalls               int // if API provides
	// Fielding
	catches, stumpings, runouts int
}

// extractPlayerPointsFromMatch parses match data for per-player stats and applies full PointCalculation.xlsx rules.
func (b *BL) extractPlayerPointsFromMatch(data CricapiMatchData) map[string]int {
	out := make(map[string]int)
	if data.Scorecard == nil {
		return out
	}
	if arr, ok := data.Scorecard.([]interface{}); ok {
		acc := make(map[string]*playerStats)
		getName := func(m map[string]interface{}, key string) string {
			v, _ := m[key].(map[string]interface{})
			if v == nil {
				return ""
			}
			n, _ := v["name"].(string)
			return n
		}
		for _, inning := range arr {
			inn, _ := inning.(map[string]interface{})
			if inn == nil {
				continue
			}
			// Batting: r, b, 4s, 6s, dismissal; track bowler for lbw/bowled
			if raw, _ := inn["batting"].([]interface{}); raw != nil {
				for _, it := range raw {
					p, _ := it.(map[string]interface{})
					if p == nil {
						continue
					}
					name := getName(p, "batsman")
					if name == "" {
						continue
					}
					if acc[name] == nil {
						acc[name] = &playerStats{}
					}
					s := acc[name]
					if v, ok := p["r"].(float64); ok {
						s.runs += int(v)
					}
					if v, ok := p["b"].(float64); ok {
						s.balls += int(v)
					}
					if v, ok := p["4s"].(float64); ok {
						s.fours += int(v)
					}
					if v, ok := p["6s"].(float64); ok {
						s.sixes += int(v)
					}
					dismissal, _ := p["dismissal"].(string)
					if dismissal != "" && dismissal != "not out" {
						s.out = true
						bowlerName := getName(p, "bowler")
						if bowlerName != "" && (dismissal == "lbw" || dismissal == "bowled") {
							if acc[bowlerName] == nil {
								acc[bowlerName] = &playerStats{}
							}
							acc[bowlerName].lbwBowledCount++
						}
					}
				}
			}
			// Bowling: o, m, r, w
			if raw, _ := inn["bowling"].([]interface{}); raw != nil {
				for _, it := range raw {
					p, _ := it.(map[string]interface{})
					if p == nil {
						continue
					}
					name := getName(p, "bowler")
					if name == "" {
						continue
					}
					if acc[name] == nil {
						acc[name] = &playerStats{}
					}
					s := acc[name]
					if v, ok := p["o"].(float64); ok {
						s.overs += v
					}
					if v, ok := p["m"].(float64); ok {
						s.maidens += int(v)
					}
					if v, ok := p["r"].(float64); ok {
						s.runsConceded += int(v)
					}
					if v, ok := p["w"].(float64); ok {
						s.wickets += int(v)
					}
				}
			}
			// Catching: catch, stumped, runout
			if raw, _ := inn["catching"].([]interface{}); raw != nil {
				for _, it := range raw {
					p, _ := it.(map[string]interface{})
					if p == nil {
						continue
					}
					name := getName(p, "catcher")
					if name == "" {
						continue
					}
					if acc[name] == nil {
						acc[name] = &playerStats{}
					}
					s := acc[name]
					if v, ok := p["catch"].(float64); ok {
						s.catches += int(v)
					}
					if v, ok := p["stumped"].(float64); ok {
						s.stumpings += int(v)
					}
					if v, ok := p["runout"].(float64); ok {
						s.runouts += int(v)
					}
				}
			}
		}
		for name, s := range acc {
			out[name] = b.computePlayerPoints(s)
		}
		return out
	}
	// Legacy: flat map of arrays
	if m, ok := data.Scorecard.(map[string]interface{}); ok {
		for _, v := range m {
			if arr, ok := v.([]interface{}); ok {
				for _, item := range arr {
					if p, ok := item.(map[string]interface{}); ok {
						name, _ := p["name"].(string)
						if name == "" {
							name, _ = p["playerName"].(string)
						}
						s := &playerStats{}
						if v, ok := p["r"].(float64); ok {
							s.runs = int(v)
						}
						if v, ok := p["4s"].(float64); ok {
							s.fours = int(v)
						}
						if v, ok := p["6s"].(float64); ok {
							s.sixes = int(v)
						}
						if v, ok := p["w"].(float64); ok {
							s.wickets = int(v)
						}
						if v, ok := p["catch"].(float64); ok {
							s.catches = int(v)
						}
						if v, ok := p["runOut"].(float64); ok {
							s.runouts = int(v)
						}
						out[name] = b.computePlayerPoints(s)
					}
				}
			}
		}
	}
	return out
}

// computePlayerPoints applies PointCalculation.xlsx rules to one player's stats.
func (b *BL) computePlayerPoints(s *playerStats) int {
	cfg := &b.PointConfig
	pts := 0
	// Batting
	pts += s.runs * cfg.Run
	pts += s.fours * cfg.Four
	pts += s.sixes * cfg.Six
	if s.out && s.runs == 0 {
		pts += cfg.DuckPenalty
	}
	if s.balls >= 10 {
		sr := 0.0
		if s.balls > 0 {
			sr = float64(s.runs) / float64(s.balls) * 100
		}
		switch {
		case sr > 170:
			pts += cfg.StrikeRateBonus170
		case sr >= 150:
			pts += cfg.StrikeRateBonus150
		case sr >= 130:
			pts += cfg.StrikeRateBonus130
		case sr >= 60 && sr <= 70:
			pts += cfg.StrikeRatePenalty70
		case sr >= 50 && sr < 60:
			pts += cfg.StrikeRatePenalty60
		case sr < 50:
			pts += cfg.StrikeRatePenalty50
		}
	}
	switch {
	case s.runs >= 100:
		pts += cfg.RunsBonus100
	case s.runs >= 75:
		pts += cfg.RunsBonus75
	case s.runs >= 50:
		pts += cfg.RunsBonus50
	case s.runs >= 25:
		pts += cfg.RunsBonus25
	}
	// Bowling
	pts += s.wickets * cfg.Wicket
	pts += s.dotBalls * cfg.DotBall
	if s.wickets >= 5 {
		pts += cfg.WicketMilestone5
	} else if s.wickets >= 4 {
		pts += cfg.WicketMilestone4
	} else if s.wickets >= 3 {
		pts += cfg.WicketMilestone3
	}
	pts += s.lbwBowledCount * cfg.LBWBowledBonus
	pts += s.maidens * cfg.MaidenBonus
	if s.overs >= 2 && s.overs > 0 {
		eco := float64(s.runsConceded) / s.overs
		switch {
		case eco < 5:
			pts += cfg.EconomyBonusLow
		case eco < 6:
			pts += cfg.EconomyBonus5
		case eco <= 7:
			pts += cfg.EconomyBonus6
		case eco <= 11:
			pts += cfg.EconomyPenalty10
		case eco <= 12:
			pts += cfg.EconomyPenalty11
		default:
			pts += cfg.EconomyPenalty12
		}
	}
	// Fielding
	pts += s.catches * cfg.Catch
	pts += s.stumpings * cfg.Stumping
	pts += s.runouts * cfg.RunOut
	if s.catches >= 3 {
		pts += cfg.CatchBonus3
	}
	return pts
}

// UpdateLeaderboard persists the given leaderboard entries (replace for that tournament).
func (b *BL) UpdateLeaderboard(ctx context.Context, tournamentID int, entries []response.LeaderboardEntry) error {
	if err := b.LeaderboardRepo.DeleteByTournamentID(ctx, tournamentID); err != nil {
		return err
	}
	for _, e := range entries {
		_, err := b.LeaderboardRepo.Upsert(ctx, tournamentID, e.TeamID, e.Points, e.Rank)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateSnapshot saves current leaderboard for the tournament on the given date.
func (b *BL) CreateSnapshot(ctx context.Context, tournamentID int, snapshotDate string) (*response.Snapshot, error) {
	entries, err := b.GetLeaderboard(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(entries)
	if err != nil {
		return nil, err
	}
	id, err := b.SnapshotRepo.Create(ctx, snapshotDate, tournamentID, payload)
	if err != nil {
		return nil, err
	}
	return &response.Snapshot{
		ID:              int(id),
		SnapshotDate:     snapshotDate,
		TournamentID:     tournamentID,
		LeaderboardJSON: entries,
	}, nil
}
