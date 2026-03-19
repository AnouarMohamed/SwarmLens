import type {
  SwarmInfo, Node, Stack, Service, Task, Network,
  Volume, Secret, Config, SwarmEvent, Finding,
  Incident, AuditEntry, OpsMetrics, OpsInsights, ActionOutcome,
  ListResponse, ItemResponse,
} from '../types'

const BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? '/api/v1'

class APIError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message)
    this.name = 'APIError'
  }
}

async function get<T>(path: string): Promise<T> {
  const token = localStorage.getItem('sl_token')
  const res = await fetch(`${BASE}${path}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText, code: String(res.status) }))
    throw new APIError(res.status, err.code, err.error)
  }
  return res.json()
}

async function post<T>(path: string, body?: unknown): Promise<T> {
  const token = localStorage.getItem('sl_token')
  const res = await fetch(`${BASE}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText, code: String(res.status) }))
    throw new APIError(res.status, err.code, err.error)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

async function put<T>(path: string, body: unknown): Promise<T> {
  const token = localStorage.getItem('sl_token')
  const res = await fetch(`${BASE}${path}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText, code: String(res.status) }))
    throw new APIError(res.status, err.code, err.error)
  }
  return res.json()
}

async function del<T>(path: string): Promise<T> {
  const token = localStorage.getItem('sl_token')
  const res = await fetch(`${BASE}${path}`, {
    method: 'DELETE',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText, code: String(res.status) }))
    throw new APIError(res.status, err.code, err.error)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

function sse(path: string): EventSource {
  const token = localStorage.getItem('sl_token')
  const url = token
    ? `${BASE}${path}?token=${encodeURIComponent(token)}`
    : `${BASE}${path}`
  return new EventSource(url)
}

export const api = {
  swarm: {
    get: () => get<ItemResponse<SwarmInfo>>('/swarm').then(r => r.data),
  },
  nodes: {
    list: () => get<ListResponse<Node>>('/nodes').then(r => r.data),
    get: (id: string) => get<ItemResponse<Node>>(`/nodes/${id}`).then(r => r.data),
    drain: (id: string) => post(`/nodes/${id}/drain`),
    activate: (id: string) => post(`/nodes/${id}/activate`),
  },
  stacks: {
    list: () => get<ListResponse<Stack>>('/stacks').then(r => r.data),
    get: (name: string) => get<ItemResponse<Stack>>(`/stacks/${name}`).then(r => r.data),
    deploy: (name: string, body: unknown) => post(`/stacks/${name}/deploy`, body),
    remove: (name: string) => del(`/stacks/${name}`),
  },
  services: {
    list: (stack?: string) => get<ListResponse<Service>>(`/services${stack ? `?stack=${stack}` : ''}`).then(r => r.data),
    get: (id: string) => get<ItemResponse<Service>>(`/services/${id}`).then(r => r.data),
    scale: (id: string, replicas: number) => post(`/services/${id}/scale`, { replicas }),
    restart: (id: string) => post(`/services/${id}/restart`),
    update: (id: string, body: unknown) => post(`/services/${id}/update`, body),
    rollback: (id: string) => post(`/services/${id}/rollback`),
  },
  tasks: {
    list: (params?: Record<string, string>) => {
      const qs = params ? '?' + new URLSearchParams(params).toString() : ''
      return get<ListResponse<Task>>(`/tasks${qs}`).then(r => r.data)
    },
    get: (id: string) => get<ItemResponse<Task>>(`/tasks/${id}`).then(r => r.data),
  },
  networks: {
    list: () => get<ListResponse<Network>>('/networks').then(r => r.data),
  },
  volumes: {
    list: () => get<ListResponse<Volume>>('/volumes').then(r => r.data),
  },
  secrets: {
    list: () => get<ListResponse<Secret>>('/secrets').then(r => r.data),
  },
  configs: {
    list: () => get<ListResponse<Config>>('/configs').then(r => r.data),
  },
  events: {
    list: (type?: string) => get<ListResponse<SwarmEvent>>(`/events${type ? `?type=${type}` : ''}`).then(r => r.data),
    stream: () => sse('/stream/events'),
  },
  diagnostics: {
    list: (severity?: string) => get<ListResponse<Finding>>(`/diagnostics${severity ? `?severity=${severity}` : ''}`).then(r => r.data),
    run: () => post<ListResponse<Finding>>('/diagnostics/run').then(r => r.data),
    get: (id: string) => get<ItemResponse<Finding>>(`/diagnostics/${id}`).then(r => r.data),
  },
  incidents: {
    list: () => get<ListResponse<Incident>>('/incidents').then(r => r.data),
    create: (body: unknown) => post<ItemResponse<Incident>>('/incidents', body).then(r => r.data),
    get: (id: string) => get<ItemResponse<Incident>>(`/incidents/${id}`).then(r => r.data),
    update: (id: string, body: unknown) => put<ItemResponse<Incident>>(`/incidents/${id}`, body).then(r => r.data),
    resolve: (id: string) => post<ItemResponse<Incident>>(`/incidents/${id}/resolve`).then(r => r.data),
  },
  audit: {
    list: (limit = 50, offset = 0) => get<ListResponse<AuditEntry>>(`/audit?limit=${limit}&offset=${offset}`).then(r => r.data),
  },
  ops: {
    metrics: () => get<ItemResponse<OpsMetrics>>('/ops/metrics').then(r => r.data),
    insights: () => get<ItemResponse<OpsInsights>>('/ops/insights').then(r => r.data),
  },
  actions: {
    execute: (payload: { action: string; resource?: string; resourceID?: string; params?: Record<string, unknown> }) =>
      post<ItemResponse<ActionOutcome>>('/actions/execute', payload).then(r => r.data),
  },
  assistant: {
    chatStream: async (
      prompt: string,
      handlers: {
        onEvent?: (event: string, payload: unknown) => void
        onDone?: () => void
        onError?: (message: string) => void
      } = {},
    ) => {
      const token = localStorage.getItem('sl_token')
      const res = await fetch(`${BASE}/assistant/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({ prompt }),
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
          let payload: unknown = data
          try {
            payload = JSON.parse(data)
          } catch {
            payload = data
          }
          handlers.onEvent?.(event, payload)
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
