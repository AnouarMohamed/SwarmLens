import { useEffect, useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Topbar } from './Topbar'
import { useClusterStore } from '../../store/clusterStore'
import { useControlPlaneStore } from '../../store/controlPlaneStore'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import { useSessionStore } from '../../store/sessionStore'
import { useEventStream } from '../../hooks/useEventStream'

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

export function Layout() {
  const fetchAll = useClusterStore((s) => s.fetchAll)
  const fetchDiagnostics = useDiagnosticsStore((s) => s.fetch)
  const fetchSession = useSessionStore((s) => s.fetch)
  const fetchClusters = useControlPlaneStore((s) => s.fetchClusters)
  const refreshWorkflow = useControlPlaneStore((s) => s.refreshWorkflow)
  const selectedClusterID = useControlPlaneStore((s) => s.selectedClusterID)
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const location = useLocation()

  useEventStream()

  useEffect(() => {
    void fetchSession()
    void fetchClusters()
  }, [fetchClusters, fetchSession])

  useEffect(() => {
    void fetchAll()
    void fetchDiagnostics()
    void refreshWorkflow()
  }, [fetchAll, fetchDiagnostics, refreshWorkflow, selectedClusterID])

  useEffect(() => {
    setSidebarOpen(false)
  }, [location.pathname])

  const isOverview = location.pathname === '/'

  return (
    <div className="min-h-screen bg-app text-text-primary">
      <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />

      <div className="min-h-screen lg:pl-[256px]">
        <Topbar onOpenSidebar={() => setSidebarOpen(true)} />

        <main
          className={cn(
            'transition-opacity duration-150',
            isOverview ? 'px-0 pb-16 pt-0' : 'px-4 pb-12 pt-6 sm:px-6 lg:px-8',
          )}
        >
          <div className={cn(isOverview ? 'w-full' : 'mx-auto w-full max-w-shell')}>
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
