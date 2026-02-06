import { useState, useEffect } from 'react'
import { getTournaments, getLeaderboard, getMatches, getFantasyLeagues } from './api'
import Stats from './components/Stats'
import Leaderboard from './components/Leaderboard'
import TournamentSelector from './components/TournamentSelector'
import './App.css'

function App() {
  const [tournaments, setTournaments] = useState([])
  const [selectedId, setSelectedId] = useState(null)
  const [leaderboard, setLeaderboard] = useState([])
  const [matches, setMatches] = useState([])
  const [teams, setTeams] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const selected = tournaments.find((t) => t.id === selectedId)
  const seriesId = selected?.series_id ?? null

  useEffect(() => {
    let cancelled = false
    getTournaments()
      .then((data) => {
        if (!cancelled) {
          setTournaments(Array.isArray(data) ? data : [])
          if (data?.length && !selectedId) setSelectedId(data[0].id)
        }
      })
      .catch((e) => !cancelled && setError(e.message))
      .finally(() => !cancelled && setLoading(false))
    return () => { cancelled = true }
  }, [])

  useEffect(() => {
    if (!selectedId) {
      setLeaderboard([])
      setMatches([])
      setTeams([])
      return
    }
    setError(null)
    Promise.all([
      getLeaderboard(selectedId),
      seriesId ? getMatches(seriesId) : Promise.resolve([]),
      seriesId ? getFantasyLeagues(seriesId) : Promise.resolve([]),
    ])
      .then(([lb, m, t]) => {
        setLeaderboard(Array.isArray(lb) ? lb : [])
        setMatches(Array.isArray(m) ? m : [])
        setTeams(Array.isArray(t) ? t : [])
      })
      .catch((e) => setError(e.message))
  }, [selectedId, seriesId])

  if (loading) return <div className="app-loading">Loading…</div>
  if (error) return <div className="app-error">Error: {error}</div>

  return (
    <div className="app">
      <header className="app-header">
        <img src="/logo.png" alt="" className="app-logo" aria-hidden="true" />
        <h1 className="app-title">Kirkit</h1>
        <p className="app-tagline">Fantasy Cricket Leaderboard</p>
      </header>
      <main className="app-main">
        <TournamentSelector
          tournaments={tournaments}
          selectedId={selectedId}
          onSelect={setSelectedId}
        />
        <Stats
          tournament={selected}
          leaderboard={leaderboard}
          matches={matches}
          teams={teams}
        />
        <Leaderboard entries={leaderboard} />
      </main>
    </div>
  )
}

export default App
