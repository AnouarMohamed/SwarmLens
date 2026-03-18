import { create } from 'zustand'
import { api } from '../lib/api'
import type { SwarmInfo, Node, Stack, Service, Task, Network, Volume, Secret, Config, SwarmEvent } from '../types'

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

  fetchAll: () => Promise<void>
  fetchSwarm: () => Promise<void>
  fetchNodes: () => Promise<void>
  fetchServices: () => Promise<void>
  fetchTasks: () => Promise<void>
  pushEvent: (evt: SwarmEvent) => void
}

export const useClusterStore = create<ClusterState>((set, get) => ({
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

  fetchAll: async () => {
    set({ loading: true, error: null })
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
      set({ swarm, nodes, stacks, services, tasks, networks, volumes, secrets, configs, events, loading: false, lastRefresh: Date.now() })
    } catch (err) {
      set({ loading: false, error: err instanceof Error ? err.message : 'fetch failed' })
    }
  },

  fetchSwarm: async () => {
    const swarm = await api.swarm.get()
    set({ swarm })
  },

  fetchNodes: async () => {
    const nodes = await api.nodes.list()
    set({ nodes })
  },

  fetchServices: async () => {
    const [stacks, services] = await Promise.all([api.stacks.list(), api.services.list()])
    set({ stacks, services })
  },

  fetchTasks: async () => {
    const tasks = await api.tasks.list()
    set({ tasks })
  },

  pushEvent: (evt: SwarmEvent) => {
    set(state => ({ events: [evt, ...state.events].slice(0, 200) }))
  },
}))
