package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"kirkit/bl"
	"kirkit/dl/mysql"
	"kirkit/endpoints"
	"kirkit/spec/request"
	"kirkit/spec/response"
)

const (
	t20WorldCupSeriesID = "c4ca5cd5-e25c-4d83-bb77-2d193d93475a"
	e2eMatchIDPrefix   = "e2e-match-t20wc-"
)

// TestE2E_Leaderboard_T20WorldCup is an end-to-end component test:
// 1. Add tournament (T20 World Cup)
// 2. Add fantasy leagues (2 teams) with squads (captain/vice)
// 3. Add a match with scorecard (player stats)
// 4. Compute leaderboard for that match
// 5. Update leaderboard (persist)
// 6. GET leaderboard and assert order/points
//
// Leaderboard logic (see docs/LEADERBOARD_LOGIC.md):
//   - Player points from scorecard: runs×1 + wickets×25 + catch×8 + runOut×12
//   - Captain: 2×, Vice-captain: 1.5×, else 1×
//   - Team match points = sum of (player points × multiplier) for squad players in the match
//   - Leaderboard = sort teams by total points (desc); rank 1 = highest
func TestE2E_Leaderboard_T20WorldCup(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "kirkit:kirkit_secret@tcp(localhost:3306)/kirkit?parseTime=true"
	}
	db, err := mysql.NewDB(dsn)
	if err != nil {
		t.Skipf("mysql not available: %v (start MySQL with: docker-compose up -d)", err)
	}
	defer db.Close()

	svc := bl.NewBL(
		mysql.NewTournamentRepo(db),
		mysql.NewFantasyLeagueRepo(db),
		mysql.NewSquadRepo(db),
		mysql.NewMatchRepo(db),
		mysql.NewSnapshotRepo(db),
		mysql.NewLeaderboardRepo(db),
	)
	eps := endpoints.New(svc)
	handler := NewHTTPHandler(eps)

	// --- Print logic (for readability in test output) ---
	t.Log("=== Leaderboard calculation logic ===")
	t.Log("1. Player points from match scorecard: runs×1 + wickets×25 + catch×8 + runOut×12")
	t.Log("2. Captain: 2× multiplier; Vice-captain: 1.5×; Others: 1×")
	t.Log("3. Team match points = sum of (player_points × multiplier) for each squad player in the match")
	t.Log("4. Leaderboard total = prior total + match points; then sort by points desc → rank 1, 2, ...")
	t.Log("   (See docs/LEADERBOARD_LOGIC.md for full description)")
	t.Log("")

	// --- 1. Use existing T20 World Cup tournament (seeded by migration 002) ---
	listResp := doJSON(t, handler, http.MethodGet, "/api/v1/tournaments", nil)
	data, ok := listResp["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Fatalf("expected at least one tournament (run go run ./cmd/upgrade); got %+v", listResp)
	}
	var tournamentID int
	for _, v := range data {
		m := v.(map[string]interface{})
		if m["series_id"] == t20WorldCupSeriesID {
			tournamentID = int(m["id"].(float64))
			break
		}
	}
	if tournamentID == 0 {
		t.Fatalf("tournament with series_id %s not found", t20WorldCupSeriesID)
	}
	t.Logf("Using tournament id=%d, series_id=%s", tournamentID, t20WorldCupSeriesID)

	// --- 2. Create two fantasy leagues and add squads (unique names per run for reruns) ---
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	teamAName := "Team A " + suffix
	teamBName := "Team B " + suffix
	// Team A: Player1 (captain), Player2 (vice), Player3
	teamABody := request.FantasyLeagueCreate{
		SeriesID:  t20WorldCupSeriesID,
		TeamName:  teamAName,
		TeamOwner: "Owner A",
	}
	teamAResp := doJSON(t, handler, http.MethodPost, "/api/v1/fantasy-leagues", teamABody)
	teamAID := int(idFromData(t, teamAResp, "fantasy_league"))

	squadA := []request.SquadMemberCreate{
		{PlayerName: "Player1", IsCaptain: true, IsViceCaptain: false},
		{PlayerName: "Player2", IsCaptain: false, IsViceCaptain: true},
		{PlayerName: "Player3", IsCaptain: false, IsViceCaptain: false},
	}
	doJSON(t, handler, http.MethodPost, "/api/v1/fantasy-leagues/"+strconv.Itoa(teamAID)+"/squad", map[string]interface{}{"members": squadA})
	t.Logf("Created Team A (id=%d) with squad: Player1 (C), Player2 (VC), Player3", teamAID)

	// Team B: Player1, Player2, Player4
	teamBBody := request.FantasyLeagueCreate{
		SeriesID:  t20WorldCupSeriesID,
		TeamName:  teamBName,
		TeamOwner: "Owner B",
	}
	teamBResp := doJSON(t, handler, http.MethodPost, "/api/v1/fantasy-leagues", teamBBody)
	teamBID := int(idFromData(t, teamBResp, "fantasy_league"))

	squadB := []request.SquadMemberCreate{
		{PlayerName: "Player1", IsCaptain: false, IsViceCaptain: false},
		{PlayerName: "Player2", IsCaptain: false, IsViceCaptain: false},
		{PlayerName: "Player4", IsCaptain: false, IsViceCaptain: false},
	}
	doJSON(t, handler, http.MethodPost, "/api/v1/fantasy-leagues/"+strconv.Itoa(teamBID)+"/squad", map[string]interface{}{"members": squadB})
	t.Logf("Created Team B (id=%d) with squad: Player1, Player2, Player4", teamBID)

	// --- 3. Add match with scorecard (player stats for this match) ---
	e2eMatchID := e2eMatchIDPrefix + suffix
	// Player1: 50 runs → 50 pts; Player2: 30 runs + 2 wkts → 30+50=80; Player3: 20 runs → 20; Player4: 1 catch → 8
	matchInfo := map[string]interface{}{
		"id":         e2eMatchID,
		"name":       "E2E T20 Match",
		"matchEnded": true,
		"teams":      []string{"Team X", "Team Y"},
		"scorecard": map[string]interface{}{
			"batsmen": []interface{}{
				map[string]interface{}{"name": "Player1", "r": float64(50), "w": float64(0), "catch": float64(0), "runOut": float64(0)},
				map[string]interface{}{"name": "Player2", "r": float64(30), "w": float64(0), "catch": float64(0), "runOut": float64(0)},
				map[string]interface{}{"name": "Player3", "r": float64(20), "w": float64(0), "catch": float64(0), "runOut": float64(0)},
			},
			"bowlers": []interface{}{
				map[string]interface{}{"name": "Player2", "r": float64(0), "w": float64(2), "catch": float64(0), "runOut": float64(0)},
				map[string]interface{}{"name": "Player4", "r": float64(0), "w": float64(0), "catch": float64(1), "runOut": float64(0)},
			},
		},
	}
	// Note: same player can appear in batsmen and bowlers; we overwrite by name. Bowlers give Player2 w=2, Player4 catch=1.
	// Our extractPlayerPointsFromMatch iterates all arrays and sets out[name]=pts. So last occurrence wins. We need each player
	// to have one combined row. Looking at the code, it does for each item: pts = runs*1 + wickets*25 + catch*8 + runOut*12, out[name]=pts.
	// So if Player2 appears in batsmen (30,0,0,0) and bowlers (0,2,0,0), we get two entries and the second overwrites: first 30, second 0+50=50. So we'd lose runs. So we need one entry per player with combined stats. Let me use a single list "players" or put combined stats in one place. Actually re-reading the code: for each array in scorecard, for each item we set out[name]=pts (overwriting). So we need each player to appear once with total stats: Player1: 50; Player2: 30+2*25=80; Player3: 20; Player4: 8. So use one array "players" with these four.
	matchInfo = map[string]interface{}{
		"id":         e2eMatchID,
		"name":       "E2E T20 Match",
		"matchEnded": true,
		"teams":      []string{"Team X", "Team Y"},
		"scorecard": map[string]interface{}{
			"players": []interface{}{
				map[string]interface{}{"name": "Player1", "r": float64(50), "w": float64(0), "catch": float64(0), "runOut": float64(0)},
				map[string]interface{}{"name": "Player2", "r": float64(30), "w": float64(2), "catch": float64(0), "runOut": float64(0)},
				map[string]interface{}{"name": "Player3", "r": float64(20), "w": float64(0), "catch": float64(0), "runOut": float64(0)},
				map[string]interface{}{"name": "Player4", "r": float64(0), "w": float64(0), "catch": float64(1), "runOut": float64(0)},
			},
		},
	}
	matchBody := request.MatchCreate{
		SeriesID:  t20WorldCupSeriesID,
		MatchID:   e2eMatchID,
		MatchInfo: matchInfo,
	}
	doJSON(t, handler, http.MethodPost, "/api/v1/matches", matchBody)
	t.Logf("Created match id=%s with scorecard: Player1(50), Player2(30+2w=80), Player3(20), Player4(1c=8)", e2eMatchID)

	// --- 4. Compute leaderboard for this match ---
	computeBody := request.ComputeLeaderboardRequest{TournamentID: tournamentID, MatchID: e2eMatchID}
	computeResp := doJSON(t, handler, http.MethodPost, "/api/v1/engine/compute-leaderboard", computeBody)
	entriesRaw := computeResp["data"]
	if entriesRaw == nil {
		t.Fatalf("compute-leaderboard did not return data: %+v", computeResp)
	}
	entriesSlice, ok := entriesRaw.([]interface{})
	if !ok {
		t.Fatalf("compute-leaderboard data not array: %T", entriesRaw)
	}
	var entries []response.LeaderboardEntry
	for _, e := range entriesSlice {
		em := e.(map[string]interface{})
		entries = append(entries, response.LeaderboardEntry{
			TournamentID: tournamentID,
			TeamID:       int(em["team_id"].(float64)),
			TeamName:     stringVal(em, "team_name"),
			TeamOwner:    stringVal(em, "team_owner"),
			Points:       int(em["points"].(float64)),
			Rank:         int(em["rank"].(float64)),
		})
	}
	t.Logf("Computed leaderboard: %d entries", len(entries))
	for _, e := range entries {
		t.Logf("  Rank %d: %s (id=%d) %d pts", e.Rank, e.TeamName, e.TeamID, e.Points)
	}

	// --- 5. Update leaderboard (persist) ---
	updateBody := map[string]interface{}{
		"tournament_id": tournamentID,
		"entries":      entries,
	}
	doJSON(t, handler, http.MethodPost, "/api/v1/engine/update-leaderboard", updateBody)
	t.Log("Updated leaderboard persisted")

	// --- 6. GET leaderboard and assert our two teams have correct points and order ---
	getResp := doJSON(t, handler, http.MethodGet, "/api/v1/leaderboard/"+strconv.Itoa(tournamentID), nil)
	lbData, ok := getResp["data"]
	if !ok {
		t.Fatalf("GET leaderboard no data: %+v", getResp)
	}
	lbArr, ok := lbData.([]interface{})
	if !ok {
		t.Fatalf("GET leaderboard data not array: %T", lbData)
	}
	var entryA, entryB map[string]interface{}
	for _, v := range lbArr {
		m := v.(map[string]interface{})
		if int(m["team_id"].(float64)) == teamAID {
			entryA = m
		}
		if int(m["team_id"].(float64)) == teamBID {
			entryB = m
		}
	}
	if entryA == nil {
		t.Fatalf("our Team A (id=%d) not found in leaderboard", teamAID)
	}
	if entryB == nil {
		t.Fatalf("our Team B (id=%d) not found in leaderboard", teamBID)
	}
	ptsA := int(entryA["points"].(float64))
	ptsB := int(entryB["points"].(float64))
	rankA := int(entryA["rank"].(float64))
	rankB := int(entryB["rank"].(float64))
	// PointCalculation.xlsx: runs + runs bonus (25-49 +4, 50-74 +8), wickets 25, catch 8, captain 2x, vice 1.5x
	// Player1: 50 -> 58 (50+8); Player2: 30+2w -> 38+50=88; Player3: 20; Player4: 1c -> 8
	// Team A: 58*2 + 88*1.5 + 20 = 116+132+20 = 268; Team B: 58+88+8 = 154
	if ptsA != 268 || ptsB != 154 {
		t.Errorf("expected Team A 268 pts, Team B 154 pts; got %s %d pts, %s %d pts", teamAName, ptsA, teamBName, ptsB)
	}
	if rankA >= rankB {
		t.Errorf("expected Team A rank (%d) better than Team B rank (%d)", rankA, rankB)
	}
	t.Log("=== Assertions passed: Team A rank 1 (268 pts), Team B rank 2 (154 pts) ===")
}

// doJSON sends a JSON request and decodes the response into a map. For GET without body, pass nil.
func doJSON(t *testing.T, handler http.Handler, method, path string, body interface{}) map[string]interface{} {
	t.Helper()
	var req *http.Request
	if body == nil {
		req = httptest.NewRequest(method, path, nil)
	} else {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Fatalf("%s %s: status %d, body %s", method, path, w.Code, w.Body.String())
	}
	if w.Code == http.StatusNoContent || w.Body.Len() == 0 {
		return nil
	}
	var out map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// If response is { "data": ... }, return the whole map so caller can use data
	return out
}

func stringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// idFromData returns the "id" field from the API response envelope {"data": {"id": N, ...}}.
func idFromData(t *testing.T, resp map[string]interface{}, _ string) float64 {
	t.Helper()
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("response has no data object: %+v", resp)
	}
	id, ok := data["id"].(float64)
	if !ok {
		t.Fatalf("data has no id: %+v", data)
	}
	return id
}
