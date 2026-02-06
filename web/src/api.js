const API_BASE = import.meta.env.VITE_API_BASE || ''

async function fetchApi(path) {
  const res = await fetch(`${API_BASE}${path}`)
  const json = await res.json()
  if (json.error) throw new Error(json.error)
  return json.data
}

export async function getTournaments() {
  return fetchApi('/api/v1/tournaments')
}

export async function getLeaderboard(tournamentId) {
  return fetchApi(`/api/v1/leaderboard/${tournamentId}`)
}

export async function getMatches(seriesId) {
  return fetchApi(`/api/v1/matches?series_id=${encodeURIComponent(seriesId)}`)
}

export async function getFantasyLeagues(seriesId) {
  return fetchApi(`/api/v1/fantasy-leagues?series_id=${encodeURIComponent(seriesId)}`)
}
