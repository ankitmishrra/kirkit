package dl

import (
	"context"
	"database/sql"
)

// TournamentRow represents a tournament row from DB.
type TournamentRow struct {
	ID       int
	SeriesID string
	Name     string
	Status   int
}

// FantasyLeagueRow represents a fantasy_league row.
type FantasyLeagueRow struct {
	ID        int
	SeriesID  string
	TeamName  string
	TeamOwner string
}

// SquadRow represents a squad row.
type SquadRow struct {
	ID            int
	TeamID        int
	PlayerName    string
	IsCaptain     bool
	IsViceCaptain bool
	PlayerType    sql.NullString
}

// MatchRow represents a match row.
type MatchRow struct {
	ID        int
	SeriesID  string
	MatchID   string
	MatchInfo []byte // JSON
	Scorecard []byte // JSON, nullable; updated after the match (e.g. from CricAPI match_scorecard)
}

// SnapshotRow represents a snapshot row.
type SnapshotRow struct {
	ID              int
	SnapshotDate    string
	TournamentID    int
	LeaderboardJSON []byte
}

// LeaderboardRow represents a leaderboard row.
type LeaderboardRow struct {
	ID           int
	TournamentID int
	TeamID       int
	Points       int
	Rank         int
}

// TournamentRepo defines tournament data access.
type TournamentRepo interface {
	Create(ctx context.Context, seriesID, name string, status int) (int64, error)
	GetByID(ctx context.Context, id int) (*TournamentRow, error)
	GetBySeriesID(ctx context.Context, seriesID string) (*TournamentRow, error)
	List(ctx context.Context) ([]TournamentRow, error)
	Update(ctx context.Context, id int, name *string, status *int) error
	Delete(ctx context.Context, id int) error
}

// FantasyLeagueRepo defines fantasy league data access.
type FantasyLeagueRepo interface {
	Create(ctx context.Context, seriesID, teamName, teamOwner string) (int64, error)
	GetByID(ctx context.Context, id int) (*FantasyLeagueRow, error)
	ListBySeriesID(ctx context.Context, seriesID string) ([]FantasyLeagueRow, error)
	Update(ctx context.Context, id int, teamName, teamOwner *string) error
	Delete(ctx context.Context, id int) error
}

// SquadRepo defines squad data access.
type SquadRepo interface {
	Create(ctx context.Context, teamID int, playerName string, isCaptain, isViceCaptain bool, playerType string) (int64, error)
	GetByTeamID(ctx context.Context, teamID int) ([]SquadRow, error)
	DeleteByTeamID(ctx context.Context, teamID int) error
	DeleteByTeamIDAndPlayer(ctx context.Context, teamID int, playerName string) error
}

// MatchRepo defines match data access.
type MatchRepo interface {
	Create(ctx context.Context, seriesID, matchID string, matchInfo []byte) (int64, error)
	Upsert(ctx context.Context, seriesID, matchID string, matchInfo []byte) (int64, error)
	UpdateScorecard(ctx context.Context, seriesID, matchID string, scorecard []byte) error
	GetByID(ctx context.Context, id int) (*MatchRow, error)
	GetByMatchID(ctx context.Context, seriesID, matchID string) (*MatchRow, error)
	ListBySeriesID(ctx context.Context, seriesID string) ([]MatchRow, error)
	Delete(ctx context.Context, id int) error
}

// SnapshotRepo defines snapshot data access.
type SnapshotRepo interface {
	Create(ctx context.Context, snapshotDate string, tournamentID int, leaderboardJSON []byte) (int64, error)
	GetByDateAndTournament(ctx context.Context, snapshotDate string, tournamentID int) (*SnapshotRow, error)
	ListByTournament(ctx context.Context, tournamentID int) ([]SnapshotRow, error)
}

// LeaderboardRepo defines leaderboard data access.
type LeaderboardRepo interface {
	Upsert(ctx context.Context, tournamentID, teamID, points, rank int) (int64, error)
	GetByTournamentID(ctx context.Context, tournamentID int) ([]LeaderboardRow, error)
	DeleteByTournamentID(ctx context.Context, tournamentID int) error
}

// VersionRepo for DB version.
type VersionRepo interface {
	Get(ctx context.Context) (int, error)
	Set(ctx context.Context, version int) error
}
