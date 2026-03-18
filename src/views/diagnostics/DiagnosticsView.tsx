import { useEffect, useState } from 'react'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import { useIncidentStore } from '../../store/incidentStore'
import { FindingRow } from '../../components/ui/FindingRow'
import type { Finding, Severity } from '../../types'

const SEVERITIES: Array<Severity | 'all'> = ['all', 'critical', 'high', 'medium', 'low', 'info']

export function DiagnosticsView() {
  const { findings, loading, running, run, fetch } = useDiagnosticsStore()
  const { create: createIncident } = useIncidentStore()
  const [sevFilter, setSevFilter] = useState<Severity | 'all'>('all')

  useEffect(() => { fetch() }, [fetch])

  const filtered = sevFilter === 'all' ? findings : findings.filter(f => f.severity === sevFilter)

  const counts = SEVERITIES.reduce((acc, s) => {
    acc[s] = s === 'all' ? findings.length : findings.filter(f => f.severity === s).length
    return acc
  }, {} as Record<string, number>)

  async function handleCreateIncident(f: Finding) {
    await createIncident({
      title: f.message,
      description: f.evidence.join('\n'),
      severity: f.severity,
      affectedServices: [f.resource],
      diagnosticRefs: [f.id],
    })
  }

  return (
    <div className="view">
      <div className="diag-toolbar">
        <div className="filter-pills">
          {SEVERITIES.map(s => (
            <button
              key={s}
              className={'pill' + (sevFilter === s ? ' active' : '') + (s !== 'all' ? ` pill-${s}` : '')}
              onClick={() => setSevFilter(s)}
            >
              {s} {counts[s] > 0 && <span className="pill-count">{counts[s]}</span>}
            </button>
          ))}
        </div>
        <button className="btn-primary btn-sm" onClick={run} disabled={running}>
          {running ? '↻ running…' : '↻ run now'}
        </button>
      </div>

      {(loading || running) && <div className="loading-bar" />}

      {filtered.length === 0 && !loading && (
        <div className="empty-state">
          {findings.length === 0
            ? 'No findings yet. Click "run now" to analyze the cluster.'
            : `No ${sevFilter} findings.`}
        </div>
      )}

      {filtered.map(f => (
        <FindingRow key={f.id} finding={f} onCreateIncident={handleCreateIncident} />
      ))}
    </div>
  )
}
