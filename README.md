# Kirkit – Fantasy Cricket Backend

Backend for a friends-only fantasy cricket league. Uses [cricapi](https://www.cricapi.com/) for match/series data and runs daily to compute leaderboards from yesterday’s matches.

## Phase 1 (current)

- **No auth** – open backend, intended to run locally.
- **Input**: CSV with `team_name`, `team_owner`, `players` (pipe-separated).
- **DB**: Docker-based MySQL with a persistent volume.
- **APIs**: Tournament CRUD, Fantasy League CRUD, Match CRUD, engine (compute/update leaderboard, snapshot).

## Prerequisites

- Go 1.23+
- Docker & Docker Compose (for MySQL)

## How to run the application

1. **Install dependencies**

   ```bash
   go mod tidy
   ```

2. **Start MySQL**

   ```bash
   make up
   # or: docker-compose up -d mysql
   ```

   Data is stored in the `kirkit_mysql_data` volume so it survives container restarts.

3. **Run DB migrations**

   ```bash
   make upgrade
   # or: go run ./cmd/upgrade
   ```

   Migrations are Go functions in `upgrade/migrations/`. The runner ensures the `version` table exists, then runs each migration with version &gt; current.

4. **Start the API server**

   ```bash
   make run
   # or: go run ./cmd/main
   ```

   Server listens on **http://localhost:8080** (override with `PORT`). Set `MYSQL_DSN` if your DB is elsewhere.

5. **Web app (optional)** – single-page React app for leaderboard and stats:

   ```bash
   cd web && npm install && npm run dev
   # or: make web  (from repo root after npm install in web/)
   ```

   Open **http://localhost:3000**. The app shows tournament selector, stats (derived from tournament, leaderboard, game, fantasy_league), and leaderboard table. See [web/README.md](web/README.md).

## Daily leaderboard update (fetch last match scorecards)

After matches are played, the leaderboard should be updated using **scorecard** data (runs, wickets, catches, etc.). The app does not poll CricAPI on its own; you run a **daily job** that:

1. For each **live tournament** (status = ongoing) in the DB, lists **matches** (games) for that series that **do not yet have a scorecard**.
2. For each such match, **fetches** the scorecard from CricAPI **match_scorecard** and saves it into `game.scorecard`.
3. **Computes** the leaderboard for that match (adds points to running total) and **persists** the updated leaderboard.
4. **Emails** the current leaderboard for all live tournaments to the configured address (default: ankitmishra7991@gmail.com).

**Run the daily sync once per day** (e.g. via cron):

```bash
make daily
# or: CRICAPI_KEY=yourkey go run ./cmd/daily
```

Requires **CRICAPI_KEY** (and optional **MYSQL_DSN**). The job processes all tournaments; for each match missing a scorecard it calls the API, updates the DB, then recomputes and saves the leaderboard. Matches are processed in table order; each match’s points are added to the cumulative leaderboard.

**Cron example** (run at 6 AM daily):

```cron
0 6 * * * cd /path/to/kirkit && make daily
```

**Email (daily job):** To send the leaderboard after the sync, set in `.env` or env:

| Variable     | Description |
|-------------|--------------|
| `SMTP_HOST` | e.g. `smtp.gmail.com` |
| `SMTP_PORT` | e.g. `587` |
| `SMTP_USER` | SMTP username |
| `SMTP_PASS` | SMTP password (Gmail: use an [App Password](https://support.google.com/accounts/answer/185833)) |
| `SMTP_FROM` | From address (e.g. your Gmail) |
| `EMAIL_TO`  | Recipient (default: `ankitmishra7991@gmail.com`) |

If SMTP vars are not set, the job still runs but skips sending email and logs "email skipped".

## Environment

| Variable     | Default (example)                          | Description              |
|-------------|---------------------------------------------|--------------------------|
| `MYSQL_DSN` | `kirkit:kirkit_secret@tcp(localhost:3306)/kirkit?parseTime=true` | MySQL connection string  |
| `PORT`      | `8080`                                      | HTTP server port         |
| `CRICAPI_KEY` | (none)                                    | cricapi API key          |

## API Overview

Base path: `/api/v1`.

- **Tournaments**: `POST/GET /tournaments`, `GET/PUT/PATCH/DELETE /tournaments/{id}`
- **Fantasy leagues**: `POST/GET /fantasy-leagues`, `GET/PUT/PATCH/DELETE /fantasy-leagues/{id}`, `GET/POST /fantasy-leagues/{id}/squad`
- **Matches**: `POST /matches`, `PUT /matches` (upsert), `GET /matches?series_id=...`, `GET/DELETE /matches/{id}`
- **Engine**: `POST /engine/compute-leaderboard`, `POST /engine/update-leaderboard`, `POST /engine/snapshot`
- **Leaderboard**: `GET /leaderboard/{tournament_id}`

## DB and migrations

- **Squirrel** is used for all MySQL queries in `dl/mysql` (parameterized; SQL injection–safe). DDL in migrations uses constant strings only.
- **Migrations** are Go files under `upgrade/migrations/` (e.g. `001_initial_schema.go`, `002_initial_tournament.go`). Version is stored in the `version` table; only migrations with version &gt; current are run.

Schema:

- `version` – single row, current schema version (integer)
- `tournament` – id, series_id, name, status (1=ongoing, 2=done)
- `fantasy_league` – id, series_id, team_name, team_owner
- `squad` – id, team_id, player_name, is_captain, is_vice_captain, player_type
- `game` – id, series_id, match_id, match_info (JSON from cricapi; table named game to avoid MySQL reserved word match)
- `snapshot` – date, tournament_id, leaderboard_json (daily snapshots)
- `leaderboard` – tournament_id, team_id, points, rank

Default tournament: **Men’s T20 World Cup** with `series_id = c4ca5cd5-e25c-4d83-bb77-2d193d93475a`.

## Point rules (PointCalculation.xlsx; configurable in BL)

- **Batting:** Runs 1, Fours 4, Sixes 6, Duck penalty −2, strike rate bonus/penalty (min 10 balls), runs bonus (25–49 +4, 50–74 +8, 75–99 +12, ≥100 +16).
- **Bowling:** Wicket 25, dot ball 1, wicket milestones (3/4/5 wkts), maiden 8, economy bonus/penalty (min 2 overs).
- **Fielding:** Catch 8, Stumping 12, Run out 6, catch bonus +4 if ≥3 catches.
- **Captain** 2×, **Vice-captain** 1.5×.

Leaderboard is computed from match scorecard (in `game.scorecard` or `match_info`); see [docs/POINT_CALCULATION_GAP.md](docs/POINT_CALCULATION_GAP.md).

## Tests

```bash
# Unit tests (no DB)
go test ./bl/ ./endpoints/ ./upgrade/ -v

# Component tests (MySQL must be running)
MYSQL_DSN="kirkit:kirkit_secret@tcp(localhost:3306)/kirkit?parseTime=true" go test ./dl/mysql/ ./transport/ -v -run TestComponent

# E2E: tournament → fantasy leagues + squads → match with scorecard → compute & update leaderboard → assert
go test ./transport/ -v -run TestE2E_Leaderboard
```

**Leaderboard calculation logic** is documented in [docs/LEADERBOARD_LOGIC.md](docs/LEADERBOARD_LOGIC.md). The e2e test prints a short summary and asserts Team A (240 pts) and Team B (138 pts) from the example scorecard.

## Project layout

- `cmd/main` – REST server entrypoint  
- `cmd/upgrade` – DB upgrade runner (install mode)  
- `upgrade/` – runner and Go migrations (`migrations/001_initial_schema.go`, `002_initial_tournament.go`)  
- `spec/` – request/response, params, paths  
- `dl/` – data layer (repo interfaces + `dl/mysql` impl)  
- `bl/` – business logic + engine (leaderboard, snapshot)  
- `endpoints/` – go-kit endpoints  
- `transport/` – HTTP handlers (gorilla/mux)  
- `middleware/` – logging, recover  
- `sdk/cricapi/` – cricapi client (series_info, match_info)

Phase 2 will add Dockerization of the application.
