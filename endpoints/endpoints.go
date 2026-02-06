package endpoints

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"

	"kirkit/bl"
	"kirkit/spec/request"
	"kirkit/spec/response"
)

// Endpoints holds all go-kit endpoints.
type Endpoints struct {
	// Tournament
	CreateTournament endpoint.Endpoint
	GetTournament    endpoint.Endpoint
	ListTournaments  endpoint.Endpoint
	UpdateTournament endpoint.Endpoint
	DeleteTournament endpoint.Endpoint

	// Fantasy League
	CreateFantasyLeague endpoint.Endpoint
	GetFantasyLeague    endpoint.Endpoint
	ListFantasyLeagues  endpoint.Endpoint
	UpdateFantasyLeague endpoint.Endpoint
	DeleteFantasyLeague endpoint.Endpoint
	GetSquad            endpoint.Endpoint
	AddSquadMembers     endpoint.Endpoint

	// Match
	CreateMatch endpoint.Endpoint
	UpsertMatch endpoint.Endpoint
	GetMatch    endpoint.Endpoint
	ListMatches endpoint.Endpoint
	DeleteMatch endpoint.Endpoint

	// Engine
	ComputeLeaderboard endpoint.Endpoint
	UpdateLeaderboard  endpoint.Endpoint
	CreateSnapshot     endpoint.Endpoint
	GetLeaderboard    endpoint.Endpoint
}

// New returns Endpoints wired to the service.
func New(svc bl.Service) Endpoints {
	return Endpoints{
		CreateTournament: makeCreateTournament(svc),
		GetTournament:    makeGetTournament(svc),
		ListTournaments:  makeListTournaments(svc),
		UpdateTournament: makeUpdateTournament(svc),
		DeleteTournament: makeDeleteTournament(svc),

		CreateFantasyLeague: makeCreateFantasyLeague(svc),
		GetFantasyLeague:    makeGetFantasyLeague(svc),
		ListFantasyLeagues:  makeListFantasyLeagues(svc),
		UpdateFantasyLeague: makeUpdateFantasyLeague(svc),
		DeleteFantasyLeague: makeDeleteFantasyLeague(svc),
		GetSquad:            makeGetSquad(svc),
		AddSquadMembers:     makeAddSquadMembers(svc),

		CreateMatch: makeCreateMatch(svc),
		UpsertMatch: makeUpsertMatch(svc),
		GetMatch:    makeGetMatch(svc),
		ListMatches: makeListMatches(svc),
		DeleteMatch: makeDeleteMatch(svc),

		ComputeLeaderboard: makeComputeLeaderboard(svc),
		UpdateLeaderboard:  makeUpdateLeaderboard(svc),
		CreateSnapshot:     makeCreateSnapshot(svc),
		GetLeaderboard:    makeGetLeaderboard(svc),
	}
}

func makeCreateTournament(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.CreateTournament(ctx, *req.(*request.TournamentCreate))
	}
}
func makeGetTournament(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.GetTournament(ctx, req.(int))
	}
}
func makeListTournaments(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.ListTournaments(ctx)
	}
}
func makeUpdateTournament(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		r := req.(UpdateTournamentRequest)
		return svc.UpdateTournament(ctx, r.ID, r.Body)
	}
}
func makeDeleteTournament(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, svc.DeleteTournament(ctx, req.(int))
	}
}

// UpdateTournamentRequest is the request for UpdateTournament endpoint.
type UpdateTournamentRequest struct {
	ID   int
	Body request.TournamentUpdate
}

func makeCreateFantasyLeague(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.CreateFantasyLeague(ctx, *req.(*request.FantasyLeagueCreate))
	}
}
func makeGetFantasyLeague(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.GetFantasyLeague(ctx, req.(int))
	}
}
func makeListFantasyLeagues(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.ListFantasyLeagues(ctx, req.(string))
	}
}
func makeUpdateFantasyLeague(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		r := req.(UpdateFantasyLeagueRequest)
		return svc.UpdateFantasyLeague(ctx, r.ID, r.Body)
	}
}
func makeDeleteFantasyLeague(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, svc.DeleteFantasyLeague(ctx, req.(int))
	}
}
func makeGetSquad(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.GetSquad(ctx, req.(int))
	}
}
func makeAddSquadMembers(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		r := req.(AddSquadMembersRequest)
		return nil, svc.AddSquadMembers(ctx, r.TeamID, r.Members)
	}
}

// UpdateFantasyLeagueRequest is the request for UpdateFantasyLeague endpoint.
type UpdateFantasyLeagueRequest struct {
	ID   int
	Body request.FantasyLeagueUpdate
}

// AddSquadMembersRequest is the request for AddSquadMembers endpoint.
type AddSquadMembersRequest struct {
	TeamID  int
	Members []request.SquadMemberCreate
}

func makeCreateMatch(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.CreateMatch(ctx, *req.(*request.MatchCreate))
	}
}
func makeUpsertMatch(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.UpsertMatch(ctx, *req.(*request.MatchCreate))
	}
}
func makeGetMatch(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.GetMatch(ctx, req.(int))
	}
}
func makeListMatches(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.ListMatches(ctx, req.(string))
	}
}
func makeDeleteMatch(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, svc.DeleteMatch(ctx, req.(int))
	}
}

func makeComputeLeaderboard(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		r := req.(*request.ComputeLeaderboardRequest)
		return svc.ComputeLeaderboardForMatch(ctx, r.TournamentID, r.MatchID)
	}
}
func makeUpdateLeaderboard(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		r := req.(*UpdateLeaderboardRequest)
		return nil, svc.UpdateLeaderboard(ctx, r.TournamentID, r.Entries)
	}
}
func makeCreateSnapshot(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		r := req.(*request.SnapshotRequest)
		date := r.SnapshotDate
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		return svc.CreateSnapshot(ctx, r.TournamentID, date)
	}
}
func makeGetLeaderboard(svc bl.Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return svc.GetLeaderboard(ctx, req.(int))
	}
}

// UpdateLeaderboardRequest is the request for UpdateLeaderboard endpoint.
type UpdateLeaderboardRequest struct {
	TournamentID int                          `json:"tournament_id"`
	Entries      []response.LeaderboardEntry `json:"entries"`
}

