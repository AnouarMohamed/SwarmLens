import { useClusterStore } from '../../store/clusterStore'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import { StatCard } from '../../components/ui/StatCard'
import { FindingRow } from '../../components/ui/FindingRow'
import { EventRow } from '../../components/ui/EventRow'
import { pct } from '../../lib/utils'

export function OverviewView() {
  const { swarm, nodes, services, tasks, events } = useClusterStore()
  const { findings, fetch: fetchDiag } = useDiagnosticsStore()

  const runningTasks = tasks.filter(t => t.currentState === 'running').length
  const failedTasks  = tasks.filter(t => t.currentState === 'failed').length
  const criticalCount = findings.filter(f => f.severity === 'critical').length
  const highCount     = findings.filter(f => f.severity === 'high').length

  const totalDesired  = services.reduce((s, svc) => s + svc.desiredReplicas, 0)
  const totalRunning  = services.reduce((s, svc) => s + svc.runningTasks, 0)
  const healthPct     = totalDesired > 0 ? pct(totalRunning, totalDesired) : 100

  const managerCount = nodes.filter(n => n.role === 'manager').length
  const workerCount  = nodes.filter(n => n.role === 'worker').length

  return (
    <div className="view">
      {/* Stat row */}
      <div className="stat-row">
        <StatCard
          label="Cluster health"
          value={healthPct}
          unit="%"
          accent={healthPct >= 90 ? 'ok' : healthPct >= 60 ? 'warn' : 'bad'}
          sub={`${totalRunning}/${totalDesired} replicas`}
        />
        <StatCard
          label="Nodes"
          value={managerCount + workerCount}
          sub={`${managerCount} managers · ${workerCount} workers`}
          accent="neutral"
        />
        <StatCard
          label="Running tasks"
          value={runningTasks}
          sub={failedTasks > 0 ? `${failedTasks} failed` : 'all healthy'}
          accent={failedTasks > 0 ? 'bad' : 'ok'}
        />
        <StatCard
          label="Active findings"
          value={findings.length}
          sub={criticalCount > 0 ? `${criticalCount} critical` : highCount > 0 ? `${highCount} high` : 'none critical'}
          accent={criticalCount > 0 ? 'bad' : highCount > 0 ? 'warn' : 'ok'}
        />
      </div>

      {/* Quorum alert */}
      {swarm && !swarm.quorumHealthy && (
        <div className="alert alert-critical">
          ⚠ Quorum at risk — {swarm.managers} manager{swarm.managers !== 1 ? 's' : ''} detected.
          Run diagnostics for details.
        </div>
      )}

      {/* Top findings */}
      <section className="section">
        <div className="section-header">
          <h2 className="section-title">Active findings</h2>
          <button className="btn-ghost btn-sm" onClick={fetchDiag}>view all</button>
        </div>
        {findings.length === 0 && (
          <div className="empty-state">No findings. Run diagnostics to check cluster health.</div>
        )}
        {findings.slice(0, 6).map(f => (
          <FindingRow key={f.id} finding={f} />
        ))}
      </section>

      {/* Recent events */}
      <section className="section">
        <div className="section-header">
          <h2 className="section-title">Recent events</h2>
        </div>
        {events.length === 0 && (
          <div className="empty-state">No recent events.</div>
        )}
        {events.slice(0, 8).map((e, i) => (
          <EventRow key={i} event={e} />
        ))}
      </section>
    </div>
  )
}
