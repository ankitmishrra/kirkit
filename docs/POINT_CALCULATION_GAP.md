# PointCalculation.xlsx – Implementation Status

All rules from **PointCalculation.xlsx** are now implemented in `bl/bl.go` (PointConfig) and `bl/engine.go` (extractPlayerPointsFromMatch, computePlayerPoints).

## Implemented rules

### Batting
- **Runs** – 1 pt per run
- **Fours** – 4 pts each
- **Sixes** – 6 pts each
- **Duck** – penalty when batsman is out for 0 (default −2)
- **Strike rate bonus** (min 10 balls faced): >170 +6, 150–170 +4, 130–150 +2, 60–70 −2, 50–60 −4, <50 −6
- **Runs bonus**: 25–49 +4, 50–74 +8, 75–99 +12, ≥100 +16

### Bowling
- **Wickets** – 25 pts each
- **Dot balls** – 1 pt each (when API provides; CricAPI match_scorecard may not expose per-bowler dots)
- **Wicket milestones**: 3 wkts +4, 4 wkts +8, 5 wkts +12
- **LBW/Bowled bonus** – configurable per such dismissal (default 0)
- **Maiden overs** – 8 pts per maiden
- **Economy rate bonus** (min 2 overs): <5 +6, 5–5.99 +4, 6–7 +2, 10–11 −2, 11.01–12 −4, >12 −6

### Fielding
- **Catch** – 8 pts each
- **Stumping** – 12 pts each
- **Run out** – 6 pts each (sheet value)
- **Catch bonus** – +4 if player has ≥3 catches in the match

### Captain / Vice
- **Captain** – 2×
- **Vice-captain** – 1.5×

## Config

All values are in `bl.PointConfig` and can be overridden. Defaults match the sheet; see `DefaultPointConfig()` in `bl/bl.go`.

## Data source

- **match_scorecard** format (array of innings with `batting[]`, `bowling[]`, `catching[]`): full stats (r, b, 4s, 6s, dismissal, o, m, r, w, catch, stumped, runout) are read and all bonuses apply.
- **Legacy** flat map (e.g. single `players` array with r, w, catch, runOut): basic stats plus runs bonus; strike rate / economy / maidens / LBW-bowled etc. require the full scorecard format.
