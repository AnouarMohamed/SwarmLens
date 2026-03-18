import { create } from 'zustand'
import { api } from '../lib/api'
import type { Incident } from '../types'

interface IncidentState {
  incidents: Incident[]
  loading: boolean

  fetch: () => Promise<void>
  create: (body: unknown) => Promise<Incident>
  resolve: (id: string) => Promise<void>
}

export const useIncidentStore = create<IncidentState>((set, get) => ({
  incidents: [],
  loading: false,

  fetch: async () => {
    set({ loading: true })
    try {
      const incidents = await api.incidents.list()
      set({ incidents, loading: false })
    } catch {
      set({ loading: false })
    }
  },

  create: async (body: unknown) => {
    const inc = await api.incidents.create(body)
    set(state => ({ incidents: [inc, ...state.incidents] }))
    return inc
  },

  resolve: async (id: string) => {
    const updated = await api.incidents.resolve(id)
    set(state => ({
      incidents: state.incidents.map(i => i.id === id ? updated : i),
    }))
  },
}))
