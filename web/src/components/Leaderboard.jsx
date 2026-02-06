import './Leaderboard.css'

export default function Leaderboard({ entries }) {
  if (!entries?.length) {
    return (
      <section className="leaderboard">
        <h2 className="leaderboard__title">Leaderboard</h2>
        <p className="leaderboard__empty">No entries yet.</p>
      </section>
    )
  }

  return (
    <section className="leaderboard">
      <h2 className="leaderboard__title">Leaderboard</h2>
      <div className="leaderboard__table-wrap">
        <table className="leaderboard__table">
          <thead>
            <tr>
              <th>Rank</th>
              <th>Team</th>
              <th>Owner</th>
              <th className="leaderboard__th-points">Points</th>
            </tr>
          </thead>
          <tbody>
            {entries.map((e) => (
              <tr
                key={e.team_id}
                className={e.rank === 1 ? 'leaderboard__row--first' : ''}
              >
                <td className="leaderboard__rank">
                  {e.rank === 1 && <span className="leaderboard__medal">🥇</span>}
                  {e.rank === 2 && <span className="leaderboard__medal">🥈</span>}
                  {e.rank === 3 && <span className="leaderboard__medal">🥉</span>}
                  {e.rank}
                </td>
                <td className="leaderboard__team">
                  {e.team_name || `Team ${e.team_id}`}
                </td>
                <td className="leaderboard__owner">{e.team_owner || '—'}</td>
                <td className="leaderboard__points">{e.points?.toLocaleString() ?? 0}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
