import { StatusBadge, type StatusBadgeVariant } from './Badge'
import { ClockIcon, DiagnosticsIcon } from './icons'

interface DiagnosticsWidgetProps {
  lastRun: number | null
  status: StatusBadgeVariant
  durationMs: number | null
  findingsCount: number
  running?: boolean
  disabled?: boolean
  onRun: () => void
}

function formatRunTime(timestamp: number | null) {
  if (!timestamp) return 'Never'
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function formatDuration(durationMs: number | null) {
  if (!durationMs) return 'n/a'
  return durationMs >= 1000 ? `${(durationMs / 1000).toFixed(1)}s` : `${durationMs}ms`
}

export function DiagnosticsWidget({
  lastRun,
  status,
  durationMs,
  findingsCount,
  running = false,
  disabled = false,
  onRun,
}: DiagnosticsWidgetProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between rounded-card border border-border-muted bg-surface-2 px-3 py-2">
        <div>
          <p className="text-xs uppercase tracking-[0.08em] text-text-tertiary">Current status</p>
          <div className="mt-1">
            <StatusBadge variant={status} label={status} />
          </div>
        </div>
        <div className="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border bg-surface-3 text-text-secondary">
          <DiagnosticsIcon className="h-4 w-4" />
        </div>
      </div>

      <dl className="grid grid-cols-2 gap-2 text-sm">
        <div className="rounded-card border border-border-muted bg-surface-2 px-3 py-2">
          <dt className="text-xs text-text-tertiary">Last run</dt>
          <dd className="mt-1 inline-flex items-center gap-1.5 text-text-primary">
            <ClockIcon className="h-3.5 w-3.5 text-text-secondary" />
            {formatRunTime(lastRun)}
          </dd>
        </div>
        <div className="rounded-card border border-border-muted bg-surface-2 px-3 py-2">
          <dt className="text-xs text-text-tertiary">Duration</dt>
          <dd className="mt-1 text-text-primary">{formatDuration(durationMs)}</dd>
        </div>
        <div className="col-span-2 rounded-card border border-border-muted bg-surface-2 px-3 py-2">
          <dt className="text-xs text-text-tertiary">Findings in latest run</dt>
          <dd className="mt-1 text-sm font-medium text-text-primary">{findingsCount}</dd>
        </div>
      </dl>

      <button
        type="button"
        onClick={onRun}
        disabled={disabled || running}
        className="inline-flex w-full items-center justify-center rounded-card border border-state-info/50 bg-state-info/15 px-3 py-2 text-sm font-medium text-state-info transition hover:bg-state-info/25 disabled:cursor-not-allowed disabled:opacity-45"
      >
        {running ? 'Running diagnostics...' : 'Run diagnostics now'}
      </button>
    </div>
  )
}
