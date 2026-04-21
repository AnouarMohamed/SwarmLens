import type { paths } from '../types/controlplane.generated'
import type {
  ActionExecuteRequest,
  ActionReasonRequest,
  ApprovalStatus,
  AssistantChatRequest,
  AssistantSessionCreateRequest,
  ClusterCreateRequest,
  ClusterUpdateRequest,
  IncidentCreateRequest,
  IncidentUpdateRequest,
  ServiceScaleRequest,
  ServiceUpdateRequest,
} from '../types'

const BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? '/api/v1'

type ContractPath = keyof paths
type HttpMethod = 'get' | 'post' | 'put' | 'delete'

type Operation<Path extends ContractPath, Method extends HttpMethod> = NonNullable<paths[Path][Method]>

type JsonRequestBody<Path extends ContractPath, Method extends HttpMethod> =
  Operation<Path, Method> extends {
    requestBody?: {
      content: {
        'application/json': infer Body
      }
    }
  }
    ? Body
    : never

type JsonResponse<
  Path extends ContractPath,
  Method extends HttpMethod,
  Status extends number,
> = Operation<Path, Method> extends { responses: infer Responses }
  ? Status extends keyof Responses
    ? Responses[Status] extends {
        content: {
          'application/json': infer Body
        }
      }
      ? Body
      : never
    : never
  : never

type QueryParams<Path extends ContractPath, Method extends HttpMethod> =
  Operation<Path, Method> extends { parameters: { query?: infer Query } } ? Query : never

type ServiceListQuery = QueryParams<'/clusters/{clusterID}/services', 'get'>
type TaskListQuery = QueryParams<'/clusters/{clusterID}/tasks', 'get'>
type EventListQuery = QueryParams<'/clusters/{clusterID}/events', 'get'>
type DiagnosticsListQuery = QueryParams<'/clusters/{clusterID}/diagnostics', 'get'>
type AuditListQuery = QueryParams<'/clusters/{clusterID}/audit', 'get'>
type ApprovalListQuery = QueryParams<'/clusters/{clusterID}/approvals', 'get'>
type ActionOutcomeData = JsonResponse<'/clusters/{clusterID}/actions/execute', 'post', 200>['data']

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

function withQuery(
  path: string,
  query?: Record<string, string | number | boolean | null | undefined>,
) {
  if (!query) return path
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null || value === '') continue
    search.set(key, String(value))
  }
  const encoded = search.toString()
  return encoded ? `${path}?${encoded}` : path
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

function getContract<Path extends ContractPath>(path: string) {
  return get<JsonResponse<Path, 'get', 200>>(path)
}

function postContract<Path extends ContractPath, Status extends number = 200>(
  path: string,
  body?: JsonRequestBody<Path, 'post'>,
) {
  return post<JsonResponse<Path, 'post', Status>>(path, body)
}

function putContract<Path extends ContractPath>(path: string, body: JsonRequestBody<Path, 'put'>) {
  return put<JsonResponse<Path, 'put', 200>>(path, body)
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function readReplicaCount(params: ActionExecuteRequest['params']) {
  const replicas = params?.replicas
  return typeof replicas === 'number' && Number.isFinite(replicas) ? replicas : null
}

function routeKnownAction(payload: ActionExecuteRequest): Promise<ActionOutcomeData> | null {
  if (!payload.resourceID) return null

  switch (payload.action) {
    case 'node.drain':
      return postContract<'/clusters/{clusterID}/nodes/{id}/drain'>(
        clusterPath(`/nodes/${payload.resourceID}/drain`),
        { reason: payload.reason },
      ).then((r) => r.data)
    case 'node.activate':
      return postContract<'/clusters/{clusterID}/nodes/{id}/activate'>(
        clusterPath(`/nodes/${payload.resourceID}/activate`),
        { reason: payload.reason },
      ).then((r) => r.data)
    case 'service.restart':
      return postContract<'/clusters/{clusterID}/services/{id}/restart'>(
        clusterPath(`/services/${payload.resourceID}/restart`),
        { reason: payload.reason },
      ).then((r) => r.data)
    case 'service.scale': {
      const replicas = readReplicaCount(payload.params)
      if (replicas === null) return null
      return postContract<'/clusters/{clusterID}/services/{id}/scale'>(
        clusterPath(`/services/${payload.resourceID}/scale`),
        { replicas, reason: payload.reason },
      ).then((r) => r.data)
    }
    case 'service.update':
      if (!isRecord(payload.params)) return null
      return postContract<'/clusters/{clusterID}/services/{id}/update'>(
        clusterPath(`/services/${payload.resourceID}/update`),
        { ...payload.params, reason: payload.reason } as ServiceUpdateRequest,
      ).then((r) => r.data)
    case 'service.rollback':
      return postContract<'/clusters/{clusterID}/services/{id}/rollback'>(
        clusterPath(`/services/${payload.resourceID}/rollback`),
        { reason: payload.reason },
      ).then((r) => r.data)
    case 'task.restart':
      return postContract<'/clusters/{clusterID}/tasks/{id}/restart'>(
        clusterPath(`/tasks/${payload.resourceID}/restart`),
        { reason: payload.reason },
      ).then((r) => r.data)
    default:
      return null
  }
}

function sse(path: string): EventSource {
  const token = readToken()
  const url = token ? `${BASE}${path}?token=${encodeURIComponent(token)}` : `${BASE}${path}`
  return new EventSource(url)
}

export const api = {
  auth: {
    me: () => getContract<'/auth/me'>('/auth/me').then((r) => r.data),
    loginUrl: (returnTo = window.location.pathname + window.location.search) =>
      `${BASE}/auth/login?returnTo=${encodeURIComponent(returnTo)}`,
    logout: () => postContract<'/auth/logout'>('/auth/logout').then((r) => r.data),
  },
  clusters: {
    list: () => getContract<'/clusters'>('/clusters').then((r) => r.data),
    get: (id: string) => getContract<'/clusters/{clusterID}'>(`/clusters/${id}`).then((r) => r.data),
    create: (body: ClusterCreateRequest) =>
      postContract<'/clusters', 201>('/clusters', body).then((r) => r.data),
    update: (id: string, body: ClusterUpdateRequest) =>
      putContract<'/clusters/{clusterID}'>(`/clusters/${id}`, body).then((r) => r.data),
  },
  swarm: {
    get: () => getContract<'/clusters/{clusterID}/swarm'>(clusterPath('/swarm')).then((r) => r.data),
  },
  nodes: {
    list: () => getContract<'/clusters/{clusterID}/nodes'>(clusterPath('/nodes')).then((r) => r.data),
    get: (id: string) => getContract<'/clusters/{clusterID}/nodes/{id}'>(clusterPath(`/nodes/${id}`)).then((r) => r.data),
    drain: (id: string, body: ActionReasonRequest) =>
      postContract<'/clusters/{clusterID}/nodes/{id}/drain'>(clusterPath(`/nodes/${id}/drain`), body).then((r) => r.data),
    activate: (id: string, body: ActionReasonRequest) =>
      postContract<'/clusters/{clusterID}/nodes/{id}/activate'>(clusterPath(`/nodes/${id}/activate`), body).then((r) => r.data),
  },
  stacks: {
    list: () => getContract<'/clusters/{clusterID}/stacks'>(clusterPath('/stacks')).then((r) => r.data),
    get: (name: string) =>
      getContract<'/clusters/{clusterID}/stacks/{name}'>(clusterPath(`/stacks/${name}`)).then((r) => r.data),
    deploy: (name: string, body: unknown) => post(clusterPath(`/stacks/${name}/deploy`), body),
    remove: (name: string) => del(clusterPath(`/stacks/${name}`)),
  },
  services: {
    list: (stack?: ServiceListQuery extends { stack?: infer Value } ? Value : never) =>
      getContract<'/clusters/{clusterID}/services'>(withQuery(clusterPath('/services'), { stack })).then((r) => r.data),
    get: (id: string) =>
      getContract<'/clusters/{clusterID}/services/{id}'>(clusterPath(`/services/${id}`)).then((r) => r.data),
    scale: (id: string, body: ServiceScaleRequest) =>
      postContract<'/clusters/{clusterID}/services/{id}/scale'>(clusterPath(`/services/${id}/scale`), body).then((r) => r.data),
    restart: (id: string, body: ActionReasonRequest) =>
      postContract<'/clusters/{clusterID}/services/{id}/restart'>(clusterPath(`/services/${id}/restart`), body).then((r) => r.data),
    update: (id: string, body: ServiceUpdateRequest) =>
      postContract<'/clusters/{clusterID}/services/{id}/update'>(clusterPath(`/services/${id}/update`), body).then((r) => r.data),
    rollback: (id: string, body: ActionReasonRequest) =>
      postContract<'/clusters/{clusterID}/services/{id}/rollback'>(clusterPath(`/services/${id}/rollback`), body).then((r) => r.data),
  },
  tasks: {
    list: (query?: TaskListQuery) =>
      getContract<'/clusters/{clusterID}/tasks'>(withQuery(clusterPath('/tasks'), query)).then((r) => r.data),
    get: (id: string) =>
      getContract<'/clusters/{clusterID}/tasks/{id}'>(clusterPath(`/tasks/${id}`)).then((r) => r.data),
    restart: (id: string, body: ActionReasonRequest) =>
      postContract<'/clusters/{clusterID}/tasks/{id}/restart'>(clusterPath(`/tasks/${id}/restart`), body).then((r) => r.data),
  },
  networks: {
    list: () => getContract<'/clusters/{clusterID}/networks'>(clusterPath('/networks')).then((r) => r.data),
  },
  volumes: {
    list: () => getContract<'/clusters/{clusterID}/volumes'>(clusterPath('/volumes')).then((r) => r.data),
  },
  secrets: {
    list: () => getContract<'/clusters/{clusterID}/secrets'>(clusterPath('/secrets')).then((r) => r.data),
  },
  configs: {
    list: () => getContract<'/clusters/{clusterID}/configs'>(clusterPath('/configs')).then((r) => r.data),
  },
  events: {
    list: (type?: EventListQuery extends { type?: infer Value } ? Value : never) =>
      getContract<'/clusters/{clusterID}/events'>(withQuery(clusterPath('/events'), { type })).then((r) => r.data),
    stream: () => sse(clusterPath('/stream/events')),
  },
  diagnostics: {
    list: (severity?: DiagnosticsListQuery extends { severity?: infer Value } ? Value : never) =>
      getContract<'/clusters/{clusterID}/diagnostics'>(withQuery(clusterPath('/diagnostics'), { severity })).then((r) => r.data),
    run: (reason: string) =>
      postContract<'/clusters/{clusterID}/actions/execute'>(clusterPath('/actions/execute'), { action: 'diagnostics.run', reason }).then((r) => r.data),
    get: (id: string) =>
      getContract<'/clusters/{clusterID}/diagnostics/{id}'>(clusterPath(`/diagnostics/${id}`)).then((r) => r.data),
  },
  incidents: {
    list: () => getContract<'/clusters/{clusterID}/incidents'>(clusterPath('/incidents')).then((r) => r.data),
    create: (body: IncidentCreateRequest) =>
      postContract<'/clusters/{clusterID}/incidents', 201>(clusterPath('/incidents'), body).then((r) => r.data),
    get: (id: string) =>
      getContract<'/clusters/{clusterID}/incidents/{id}'>(clusterPath(`/incidents/${id}`)).then((r) => r.data),
    update: (id: string, body: IncidentUpdateRequest) =>
      putContract<'/clusters/{clusterID}/incidents/{id}'>(clusterPath(`/incidents/${id}`), body).then((r) => r.data),
    resolve: (id: string) =>
      postContract<'/clusters/{clusterID}/incidents/{id}/resolve'>(clusterPath(`/incidents/${id}/resolve`)).then((r) => r.data),
  },
  audit: {
    list: (limit = 50, offset = 0) =>
      getContract<'/clusters/{clusterID}/audit'>(
        withQuery(clusterPath('/audit'), { limit, offset } satisfies AuditListQuery),
      ).then((r) => r.data),
  },
  ops: {
    metrics: () => getContract<'/clusters/{clusterID}/ops/metrics'>(clusterPath('/ops/metrics')).then((r) => r.data),
    insights: () => getContract<'/clusters/{clusterID}/ops/insights'>(clusterPath('/ops/insights')).then((r) => r.data),
  },
  actions: {
    list: () => getContract<'/clusters/{clusterID}/actions'>(clusterPath('/actions')).then((r) => r.data),
    execute: (payload: ActionExecuteRequest) => {
      const direct = routeKnownAction(payload)
      if (direct) return direct
      return postContract<'/clusters/{clusterID}/actions/execute'>(clusterPath('/actions/execute'), payload).then((r) => r.data)
    },
  },
  approvals: {
    list: (status?: ApprovalStatus) =>
      getContract<'/clusters/{clusterID}/approvals'>(
        withQuery(clusterPath('/approvals'), { status } satisfies ApprovalListQuery),
      ).then((r) => r.data),
    approve: (id: string) =>
      postContract<'/clusters/{clusterID}/approvals/{id}/approve'>(clusterPath(`/approvals/${id}/approve`)).then((r) => r.data),
    reject: (id: string) =>
      postContract<'/clusters/{clusterID}/approvals/{id}/reject'>(clusterPath(`/approvals/${id}/reject`)).then((r) => r.data),
  },
  assistant: {
    sessions: () =>
      getContract<'/clusters/{clusterID}/assistant/sessions'>(clusterPath('/assistant/sessions')).then((r) => r.data),
    createSession: (body?: AssistantSessionCreateRequest) =>
      postContract<'/clusters/{clusterID}/assistant/sessions', 201>(clusterPath('/assistant/sessions'), body ?? {}).then((r) => r.data),
    getSession: (id: string) =>
      getContract<'/clusters/{clusterID}/assistant/sessions/{id}'>(clusterPath(`/assistant/sessions/${id}`)).then((r) => r.data),
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
