// Package cricapi provides a client for cricapi.com (series_info, match_info, match_scorecard).
package cricapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Client for cricapi.
type Client struct {
	APIKey string
	Base   string
	HTTP   *http.Client
}

// NewClient returns a cricapi client.
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.cricapi.com/v1"
	}
	return &Client{APIKey: apiKey, Base: baseURL, HTTP: http.DefaultClient}
}

// SeriesInfoResponse is the response from series_info API.
type SeriesInfoResponse struct {
	Data   *SeriesInfoData `json:"data"`
	Status string          `json:"status"`
}

// SeriesInfoData contains info and match list (API may use "matchList" or "matches").
type SeriesInfoData struct {
	Info      *SeriesInfo   `json:"info"`
	MatchList []MatchSummary `json:"matchList"`
	Matches   []MatchSummary `json:"matches"`
}

// MatchList returns all matches from the response (matchList or matches, whichever is present).
func (d *SeriesInfoData) MatchListOrMatches() []MatchSummary {
	if d == nil {
		return nil
	}
	if len(d.MatchList) > 0 {
		return d.MatchList
	}
	return d.Matches
}

// SeriesInfo is the series metadata.
type SeriesInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"startdate"`
	EndDate   string `json:"enddate"`
	T20       int    `json:"t20"`
}

// MatchSummary is a match in the series list.
type MatchSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MatchType string `json:"matchType"`
	Status   string `json:"status"`
	Date     string `json:"date"`
	Teams    []string `json:"teams"`
	MatchStarted bool `json:"matchStarted"`
	MatchEnded   bool `json:"matchEnded"`
}

// MatchInfoResponse is the response from match_info API.
type MatchInfoResponse struct {
	Data   interface{} `json:"data"`
	Status string      `json:"status"`
}

// SeriesInfo fetches series info by series ID. Use offset for pagination (0-based).
// The API returns a limited number of matches per call; call repeatedly with
// increasing offset until matchList is empty to get all matches.
func (c *Client) SeriesInfo(ctx context.Context, seriesID string, offset int) (*SeriesInfoResponse, error) {
	url := fmt.Sprintf("%s/series_info?apikey=%s&id=%s&offset=%d", c.Base, c.APIKey, seriesID, offset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var out SeriesInfoResponse
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("series_info decode: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("series_info HTTP %d", resp.StatusCode)
	}
	if strings.EqualFold(out.Status, "fail") || strings.EqualFold(out.Status, "error") {
		return nil, fmt.Errorf("series_info API status=%s (check series id and api key)", out.Status)
	}
	return &out, nil
}

// MatchInfo fetches match info by match ID. Returns the raw "data" object for storage.
func (c *Client) MatchInfo(ctx context.Context, matchID string) (interface{}, error) {
	url := fmt.Sprintf("%s/match_info?apikey=%s&id=%s", c.Base, c.APIKey, matchID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Data   interface{} `json:"data"`
		Status string      `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// SeriesSquadResponse is the response from series_squad API.
type SeriesSquadResponse struct {
	Data   *SeriesSquadData `json:"data"`
	Status string           `json:"status"`
}

type SeriesSquadData struct {
	Squad []TeamSquad `json:"squad"`
}

type TeamSquad struct {
	Name    string              `json:"name"`
	Country string              `json:"country"`
	Players []SeriesSquadPlayer `json:"players"`
}

type SeriesSquadPlayer struct {
	Name        string `json:"name"`
	Pid         int64  `json:"pid"`
	Country     string `json:"country"`
	Nationality string `json:"nationality"`
}

func (p *SeriesSquadPlayer) NationalityOrCountry(teamName string) string {
	if p.Nationality != "" {
		return p.Nationality
	}
	if p.Country != "" {
		return p.Country
	}
	return teamName
}

// SeriesSquad fetches series squad by series ID.
func (c *Client) SeriesSquad(ctx context.Context, seriesID string) (*SeriesSquadResponse, error) {
	url := fmt.Sprintf("%s/series_squad?apikey=%s&id=%s", c.Base, c.APIKey, seriesID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var out SeriesSquadResponse
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("series_squad decode: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("series_squad HTTP %d", resp.StatusCode)
	}
	if strings.EqualFold(out.Status, "fail") || strings.EqualFold(out.Status, "error") {
		return nil, fmt.Errorf("series_squad API status=%s (check series id and api key)", out.Status)
	}
	return &out, nil
}

// MatchScorecard fetches match scorecard by match ID (player-level batting, bowling, catching).
// Returns the raw "data" object; use this payload in game.match_info for leaderboard computation.
// Response shape: data.id, data.name, data.scorecard (array of innings with batting[], bowling[], catching[]).
func (c *Client) MatchScorecard(ctx context.Context, matchID string) (interface{}, error) {
	url := fmt.Sprintf("%s/match_scorecard?apikey=%s&id=%s", c.Base, c.APIKey, matchID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Data   interface{} `json:"data"`
		Status string      `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Data, nil
}
