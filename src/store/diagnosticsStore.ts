import { create } from 'zustand'
import { api } from '../lib/api'
import type { Finding } from '../types'

interface DiagnosticsState {
  findings: Finding[]
  loading: boolean
  running: boolean
  lastRun: number | null

  fetch: () => Promise<void>
  run: () => Promise<void>
}

export const useDiagnosticsStore = create<DiagnosticsState>((set) => ({
  findings: [],
  loading: false,
  running: false,
  lastRun: null,

  fetch: async () => {
    set({ loading: true })
    try {
      const findings = await api.diagnostics.list()
      set({ findings, loading: false })
    } catch {
      set({ loading: false })
    }
  },

  run: async () => {
    set({ running: true })
    try {
      const findings = await api.diagnostics.run()
      set({ findings: findings ?? [], running: false, lastRun: Date.now() })
    } catch {
      set({ running: false })
    }
  },
}))
