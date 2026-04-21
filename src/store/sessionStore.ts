import { create } from 'zustand'
import { api } from '../lib/api'
import type { AuthIdentity } from '../types'

interface SessionState {
  me: AuthIdentity | null
  loading: boolean
  error: string | null
  fetch: () => Promise<void>
  login: () => void
  logout: () => Promise<void>
}

export const useSessionStore = create<SessionState>((set) => ({
  me: null,
  loading: false,
  error: null,

  fetch: async () => {
    set({ loading: true, error: null })
    try {
      const me = await api.auth.me()
      if (me.csrfToken) {
        sessionStorage.setItem('sl_csrf', me.csrfToken)
      } else {
        sessionStorage.removeItem('sl_csrf')
      }
      set({ me, loading: false, error: null })
    } catch (err) {
      sessionStorage.removeItem('sl_csrf')
      set({
        me: { authenticated: false },
        loading: false,
        error: err instanceof Error ? err.message : 'Unable to load session.',
      })
    }
  },

  login: () => {
    window.location.href = api.auth.loginUrl()
  },

  logout: async () => {
    await api.auth.logout()
    sessionStorage.removeItem('sl_csrf')
    set({ me: { authenticated: false }, error: null })
  },
}))
