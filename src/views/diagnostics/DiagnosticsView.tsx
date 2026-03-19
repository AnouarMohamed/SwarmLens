import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { GrafanaEmbed } from '../../components/charts/GrafanaEmbed'
import { GrafanaAreaSeries, GrafanaBarSeries } from '../../components/charts/GrafanaCharts'
import { grafanaConfig } from '../../lib/grafana'
import { buildMockDiagnosticsFindings } from '../../lib/mockData'
import { buildDiagnosticsTelemetry } from '../../lib/telemetry'
import { relativeTime } from '../../lib/utils'
import { useClusterStore } from '../../store/clusterStore'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import { useIncidentStore } from '../../store/incidentStore'
import type { Finding, Severity } from '../../types'

const SEVERITIES: Array<Severity | 'all'> = ['all', 'critical', 'high', 'medium', 'low', 'info']

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

function severityTone(severity: Severity) {
  return severity === 'critical' || severity === 'high' || severity === 'medium'
    ? 'text-state-danger'
    : 'text-text-secondary'
}

export function DiagnosticsView() {
  const navigate = useNavigate()
  const {
    findings,
    loading,
    running,
    lastRun,
    lastDurationMs,
    run,
    fetch,
    error: diagnosticsError,
  } = useDiagnosticsStore()
  const { create: createIncident } = useIncidentStore()
  const { swarm, connectionState, error: clusterError } = useClusterStore()

  const [sevFilter, setSevFilter] = useState<Severity | 'all'>('all')
  const [query, setQuery] = useState('')
  const [openIds, setOpenIds] = useState<Record<string, boolean>>({})
  const [bulkCreating, setBulkCreating] = useState(false)

  useEffect(() => {
    void fetch()
  }, [fetch])

  const disconnected = connectionState === 'disconnected' || Boolean(clusterError)
  const demoFindings = useMemo(() => buildMockDiagnosticsFindings(), [])
  const isDemoCluster = (swarm?.mode ?? '').toLowerCase() === 'demo'
  const useDemo = findings.length === 0 && !loading && isDemoCluster
  const showDemoDetails = isDemoCluster || useDemo
  const dataset = useDemo ? demoFindings : findings

  const filtered = useMemo(() => {
    return dataset.filter((finding) => {
      const severityOk = sevFilter === 'all' || finding.severity === sevFilter
      if (!severityOk) return false
      if (!query.trim()) return true
      const haystack = `${finding.resource} ${finding.message} ${finding.scope}`.toLowerCase()
      return haystack.includes(query.toLowerCase())
    })
  }, [dataset, sevFilter, query])

  const counts = useMemo(
    () =>
      SEVERITIES.reduce(
        (acc, severity) => {
          acc[severity] =
            severity === 'all'
              ? dataset.length
              : dataset.filter((finding) => finding.severity === severity).length
          return acc
        },
        {} as Record<string, number>,
      ),
    [dataset],
  )

  const telemetry = useMemo(
    () => buildDiagnosticsTelemetry({ findings: dataset, disconnected }),
    [dataset, disconnected],
  )
  const grafana = grafanaConfig()

  async function handleCreateIncident(finding: Finding) {
    await createIncident({
      title: finding.message,
      description: finding.evidence.join('\n'),
      severity: finding.severity,
      affectedServices: [finding.resource],
      diagnosticRefs: [finding.id],
    })
  }

  async function handleCreateCriticalBatch() {
    const critical = filtered.filter((finding) => finding.severity === 'critical')
    if (critical.length === 0) return
    setBulkCreating(true)
    try {
      for (const finding of critical) {
        await handleCreateIncident(finding)
      }
      navigate('/incidents')
    } finally {
      setBulkCreating(false)
    }
  }

  function exportFindings() {
    const payload = JSON.stringify(filtered, null, 2)
    const blob = new Blob([payload], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `swarmlens-findings-${Date.now()}.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-10">
      <section className="industrial-section border-b-0 pt-0">
        <p className="industrial-label">Diagnostics Control</p>
        <h2 className="mt-3 font-heading text-[2rem] uppercase leading-none tracking-[0.04em]">
          Findings Pipeline
        </h2>
        <p className="mt-3 text-sm text-text-secondary">
          {disconnected
            ? 'Cluster is disconnected. Diagnostics are in read-only mode with last known data.'
            : useDemo
              ? 'Demo mode is active with enriched synthetic findings and telemetry.'
              : `Diagnostics last run ${lastRun ? relativeTime(new Date(lastRun).toISOString()) : 'not available'}.`}
        </p>
        <div className="mt-5 flex flex-wrap items-center gap-8">
          <button
            type="button"
            onClick={() => {
              void run()
            }}
            disabled={running || disconnected}
            className={cn(
              'industrial-action industrial-action-accent',
              (running || disconnected) && 'cursor-not-allowed opacity-35',
            )}
          >
            {running ? 'Running Diagnostics' : 'Run Diagnostics'}
          </button>
          <button
            type="button"
            onClick={() => {
              void fetch()
            }}
            className="industrial-action"
          >
            Refresh Findings
          </button>
          <button
            type="button"
            onClick={handleCreateCriticalBatch}
            disabled={
              bulkCreating || filtered.filter((f) => f.severity === 'critical').length === 0
            }
            className={cn(
              'industrial-action',
              (bulkCreating || filtered.filter((f) => f.severity === 'critical').length === 0) &&
                'cursor-not-allowed opacity-35',
            )}
          >
            {bulkCreating ? 'Creating Incidents...' : 'Create Incidents for Critical'}
          </button>
          <button type="button" onClick={exportFindings} className="industrial-action">
            Export Findings JSON
          </button>
          <span className="industrial-data text-xs text-text-tertiary">
            Duration {lastDurationMs ? `${lastDurationMs}ms` : 'n/a'}
          </span>
        </div>
      </section>

      {showDemoDetails ? (
        <section className="border-t border-white/10 pt-6">
          <p className="industrial-label">Demo Mode</p>
          <h3 className="mt-2 font-heading text-[1.55rem] uppercase leading-none tracking-[0.05em]">
            How Diagnostics Demo Works
          </h3>
          <p className="mt-3 text-sm text-text-secondary">
            Diagnostics demo is designed to teach triage flow: synthetic findings, synthetic telemetry trend lines, and the same incident/action handlers used in live mode.
          </p>
          <ul className="mt-4 space-y-2 text-sm text-text-secondary">
            <li className="flex items-start gap-3">
              <span className="industrial-label text-text-tertiary">Dataset</span>
              <span>
                {useDemo
                  ? `Using ${demoFindings.length} synthetic findings because live diagnostics returned no data.`
                  : 'Live findings are available; demo scenario controls remain available from Overview.'}
              </span>
            </li>
            <li className="flex items-start gap-3">
              <span className="industrial-label text-text-tertiary">Telemetry</span>
              <span>Severity trend and scope bars are generated from deterministic synthetic history so charts stay meaningful in demo sessions.</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="industrial-label text-text-tertiary">Actions</span>
              <span>Run diagnostics, create incident, bulk create, and export all execute the same handlers as production mode.</span>
            </li>
          </ul>
          <div className="mt-5 border-t border-white/10 pt-4">
            <p className="industrial-label text-text-secondary">Scenario Walkthrough</p>
            <div className="mt-3 flex flex-wrap items-center gap-5">
              <button type="button" onClick={() => navigate('/?scenario=healthy')} className="industrial-action">
                Open Healthy Overview
              </button>
              <button type="button" onClick={() => navigate('/?scenario=degraded')} className="industrial-action">
                Open Degraded Overview
              </button>
              <button type="button" onClick={() => navigate('/?scenario=disconnected')} className="industrial-action">
                Open Disconnected Overview
              </button>
            </div>
          </div>
        </section>
      ) : null}

      {grafana.enabled ? (
        <section className="grid grid-cols-1 gap-10 xl:grid-cols-2">
          <GrafanaEmbed
            panelId={11}
            title="Severity Trend"
            subtitle="Live diagnostics trend from Grafana"
          />
          <GrafanaEmbed panelId={12} title="Findings by Scope" subtitle="Top diagnostic scopes" />
        </section>
      ) : (
        <section className="grid grid-cols-1 gap-10 xl:grid-cols-2">
          <GrafanaAreaSeries
            title="Severity Trend"
            subtitle="Synthetic run history for quick trend scanning"
            data={telemetry.severityTrend}
            xKey="time"
            areas={[
              { key: 'critical', label: 'Critical', color: '#F5A623' },
              { key: 'high', label: 'High', color: 'rgba(255,255,255,0.82)' },
              { key: 'medium', label: 'Medium', color: 'rgba(255,255,255,0.6)' },
            ]}
          />
          <GrafanaBarSeries
            title="Findings by Scope"
            subtitle="Top scopes requiring attention"
            data={
              telemetry.scopeBars.length > 0
                ? telemetry.scopeBars
                : [{ scope: 'none', findings: 0 }]
            }
            xKey="scope"
            bars={[{ key: 'findings', label: 'Findings', color: '#F5A623' }]}
          />
        </section>
      )}

      <section>
        <div className="flex flex-wrap items-end justify-between gap-6">
          <div>
            <p className="industrial-label">Finding Explorer</p>
            <h3 className="mt-2 font-heading text-[1.85rem] uppercase leading-none tracking-[0.04em]">
              Active Findings
            </h3>
          </div>
          <input
            className="h-9 w-72 border-b border-border-muted bg-transparent px-0 text-sm text-text-primary placeholder:text-text-tertiary focus:outline-none"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Filter by service, node, message..."
          />
        </div>

        <div className="mt-5 flex flex-wrap items-center gap-2">
          {SEVERITIES.map((severity) => (
            <button
              key={severity}
              type="button"
              onClick={() => setSevFilter(severity)}
              className={cn(
                'industrial-action',
                sevFilter === severity && 'industrial-action-accent',
              )}
            >
              {severity.toUpperCase()} ({counts[severity] ?? 0})
            </button>
          ))}
        </div>

        {diagnosticsError ? (
          <div className="mt-6 border-t border-state-danger/50 pt-6">
            <p className="text-sm text-state-danger">Unable to load diagnostics findings</p>
            <p className="mt-1 text-sm text-text-secondary">{diagnosticsError}</p>
          </div>
        ) : filtered.length === 0 ? (
          <div className="mt-6 border-t border-border-muted pt-6">
            <p className="text-sm text-text-primary">
              {dataset.length === 0
                ? 'No findings available.'
                : `No ${sevFilter} findings for this filter.`}
            </p>
            <p className="mt-1 text-sm text-text-secondary">
              Run diagnostics or broaden the filter to inspect full cluster posture.
            </p>
          </div>
        ) : (
          <ul className="mt-6 divide-y divide-border-muted">
            {filtered.map((finding) => {
              const open = Boolean(openIds[finding.id])
              return (
                <li key={finding.id}>
                  <button
                    type="button"
                    onClick={() => setOpenIds((state) => ({ ...state, [finding.id]: !open }))}
                    className="industrial-row w-full text-left focus-visible:outline-none"
                  >
                    <div className="flex items-start justify-between gap-6">
                      <div className="min-w-0">
                        <p className={cn('text-sm', severityTone(finding.severity))}>
                          {finding.message}
                        </p>
                        <p className="mt-1 text-xs text-text-secondary">
                          <span className="font-mono">{finding.resource}</span>
                          <span className="mx-2">|</span>
                          <span className="uppercase tracking-[0.08em]">{finding.severity}</span>
                          <span className="mx-2">|</span>
                          <span>{relativeTime(finding.detectedAt)}</span>
                        </p>
                      </div>
                      <span className="industrial-label text-text-secondary">
                        {open ? 'Collapse' : 'Expand'}
                      </span>
                    </div>
                  </button>
                  {open ? (
                    <div className="px-4 pb-4 pl-8">
                      <p className="text-sm text-text-secondary">{finding.recommendation}</p>
                      {finding.evidence.length > 0 ? (
                        <ul className="mt-2 space-y-1 text-xs text-text-tertiary">
                          {finding.evidence.map((evidence) => (
                            <li key={evidence} className="font-mono">
                              - {evidence}
                            </li>
                          ))}
                        </ul>
                      ) : null}
                      <div className="mt-4 flex flex-wrap items-center gap-6">
                        <button
                          type="button"
                          onClick={() => {
                            void handleCreateIncident(finding)
                          }}
                          className="industrial-action industrial-action-accent"
                        >
                          Create Incident
                        </button>
                        <button
                          type="button"
                          onClick={() => navigate('/incidents')}
                          className="industrial-action"
                        >
                          Open Incidents
                        </button>
                      </div>
                    </div>
                  ) : null}
                </li>
              )
            })}
          </ul>
        )}
      </section>
    </div>
  )
}
