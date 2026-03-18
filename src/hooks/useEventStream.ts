import { useEffect } from 'react'
import { api } from '../lib/api'
import { useClusterStore } from '../store/clusterStore'
import type { SwarmEvent } from '../types'

export function useEventStream() {
  const pushEvent = useClusterStore((s) => s.pushEvent)
  const setConnectionState = useClusterStore((s) => s.setConnectionState)

  useEffect(() => {
    let source: EventSource | null = null
    let retryTimeout: ReturnType<typeof setTimeout>

    function connect() {
      setConnectionState('connecting')
      source = api.events.stream()

      source.onopen = () => {
        setConnectionState('connected')
      }

      source.addEventListener('swarm', (e: MessageEvent) => {
        try {
          const evt = JSON.parse(e.data) as SwarmEvent
          pushEvent(evt)
          setConnectionState('connected')
        } catch {
          /* ignore parse errors */
        }
      })

      source.onerror = () => {
        setConnectionState('disconnected')
        source?.close()
        retryTimeout = setTimeout(connect, 5000)
      }
    }

    connect()
    return () => {
      source?.close()
      clearTimeout(retryTimeout)
    }
  }, [pushEvent, setConnectionState])
}
