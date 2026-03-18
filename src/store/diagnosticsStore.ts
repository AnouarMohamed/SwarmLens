import { create } from 'zustand'
import { api } from '../lib/api'
import type { Finding } from '../types'

interface DiagnosticsState {
  findings: Finding[]
  loading: boolean
  running: boolean
  lastRun: number | null
  lastDurationMs: number | null
  error: string | null

  fetch: () => Promise<void>
  run: () => Promise<void>
}

export const useDiagnosticsStore = create<DiagnosticsState>((set) => ({
  findings: [],
  loading: false,
  running: false,
  lastRun: null,
  lastDurationMs: null,
  error: null,

  fetch: async () => {
    set({ loading: true, error: null })
    try {
      const findings = await api.diagnostics.list()
      set({ findings, loading: false, error: null })
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : 'Unable to load diagnostics findings.',
      })
    }
  },

  run: async () => {
    set({ running: true, error: null })
    const start = Date.now()
    try {
      const findings = await api.diagnostics.run()
      const finish = Date.now()
      set({
        findings: findings ?? [],
        running: false,
        lastRun: finish,
        lastDurationMs: finish - start,
        error: null,
      })
    } catch (err) {
      set({
        running: false,
        error: err instanceof Error ? err.message : 'Diagnostics run failed.',
      })
    }
  },
}))
