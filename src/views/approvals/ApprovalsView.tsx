import { useEffect, useState } from 'react'
import { useControlPlaneStore } from '../../store/controlPlaneStore'
import { relativeTime } from '../../lib/utils'

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

export function ApprovalsView() {
  const {
    clusters,
    selectedClusterID,
    approvals,
    actionRuns,
    loading,
    error,
    refreshWorkflow,
    approve,
    reject,
  } = useControlPlaneStore()
  const [workingID, setWorkingID] = useState('')

  useEffect(() => {
    void refreshWorkflow()
  }, [refreshWorkflow, selectedClusterID])

  const activeCluster =
    clusters.find((cluster) => cluster.id === selectedClusterID) ??
    clusters.find((cluster) => cluster.default) ??
    clusters[0] ??
    null

  async function handleApproval(id: string, decision: 'approve' | 'reject') {
    setWorkingID(id)
    try {
      if (decision === 'approve') {
        await approve(id)
      } else {
        await reject(id)
      }
    } finally {
      setWorkingID('')
    }
  }

  return (
    <div className="space-y-10">
      <section className="industrial-section border-b-0 pt-0">
        <p className="industrial-label">Action Center</p>
        <h2 className="mt-3 font-heading text-[2rem] uppercase leading-none tracking-[0.04em]">
          Approvals And Recent Writes
        </h2>
        <p className="mt-3 text-sm text-text-secondary">
          {activeCluster ? `Cluster ${activeCluster.name}` : 'Cluster context unavailable.'} Pending approvals must be resolved by an admin before risky actions execute.
        </p>
        {error ? <p className="mt-2 text-sm text-state-danger">{error}</p> : null}
      </section>

      <section className="border-t border-border-muted pt-5">
        <div className="view-header">
          <span className="count-label">{approvals.length} pending approvals</span>
        </div>
        {loading ? <div className="loading-bar" /> : null}
        {approvals.length === 0 ? (
          <div className="empty-state">No pending approvals for this cluster.</div>
        ) : (
          <ul className="divide-y divide-border-muted">
            {approvals.map((approval) => (
              <li key={approval.id} className="industrial-row">
                <div className="flex flex-wrap items-start justify-between gap-4">
                  <div className="min-w-0">
                    <p className="industrial-label text-text-secondary">
                      {approval.action} · {approval.resource}/{approval.resourceID}
                    </p>
                    <p className="mt-2 text-sm text-text-primary">{approval.reason}</p>
                    <p className="mt-2 industrial-data text-xs text-text-tertiary">
                      Requested by {approval.requestedBy} ({approval.requestedRole}) · {relativeTime(approval.createdAt)}
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <button
                      type="button"
                      onClick={() => {
                        void handleApproval(approval.id, 'approve')
                      }}
                      disabled={workingID === approval.id}
                      className={cn('industrial-action industrial-action-accent', workingID === approval.id && 'cursor-not-allowed opacity-35')}
                    >
                      Approve
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        void handleApproval(approval.id, 'reject')
                      }}
                      disabled={workingID === approval.id}
                      className={cn('industrial-action', workingID === approval.id && 'cursor-not-allowed opacity-35')}
                    >
                      Reject
                    </button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="border-t border-border-muted pt-5">
        <div className="view-header">
          <span className="count-label">{actionRuns.length} recent action runs</span>
        </div>
        {actionRuns.length === 0 ? (
          <div className="empty-state">No action runs recorded for this cluster yet.</div>
        ) : (
          <ul className="divide-y divide-border-muted">
            {actionRuns.slice(0, 20).map((run) => (
              <li key={run.id} className="industrial-row">
                <p className="industrial-label text-text-secondary">
                  {run.action} · {run.resource}/{run.resourceID}
                </p>
                <p className="mt-2 text-sm text-text-primary">
                  {run.message || run.reason}
                </p>
                <p className="mt-2 industrial-data text-xs text-text-tertiary">
                  {run.status} · requested by {run.requestedBy} · {relativeTime(run.createdAt)}
                </p>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
