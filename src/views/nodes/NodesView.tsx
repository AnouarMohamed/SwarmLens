import { useClusterStore } from '../../store/clusterStore'
import { ResourceTable, type Column } from '../../components/ui/ResourceTable'
import { StatusBadge } from '../../components/ui/Badge'
import { fmtBytes, fmtNanoCPU, pct } from '../../lib/utils'
import type { Node } from '../../types'

export function NodesView() {
  const { nodes, loading } = useClusterStore()

  const columns: Column<Node>[] = [
    {
      key: 'hostname', header: 'Hostname',
      render: n => (
        <span className="mono">
          {n.hostname}
          {n.managerStatus?.leader && <span className="leader-pill">leader</span>}
        </span>
      ),
    },
    {
      key: 'role', header: 'Role',
      render: n => <span className={`role-badge role-${n.role}`}>{n.role}</span>,
      width: '80px',
    },
    {
      key: 'availability', header: 'Availability',
      render: n => <StatusBadge status={n.availability} />,
      width: '110px',
    },
    {
      key: 'state', header: 'State',
      render: n => <StatusBadge status={n.state} />,
      width: '90px',
    },
    {
      key: 'cpu', header: 'CPU reserved',
      render: n => n.cpuTotal > 0
        ? <ProgressBar value={pct(n.cpuReserved, n.cpuTotal)} label={`${fmtNanoCPU(n.cpuReserved)} / ${fmtNanoCPU(n.cpuTotal)}`} />
        : <span className="dim">—</span>,
    },
    {
      key: 'mem', header: 'Mem reserved',
      render: n => n.memTotal > 0
        ? <ProgressBar value={pct(n.memReserved, n.memTotal)} label={`${fmtBytes(n.memReserved)} / ${fmtBytes(n.memTotal)}`} />
        : <span className="dim">—</span>,
    },
    {
      key: 'tasks', header: 'Tasks',
      render: n => <span className="mono">{n.runningTasks}</span>,
      width: '64px',
    },
    {
      key: 'engine', header: 'Engine',
      render: n => <span className="dim mono">{n.engineVersion || '—'}</span>,
      width: '90px',
    },
  ]

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">{nodes.length} nodes</span>
      </div>
      <ResourceTable
        columns={columns}
        rows={nodes}
        keyFn={n => n.id}
        loading={loading}
        empty="No nodes found."
      />
    </div>
  )
}

function ProgressBar({ value, label }: { value: number; label: string }) {
  const cls = value >= 85 ? 'bar-bad' : value >= 60 ? 'bar-warn' : 'bar-ok'
  return (
    <div className="progress-wrap" title={label}>
      <div className="progress-track">
        <div className={`progress-fill ${cls}`} style={{ width: `${Math.min(value, 100)}%` }} />
      </div>
      <span className="progress-label">{value}%</span>
    </div>
  )
}
