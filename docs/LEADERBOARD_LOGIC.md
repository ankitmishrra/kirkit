# Leaderboard Calculation Logic

## Overview

The leaderboard is computed **per tournament** and **per match**. When a match is completed, we:

1. Load the **match_info** (JSON) for that match, which may contain a **scorecard** with per-player stats.
2. For each **fantasy team** in that tournament (same `series_id`), we look at its **squad**.
3. For each **player** in the squad who appears in the match scorecard, we compute **points** from their stats, then apply **captain/vice-captain multipliers**.
4. We add these **match points** to the team’s **existing leaderboard total** (from previous matches).
5. We **sort** all teams by total points (descending) and assign **ranks**.

---

## Step 1: Player points from match (scorecard)

From the match’s **scorecard** we read per-player stats and apply **PointCalculation.xlsx** rules (see `bl.PointConfig` and `bl/engine.go`).

**Batting:** Runs (1), Fours (4), Sixes (6), Duck penalty (−2 if out for 0), strike rate bonus/penalty (min 10 balls), runs bonus (25–49 +4, 50–74 +8, 75–99 +12, ≥100 +16).

**Bowling:** Wickets (25), dot balls (1), wicket milestones (3/4/5 wkts), maiden overs (8), economy bonus/penalty (min 2 overs), LBW/Bowled bonus.

**Fielding:** Catch (8), Stumping (12), Run out (6), catch bonus (+4 if ≥3 catches).

Player is identified by `name` in the scorecard (batsman, bowler, catcher). If the match has no scorecard, every player gets **0** points.

---

## Step 2: Captain and vice-captain multipliers

Each fantasy squad marks one **captain** and one **vice-captain**:

- **Captain:** points for that match are multiplied by **2**.
- **Vice-captain:** points are multiplied by **1.5** (stored as 3/2; integer division: `(pts × 3) / 2`).
- **Others:** multiplier **1** (no change).

So for a player with **50** raw points in the match:

- If **captain**  → **100** points for that match.
- If **vice-captain** → **75** points.
- If **normal** → **50** points.

---

## Step 3: Team total for the match

For each fantasy team we:

1. Take its **squad** (list of players with captain/vice flags).
2. For each squad member, look up their **player_points** from the match scorecard (by name).
3. Apply the **multiplier** (captain 2×, vice 1.5×, else 1×).
4. **Sum** these contributions → that team’s **match points** for this match.

---

## Step 4: Running total and rank

- We load the **current leaderboard** for the tournament (points from previous matches).
- For each team we do: **new_total = current_total + match_points** (from step 3).
- Teams that had no entry before get **current_total = 0**.
- We **sort** teams by **new_total** descending.
- **Rank 1** = highest total, **Rank 2** = second, etc.

The result is the **new leaderboard** (list of entries with `tournament_id`, `team_id`, `points`, `rank`). Persisting it is done via **UpdateLeaderboard** (replaces all leaderboard rows for that tournament).

---

## Example (used in e2e test)

**Tournament:** T20 World Cup (`series_id = c4ca5cd5-e25c-4d83-bb77-2d193d93475a`).

**Fantasy teams:**

- **Team A:** Player1 (captain), Player2 (vice), Player3  
- **Team B:** Player1, Player2, Player4  

**Match scorecard:**

| Player   | r (runs) | w (wickets) | catch | runOut | Raw points      |
|----------|----------|-------------|--------|--------|-----------------|
| Player1  | 50       | 0           | 0      | 0      | 50×1 = 50       |
| Player2  | 30       | 2           | 0      | 0      | 30 + 50 = 80    |
| Player3  | 20       | 0           | 0      | 0      | 20              |
| Player4  | 0        | 0           | 1      | 0      | 8               |

**Team A (this match):**

- Player1 (captain): 50 × 2 = **100**
- Player2 (vice): 80 × 1.5 = **120**
- Player3: 20 × 1 = **20**  
→ **Match total = 240**

**Team B (this match):**

- Player1: 50 × 1 = **50**
- Player2: 80 × 1 = **80**
- Player4: 8 × 1 = **8**  
→ **Match total = 138**

**Leaderboard (single match, no prior points):**

| Rank | Team   | Points |
|------|--------|--------|
| 1    | Team A | 240    |
| 2    | Team B | 138    |

---

## API flow

1. **POST /api/v1/engine/compute-leaderboard**  
   Body: `{ "tournament_id": 1, "match_id": "<match_id>" }`  
   Returns the **computed** leaderboard entries (not yet saved).

2. **POST /api/v1/engine/update-leaderboard**  
   Body: `{ "tournament_id": 1, "entries": [ ... ] }`  
   Replaces the stored leaderboard for that tournament with the given entries.

3. **GET /api/v1/leaderboard/{tournament_id}**  
   Returns the **current** leaderboard (what was last updated).

So the typical “after a match” flow is: **compute-leaderboard** → **update-leaderboard** → (optional) **snapshot** for that date.
