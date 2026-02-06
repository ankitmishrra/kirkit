package response

// Tournament is the API response for a tournament.
type Tournament struct {
	ID       int    `json:"id"`
	SeriesID string `json:"series_id"`
	Name     string `json:"name"`
	Status   int    `json:"status"`
}

// FantasyLeague is the API response for a fantasy league team.
type FantasyLeague struct {
	ID        int    `json:"id"`
	SeriesID  string `json:"series_id"`
	TeamName  string `json:"team_name"`
	TeamOwner string `json:"team_owner"`
}

// SquadMember is a squad member in API response.
type SquadMember struct {
	ID            int    `json:"id"`
	TeamID        int    `json:"team_id"`
	PlayerName    string `json:"player_name"`
	IsCaptain     bool   `json:"is_captain"`
	IsViceCaptain bool   `json:"is_vice_captain"`
	PlayerType    string `json:"player_type,omitempty"`
}

// Match is the API response for a match.
type Match struct {
	ID        int         `json:"id"`
	SeriesID  string      `json:"series_id"`
	MatchID   string      `json:"match_id"`
	MatchInfo interface{} `json:"match_info"`
	Scorecard interface{} `json:"scorecard,omitempty"` // populated after match (e.g. from CricAPI match_scorecard)
}

// LeaderboardEntry is one row in the leaderboard.
type LeaderboardEntry struct {
	ID           int    `json:"id"`
	TournamentID int    `json:"tournament_id"`
	TeamID       int    `json:"team_id"`
	TeamName     string `json:"team_name,omitempty"`
	TeamOwner    string `json:"team_owner,omitempty"`
	Points       int    `json:"points"`
	Rank         int    `json:"rank"`
}

// Snapshot is the API response for a snapshot.
type Snapshot struct {
	ID             int         `json:"id"`
	SnapshotDate   string      `json:"snapshot_date"`
	TournamentID   int         `json:"tournament_id"`
	LeaderboardJSON interface{} `json:"leaderboard_json"`
}

// APIResponse is the standard envelope for success.
type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error string     `json:"error,omitempty"`
}
