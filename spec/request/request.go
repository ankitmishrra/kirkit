package request

// TournamentCreate is the request body for creating a tournament.
type TournamentCreate struct {
	SeriesID string `json:"series_id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Status   int    `json:"status"` // 1=ongoing, 2=done
}

// TournamentUpdate is the request body for updating a tournament.
type TournamentUpdate struct {
	Name   *string `json:"name,omitempty"`
	Status *int    `json:"status,omitempty"`
}

// FantasyLeagueCreate is the request body for creating a fantasy league team.
type FantasyLeagueCreate struct {
	SeriesID   string   `json:"series_id" validate:"required"`
	TeamName   string   `json:"team_name" validate:"required"`
	TeamOwner  string   `json:"team_owner" validate:"required"`
	PlayerNames []string `json:"player_names,omitempty"` // optional, can add via squad later
}

// FantasyLeagueUpdate is the request body for updating a fantasy league.
type FantasyLeagueUpdate struct {
	TeamName  *string  `json:"team_name,omitempty"`
	TeamOwner *string  `json:"team_owner,omitempty"`
}

// SquadMemberCreate is a single squad member.
type SquadMemberCreate struct {
	PlayerName    string `json:"player_name" validate:"required"`
	IsCaptain     bool   `json:"is_captain"`
	IsViceCaptain bool   `json:"is_vice_captain"`
	PlayerType    string `json:"player_type,omitempty"`
}

// MatchCreate is the request for creating/upserting a match (e.g. from cricapi).
type MatchCreate struct {
	SeriesID  string          `json:"series_id" validate:"required"`
	MatchID   string          `json:"match_id" validate:"required"`
	MatchInfo interface{}     `json:"match_info" validate:"required"` // JSON from cricapi match_info
}

// UpdateScorecardRequest is the request for updating a match's scorecard (after the match).
type UpdateScorecardRequest struct {
	SeriesID  string      `json:"series_id" validate:"required"`
	MatchID   string      `json:"match_id" validate:"required"`
	Scorecard interface{} `json:"scorecard" validate:"required"` // JSON from CricAPI match_scorecard data
}

// ComputeLeaderboardRequest is the request for computing leaderboard for a match.
type ComputeLeaderboardRequest struct {
	TournamentID int    `json:"tournament_id" validate:"required"`
	MatchID      string `json:"match_id" validate:"required"`
}

// SnapshotRequest is the request for creating a daily snapshot.
type SnapshotRequest struct {
	TournamentID int    `json:"tournament_id" validate:"required"`
	SnapshotDate string `json:"snapshot_date"` // YYYY-MM-DD, default today
}
