# Kirkit Web – Leaderboard & Stats

Single-page React app that shows the fantasy cricket **leaderboard** and **stats** derived from the Kirkit API (and thus from the DB schema: tournament, leaderboard, game, fantasy_league).

## Prerequisites

- Node 18+
- Kirkit API running at `http://localhost:8080` (e.g. `make run` from repo root)

## Setup

```bash
cd web
npm install
```

## Run

```bash
npm run dev
```

Open http://localhost:3000. The dev server proxies `/api` to `http://localhost:8080`, so the app talks to the Kirkit backend without CORS.

## Build

```bash
npm run build
npm run preview   # serve dist/
```

For production, set `VITE_API_BASE` to your API base URL (e.g. `https://api.example.com`) if the app is not served behind the same host as the API.

## What it shows

- **Tournament selector** – pick a tournament (from `tournament` table).
- **Stats** – derived from APIs:
  - Teams in leaderboard (count from `leaderboard`).
  - Total points (sum of `leaderboard.points`).
  - Matches in series (count from `game` by `series_id`).
  - Fantasy teams (count from `fantasy_league` by `series_id`).
  - Leader (rank 1 team and points).
- **Leaderboard** – table of rank, team name, owner, points (from `leaderboard`).
