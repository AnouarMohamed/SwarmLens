import type { Severity } from '../types'

export function severityColor(s: Severity): string {
  switch (s) {
    case 'critical': return 'var(--sev-critical)'
    case 'high':     return 'var(--sev-high)'
    case 'medium':   return 'var(--sev-medium)'
    case 'low':      return 'var(--sev-low)'
    default:         return 'var(--sev-info)'
  }
}

export function fmtBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`
}

export function fmtNanoCPU(nano: number): string {
  return `${(nano / 1e9).toFixed(2)} CPU`
}

export function pct(used: number, total: number): number {
  if (total === 0) return 0
  return Math.round((used / total) * 100)
}

export function relativeTime(ts: string): string {
  const diff = Date.now() - new Date(ts).getTime()
  const s = Math.floor(diff / 1000)
  if (s < 60) return `${s}s ago`
  if (s < 3600) return `${Math.floor(s / 60)}m ago`
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`
  return `${Math.floor(s / 86400)}d ago`
}

export function taskStateClass(state: string): string {
  if (state === 'running') return 'state-running'
  if (state === 'failed' || state === 'rejected') return 'state-failed'
  if (state === 'complete' || state === 'shutdown') return 'state-done'
  return 'state-pending'
}

export function healthScoreClass(score: number): string {
  if (score >= 90) return 'health-good'
  if (score >= 60) return 'health-warn'
  return 'health-bad'
}
