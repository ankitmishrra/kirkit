import './TournamentSelector.css'

export default function TournamentSelector({ tournaments, selectedId, onSelect }) {
  if (!tournaments?.length) return null
  return (
    <section className="tournament-selector">
      <label htmlFor="tournament" className="tournament-selector__label">
        Tournament
      </label>
      <select
        id="tournament"
        className="tournament-selector__select"
        value={selectedId ?? ''}
        onChange={(e) => onSelect(e.target.value ? Number(e.target.value) : null)}
      >
        {tournaments.map((t) => (
          <option key={t.id} value={t.id}>
            {t.name} {t.status === 2 ? '(Done)' : ''}
          </option>
        ))}
      </select>
    </section>
  )
}
