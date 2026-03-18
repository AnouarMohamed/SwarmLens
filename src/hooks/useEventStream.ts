import { useEffect } from 'react'
import { api } from '../lib/api'
import { useClusterStore } from '../store/clusterStore'
import type { SwarmEvent } from '../types'

export function useEventStream() {
  const pushEvent = useClusterStore(s => s.pushEvent)

  useEffect(() => {
    let source: EventSource | null = null
    let retryTimeout: ReturnType<typeof setTimeout>

    function connect() {
      source = api.events.stream()

      source.addEventListener('swarm', (e: MessageEvent) => {
        try {
          const evt = JSON.parse(e.data) as SwarmEvent
          pushEvent(evt)
        } catch { /* ignore parse errors */ }
      })

      source.onerror = () => {
        source?.close()
        retryTimeout = setTimeout(connect, 5000)
      }
    }

    connect()
    return () => {
      source?.close()
      clearTimeout(retryTimeout)
    }
  }, [pushEvent])
}
