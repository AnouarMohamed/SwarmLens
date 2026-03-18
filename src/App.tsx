import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from './components/layout/Layout'
import { OverviewView }    from './views/overview/OverviewView'
import { NodesView }       from './views/nodes/NodesView'
import { StacksView }      from './views/stacks/StacksView'
import { ServicesView }    from './views/services/ServicesView'
import { TasksView }       from './views/tasks/TasksView'
import { NetworksView }    from './views/networks/NetworksView'
import { VolumesView }     from './views/volumes/VolumesView'
import { SecretsView }     from './views/secrets/SecretsView'
import { DiagnosticsView } from './views/diagnostics/DiagnosticsView'
import { IncidentsView }   from './views/incidents/IncidentsView'
import { AuditView }       from './views/audit/AuditView'
import { AssistantPanel }  from './views/assistant/AssistantPanel'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index          element={<OverviewView />} />
          <Route path="stacks"      element={<StacksView />} />
          <Route path="services"    element={<ServicesView />} />
          <Route path="tasks"       element={<TasksView />} />
          <Route path="nodes"       element={<NodesView />} />
          <Route path="networks"    element={<NetworksView />} />
          <Route path="volumes"     element={<VolumesView />} />
          <Route path="secrets"     element={<SecretsView />} />
          <Route path="diagnostics" element={<DiagnosticsView />} />
          <Route path="incidents"   element={<IncidentsView />} />
          <Route path="audit"       element={<AuditView />} />
          <Route path="assistant"   element={<AssistantPanel />} />
          <Route path="*"           element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
