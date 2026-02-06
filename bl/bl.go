package bl

import (
	"context"
	"encoding/json"

	"kirkit/dl"
	"kirkit/spec/request"
	"kirkit/spec/response"
)

// PointConfig holds fantasy point rules (aligned with PointCalculation.xlsx).
type PointConfig struct {
	// Batting
	Run       int `json:"run"`        // per run
	Four      int `json:"four"`       // per four
	Six       int `json:"six"`        // per six
	DuckPenalty int `json:"duck_penalty"` // batsman out for 0 (e.g. -2)
	// Strike rate bonus (min 10 balls): >170 +6, 150-170 +4, 130-150 +2, 60-70 -2, 50-60 -4, <50 -6
	StrikeRateBonus170  int `json:"strike_rate_bonus_170"`  // SR > 170
	StrikeRateBonus150  int `json:"strike_rate_bonus_150"`  // 150-170
	StrikeRateBonus130  int `json:"strike_rate_bonus_130"`  // 130-150
	StrikeRatePenalty70 int `json:"strike_rate_penalty_70"` // 60-70
	StrikeRatePenalty60 int `json:"strike_rate_penalty_60"` // 50-60
	StrikeRatePenalty50 int `json:"strike_rate_penalty_50"` // <50
	// Runs bonus: 25-49 +4, 50-74 +8, 75-99 +12, >=100 +16
	RunsBonus25  int `json:"runs_bonus_25"`  // 25-49
	RunsBonus50  int `json:"runs_bonus_50"`  // 50-74
	RunsBonus75  int `json:"runs_bonus_75"`  // 75-99
	RunsBonus100 int `json:"runs_bonus_100"` // >=100
	// Bowling
	Wicket       int `json:"wicket"`        // per wicket
	DotBall      int `json:"dot_ball"`      // per dot ball (if API provides)
	WicketMilestone3 int `json:"wicket_milestone_3"` // 3 wkts
	WicketMilestone4 int `json:"wicket_milestone_4"` // 4 wkts
	WicketMilestone5 int `json:"wicket_milestone_5"` // 5 wkts
	LBWBowledBonus   int `json:"lbw_bowled_bonus"`   // per LBW/bowled dismissal
	MaidenBonus      int `json:"maiden_bonus"`       // per maiden over
	// Economy bonus (min 2 overs): <5 +6, 5-5.99 +4, 6-7 +2, 10-11 -2, 11.01-12 -4, >12 -6
	EconomyBonusLow  int `json:"economy_bonus_low"`  // <5
	EconomyBonus5    int `json:"economy_bonus_5"`     // 5-5.99
	EconomyBonus6    int `json:"economy_bonus_6"`    // 6-7
	EconomyPenalty10 int `json:"economy_penalty_10"`  // 10-11
	EconomyPenalty11 int `json:"economy_penalty_11"`  // 11.01-12
	EconomyPenalty12 int `json:"economy_penalty_12"`  // >12
	// Fielding
	Catch       int `json:"catch"`        // per catch
	Stumping    int `json:"stumping"`     // per stumping
	RunOut      int `json:"run_out"`      // per run out (sheet: 6)
	CatchBonus3 int `json:"catch_bonus_3"` // +4 if >=3 catches in match
	// Captain / Vice
	CaptainMult int `json:"captain_mult"`
	ViceMultNum int `json:"vice_mult_num"`
	ViceMultDen int `json:"vice_mult_den"`
}

// DefaultPointConfig returns scoring aligned with PointCalculation.xlsx.
func DefaultPointConfig() PointConfig {
	return PointConfig{
		Run: 1, Four: 4, Six: 6, DuckPenalty: -2,
		StrikeRateBonus170: 6, StrikeRateBonus150: 4, StrikeRateBonus130: 2,
		StrikeRatePenalty70: -2, StrikeRatePenalty60: -4, StrikeRatePenalty50: -6,
		RunsBonus25: 4, RunsBonus50: 8, RunsBonus75: 12, RunsBonus100: 16,
		Wicket: 25, DotBall: 1,
		WicketMilestone3: 4, WicketMilestone4: 8, WicketMilestone5: 12,
		LBWBowledBonus: 0, MaidenBonus: 8,
		EconomyBonusLow: 6, EconomyBonus5: 4, EconomyBonus6: 2,
		EconomyPenalty10: -2, EconomyPenalty11: -4, EconomyPenalty12: -6,
		Catch: 8, Stumping: 12, RunOut: 6, CatchBonus3: 4,
		CaptainMult: 2, ViceMultNum: 3, ViceMultDen: 2,
	}
}

// Service defines the business logic interface.
type Service interface {
	// Tournament
	CreateTournament(ctx context.Context, req request.TournamentCreate) (*response.Tournament, error)
	GetTournament(ctx context.Context, id int) (*response.Tournament, error)
	ListTournaments(ctx context.Context) ([]response.Tournament, error)
	UpdateTournament(ctx context.Context, id int, req request.TournamentUpdate) (*response.Tournament, error)
	DeleteTournament(ctx context.Context, id int) error

	// Fantasy League
	CreateFantasyLeague(ctx context.Context, req request.FantasyLeagueCreate) (*response.FantasyLeague, error)
	GetFantasyLeague(ctx context.Context, id int) (*response.FantasyLeague, error)
	ListFantasyLeagues(ctx context.Context, seriesID string) ([]response.FantasyLeague, error)
	UpdateFantasyLeague(ctx context.Context, id int, req request.FantasyLeagueUpdate) (*response.FantasyLeague, error)
	DeleteFantasyLeague(ctx context.Context, id int) error

	// Squad
	GetSquad(ctx context.Context, teamID int) ([]response.SquadMember, error)
	AddSquadMembers(ctx context.Context, teamID int, members []request.SquadMemberCreate) error

	// Match
	CreateMatch(ctx context.Context, req request.MatchCreate) (*response.Match, error)
	UpsertMatch(ctx context.Context, req request.MatchCreate) (*response.Match, error)
	GetMatch(ctx context.Context, id int) (*response.Match, error)
	ListMatches(ctx context.Context, seriesID string) ([]response.Match, error)
	UpdateMatchScorecard(ctx context.Context, seriesID, matchID string, scorecard interface{}) error
	DeleteMatch(ctx context.Context, id int) error

	// Engine
	ComputeLeaderboardForMatch(ctx context.Context, tournamentID int, matchID string) ([]response.LeaderboardEntry, error)
	UpdateLeaderboard(ctx context.Context, tournamentID int, entries []response.LeaderboardEntry) error
	CreateSnapshot(ctx context.Context, tournamentID int, snapshotDate string) (*response.Snapshot, error)
	GetLeaderboard(ctx context.Context, tournamentID int) ([]response.LeaderboardEntry, error)
}

// BL implements Service using repos.
type BL struct {
	TournamentRepo   dl.TournamentRepo
	FantasyLeagueRepo dl.FantasyLeagueRepo
	SquadRepo        dl.SquadRepo
	MatchRepo        dl.MatchRepo
	SnapshotRepo     dl.SnapshotRepo
	LeaderboardRepo  dl.LeaderboardRepo
	PointConfig      PointConfig
}

// NewBL returns a new BL with default point config.
func NewBL(tr dl.TournamentRepo, flr dl.FantasyLeagueRepo, sr dl.SquadRepo, mr dl.MatchRepo, snapr dl.SnapshotRepo, lbr dl.LeaderboardRepo) *BL {
	return &BL{
		TournamentRepo:   tr,
		FantasyLeagueRepo: flr,
		SquadRepo:        sr,
		MatchRepo:        mr,
		SnapshotRepo:     snapr,
		LeaderboardRepo:  lbr,
		PointConfig:      DefaultPointConfig(),
	}
}

// CreateTournament creates a tournament.
func (b *BL) CreateTournament(ctx context.Context, req request.TournamentCreate) (*response.Tournament, error) {
	status := req.Status
	if status == 0 {
		status = 1
	}
	id, err := b.TournamentRepo.Create(ctx, req.SeriesID, req.Name, status)
	if err != nil {
		return nil, err
	}
	return b.GetTournament(ctx, int(id))
}

// GetTournament returns a tournament by ID.
func (b *BL) GetTournament(ctx context.Context, id int) (*response.Tournament, error) {
	row, err := b.TournamentRepo.GetByID(ctx, id)
	if err != nil || row == nil {
		return nil, err
	}
	return &response.Tournament{ID: row.ID, SeriesID: row.SeriesID, Name: row.Name, Status: row.Status}, nil
}

// ListTournaments returns all tournaments.
func (b *BL) ListTournaments(ctx context.Context) ([]response.Tournament, error) {
	rows, err := b.TournamentRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]response.Tournament, len(rows))
	for i := range rows {
		out[i] = response.Tournament{ID: rows[i].ID, SeriesID: rows[i].SeriesID, Name: rows[i].Name, Status: rows[i].Status}
	}
	return out, nil
}

// UpdateTournament updates a tournament.
func (b *BL) UpdateTournament(ctx context.Context, id int, req request.TournamentUpdate) (*response.Tournament, error) {
	if err := b.TournamentRepo.Update(ctx, id, req.Name, req.Status); err != nil {
		return nil, err
	}
	return b.GetTournament(ctx, id)
}

// DeleteTournament deletes a tournament.
func (b *BL) DeleteTournament(ctx context.Context, id int) error {
	return b.TournamentRepo.Delete(ctx, id)
}

// CreateFantasyLeague creates a fantasy league team.
func (b *BL) CreateFantasyLeague(ctx context.Context, req request.FantasyLeagueCreate) (*response.FantasyLeague, error) {
	id, err := b.FantasyLeagueRepo.Create(ctx, req.SeriesID, req.TeamName, req.TeamOwner)
	if err != nil {
		return nil, err
	}
	if len(req.PlayerNames) > 0 {
		for _, p := range req.PlayerNames {
			_, err = b.SquadRepo.Create(ctx, int(id), p, false, false, "")
			if err != nil {
				return nil, err
			}
		}
	}
	return b.GetFantasyLeague(ctx, int(id))
}

// GetFantasyLeague returns a fantasy league by ID.
func (b *BL) GetFantasyLeague(ctx context.Context, id int) (*response.FantasyLeague, error) {
	row, err := b.FantasyLeagueRepo.GetByID(ctx, id)
	if err != nil || row == nil {
		return nil, err
	}
	return &response.FantasyLeague{ID: row.ID, SeriesID: row.SeriesID, TeamName: row.TeamName, TeamOwner: row.TeamOwner}, nil
}

// ListFantasyLeagues returns fantasy leagues for a series.
func (b *BL) ListFantasyLeagues(ctx context.Context, seriesID string) ([]response.FantasyLeague, error) {
	rows, err := b.FantasyLeagueRepo.ListBySeriesID(ctx, seriesID)
	if err != nil {
		return nil, err
	}
	out := make([]response.FantasyLeague, len(rows))
	for i := range rows {
		out[i] = response.FantasyLeague{ID: rows[i].ID, SeriesID: rows[i].SeriesID, TeamName: rows[i].TeamName, TeamOwner: rows[i].TeamOwner}
	}
	return out, nil
}

// UpdateFantasyLeague updates a fantasy league.
func (b *BL) UpdateFantasyLeague(ctx context.Context, id int, req request.FantasyLeagueUpdate) (*response.FantasyLeague, error) {
	if err := b.FantasyLeagueRepo.Update(ctx, id, req.TeamName, req.TeamOwner); err != nil {
		return nil, err
	}
	return b.GetFantasyLeague(ctx, id)
}

// DeleteFantasyLeague deletes a fantasy league.
func (b *BL) DeleteFantasyLeague(ctx context.Context, id int) error {
	return b.FantasyLeagueRepo.Delete(ctx, id)
}

// GetSquad returns squad for a team.
func (b *BL) GetSquad(ctx context.Context, teamID int) ([]response.SquadMember, error) {
	rows, err := b.SquadRepo.GetByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	out := make([]response.SquadMember, len(rows))
	for i := range rows {
		pt := ""
		if rows[i].PlayerType.Valid {
			pt = rows[i].PlayerType.String
		}
		out[i] = response.SquadMember{
			ID: rows[i].ID, TeamID: rows[i].TeamID, PlayerName: rows[i].PlayerName,
			IsCaptain: rows[i].IsCaptain, IsViceCaptain: rows[i].IsViceCaptain, PlayerType: pt,
		}
	}
	return out, nil
}

// AddSquadMembers adds members to a team squad.
func (b *BL) AddSquadMembers(ctx context.Context, teamID int, members []request.SquadMemberCreate) error {
	for _, m := range members {
		_, err := b.SquadRepo.Create(ctx, teamID, m.PlayerName, m.IsCaptain, m.IsViceCaptain, m.PlayerType)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateMatch creates a match.
func (b *BL) CreateMatch(ctx context.Context, req request.MatchCreate) (*response.Match, error) {
	infoJSON, err := json.Marshal(req.MatchInfo)
	if err != nil {
		return nil, err
	}
	id, err := b.MatchRepo.Create(ctx, req.SeriesID, req.MatchID, infoJSON)
	if err != nil {
		return nil, err
	}
	return b.GetMatch(ctx, int(id))
}

// UpsertMatch creates or updates a match.
func (b *BL) UpsertMatch(ctx context.Context, req request.MatchCreate) (*response.Match, error) {
	infoJSON, err := json.Marshal(req.MatchInfo)
	if err != nil {
		return nil, err
	}
	id, err := b.MatchRepo.Upsert(ctx, req.SeriesID, req.MatchID, infoJSON)
	if err != nil {
		return nil, err
	}
	return b.GetMatch(ctx, int(id))
}

// GetMatch returns a match by ID.
func (b *BL) GetMatch(ctx context.Context, id int) (*response.Match, error) {
	row, err := b.MatchRepo.GetByID(ctx, id)
	if err != nil || row == nil {
		return nil, err
	}
	var info, scorecard interface{}
	_ = json.Unmarshal(row.MatchInfo, &info)
	if len(row.Scorecard) > 0 {
		_ = json.Unmarshal(row.Scorecard, &scorecard)
	}
	return &response.Match{ID: row.ID, SeriesID: row.SeriesID, MatchID: row.MatchID, MatchInfo: info, Scorecard: scorecard}, nil
}

// ListMatches returns matches for a series.
func (b *BL) ListMatches(ctx context.Context, seriesID string) ([]response.Match, error) {
	rows, err := b.MatchRepo.ListBySeriesID(ctx, seriesID)
	if err != nil {
		return nil, err
	}
	out := make([]response.Match, len(rows))
	for i := range rows {
		var info, scorecard interface{}
		_ = json.Unmarshal(rows[i].MatchInfo, &info)
		if len(rows[i].Scorecard) > 0 {
			_ = json.Unmarshal(rows[i].Scorecard, &scorecard)
		}
		out[i] = response.Match{ID: rows[i].ID, SeriesID: rows[i].SeriesID, MatchID: rows[i].MatchID, MatchInfo: info, Scorecard: scorecard}
	}
	return out, nil
}

// UpdateMatchScorecard sets the scorecard for a match (e.g. after the match, from CricAPI match_scorecard).
func (b *BL) UpdateMatchScorecard(ctx context.Context, seriesID, matchID string, scorecard interface{}) error {
	scorecardJSON, err := json.Marshal(scorecard)
	if err != nil {
		return err
	}
	return b.MatchRepo.UpdateScorecard(ctx, seriesID, matchID, scorecardJSON)
}

// DeleteMatch deletes a match.
func (b *BL) DeleteMatch(ctx context.Context, id int) error {
	return b.MatchRepo.Delete(ctx, id)
}

// GetLeaderboard returns current leaderboard for a tournament.
func (b *BL) GetLeaderboard(ctx context.Context, tournamentID int) ([]response.LeaderboardEntry, error) {
	rows, err := b.LeaderboardRepo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	out := make([]response.LeaderboardEntry, len(rows))
	for i := range rows {
		out[i] = response.LeaderboardEntry{ID: rows[i].ID, TournamentID: rows[i].TournamentID, TeamID: rows[i].TeamID, Points: rows[i].Points, Rank: rows[i].Rank}
	}
	// Enrich with team names
	for i := range out {
		fl, _ := b.FantasyLeagueRepo.GetByID(ctx, out[i].TeamID)
		if fl != nil {
			out[i].TeamName = fl.TeamName
			out[i].TeamOwner = fl.TeamOwner
		}
	}
	return out, nil
}
