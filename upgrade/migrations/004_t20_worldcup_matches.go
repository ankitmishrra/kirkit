package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"

	"kirkit/sdk/cricapi"
)

// seriesIDForT20Matches is the series used to fetch match list from CricAPI series_info.
// Example: https://api.cricapi.com/v1/series_info?apikey=...&id=<series_id>
const seriesIDForT20Matches = "0cdf6736-ad9b-4e95-a647-5ee3a99c5510"

// Up004T20WorldCupMatches fetches match list from CricAPI series_info for the given series
// and inserts them into the game table. It also creates fantasy league teams and their squads.
// If CRICAPI_KEY is not set, the migration still creates fantasy league teams (matches can be filled later).
func Up004T20WorldCupMatches(ctx context.Context, db *sql.DB) error {
	// Create fantasy league teams and squads
	if err := createFantasyLeagueTeams(ctx, db); err != nil {
		return err
	}

	apiKey := os.Getenv("CRICAPI_KEY")
	if apiKey == "" {
		return nil // no-op when key not set (e.g. CI); matches can be filled later
	}

	client := cricapi.NewClient(apiKey, os.Getenv("CRICAPI_BASE"))

	stmt, err := db.PrepareContext(ctx,
		`INSERT INTO game (series_id, match_id, match_info) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE match_info = VALUES(match_info), updated_at = CURRENT_TIMESTAMP`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	const pageSize = 20
	for offset := 0; ; offset += pageSize {
		resp, err := client.SeriesInfo(ctx, seriesIDForT20Matches, offset)
		if err != nil {
			return err
		}
		if resp.Data == nil {
			break
		}
		// Ensure tournament exists for this series (so it appears in app and matches can be listed).
		if offset == 0 {
			seriesName := "T20 Series"
			if resp.Data.Info != nil && resp.Data.Info.Name != "" {
				seriesName = resp.Data.Info.Name
			}
			_, err = db.ExecContext(ctx,
				`INSERT INTO tournament (series_id, name, status) VALUES (?, ?, ?)
				 ON DUPLICATE KEY UPDATE name = VALUES(name), status = VALUES(status)`,
				seriesIDForT20Matches, seriesName, 1)
			if err != nil {
				return err
			}
		}
		list := resp.Data.MatchListOrMatches()
		if len(list) == 0 {
			break
		}
		for _, m := range list {
			infoJSON, err := json.Marshal(m)
			if err != nil {
				return err
			}
			_, err = stmt.ExecContext(ctx, seriesIDForT20Matches, m.ID, infoJSON)
			if err != nil {
				return err
			}
		}
		if len(list) < pageSize {
			break
		}
	}
	return nil
}

// createFantasyLeagueTeams creates fantasy league teams and their squads for the T20 World Cup.
func createFantasyLeagueTeams(ctx context.Context, db *sql.DB) error {
	// Team data: owner name (which is also the team name), captain, vice captain, and full squad
	teams := []struct {
		owner       string
		captain     string
		viceCaptain string
		squad       string
	}{
		{"Anand/Sagar", "Hardik Pandya", "Mitchell Marsh", "Harry Tector|Khawaja Nafay|Sediqullah Atal|Mitchell Marsh|Will Jacks|George Linde|Hardik Pandya|Johnson Charles|Kusal Janith Perera|Jasprit Bumrah|Kwena Maphaka|Noor Ahmad|Jacob Duffy|Dushmantha Chameera|Shamar Joseph"},
		{"Anupam/Ruturaj", "Suryakumar Yadav", "Salman Agha", "Ibrahim Zadran|Daryl Mitchell|Suryakumar Yadav|Finn Allen|Sherfane Rutherford|Sikandar Raza|Salman Agha|Shivam Dube|Corbin Bosch|Dasun Shanaka|Kusal Mendis|Shaheen Afridi|Xavier Bartlett|Luke Wood|Fazalhaq Farooqi"},
		{"Arpit", "Saim Ayub", "Mitchell Santner", "Paul Stirling|Rovman Powell|Jason Smith|Dewald Brevis|Saim Ayub|Mitchell Santner|Dunith Wellalage|Marcus Stoinis|Rahmanullah Gurbaz|Rehan Ahmed|Adam Zampa|Arshdeep Singh|Abrar Ahmed|Gudakesh Motie|Jamie Overton"},
		{"Mayur", "Ishan Kishan", "Tim Seifert", "David Miller|Tim David|Kamil Mishara|Sahibzada Farhan|Roston Chase|Sam Curran|Cameron Green|Shadab Khan|Ishan Kishan|Phil Salt|Tim Seifert|Ish Sodhi|Kuldeep Yadav|Mujeeb Ur Rahman|Kagiso Rabada"},
		{"Pradeep", "Axar Patel", "Aiden Markram", "Max ODowd|Brian Bennett|Fakhar Zaman|Aiden Markram|Romario Shepherd|Axar Patel|Rachin Ravindra|Azmatullah Omarzai|Kamindu Mendis|Shai Hope|Jos Buttler|Nathan Ellis|Matt Henry|Keshav Maharaj|Varun Chakaravarthy"},
		{"Vaibhav", "Abhishek Sharma", "Shimron Hetmyer", "Shimron Hetmyer|Rinku Singh|Abhishek Sharma|Babar Azam|Glenn Maxwell|Marco Jansen|Jacob Bethell|Devon Conway|Mohammad Ishaq|Maheesh Theekshana|Naseem Shah|Anrich Nortje|Kyle Jamieson|Jofra Archer|Josh Hazlewood"},
		{"Vikram", "Travis Head", "Ryan Rickelton", "Tilak Varma|Matt Renshaw|Tristan Stubbs|Harry Brook|Travis Head|Bas de Leede|Mohammad Nabi|Jason Holder|Ryan Rickelton|Tom Banton|Eshan Malinga|Rashid Khan|Lockie Ferguson|Matheesha Pathirana|Harshit Rana"},
		{"Vishal", "Quinton de Kock", "Josh Inglis", "Brandon King|Pathum Nissanka|Darwish Rasooli|Ben Duckett|Glenn Phillips|Nadeem Khan|Brandon McMullen|Wanindu Hasaranga|Sanju Samson|Josh Inglis|Quinton de Kock|Lungi Ngidi|Ruben Trumpelmann|Adil Rashid|Akeal Hosein"},
	}

	// Prepare statements
	teamStmt, err := db.PrepareContext(ctx,
		`INSERT INTO fantasy_league (series_id, team_name, team_owner) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE team_owner = VALUES(team_owner), updated_at = CURRENT_TIMESTAMP`)
	if err != nil {
		return err
	}
	defer teamStmt.Close()

	squadStmt, err := db.PrepareContext(ctx,
		`INSERT INTO squad (team_id, player_name, is_captain, is_vice_captain) VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE is_captain = VALUES(is_captain), is_vice_captain = VALUES(is_vice_captain)`)
	if err != nil {
		return err
	}
	defer squadStmt.Close()

	// Insert teams and their squads
	for _, team := range teams {
		// Insert team (team_name = team_owner as specified)
		_, err := teamStmt.ExecContext(ctx, seriesIDForT20Matches, team.owner, team.owner)
		if err != nil {
			return err
		}

		// Get the team ID (works for both new inserts and updates)
		var teamID int64
		err = db.QueryRowContext(ctx,
			`SELECT id FROM fantasy_league WHERE series_id = ? AND team_name = ?`,
			seriesIDForT20Matches, team.owner).Scan(&teamID)
		if err != nil {
			return err
		}

		// Parse squad players (pipe-separated)
		players := strings.Split(team.squad, "|")
		for _, player := range players {
			player = strings.TrimSpace(player)
			if player == "" {
				continue
			}

			// Determine if this player is captain or vice captain
			isCaptain := 0
			isViceCaptain := 0
			if player == team.captain {
				isCaptain = 1
			}
			if player == team.viceCaptain {
				isViceCaptain = 1
			}

			// Insert squad member
			_, err = squadStmt.ExecContext(ctx, teamID, player, isCaptain, isViceCaptain)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
