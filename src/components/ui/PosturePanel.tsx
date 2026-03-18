import { StatusBadge } from './Badge'

interface PosturePanelProps {
  managersOnline: number
  workersOnline: number
  totalReplicas: number
  pendingTasks: number
  unhealthyServices: number
  quorumState?: 'healthy' | 'degraded' | 'single' | 'unknown'
}

interface MetricRow {
  label: string
  value: string | number
  note?: string
}

function quorumVariant(state: PosturePanelProps['quorumState']) {
  if (state === 'healthy') return 'healthy'
  if (state === 'degraded' || state === 'single') return 'warning'
  return 'neutral'
}

export function PosturePanel({
  managersOnline,
  workersOnline,
  totalReplicas,
  pendingTasks,
  unhealthyServices,
  quorumState = 'unknown',
}: PosturePanelProps) {
  const rows: MetricRow[] = [
    { label: 'Managers online', value: managersOnline },
    { label: 'Workers online', value: workersOnline },
    { label: 'Total replicas', value: totalReplicas },
    {
      label: 'Pending tasks',
      value: pendingTasks,
      note: pendingTasks > 0 ? 'Needs scheduling' : 'None pending',
    },
    {
      label: 'Unhealthy services',
      value: unhealthyServices,
      note: unhealthyServices > 0 ? 'Investigate drift' : 'All services nominal',
    },
  ]

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between rounded-card border border-border-muted bg-surface-2 px-3 py-2">
        <p className="text-xs uppercase tracking-[0.08em] text-text-tertiary">Swarm quorum</p>
        <StatusBadge variant={quorumVariant(quorumState)} label={quorumState} />
      </div>
      <ul className="space-y-2">
        {rows.map((row) => (
          <li
            key={row.label}
            className="flex items-center justify-between rounded-card border border-border-muted bg-surface-2 px-3 py-2"
          >
            <div>
              <p className="text-sm text-text-primary">{row.label}</p>
              {row.note && <p className="text-xs text-text-tertiary">{row.note}</p>}
            </div>
            <span className="font-mono text-sm text-text-secondary">{row.value}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}
