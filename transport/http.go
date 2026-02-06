package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"

	"kirkit/endpoints"
	"kirkit/spec/request"
	"kirkit/spec/response"
)

// NewHTTPHandler returns the root HTTP handler for the API.
func NewHTTPHandler(eps endpoints.Endpoints) http.Handler {
	r := mux.NewRouter()

	// Tournament
	r.HandleFunc("/api/v1/tournaments", decodeCreateTournament(eps.CreateTournament)).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/tournaments", decodeListTournaments(eps.ListTournaments)).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/tournaments/{id}", decodeID(eps.GetTournament)).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/tournaments/{id}", decodeUpdateTournament(eps.UpdateTournament)).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/api/v1/tournaments/{id}", decodeID(eps.DeleteTournament)).Methods(http.MethodDelete)

	// Fantasy League
	r.HandleFunc("/api/v1/fantasy-leagues", decodeCreateFantasyLeague(eps.CreateFantasyLeague)).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/fantasy-leagues", decodeListFantasyLeagues(eps.ListFantasyLeagues)).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/fantasy-leagues/{id}", decodeID(eps.GetFantasyLeague)).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/fantasy-leagues/{id}", decodeUpdateFantasyLeague(eps.UpdateFantasyLeague)).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/api/v1/fantasy-leagues/{id}", decodeID(eps.DeleteFantasyLeague)).Methods(http.MethodDelete)
	r.HandleFunc("/api/v1/fantasy-leagues/{id}/squad", decodeID(eps.GetSquad)).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/fantasy-leagues/{id}/squad", decodeAddSquadMembers(eps.AddSquadMembers)).Methods(http.MethodPost)

	// Match
	r.HandleFunc("/api/v1/matches", decodeCreateMatch(eps.CreateMatch)).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/matches", decodeUpsertMatch(eps.UpsertMatch)).Methods(http.MethodPut)
	r.HandleFunc("/api/v1/matches/{id}", decodeID(eps.GetMatch)).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/matches", decodeListMatches(eps.ListMatches)).Methods(http.MethodGet).Queries("series_id", "{series_id}")
	r.HandleFunc("/api/v1/matches/{id}", decodeID(eps.DeleteMatch)).Methods(http.MethodDelete)

	// Engine
	r.HandleFunc("/api/v1/engine/compute-leaderboard", decodeComputeLeaderboard(eps.ComputeLeaderboard)).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/engine/update-leaderboard", decodeUpdateLeaderboard(eps.UpdateLeaderboard)).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/engine/snapshot", decodeCreateSnapshot(eps.CreateSnapshot)).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/leaderboard/{tournament_id}", decodeTournamentID(eps.GetLeaderboard)).Methods(http.MethodGet)

	return r
}

func encodeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func encodeError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response.APIResponse{Error: err.Error()})
}

func decodeID(ep endpoint.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			encodeError(w, err, http.StatusBadRequest)
			return
		}
		resp, err := ep(r.Context(), id)
		if err != nil {
			encodeError(w, err, http.StatusInternalServerError)
			return
		}
		if resp == nil {
			encodeJSON(w, http.StatusNoContent, nil)
			return
		}
		encodeJSON(w, http.StatusOK, response.APIResponse{Data: resp})
	}
}

func decodeTournamentID(ep endpoint.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["tournament_id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			encodeError(w, err, http.StatusBadRequest)
			return
		}
		resp, err := ep(r.Context(), id)
		if err != nil {
			encodeError(w, err, http.StatusInternalServerError)
			return
		}
		encodeJSON(w, http.StatusOK, response.APIResponse{Data: resp})
	}
}

func decodeCreateTournament(ep endpoint.Endpoint) http.HandlerFunc {
	return decodeJSONBody(ep, func() interface{} { return &request.TournamentCreate{} })
}
func decodeListTournaments(ep endpoint.Endpoint) http.HandlerFunc {
	return invokeEndpoint(ep, nil)
}
func decodeUpdateTournament(ep endpoint.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			encodeError(w, err, http.StatusBadRequest)
			return
		}
		var body request.TournamentUpdate
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			encodeError(w, err, http.StatusBadRequest)
			return
		}
		resp, err := ep(r.Context(), endpoints.UpdateTournamentRequest{ID: id, Body: body})
		if err != nil {
			encodeError(w, err, http.StatusInternalServerError)
			return
		}
		encodeJSON(w, http.StatusOK, response.APIResponse{Data: resp})
	}
}

func decodeCreateFantasyLeague(ep endpoint.Endpoint) http.HandlerFunc {
	return decodeJSONBody(ep, func() interface{} { return &request.FantasyLeagueCreate{} })
}
func decodeListFantasyLeagues(ep endpoint.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		seriesID := r.URL.Query().Get("series_id")
		resp, err := ep(r.Context(), seriesID)
		if err != nil {
			encodeError(w, err, http.StatusInternalServerError)
			return
		}
		encodeJSON(w, http.StatusOK, response.APIResponse{Data: resp})
	}
}
func decodeUpdateFantasyLeague(ep endpoint.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			encodeError(w, err, http.StatusBadRequest)
			return
		}
		var body request.FantasyLeagueUpdate
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			encodeError(w, err, http.StatusBadRequest)
			return
		}
		resp, err := ep(r.Context(), endpoints.UpdateFantasyLeagueRequest{ID: id, Body: body})
		if err != nil {
			encodeError(w, err, http.StatusInternalServerError)
			return
		}
		encodeJSON(w, http.StatusOK, response.APIResponse{Data: resp})
	}
}
func decodeAddSquadMembers(ep endpoint.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		teamID, err := strconv.Atoi(idStr)
		if err != nil {
			encodeError(w, err, http.StatusBadRequest)
			return
		}
		var body struct {
			Members []request.SquadMemberCreate `json:"members"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			encodeError(w, err, http.StatusBadRequest)
			return
		}
		resp, err := ep(r.Context(), endpoints.AddSquadMembersRequest{TeamID: teamID, Members: body.Members})
		if err != nil {
			encodeError(w, err, http.StatusInternalServerError)
			return
		}
		if resp != nil {
			encodeJSON(w, http.StatusOK, response.APIResponse{Data: resp})
		} else {
			encodeJSON(w, http.StatusNoContent, nil)
		}
	}
}

func decodeCreateMatch(ep endpoint.Endpoint) http.HandlerFunc {
	return decodeJSONBody(ep, func() interface{} { return &request.MatchCreate{} })
}
func decodeUpsertMatch(ep endpoint.Endpoint) http.HandlerFunc {
	return decodeJSONBody(ep, func() interface{} { return &request.MatchCreate{} })
}
func decodeListMatches(ep endpoint.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		seriesID := r.URL.Query().Get("series_id")
		resp, err := ep(r.Context(), seriesID)
		if err != nil {
			encodeError(w, err, http.StatusInternalServerError)
			return
		}
		encodeJSON(w, http.StatusOK, response.APIResponse{Data: resp})
	}
}

func decodeComputeLeaderboard(ep endpoint.Endpoint) http.HandlerFunc {
	return decodeJSONBody(ep, func() interface{} { return &request.ComputeLeaderboardRequest{} })
}
func decodeUpdateLeaderboard(ep endpoint.Endpoint) http.HandlerFunc {
	return decodeJSONBody(ep, func() interface{} { return &endpoints.UpdateLeaderboardRequest{} })
}
func decodeCreateSnapshot(ep endpoint.Endpoint) http.HandlerFunc {
	return decodeJSONBody(ep, func() interface{} { return &request.SnapshotRequest{} })
}

func decodeJSONBody(ep endpoint.Endpoint, newReq func() interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := newReq()
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			encodeError(w, err, http.StatusBadRequest)
			return
		}
		resp, err := ep(r.Context(), req)
		if err != nil {
			encodeError(w, err, http.StatusInternalServerError)
			return
		}
		if resp == nil {
			encodeJSON(w, http.StatusNoContent, nil)
			return
		}
		encodeJSON(w, http.StatusOK, response.APIResponse{Data: resp})
	}
}

func invokeEndpoint(ep endpoint.Endpoint, req interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := ep(r.Context(), req)
		if err != nil {
			encodeError(w, err, http.StatusInternalServerError)
			return
		}
		encodeJSON(w, http.StatusOK, response.APIResponse{Data: resp})
	}
}
