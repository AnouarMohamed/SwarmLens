import { useEffect } from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Topbar } from './Topbar'
import { useClusterStore } from '../../store/clusterStore'
import { useEventStream } from '../../hooks/useEventStream'

export function Layout() {
  const fetchAll = useClusterStore(s => s.fetchAll)
  useEventStream()

  useEffect(() => {
    fetchAll()
  }, [fetchAll])

  return (
    <div className="layout">
      <Sidebar />
      <div className="main">
        <Topbar />
        <main className="content">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
