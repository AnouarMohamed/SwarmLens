type GrafanaVars = Record<string, string | number | boolean>

function trimTrailingSlash(value: string) {
  return value.replace(/\/+$/, '')
}

function appBase() {
  const base = import.meta.env.BASE_URL || '/'
  if (base.startsWith('http')) return base
  if (typeof window === 'undefined') return base
  return new URL(base, window.location.origin).toString()
}

export function grafanaConfig() {
  const baseUrl = (import.meta.env.VITE_GRAFANA_URL as string | undefined)?.trim()
  const dashboardUid = (import.meta.env.VITE_GRAFANA_DASHBOARD_UID as string | undefined)?.trim()
  const orgId = (import.meta.env.VITE_GRAFANA_ORG_ID as string | undefined)?.trim() ?? '1'
  const theme = (
    (import.meta.env.VITE_GRAFANA_THEME as string | undefined)?.trim() || 'dark'
  ).toLowerCase()
  const from = (import.meta.env.VITE_GRAFANA_FROM as string | undefined)?.trim() || 'now-6h'
  const to = (import.meta.env.VITE_GRAFANA_TO as string | undefined)?.trim() || 'now'
  const refresh = (import.meta.env.VITE_GRAFANA_REFRESH as string | undefined)?.trim() || '30s'
  const kiosk = (import.meta.env.VITE_GRAFANA_KIOSK as string | undefined)?.trim() || 'tv'

  return {
    enabled: Boolean(baseUrl && dashboardUid),
    baseUrl: baseUrl ? trimTrailingSlash(baseUrl) : '',
    dashboardUid: dashboardUid ?? '',
    orgId,
    theme,
    from,
    to,
    refresh,
    kiosk,
  }
}

export function grafanaPanelUrl({ panelId, vars = {} }: { panelId: number; vars?: GrafanaVars }) {
  const cfg = grafanaConfig()
  if (!cfg.enabled) return null

  const params = new URLSearchParams({
    orgId: cfg.orgId,
    from: cfg.from,
    to: cfg.to,
    theme: cfg.theme,
    refresh: cfg.refresh,
    kiosk: cfg.kiosk,
    panelId: String(panelId),
  })

  Object.entries(vars).forEach(([key, value]) => {
    params.set(`var-${key}`, String(value))
  })

  return `${cfg.baseUrl}/d-solo/${cfg.dashboardUid}/swarmlens?${params.toString()}`
}

export function grafanaDashboardUrl(vars: GrafanaVars = {}) {
  const cfg = grafanaConfig()
  if (!cfg.enabled) return null

  const params = new URLSearchParams({
    orgId: cfg.orgId,
    from: cfg.from,
    to: cfg.to,
    theme: cfg.theme,
    refresh: cfg.refresh,
  })

  Object.entries(vars).forEach(([key, value]) => {
    params.set(`var-${key}`, String(value))
  })

  return `${cfg.baseUrl}/d/${cfg.dashboardUid}/swarmlens?${params.toString()}`
}

export function grafanaExploreLink(query: string) {
  const cfg = grafanaConfig()
  if (!cfg.enabled) return null

  const left = {
    datasource: import.meta.env.VITE_GRAFANA_DATASOURCE || 'Prometheus',
    queries: [{ refId: 'A', expr: query }],
    range: { from: cfg.from, to: cfg.to },
  }

  const params = new URLSearchParams({
    left: JSON.stringify(left),
    orgId: cfg.orgId,
  })

  return `${cfg.baseUrl}/explore?${params.toString()}`
}

export function grafanaProvisionNote() {
  return `Set VITE_GRAFANA_URL and VITE_GRAFANA_DASHBOARD_UID in ${appBase()}`
}
