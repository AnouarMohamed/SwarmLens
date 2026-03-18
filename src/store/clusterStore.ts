import { create } from 'zustand'
import { api } from '../lib/api'
import type {
  SwarmInfo,
  Node,
  Stack,
  Service,
  Task,
  Network,
  Volume,
  Secret,
  Config,
  SwarmEvent,
} from '../types'

export type ConnectionState = 'connecting' | 'connected' | 'disconnected'

interface ClusterState {
  swarm: SwarmInfo | null
  nodes: Node[]
  stacks: Stack[]
  services: Service[]
  tasks: Task[]
  networks: Network[]
  volumes: Volume[]
  secrets: Secret[]
  configs: Config[]
  events: SwarmEvent[]
  loading: boolean
  error: string | null
  lastRefresh: number
  connectionState: ConnectionState

  fetchAll: () => Promise<void>
  fetchSwarm: () => Promise<void>
  fetchNodes: () => Promise<void>
  fetchServices: () => Promise<void>
  fetchTasks: () => Promise<void>
  pushEvent: (evt: SwarmEvent) => void
  setConnectionState: (next: ConnectionState) => void
}

export const useClusterStore = create<ClusterState>((set) => ({
  swarm: null,
  nodes: [],
  stacks: [],
  services: [],
  tasks: [],
  networks: [],
  volumes: [],
  secrets: [],
  configs: [],
  events: [],
  loading: false,
  error: null,
  lastRefresh: 0,
  connectionState: 'connecting',

  fetchAll: async () => {
    set({ loading: true, error: null, connectionState: 'connecting' })
    try {
      const [swarm, nodes, stacks, services, tasks, networks, volumes, secrets, configs, events] =
        await Promise.all([
          api.swarm.get(),
          api.nodes.list(),
          api.stacks.list(),
          api.services.list(),
          api.tasks.list(),
          api.networks.list(),
          api.volumes.list(),
          api.secrets.list(),
          api.configs.list(),
          api.events.list(),
        ])
      set({
        swarm,
        nodes,
        stacks,
        services,
        tasks,
        networks,
        volumes,
        secrets,
        configs,
        events,
        loading: false,
        error: null,
        lastRefresh: Date.now(),
        connectionState: 'connected',
      })
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : 'fetch failed',
        connectionState: 'disconnected',
      })
    }
  },

  fetchSwarm: async () => {
    const swarm = await api.swarm.get()
    set({ swarm, connectionState: 'connected', error: null })
  },

  fetchNodes: async () => {
    const nodes = await api.nodes.list()
    set({ nodes, connectionState: 'connected', error: null })
  },

  fetchServices: async () => {
    const [stacks, services] = await Promise.all([api.stacks.list(), api.services.list()])
    set({ stacks, services, connectionState: 'connected', error: null })
  },

  fetchTasks: async () => {
    const tasks = await api.tasks.list()
    set({ tasks, connectionState: 'connected', error: null })
  },

  pushEvent: (evt: SwarmEvent) => {
    set((state) => ({ events: [evt, ...state.events].slice(0, 200) }))
  },

  setConnectionState: (next) => {
    set({ connectionState: next })
  },
}))
