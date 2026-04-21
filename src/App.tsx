import { Suspense, lazy } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from './components/layout/Layout'

const OverviewView = lazy(() =>
  import('./views/overview/OverviewView').then((module) => ({ default: module.OverviewView })),
)
const NodesView = lazy(() =>
  import('./views/nodes/NodesView').then((module) => ({ default: module.NodesView })),
)
const StacksView = lazy(() =>
  import('./views/stacks/StacksView').then((module) => ({ default: module.StacksView })),
)
const ServicesView = lazy(() =>
  import('./views/services/ServicesView').then((module) => ({ default: module.ServicesView })),
)
const TasksView = lazy(() =>
  import('./views/tasks/TasksView').then((module) => ({ default: module.TasksView })),
)
const NetworksView = lazy(() =>
  import('./views/networks/NetworksView').then((module) => ({ default: module.NetworksView })),
)
const VolumesView = lazy(() =>
  import('./views/volumes/VolumesView').then((module) => ({ default: module.VolumesView })),
)
const SecretsView = lazy(() =>
  import('./views/secrets/SecretsView').then((module) => ({ default: module.SecretsView })),
)
const DiagnosticsView = lazy(() =>
  import('./views/diagnostics/DiagnosticsView').then((module) => ({ default: module.DiagnosticsView })),
)
const IncidentsView = lazy(() =>
  import('./views/incidents/IncidentsView').then((module) => ({ default: module.IncidentsView })),
)
const AuditView = lazy(() =>
  import('./views/audit/AuditView').then((module) => ({ default: module.AuditView })),
)
const ApprovalsView = lazy(() =>
  import('./views/approvals/ApprovalsView').then((module) => ({ default: module.ApprovalsView })),
)
const AssistantPanel = lazy(() =>
  import('./views/assistant/AssistantPanel').then((module) => ({ default: module.AssistantPanel })),
)

export default function App() {
  return (
    <BrowserRouter>
      <Suspense fallback={<div className="table-loading">Loading view...</div>}>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<OverviewView />} />
            <Route path="stacks" element={<StacksView />} />
            <Route path="services" element={<ServicesView />} />
            <Route path="tasks" element={<TasksView />} />
            <Route path="nodes" element={<NodesView />} />
            <Route path="networks" element={<NetworksView />} />
            <Route path="volumes" element={<VolumesView />} />
            <Route path="secrets" element={<SecretsView />} />
            <Route path="diagnostics" element={<DiagnosticsView />} />
            <Route path="incidents" element={<IncidentsView />} />
            <Route path="audit" element={<AuditView />} />
            <Route path="approvals" element={<ApprovalsView />} />
            <Route path="assistant" element={<AssistantPanel />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}
