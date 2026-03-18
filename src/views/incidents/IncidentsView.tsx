import { useEffect } from 'react'
import { useIncidentStore } from '../../store/incidentStore'
import { SeverityBadge, StatusBadge } from '../../components/ui/Badge'
import { relativeTime } from '../../lib/utils'

export function IncidentsView() {
  const { incidents, loading, fetch, resolve } = useIncidentStore()

  useEffect(() => { fetch() }, [fetch])

  const open     = incidents.filter(i => i.status !== 'resolved')
  const resolved = incidents.filter(i => i.status === 'resolved')

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">
          {open.length} open · {resolved.length} resolved
        </span>
      </div>

      {loading && <div className="loading-bar" />}

      {incidents.length === 0 && !loading && (
        <div className="empty-state">No incidents. Create one from a diagnostic finding.</div>
      )}

      {open.map(inc => (
        <div key={inc.id} className={`incident-card sev-border-${inc.severity}`}>
          <div className="incident-header">
            <SeverityBadge severity={inc.severity} />
            <span className="incident-title">{inc.title}</span>
            <StatusBadge status={inc.status} />
            <span className="dim">{relativeTime(inc.createdAt)}</span>
            <span className="dim">by {inc.createdBy}</span>
          </div>
          {inc.affectedServices.length > 0 && (
            <div className="incident-meta">
              Affected: <span className="mono">{inc.affectedServices.join(', ')}</span>
            </div>
          )}
          {inc.timeline.length > 0 && (
            <div className="incident-timeline">
              {inc.timeline.slice(-3).map(e => (
                <div key={e.id} className="timeline-entry">
                  <span className="dim">{relativeTime(e.timestamp)}</span>
                  <span className="dim">·</span>
                  <span>{e.actor}</span>
                  <span className="dim">{e.action}</span>
                  {e.note && <span className="dim">— {e.note}</span>}
                </div>
              ))}
            </div>
          )}
          <div className="incident-actions">
            <button
              className="btn-ghost btn-sm"
              onClick={() => resolve(inc.id)}
            >
              ✓ Resolve
            </button>
          </div>
        </div>
      ))}

      {resolved.length > 0 && (
        <section className="section" style={{ marginTop: '24px' }}>
          <h2 className="section-title">Resolved</h2>
          {resolved.map(inc => (
            <div key={inc.id} className="incident-card incident-resolved">
              <div className="incident-header">
                <SeverityBadge severity={inc.severity} />
                <span className="incident-title dim">{inc.title}</span>
                <StatusBadge status={inc.status} />
                <span className="dim">{inc.resolvedAt ? relativeTime(inc.resolvedAt) : ''}</span>
              </div>
            </div>
          ))}
        </section>
      )}
    </div>
  )
}
