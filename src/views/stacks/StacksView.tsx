import { useClusterStore } from '../../store/clusterStore'
import { ResourceTable, type Column } from '../../components/ui/ResourceTable'
import type { Stack } from '../../types'

export function StacksView() {
  const { stacks, loading } = useClusterStore()

  const columns: Column<Stack>[] = [
    {
      key: 'name', header: 'Stack',
      render: s => <span className="mono fw-medium">{s.name}</span>,
    },
    {
      key: 'services', header: 'Services',
      render: s => (
        <span className={s.runningServices < s.serviceCount ? 'text-warn' : ''}>
          {s.runningServices}/{s.serviceCount}
        </span>
      ),
      width: '90px',
    },
    {
      key: 'replicas', header: 'Replicas',
      render: s => (
        <span className={s.runningReplicas < s.totalReplicas ? 'text-bad' : ''}>
          {s.runningReplicas}/{s.totalReplicas}
        </span>
      ),
      width: '90px',
    },
    {
      key: 'health', header: 'Health',
      render: s => <HealthScore score={s.healthScore} />,
      width: '130px',
    },
  ]

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">{stacks.length} stacks</span>
      </div>
      <ResourceTable
        columns={columns}
        rows={stacks}
        keyFn={s => s.name}
        loading={loading}
        empty="No stacks found."
      />
    </div>
  )
}

function HealthScore({ score }: { score: number }) {
  const cls = score >= 90 ? 'bar-ok' : score >= 60 ? 'bar-warn' : 'bar-bad'
  return (
    <div className="progress-wrap">
      <div className="progress-track">
        <div className={`progress-fill ${cls}`} style={{ width: `${score}%` }} />
      </div>
      <span className="progress-label">{score}%</span>
    </div>
  )
}
