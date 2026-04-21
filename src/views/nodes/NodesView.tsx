import { useState } from 'react'
import { api } from '../../lib/api'
import { useClusterStore } from '../../store/clusterStore'
import { useControlPlaneStore } from '../../store/controlPlaneStore'
import { useSessionStore } from '../../store/sessionStore'
import { ResourceTable, type Column } from '../../components/ui/ResourceTable'
import { StatusBadge } from '../../components/ui/Badge'
import { fmtBytes, fmtNanoCPU, pct } from '../../lib/utils'
import type { Node } from '../../types'

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

export function NodesView() {
  const { nodes, loading, fetchNodes } = useClusterStore()
  const refreshWorkflow = useControlPlaneStore((s) => s.refreshWorkflow)
  const me = useSessionStore((s) => s.me)
  const [busyKey, setBusyKey] = useState('')
  const [notice, setNotice] = useState('')
  const canOperate = me ? me.authenticated && me.role !== 'viewer' : true

  async function submitAction(node: Node, action: string, reasonDefault: string) {
    const reason = window.prompt('Reason for this action:', reasonDefault)?.trim()
    if (!reason) return

    const key = `${action}:${node.id}`
    setBusyKey(key)
    try {
      const outcome =
        action === 'node.drain'
          ? await api.nodes.drain(node.id, { reason })
          : await api.nodes.activate(node.id, { reason })
      setNotice(outcome ? `${node.hostname}: ${outcome.status} · ${outcome.message}` : `${node.hostname}: action failed`)
      await Promise.all([fetchNodes(), refreshWorkflow()])
    } finally {
      setBusyKey('')
    }
  }

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
    {
      key: 'actions',
      header: 'Actions',
      render: (node) => (
        <div className="flex flex-wrap items-center gap-3">
          {node.availability !== 'drain' ? (
            <button
              type="button"
              onClick={() => {
                void submitAction(node, 'node.drain', `Drain ${node.hostname} for maintenance.`)
              }}
              disabled={!canOperate || busyKey === `node.drain:${node.id}`}
              className={cn('industrial-action', (!canOperate || busyKey === `node.drain:${node.id}`) && 'cursor-not-allowed opacity-35')}
            >
              Drain
            </button>
          ) : (
            <button
              type="button"
              onClick={() => {
                void submitAction(node, 'node.activate', `Return ${node.hostname} to active scheduling.`)
              }}
              disabled={!canOperate || busyKey === `node.activate:${node.id}`}
              className={cn('industrial-action', (!canOperate || busyKey === `node.activate:${node.id}`) && 'cursor-not-allowed opacity-35')}
            >
              Activate
            </button>
          )}
        </div>
      ),
      width: '120px',
    },
  ]

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">{nodes.length} nodes</span>
      </div>
      {notice ? <p className="mb-3 text-sm text-text-secondary">{notice}</p> : null}
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
