import { useMemo, useState } from 'react'
import { grafanaDashboardUrl, grafanaPanelUrl, grafanaProvisionNote } from '../../lib/grafana'

interface GrafanaEmbedProps {
  panelId: number
  title: string
  subtitle?: string
  height?: number
  vars?: Record<string, string | number | boolean>
}

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

export function GrafanaEmbed({ panelId, title, subtitle, height = 280, vars }: GrafanaEmbedProps) {
  const [loaded, setLoaded] = useState(false)
  const src = useMemo(() => grafanaPanelUrl({ panelId, vars }), [panelId, vars])
  const dashboardUrl = useMemo(() => grafanaDashboardUrl(vars), [vars])

  return (
    <section className="border-t border-border-muted pt-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="industrial-label">{title}</p>
          {subtitle ? <p className="mt-1 text-xs text-text-secondary">{subtitle}</p> : null}
        </div>
        {dashboardUrl ? (
          <a href={dashboardUrl} target="_blank" rel="noreferrer" className="industrial-action">
            Open Dashboard
          </a>
        ) : null}
      </div>

      {!src ? (
        <div className="mt-4 border-t border-border-muted pt-4">
          <p className="text-sm text-text-secondary">
            Grafana panel is not configured for this environment.
          </p>
          <p className="mt-1 text-xs text-text-tertiary">{grafanaProvisionNote()}</p>
        </div>
      ) : (
        <div className={cn('mt-4 relative overflow-hidden border border-border-muted')}>
          <iframe
            title={title}
            src={src}
            className={cn(
              'w-full bg-app transition-opacity duration-150',
              loaded ? 'opacity-100' : 'opacity-0',
            )}
            style={{ height }}
            onLoad={() => setLoaded(true)}
          />
          {!loaded ? (
            <div className="absolute inset-0 animate-pulse bg-white/[0.04]" aria-hidden="true" />
          ) : null}
        </div>
      )}
    </section>
  )
}
