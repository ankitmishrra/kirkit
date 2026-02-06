package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"kirkit/bl"
	"kirkit/dl/mysql"
	"kirkit/endpoints"
	"kirkit/spec/request"
	"kirkit/spec/response"
)

// TestHTTP_TournamentCRUD runs against MySQL. Skip if not available.
func TestHTTP_TournamentCRUD(t *testing.T) {
	dsn := "kirkit:kirkit_secret@tcp(localhost:3306)/kirkit?parseTime=true"
	db, err := mysql.NewDB(dsn)
	if err != nil {
		t.Skipf("mysql not available: %v (start with docker-compose up -d)", err)
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

	// POST create tournament
	body := request.TournamentCreate{SeriesID: "test-s1", Name: "Test T", Status: 1}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tournaments", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("create tournament: status %d, body %s", w.Code, w.Body.String())
	}
	var apiResp response.APIResponse
	if err := json.NewDecoder(w.Body).Decode(&apiResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, _ := apiResp.Data.(map[string]interface{})
	idFloat, _ := data["id"].(float64)
	id := int(idFloat)

	// GET tournament by id
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/tournaments/"+strconv.Itoa(id), nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("get tournament: status %d, body %s", w2.Code, w2.Body.String())
	}
}
