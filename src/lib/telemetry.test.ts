import { describe, expect, it } from 'vitest'
import { buildDiagnosticsTelemetry, buildOverviewTelemetry } from './telemetry'
import type { Finding } from '../types'

describe('buildOverviewTelemetry', () => {
  it('includes risk and restart series fields', () => {
    const telemetry = buildOverviewTelemetry({
      runningTasks: 120,
      failedTasks: 4,
      managersOnline: 3,
      workersOnline: 5,
      criticalFindings: 2,
      warningFindings: 3,
      disconnected: false,
      degraded: true,
    })

    expect(telemetry.throughput.length).toBeGreaterThan(0)
    expect(telemetry.throughput[0]).toHaveProperty('risk')
    expect(telemetry.throughput[0]).toHaveProperty('restarts')
  })
})

describe('buildDiagnosticsTelemetry', () => {
  it('aggregates scope bars from findings', () => {
    const findings: Finding[] = [
      {
        id: '1',
        severity: 'critical',
        resource: 'svc/a',
        scope: 'services',
        message: 'critical issue',
        evidence: [],
        recommendation: 'fix',
        source: 'plugin',
        detectedAt: new Date().toISOString(),
      },
      {
        id: '2',
        severity: 'high',
        resource: 'node/b',
        scope: 'nodes',
        message: 'high issue',
        evidence: [],
        recommendation: 'fix',
        source: 'plugin',
        detectedAt: new Date().toISOString(),
      },
    ]

    const telemetry = buildDiagnosticsTelemetry({ findings, disconnected: false })
    expect(telemetry.scopeBars.length).toBeGreaterThan(0)
    expect(telemetry.scopeBars.some((row) => row.scope === 'services')).toBe(true)
  })
})
