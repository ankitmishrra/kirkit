package paths

// API path constants and swagger route tags.
const (
	Base = "/api/v1"

	// Tournament
	Tournaments      = Base + "/tournaments"
	TournamentByID   = Base + "/tournaments/{id}"

	// Fantasy League
	FantasyLeagues   = Base + "/fantasy-leagues"
	FantasyLeagueByID = Base + "/fantasy-leagues/{id}"
	SquadByTeamID    = Base + "/fantasy-leagues/{id}/squad"

	// Match
	Matches    = Base + "/matches"
	MatchByID  = Base + "/matches/{id}"

	// Engine
	ComputeLeaderboard = Base + "/engine/compute-leaderboard"
	UpdateLeaderboard  = Base + "/engine/update-leaderboard"
	Snapshot           = Base + "/engine/snapshot"
	Leaderboard        = Base + "/leaderboard/{tournament_id}"
)
