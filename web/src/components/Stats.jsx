import './Stats.css'

// Stats derived from schema/APIs: tournaments, leaderboard (teams, points), matches, fantasy_leagues
export default function Stats({ tournament, leaderboard, matches, teams }) {
  const teamCount = leaderboard?.length ?? 0
  const totalPoints = (leaderboard ?? []).reduce((sum, e) => sum + (e.points ?? 0), 0)
  const matchCount = matches?.length ?? 0
  const fantasyTeamCount = teams?.length ?? 0
  const topEntry = leaderboard?.length ? leaderboard[0] : null

  const cards = [
    {
      label: 'Teams in leaderboard',
      value: teamCount,
      desc: 'From leaderboard table',
    },
    {
      label: 'Total points',
      value: totalPoints.toLocaleString(),
      desc: 'Sum of all team points',
    },
    {
      label: 'Matches in series',
      value: matchCount,
      desc: 'From game table (series_id)',
    },
    {
      label: 'Fantasy teams',
      value: fantasyTeamCount,
      desc: 'From fantasy_league table',
    },
    {
      label: 'Leader',
      value: topEntry ? (topEntry.team_name || `Team ${topEntry.team_id}`) : '—',
      desc: topEntry ? `${topEntry.points ?? 0} pts` : 'Rank 1',
    },
  ]

  return (
    <section className="stats">
      <h2 className="stats__title">Stats</h2>
      <p className="stats__subtitle">
        Derived from tournament, leaderboard, game, and fantasy_league
      </p>
      <div className="stats__grid">
        {cards.map((card) => (
          <div key={card.label} className="stats__card">
            <div className="stats__card-value">{card.value}</div>
            <div className="stats__card-label">{card.label}</div>
            <div className="stats__card-desc">{card.desc}</div>
          </div>
        ))}
      </div>
    </section>
  )
}
