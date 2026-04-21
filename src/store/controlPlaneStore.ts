import { create } from 'zustand'
import { api } from '../lib/api'
import type { ActionOutcome, ActionRun, ApprovalRequest, Cluster } from '../types'

interface ControlPlaneState {
  clusters: Cluster[]
  selectedClusterID: string
  actionRuns: ActionRun[]
  approvals: ApprovalRequest[]
  loading: boolean
  error: string | null
  fetchClusters: () => Promise<void>
  setSelectedCluster: (clusterID: string) => Promise<void>
  refreshWorkflow: () => Promise<void>
  approve: (id: string) => Promise<ActionOutcome | null>
  reject: (id: string) => Promise<ActionOutcome | null>
}

function initialClusterID() {
  return localStorage.getItem('sl_cluster_id') ?? ''
}

export const useControlPlaneStore = create<ControlPlaneState>((set, get) => ({
  clusters: [],
  selectedClusterID: initialClusterID(),
  actionRuns: [],
  approvals: [],
  loading: false,
  error: null,

  fetchClusters: async () => {
    set({ loading: true, error: null })
    try {
      const clusters = await api.clusters.list()
      const selected = get().selectedClusterID
      const nextSelected =
        clusters.find((cluster) => cluster.id === selected)?.id ??
        clusters.find((cluster) => cluster.default)?.id ??
        clusters[0]?.id ??
        ''
      if (nextSelected) {
        localStorage.setItem('sl_cluster_id', nextSelected)
      }
      set({ clusters, selectedClusterID: nextSelected, loading: false, error: null })
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : 'Unable to load clusters.',
      })
    }
  },

  setSelectedCluster: async (clusterID) => {
    localStorage.setItem('sl_cluster_id', clusterID)
    set({ selectedClusterID: clusterID })
    await get().refreshWorkflow()
  },

  refreshWorkflow: async () => {
    set({ loading: true, error: null })
    try {
      const [actionRuns, approvals] = await Promise.all([
        api.actions.list(),
        api.approvals.list('pending'),
      ])
      set({ actionRuns, approvals, loading: false, error: null })
    } catch (err) {
      set({
        loading: false,
        error: err instanceof Error ? err.message : 'Unable to refresh workflow state.',
      })
    }
  },

  approve: async (id) => {
    try {
      const outcome = await api.approvals.approve(id)
      await get().refreshWorkflow()
      return outcome
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Approval failed.' })
      return null
    }
  },

  reject: async (id) => {
    try {
      const outcome = await api.approvals.reject(id)
      await get().refreshWorkflow()
      return outcome
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Rejection failed.' })
      return null
    }
  },
}))
