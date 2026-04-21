import { useState } from 'react'
import { api } from '../../lib/api'
import { useClusterStore } from '../../store/clusterStore'
import { useControlPlaneStore } from '../../store/controlPlaneStore'
import { useSessionStore } from '../../store/sessionStore'
import { ResourceTable, type Column } from '../../components/ui/ResourceTable'
import { StatusBadge } from '../../components/ui/Badge'
import type { Service } from '../../types'

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

export function ServicesView() {
  const { services, loading, fetchServices } = useClusterStore()
  const refreshWorkflow = useControlPlaneStore((s) => s.refreshWorkflow)
  const me = useSessionStore((s) => s.me)
  const [stackFilter, setStackFilter] = useState('')
  const [busyKey, setBusyKey] = useState('')
  const [notice, setNotice] = useState('')

  const stackNames = ['', ...Array.from(new Set(services.map((s) => s.stack).filter(Boolean)))]
  const filtered = stackFilter ? services.filter((s) => s.stack === stackFilter) : services
  const canOperate = me ? me.authenticated && me.role !== 'viewer' : true

  async function submitAction(
    service: Service,
    action: string,
    reasonDefault: string,
    params?: Record<string, unknown>,
  ) {
    const reason = window.prompt('Reason for this action:', reasonDefault)?.trim()
    if (!reason) return

    const key = `${action}:${service.id}`
    setBusyKey(key)
    try {
      const outcome =
        action === 'service.restart'
          ? await api.services.restart(service.id, { reason })
          : action === 'service.scale'
            ? await api.services.scale(service.id, {
                replicas: Number(params?.replicas ?? service.desiredReplicas),
                reason,
              })
            : await api.services.rollback(service.id, { reason })
      setNotice(outcome ? `${service.name}: ${outcome.status} · ${outcome.message}` : `${service.name}: action failed`)
      await Promise.all([fetchServices(), refreshWorkflow()])
    } finally {
      setBusyKey('')
    }
  }

  const columns: Column<Service>[] = [
    {
      key: 'name',
      header: 'Service',
      render: (s) => (
        <div>
          <span className="mono fw-medium">{s.name}</span>
          {s.stack && <span className="dim mono"> ({s.stack})</span>}
        </div>
      ),
    },
    {
      key: 'image',
      header: 'Image',
      render: (s) => <span className="mono dim">{s.image.split('/').pop()}</span>,
    },
    {
      key: 'replicas',
      header: 'Replicas',
      render: (s) => (
        <span className={s.runningTasks < s.desiredReplicas ? 'text-bad' : ''}>
          {s.runningTasks}/{s.desiredReplicas}
        </span>
      ),
      width: '90px',
    },
    {
      key: 'update',
      header: 'Update state',
      render: (s) =>
        s.updateState ? <StatusBadge status={s.updateState} /> : <span className="dim">—</span>,
      width: '140px',
    },
    {
      key: 'ports',
      header: 'Ports',
      render: (s) =>
        s.publishedPorts.length > 0 ? (
          <span className="mono dim">
            {s.publishedPorts.map((p) => p.publishedPort).join(', ')}
          </span>
        ) : (
          <span className="dim">—</span>
        ),
      width: '100px',
    },
    {
      key: 'mode',
      header: 'Mode',
      render: (s) => <span className="dim">{s.mode}</span>,
      width: '90px',
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (service) => (
        <div className="flex flex-wrap items-center gap-3">
          <button
            type="button"
            onClick={() => {
              void submitAction(
                service,
                'service.restart',
                `Restart ${service.name} to recover failed tasks.`,
              )
            }}
            disabled={!canOperate || busyKey === `service.restart:${service.id}`}
            className={cn('industrial-action', (!canOperate || busyKey === `service.restart:${service.id}`) && 'cursor-not-allowed opacity-35')}
          >
            Restart
          </button>
          {service.mode === 'replicated' ? (
            <>
              <button
                type="button"
                onClick={() => {
                  void submitAction(
                    service,
                    'service.scale',
                    `Scale ${service.name} up by one replica.`,
                    { replicas: service.desiredReplicas + 1 },
                  )
                }}
                disabled={!canOperate || busyKey === `service.scale:${service.id}`}
                className={cn('industrial-action', (!canOperate || busyKey === `service.scale:${service.id}`) && 'cursor-not-allowed opacity-35')}
              >
                +1
              </button>
              <button
                type="button"
                onClick={() => {
                  void submitAction(
                    service,
                    'service.scale',
                    `Scale ${service.name} down by one replica.`,
                    { replicas: Math.max(service.desiredReplicas - 1, 0) },
                  )
                }}
                disabled={!canOperate || service.desiredReplicas === 0 || busyKey === `service.scale:${service.id}`}
                className={cn('industrial-action', (!canOperate || service.desiredReplicas === 0 || busyKey === `service.scale:${service.id}`) && 'cursor-not-allowed opacity-35')}
              >
                -1
              </button>
            </>
          ) : null}
          <button
            type="button"
            onClick={() => {
              void submitAction(
                service,
                'service.rollback',
                `Rollback ${service.name} to the previous service spec.`,
              )
            }}
            disabled={!canOperate || busyKey === `service.rollback:${service.id}`}
            className={cn('industrial-action', (!canOperate || busyKey === `service.rollback:${service.id}`) && 'cursor-not-allowed opacity-35')}
          >
            Rollback
          </button>
        </div>
      ),
      width: '280px',
    },
  ]

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">{filtered.length} services</span>
        <div className="filter-row">
          <label className="filter-label">Stack</label>
          <select
            className="filter-select"
            value={stackFilter}
            onChange={(e) => setStackFilter(e.target.value)}
          >
            {stackNames.map((n) => (
              <option key={n} value={n}>
                {n || 'All stacks'}
              </option>
            ))}
          </select>
        </div>
      </div>
      {notice ? <p className="mb-3 text-sm text-text-secondary">{notice}</p> : null}
      <ResourceTable
        columns={columns}
        rows={filtered}
        keyFn={(s) => s.id}
        loading={loading}
        empty="No services found."
      />
    </div>
  )
}
