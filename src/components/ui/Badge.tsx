import type { Severity } from '../../types'

interface Props {
  severity: Severity
  size?: 'sm' | 'md'
}

export function SeverityBadge({ severity, size = 'sm' }: Props) {
  return (
    <span className={`badge sev-${severity} badge-${size}`}>
      {severity}
    </span>
  )
}

interface StatusProps {
  status: string
  size?: 'sm' | 'md'
}

export function StatusBadge({ status, size = 'sm' }: StatusProps) {
  const cls = {
    running: 'state-running',
    ready: 'state-running',
    active: 'state-running',
    failed: 'state-failed',
    rejected: 'state-failed',
    down: 'state-failed',
    drain: 'state-warn',
    pause: 'state-warn',
    paused: 'state-warn',
    complete: 'state-done',
    shutdown: 'state-done',
  }[status] ?? 'state-pending'

  return <span className={`badge ${cls} badge-${size}`}>{status}</span>
}
