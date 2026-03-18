import { useClusterStore } from '../store/clusterStore'

export function useWriteGate() {
  const swarm = useClusterStore(s => s.swarm)
  return {
    enabled: swarm?.writeEnabled ?? false,
    mode: swarm?.mode ?? 'demo',
  }
}
