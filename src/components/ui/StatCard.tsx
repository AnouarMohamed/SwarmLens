interface Props {
  label: string
  value: string | number
  unit?: string
  sub?: string
  accent?: 'ok' | 'warn' | 'bad' | 'neutral'
}

export function StatCard({ label, value, unit, sub, accent = 'neutral' }: Props) {
  return (
    <div className={`stat-card accent-${accent}`}>
      <div className="stat-label">{label}</div>
      <div className="stat-value">
        {value}
        {unit && <span className="stat-unit">{unit}</span>}
      </div>
      {sub && <div className="stat-sub">{sub}</div>}
    </div>
  )
}
