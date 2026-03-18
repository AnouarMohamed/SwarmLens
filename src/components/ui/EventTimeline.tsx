import { relativeTime } from '../../lib/utils'

export type TimelineSeverity = 'critical' | 'warning' | 'info' | 'neutral'

export interface TimelineEvent {
  id: string
  title: string
  timestamp: string
  source: string
  severity: TimelineSeverity
  metadata?: string
}

interface EventTimelineProps {
  events: TimelineEvent[]
}

function severityClasses(severity: TimelineSeverity) {
  switch (severity) {
    case 'critical':
      return 'border-state-danger bg-state-danger'
    case 'warning':
      return 'border-state-warning bg-state-warning'
    case 'info':
      return 'border-state-info bg-state-info'
    case 'neutral':
    default:
      return 'border-border bg-state-neutral'
  }
}

export function EventTimeline({ events }: EventTimelineProps) {
  return (
    <ol className="space-y-3">
      {events.map((event) => (
        <li
          key={event.id}
          className="grid grid-cols-[auto_1fr_auto] items-start gap-3 rounded-card border border-border-muted bg-surface-2 px-3 py-2.5"
        >
          <span
            className={`mt-1 block h-2.5 w-2.5 rounded-full border ${severityClasses(event.severity)}`}
            aria-label={`severity ${event.severity}`}
          />
          <div className="min-w-0">
            <p className="truncate text-sm text-text-primary">{event.title}</p>
            <p className="mt-0.5 text-xs text-text-secondary">
              <span className="font-mono">{event.source}</span>
              {event.metadata ? (
                <span className="ml-2 text-text-tertiary">{event.metadata}</span>
              ) : null}
            </p>
          </div>
          <time className="shrink-0 text-xs text-text-tertiary" dateTime={event.timestamp}>
            {relativeTime(event.timestamp)}
          </time>
        </li>
      ))}
    </ol>
  )
}
