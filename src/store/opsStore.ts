import { create } from 'zustand'
import { api } from '../lib/api'
import type { ActionOutcome, OpsInsights, OpsMetrics } from '../types'

interface OpsState {
  metrics: OpsMetrics | null
  insights: OpsInsights | null
  actionLog: ActionOutcome[]
  loading: boolean
  error: string | null
  refresh: () => Promise<void>
  runAction: (payload: {
    action: string
    resource?: string
    resourceID?: string
    params?: Record<string, unknown>
  }) => Promise<ActionOutcome | null>
}

export const useOpsStore = create<OpsState>((set) => ({
  metrics: null,
  insights: null,
  actionLog: [],
  loading: false,
  error: null,

  refresh: async () => {
    set({ loading: true, error: null })
    try {
      const [metrics, insights] = await Promise.all([api.ops.metrics(), api.ops.insights()])
      set({ metrics, insights, loading: false, error: null })
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : 'Unable to load operational intelligence.',
      })
    }
  },

  runAction: async (payload) => {
    try {
      const outcome = await api.actions.execute(payload)
      set((state) => ({
        actionLog: [outcome, ...state.actionLog].slice(0, 20),
      }))
      return outcome
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Action failed' })
      return null
    }
  },
}))
