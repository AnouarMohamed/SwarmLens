import { useState } from 'react'
import type { Finding } from '../../types'
import { SeverityBadge } from './Badge'

interface Props {
  finding: Finding
  onCreateIncident?: (f: Finding) => void
}

export function FindingRow({ finding, onCreateIncident }: Props) {
  const [open, setOpen] = useState(false)

  return (
    <div
      className={`finding finding-sev-${finding.severity}`}
      onClick={() => setOpen(o => !o)}
    >
      <div className="finding-header">
        <SeverityBadge severity={finding.severity} />
        <span className="finding-resource">{finding.resource}</span>
        <span className="finding-msg">{finding.message}</span>
        <span className="finding-plugin">{finding.source}</span>
        <span className="finding-chevron">{open ? '▴' : '▾'}</span>
      </div>

      {open && (
        <div className="finding-body" onClick={e => e.stopPropagation()}>
          {finding.evidence.length > 0 && (
            <div className="evidence">
              <div className="evidence-label">Evidence</div>
              {finding.evidence.map((e, i) => (
                <div key={i} className="evidence-item">· {e}</div>
              ))}
            </div>
          )}
          <div className="recommendation">
            <div className="rec-label">Recommendation</div>
            <div className="rec-text">{finding.recommendation}</div>
          </div>
          {onCreateIncident && (
            <div className="finding-actions">
              <button
                className="btn-ghost btn-sm"
                onClick={() => onCreateIncident(finding)}
              >
                + Create incident
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
