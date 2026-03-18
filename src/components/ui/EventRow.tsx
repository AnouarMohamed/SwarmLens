import type { SwarmEvent } from '../../types'
import { relativeTime } from '../../lib/utils'

interface Props { event: SwarmEvent }

export function EventRow({ event }: Props) {
  return (
    <div className="event-row">
      <span className="event-type">{event.type}</span>
      <span className="event-actor">{event.actor}</span>
      <span className="event-action">{event.action}</span>
      <span className="event-msg">{event.message}</span>
      <span className="event-time">{relativeTime(event.timestamp)}</span>
    </div>
  )
}
