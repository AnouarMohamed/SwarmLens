import type {
  SwarmInfo,
  Node,
  Stack,
  Service,
  Task,
  Network,
  Volume,
  Secret,
  Config,
  SwarmEvent,
  Finding,
  Incident,
  AuditEntry,
  OpsMetrics,
  OpsInsights,
  ActionOutcome,
  ListResponse,
  ItemResponse,
  Cluster,
  ClusterCreateRequest,
  ClusterUpdateRequest,
  AuthIdentity,
  ActionRun,
  ActionExecuteRequest,
  ApprovalStatus,
  ApprovalRequest,
  IncidentCreateRequest,
  IncidentUpdateRequest,
  AssistantSession,
  AssistantSessionCreateRequest,
  AssistantChatRequest,
} from '../types'

const BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? '/api/v1'

class APIError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message)
    this.name = 'APIError'
  }
}

function readToken() {
  return localStorage.getItem('sl_token') ?? ''
}

function readCSRFToken() {
  return sessionStorage.getItem('sl_csrf') ?? ''
}

function selectedClusterID() {
  return localStorage.getItem('sl_cluster_id')?.trim() ?? ''
}

function clusterPath(path: string) {
  const clusterID = selectedClusterID()
  if (!clusterID) return path
  return `/clusters/${encodeURIComponent(clusterID)}${path}`
}

function headersForRequest(method: 'GET' | 'POST' | 'PUT' | 'DELETE', json = false) {
  const token = readToken()
  const csrfToken = readCSRFToken()
  const headers: Record<string, string> = {}
  if (json) headers['Content-Type'] = 'application/json'
  if (token) headers.Authorization = `Bearer ${token}`
  if (method !== 'GET' && csrfToken) headers['X-CSRF-Token'] = csrfToken
  return headers
}

async function request<T>(method: 'GET' | 'POST' | 'PUT' | 'DELETE', path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method,
    credentials: 'include',
    headers: headersForRequest(method, body !== undefined),
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText, code: String(res.status) }))
    throw new APIError(res.status, err.code, err.error)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

function get<T>(path: string) {
  return request<T>('GET', path)
}

function post<T>(path: string, body?: unknown) {
  return request<T>('POST', path, body)
}

function put<T>(path: string, body: unknown) {
  return request<T>('PUT', path, body)
}

function del<T>(path: string) {
  return request<T>('DELETE', path)
}

function sse(path: string): EventSource {
  const token = readToken()
  const url = token ? `${BASE}${path}?token=${encodeURIComponent(token)}` : `${BASE}${path}`
  return new EventSource(url)
}

export const api = {
  auth: {
    me: () => get<ItemResponse<AuthIdentity>>('/auth/me').then((r) => r.data),
    loginUrl: (returnTo = window.location.pathname + window.location.search) =>
      `${BASE}/auth/login?returnTo=${encodeURIComponent(returnTo)}`,
    logout: () => post<ItemResponse<{ ok: boolean }>>('/auth/logout').then((r) => r.data),
  },
  clusters: {
    list: () => get<ListResponse<Cluster>>('/clusters').then((r) => r.data),
    get: (id: string) => get<ItemResponse<Cluster>>(`/clusters/${id}`).then((r) => r.data),
    create: (body: ClusterCreateRequest) => post<ItemResponse<Cluster>>('/clusters', body).then((r) => r.data),
    update: (id: string, body: ClusterUpdateRequest) =>
      put<ItemResponse<Cluster>>(`/clusters/${id}`, body).then((r) => r.data),
  },
  swarm: {
    get: () => get<ItemResponse<SwarmInfo>>(clusterPath('/swarm')).then((r) => r.data),
  },
  nodes: {
    list: () => get<ListResponse<Node>>(clusterPath('/nodes')).then((r) => r.data),
    get: (id: string) => get<ItemResponse<Node>>(clusterPath(`/nodes/${id}`)).then((r) => r.data),
    drain: (id: string, reason: string) => post(clusterPath(`/nodes/${id}/drain?reason=${encodeURIComponent(reason)}`)),
    activate: (id: string, reason: string) => post(clusterPath(`/nodes/${id}/activate?reason=${encodeURIComponent(reason)}`)),
  },
  stacks: {
    list: () => get<ListResponse<Stack>>(clusterPath('/stacks')).then((r) => r.data),
    get: (name: string) => get<ItemResponse<Stack>>(clusterPath(`/stacks/${name}`)).then((r) => r.data),
    deploy: (name: string, body: unknown) => post(clusterPath(`/stacks/${name}/deploy`), body),
    remove: (name: string) => del(clusterPath(`/stacks/${name}`)),
  },
  services: {
    list: (stack?: string) =>
      get<ListResponse<Service>>(clusterPath(`/services${stack ? `?stack=${stack}` : ''}`)).then((r) => r.data),
    get: (id: string) => get<ItemResponse<Service>>(clusterPath(`/services/${id}`)).then((r) => r.data),
    scale: (id: string, replicas: number, reason: string) =>
      post(clusterPath(`/services/${id}/scale`), { replicas, reason }),
    restart: (id: string, reason: string) => post(clusterPath(`/services/${id}/restart`), { reason }),
    update: (id: string, body: unknown) => post(clusterPath(`/services/${id}/update`), body),
    rollback: (id: string, reason: string) => post(clusterPath(`/services/${id}/rollback`), { reason }),
  },
  tasks: {
    list: (params?: Record<string, string>) => {
      const qs = params ? `?${new URLSearchParams(params).toString()}` : ''
      return get<ListResponse<Task>>(clusterPath(`/tasks${qs}`)).then((r) => r.data)
    },
    get: (id: string) => get<ItemResponse<Task>>(clusterPath(`/tasks/${id}`)).then((r) => r.data),
  },
  networks: {
    list: () => get<ListResponse<Network>>(clusterPath('/networks')).then((r) => r.data),
  },
  volumes: {
    list: () => get<ListResponse<Volume>>(clusterPath('/volumes')).then((r) => r.data),
  },
  secrets: {
    list: () => get<ListResponse<Secret>>(clusterPath('/secrets')).then((r) => r.data),
  },
  configs: {
    list: () => get<ListResponse<Config>>(clusterPath('/configs')).then((r) => r.data),
  },
  events: {
    list: (type?: string) =>
      get<ListResponse<SwarmEvent>>(clusterPath(`/events${type ? `?type=${type}` : ''}`)).then((r) => r.data),
    stream: () => sse(clusterPath('/stream/events')),
  },
  diagnostics: {
    list: (severity?: string) =>
      get<ListResponse<Finding>>(clusterPath(`/diagnostics${severity ? `?severity=${severity}` : ''}`)).then((r) => r.data),
    run: (reason: string) =>
      post<ItemResponse<ActionOutcome>>(clusterPath('/actions/execute'), { action: 'diagnostics.run', reason }).then((r) => r.data),
    get: (id: string) => get<ItemResponse<Finding>>(clusterPath(`/diagnostics/${id}`)).then((r) => r.data),
  },
  incidents: {
    list: () => get<ListResponse<Incident>>(clusterPath('/incidents')).then((r) => r.data),
    create: (body: IncidentCreateRequest) =>
      post<ItemResponse<Incident>>(clusterPath('/incidents'), body).then((r) => r.data),
    get: (id: string) => get<ItemResponse<Incident>>(clusterPath(`/incidents/${id}`)).then((r) => r.data),
    update: (id: string, body: IncidentUpdateRequest) =>
      put<ItemResponse<Incident>>(clusterPath(`/incidents/${id}`), body).then((r) => r.data),
    resolve: (id: string) => post<ItemResponse<Incident>>(clusterPath(`/incidents/${id}/resolve`)).then((r) => r.data),
  },
  audit: {
    list: (limit = 50, offset = 0) =>
      get<ListResponse<AuditEntry>>(clusterPath(`/audit?limit=${limit}&offset=${offset}`)).then((r) => r.data),
  },
  ops: {
    metrics: () => get<ItemResponse<OpsMetrics>>(clusterPath('/ops/metrics')).then((r) => r.data),
    insights: () => get<ItemResponse<OpsInsights>>(clusterPath('/ops/insights')).then((r) => r.data),
  },
  actions: {
    list: () => get<ListResponse<ActionRun>>(clusterPath('/actions')).then((r) => r.data),
    execute: (payload: ActionExecuteRequest) =>
      post<ItemResponse<ActionOutcome>>(clusterPath('/actions/execute'), payload).then((r) => r.data),
  },
  approvals: {
    list: (status?: ApprovalStatus) =>
      get<ListResponse<ApprovalRequest>>(clusterPath(`/approvals${status ? `?status=${status}` : ''}`)).then((r) => r.data),
    approve: (id: string) => post<ItemResponse<ActionOutcome>>(clusterPath(`/approvals/${id}/approve`)).then((r) => r.data),
    reject: (id: string) => post<ItemResponse<ActionOutcome>>(clusterPath(`/approvals/${id}/reject`)).then((r) => r.data),
  },
  assistant: {
    sessions: () => get<ListResponse<AssistantSession>>(clusterPath('/assistant/sessions')).then((r) => r.data),
    createSession: (body?: AssistantSessionCreateRequest) =>
      post<ItemResponse<AssistantSession>>(clusterPath('/assistant/sessions'), body ?? {}).then((r) => r.data),
    getSession: (id: string) => get<ItemResponse<AssistantSession>>(clusterPath(`/assistant/sessions/${id}`)).then((r) => r.data),
    chatStream: async (
      payload: AssistantChatRequest,
      handlers: {
        onEvent?: (event: string, payload: unknown) => void
        onDone?: () => void
        onError?: (message: string) => void
      } = {},
    ) => {
      const res = await fetch(`${BASE}${clusterPath('/assistant/chat')}`, {
        method: 'POST',
        credentials: 'include',
        headers: headersForRequest('POST', true),
        body: JSON.stringify(payload),
      })

      if (!res.ok || !res.body) {
        const err = await res.json().catch(() => ({ error: res.statusText }))
        handlers.onError?.(err.error || 'assistant stream failed')
        return
      }

      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const chunks = buffer.split('\n\n')
        buffer = chunks.pop() ?? ''
        for (const chunk of chunks) {
          const lines = chunk.split('\n')
          const eventLine = lines.find((line) => line.startsWith('event:'))
          const dataLine = lines.find((line) => line.startsWith('data:'))
          const event = eventLine?.slice(6).trim() ?? 'message'
          const data = dataLine?.slice(5).trim() ?? '{}'
          let payloadData: unknown = data
          try {
            payloadData = JSON.parse(data)
          } catch {
            payloadData = data
          }
          handlers.onEvent?.(event, payloadData)
          if (event === 'done') handlers.onDone?.()
        }
      }
    },
  },
  obs: {
    healthz: () => get('/healthz'),
    readyz: () => get('/readyz'),
    runtime: () => get('/runtime'),
  },
}

export { APIError }
