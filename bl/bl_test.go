package bl

import (
	"context"
	"testing"

	"kirkit/dl"
	"kirkit/spec/request"
	"kirkit/spec/response"
)

type mockTournamentRepo struct {
	create func(ctx context.Context, seriesID, name string, status int) (int64, error)
	getByID func(ctx context.Context, id int) (*dl.TournamentRow, error)
	list   func(ctx context.Context) ([]dl.TournamentRow, error)
	update func(ctx context.Context, id int, name *string, status *int) error
	delete func(ctx context.Context, id int) error
}

func (m *mockTournamentRepo) Create(ctx context.Context, seriesID, name string, status int) (int64, error) {
	if m.create != nil {
		return m.create(ctx, seriesID, name, status)
	}
	return 1, nil
}
func (m *mockTournamentRepo) GetByID(ctx context.Context, id int) (*dl.TournamentRow, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return &dl.TournamentRow{ID: id, SeriesID: "s1", Name: "T1", Status: 1}, nil
}
func (m *mockTournamentRepo) GetBySeriesID(ctx context.Context, seriesID string) (*dl.TournamentRow, error) {
	return nil, nil
}
func (m *mockTournamentRepo) List(ctx context.Context) ([]dl.TournamentRow, error) {
	if m.list != nil {
		return m.list(ctx)
	}
	return nil, nil
}
func (m *mockTournamentRepo) Update(ctx context.Context, id int, name *string, status *int) error {
	if m.update != nil {
		return m.update(ctx, id, name, status)
	}
	return nil
}
func (m *mockTournamentRepo) Delete(ctx context.Context, id int) error {
	if m.delete != nil {
		return m.delete(ctx, id)
	}
	return nil
}

type mockFantasyLeagueRepo struct {
	create func(ctx context.Context, seriesID, teamName, teamOwner string) (int64, error)
	getByID func(ctx context.Context, id int) (*dl.FantasyLeagueRow, error)
	listBySeriesID func(ctx context.Context, seriesID string) ([]dl.FantasyLeagueRow, error)
	update func(ctx context.Context, id int, teamName, teamOwner *string) error
	delete func(ctx context.Context, id int) error
}

func (m *mockFantasyLeagueRepo) Create(ctx context.Context, seriesID, teamName, teamOwner string) (int64, error) {
	if m.create != nil {
		return m.create(ctx, seriesID, teamName, teamOwner)
	}
	return 1, nil
}
func (m *mockFantasyLeagueRepo) GetByID(ctx context.Context, id int) (*dl.FantasyLeagueRow, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return &dl.FantasyLeagueRow{ID: id, SeriesID: "s1", TeamName: "Team1", TeamOwner: "Owner1"}, nil
}
func (m *mockFantasyLeagueRepo) ListBySeriesID(ctx context.Context, seriesID string) ([]dl.FantasyLeagueRow, error) {
	if m.listBySeriesID != nil {
		return m.listBySeriesID(ctx, seriesID)
	}
	return nil, nil
}
func (m *mockFantasyLeagueRepo) Update(ctx context.Context, id int, teamName, teamOwner *string) error {
	if m.update != nil {
		return m.update(ctx, id, teamName, teamOwner)
	}
	return nil
}
func (m *mockFantasyLeagueRepo) Delete(ctx context.Context, id int) error {
	if m.delete != nil {
		return m.delete(ctx, id)
	}
	return nil
}

type mockSquadRepo struct{}
func (m *mockSquadRepo) Create(ctx context.Context, teamID int, playerName string, isCaptain, isViceCaptain bool, playerType string) (int64, error) {
	return 1, nil
}
func (m *mockSquadRepo) GetByTeamID(ctx context.Context, teamID int) ([]dl.SquadRow, error) {
	return nil, nil
}
func (m *mockSquadRepo) DeleteByTeamID(ctx context.Context, teamID int) error { return nil }
func (m *mockSquadRepo) DeleteByTeamIDAndPlayer(ctx context.Context, teamID int, playerName string) error {
	return nil
}

type mockMatchRepo struct {
	getByID      func(ctx context.Context, id int) (*dl.MatchRow, error)
	getByMatchID func(ctx context.Context, seriesID, matchID string) (*dl.MatchRow, error)
}
func (m *mockMatchRepo) Create(ctx context.Context, seriesID, matchID string, matchInfo []byte) (int64, error) {
	return 1, nil
}
func (m *mockMatchRepo) Upsert(ctx context.Context, seriesID, matchID string, matchInfo []byte) (int64, error) {
	return 1, nil
}
func (m *mockMatchRepo) UpdateScorecard(ctx context.Context, seriesID, matchID string, scorecard []byte) error {
	return nil
}
func (m *mockMatchRepo) GetByID(ctx context.Context, id int) (*dl.MatchRow, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	// Default: return a row for id=1 so CreateMatch -> GetMatch(1) works
	if id == 1 {
		return &dl.MatchRow{ID: 1, SeriesID: "s1", MatchID: "m1", MatchInfo: []byte(`{}`)}, nil
	}
	return nil, nil
}
func (m *mockMatchRepo) GetByMatchID(ctx context.Context, seriesID, matchID string) (*dl.MatchRow, error) {
	if m.getByMatchID != nil {
		return m.getByMatchID(ctx, seriesID, matchID)
	}
	return nil, nil
}
func (m *mockMatchRepo) ListBySeriesID(ctx context.Context, seriesID string) ([]dl.MatchRow, error) {
	return nil, nil
}
func (m *mockMatchRepo) Delete(ctx context.Context, id int) error { return nil }

type mockSnapshotRepo struct{}
func (m *mockSnapshotRepo) Create(ctx context.Context, snapshotDate string, tournamentID int, leaderboardJSON []byte) (int64, error) {
	return 1, nil
}
func (m *mockSnapshotRepo) GetByDateAndTournament(ctx context.Context, snapshotDate string, tournamentID int) (*dl.SnapshotRow, error) {
	return nil, nil
}
func (m *mockSnapshotRepo) ListByTournament(ctx context.Context, tournamentID int) ([]dl.SnapshotRow, error) {
	return nil, nil
}

type mockLeaderboardRepo struct {
	getByTournamentID func(ctx context.Context, tournamentID int) ([]dl.LeaderboardRow, error)
	deleteByTournamentID func(ctx context.Context, tournamentID int) error
	upsert func(ctx context.Context, tournamentID, teamID, points, rank int) (int64, error)
}
func (m *mockLeaderboardRepo) GetByTournamentID(ctx context.Context, tournamentID int) ([]dl.LeaderboardRow, error) {
	if m.getByTournamentID != nil {
		return m.getByTournamentID(ctx, tournamentID)
	}
	return nil, nil
}
func (m *mockLeaderboardRepo) DeleteByTournamentID(ctx context.Context, tournamentID int) error {
	if m.deleteByTournamentID != nil {
		return m.deleteByTournamentID(ctx, tournamentID)
	}
	return nil
}
func (m *mockLeaderboardRepo) Upsert(ctx context.Context, tournamentID, teamID, points, rank int) (int64, error) {
	if m.upsert != nil {
		return m.upsert(ctx, tournamentID, teamID, points, rank)
	}
	return 1, nil
}

func TestBL_CreateTournament(t *testing.T) {
	var createdID int64
	tr := &mockTournamentRepo{
		create: func(ctx context.Context, seriesID, name string, status int) (int64, error) {
			createdID = 42
			return 42, nil
		},
		getByID: func(ctx context.Context, id int) (*dl.TournamentRow, error) {
			return &dl.TournamentRow{ID: id, SeriesID: "s1", Name: "T1", Status: 1}, nil
		},
	}
	b := NewBL(tr, &mockFantasyLeagueRepo{}, &mockSquadRepo{}, &mockMatchRepo{}, &mockSnapshotRepo{}, &mockLeaderboardRepo{})
	ctx := context.Background()
	out, err := b.CreateTournament(ctx, request.TournamentCreate{SeriesID: "s1", Name: "T1", Status: 1})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != 42 || out.SeriesID != "s1" || out.Name != "T1" {
		t.Errorf("unexpected output: %+v", out)
	}
	_ = createdID
}

func TestBL_CreateMatch(t *testing.T) {
	matchInfo := map[string]interface{}{"id": "m1", "name": "Match 1"}
	b := NewBL(&mockTournamentRepo{}, &mockFantasyLeagueRepo{}, &mockSquadRepo{}, &mockMatchRepo{}, &mockSnapshotRepo{}, &mockLeaderboardRepo{})
	ctx := context.Background()
	out, err := b.CreateMatch(ctx, request.MatchCreate{SeriesID: "s1", MatchID: "m1", MatchInfo: matchInfo})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != 1 || out.MatchID != "m1" {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestDefaultPointConfig(t *testing.T) {
	cfg := DefaultPointConfig()
	if cfg.Run != 1 || cfg.Wicket != 25 || cfg.CaptainMult != 2 || cfg.RunOut != 6 || cfg.Four != 4 || cfg.Six != 6 {
		t.Errorf("unexpected point config: %+v", cfg)
	}
}

func TestComputeLeaderboardForMatch_NotFound(t *testing.T) {
	tr := &mockTournamentRepo{
		getByID: func(ctx context.Context, id int) (*dl.TournamentRow, error) {
			return nil, nil
		},
	}
	b := NewBL(tr, &mockFantasyLeagueRepo{}, &mockSquadRepo{}, &mockMatchRepo{}, &mockSnapshotRepo{}, &mockLeaderboardRepo{})
	ctx := context.Background()
	_, err := b.ComputeLeaderboardForMatch(ctx, 1, "m1")
	if err == nil {
		t.Fatal("expected error when tournament not found")
	}
}

func TestUpdateLeaderboard(t *testing.T) {
	var deletedID int
	upserted := &[]response.LeaderboardEntry{}
	lbr := &mockLeaderboardRepo{
		deleteByTournamentID: func(ctx context.Context, tournamentID int) error {
			deletedID = tournamentID
			return nil
		},
		upsert: func(ctx context.Context, tournamentID, teamID, points, rank int) (int64, error) {
			*upserted = append(*upserted, response.LeaderboardEntry{TournamentID: tournamentID, TeamID: teamID, Points: points, Rank: rank})
			return 1, nil
		},
	}
	b := NewBL(&mockTournamentRepo{}, &mockFantasyLeagueRepo{}, &mockSquadRepo{}, &mockMatchRepo{}, &mockSnapshotRepo{}, lbr)
	ctx := context.Background()
	entries := []response.LeaderboardEntry{
		{TournamentID: 1, TeamID: 10, Points: 100, Rank: 1},
		{TournamentID: 1, TeamID: 11, Points: 80, Rank: 2},
	}
	if err := b.UpdateLeaderboard(ctx, 1, entries); err != nil {
		t.Fatal(err)
	}
	if deletedID != 1 {
		t.Errorf("expected delete tournament_id=1, got %d", deletedID)
	}
	if len(*upserted) != 2 {
		t.Errorf("expected 2 upserts, got %d", len(*upserted))
	}
}

func TestExtractPlayerPoints_EmptyScorecard(t *testing.T) {
	b := &BL{PointConfig: DefaultPointConfig()}
	data := CricapiMatchData{Score: []CricapiInningScore{{R: 100, W: 5}}}
	out := b.extractPlayerPointsFromMatch(data)
	if len(out) != 0 {
		t.Errorf("expected no player points without scorecard, got %d", len(out))
	}
}

func TestExtractPlayerPoints_WithScorecard(t *testing.T) {
	b := &BL{PointConfig: DefaultPointConfig()}
	// legacy format: map with array of player objects
	scorecard := map[string]interface{}{
		"batsmen": []interface{}{
			map[string]interface{}{"name": "Player A", "r": float64(50), "w": float64(0), "catch": float64(0), "runOut": float64(0)},
		},
	}
	data := CricapiMatchData{Scorecard: scorecard}
	out := b.extractPlayerPointsFromMatch(data)
	if len(out) == 0 {
		t.Error("expected player points from scorecard")
	}
	// 50 runs: 50*1 + RunsBonus50 (8) = 58
	if out["Player A"] != 58 {
		t.Errorf("expected 58 for Player A (50 runs + runs bonus 8), got %d", out["Player A"])
	}
}
