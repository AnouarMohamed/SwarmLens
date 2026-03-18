import type { Finding } from '../types'

interface OverviewTelemetryInput {
  runningTasks: number
  failedTasks: number
  managersOnline: number
  workersOnline: number
  criticalFindings: number
  warningFindings: number
  disconnected: boolean
  degraded: boolean
}

interface DiagnosticsTelemetryInput {
  findings: Finding[]
  disconnected: boolean
}

function labelForOffset(minutesAgo: number) {
  const dt = new Date(Date.now() - minutesAgo * 60000)
  return dt.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function wobble(index: number, magnitude: number, phase = 0) {
  return Math.sin(index * 0.55 + phase) * magnitude
}

function clamp(value: number, min: number, max: number) {
  return Math.max(min, Math.min(max, value))
}

export function buildOverviewTelemetry(input: OverviewTelemetryInput) {
  const points = 18
  const throughput = Array.from({ length: points }).map((_, idx) => {
    const age = (points - idx - 1) * 5
    const running = clamp(
      Math.round(
        input.runningTasks +
          wobble(idx, input.degraded ? 6 : 3, 0.6) -
          (input.disconnected ? 2 : 0),
      ),
      0,
      Math.max(input.runningTasks + 20, 20),
    )
    const failed = clamp(
      Math.round(
        input.failedTasks + (input.degraded ? 2 : 0) + wobble(idx, input.degraded ? 1.8 : 0.8, 1.4),
      ),
      0,
      Math.max(input.failedTasks + 8, 4),
    )
    const critical = clamp(
      Math.round(
        input.criticalFindings +
          (input.degraded ? 1 : 0) +
          wobble(idx, input.degraded ? 1.4 : 0.4, 0.3),
      ),
      0,
      Math.max(input.criticalFindings + 6, 3),
    )
    const warning = clamp(
      Math.round(input.warningFindings + (input.degraded ? 2 : 1) + wobble(idx, 1.2, 1.8)),
      0,
      Math.max(input.warningFindings + 8, 4),
    )
    return {
      time: labelForOffset(age),
      running,
      failed,
      critical,
      warning,
    }
  })

  const nodeHealth = Array.from({ length: points }).map((_, idx) => {
    const age = (points - idx - 1) * 5
    const managers = clamp(
      Math.round(input.managersOnline - (input.degraded ? Math.abs(wobble(idx, 0.6, 0.5)) : 0)),
      0,
      Math.max(input.managersOnline, 1),
    )
    const workers = clamp(
      Math.round(input.workersOnline - (input.degraded ? Math.abs(wobble(idx, 1.1, 1.1)) : 0)),
      0,
      Math.max(input.workersOnline, 1),
    )
    return {
      time: labelForOffset(age),
      managers,
      workers,
    }
  })

  return { throughput, nodeHealth }
}

export function buildDiagnosticsTelemetry(input: DiagnosticsTelemetryInput) {
  const points = 12
  const severityTrend = Array.from({ length: points }).map((_, idx) => {
    const age = (points - idx - 1) * 10
    const critical = clamp(
      Math.round(
        input.findings.filter((f) => f.severity === 'critical').length + wobble(idx, 1.5, 0.3),
      ),
      0,
      30,
    )
    const high = clamp(
      Math.round(
        input.findings.filter((f) => f.severity === 'high').length + wobble(idx, 1.2, 1.4),
      ),
      0,
      30,
    )
    const medium = clamp(
      Math.round(
        input.findings.filter((f) => f.severity === 'medium').length + wobble(idx, 1.2, 2.3),
      ),
      0,
      30,
    )
    const low = clamp(
      Math.round(input.findings.filter((f) => f.severity === 'low').length + wobble(idx, 1.2, 0.8)),
      0,
      30,
    )
    return {
      time: labelForOffset(age),
      critical: input.disconnected ? 0 : critical,
      high: input.disconnected ? 0 : high,
      medium: input.disconnected ? 0 : medium,
      low: input.disconnected ? 0 : low,
    }
  })

  const scopeCounts = input.findings.reduce<Record<string, number>>((acc, finding) => {
    const scope = finding.scope || 'other'
    acc[scope] = (acc[scope] ?? 0) + 1
    return acc
  }, {})

  const scopeBars = Object.entries(scopeCounts)
    .sort(([, a], [, b]) => b - a)
    .slice(0, 6)
    .map(([scope, count]) => ({ scope, findings: count }))

  return { severityTrend, scopeBars }
}
