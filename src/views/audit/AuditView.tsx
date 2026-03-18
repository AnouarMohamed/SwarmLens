import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import type { AuditEntry } from '../../types'
import { relativeTime } from '../../lib/utils'

export function AuditView() {
  const [entries, setEntries] = useState<AuditEntry[]>([])
  const [total, setTotal]     = useState(0)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    api.audit.list(50, 0)
      .then(data => {
        setEntries(data)
        setTotal(data.length)
      })
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">{total} entries</span>
      </div>

      {loading && <div className="loading-bar" />}

      {entries.length === 0 && !loading && (
        <div className="empty-state">No audit entries yet. Write actions will appear here.</div>
      )}

      <div className="audit-log">
        {entries.map(e => (
          <div key={e.id} className={`audit-row result-${e.result}`}>
            <span className="audit-time">{relativeTime(e.timestamp)}</span>
            <span className={`audit-result result-${e.result}`}>{e.result}</span>
            <span className="audit-actor mono">{e.actor}</span>
            <span className="dim mono">[{e.role}]</span>
            <span className="audit-action">{e.action}</span>
            <span className="mono dim">{e.resource}/{e.resourceID}</span>
            {e.reason && <span className="dim">— {e.reason}</span>}
          </div>
        ))}
      </div>
    </div>
  )
}
