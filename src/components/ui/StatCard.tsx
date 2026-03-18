import type { ReactNode } from 'react'
import type { StatusBadgeVariant } from './Badge'

interface StatCardProps {
  label: string
  value: string | number
  sublabel: string
  trend?: string
  icon: ReactNode
  tone?: StatusBadgeVariant
  href?: string
  onClick?: () => void
}

function cn(...parts: Array<string | undefined | false>) {
  return parts.filter(Boolean).join(' ')
}

function toneStripeClasses(tone: StatusBadgeVariant) {
  switch (tone) {
    case 'healthy':
      return 'bg-state-success/70'
    case 'warning':
      return 'bg-state-warning/70'
    case 'critical':
    case 'disconnected':
      return 'bg-state-danger/70'
    case 'info':
      return 'bg-state-info/70'
    case 'neutral':
    default:
      return 'bg-border'
  }
}

function iconToneClasses(tone: StatusBadgeVariant) {
  switch (tone) {
    case 'healthy':
      return 'border-state-success/35 bg-state-success/10 text-state-success'
    case 'warning':
      return 'border-state-warning/35 bg-state-warning/10 text-state-warning'
    case 'critical':
    case 'disconnected':
      return 'border-state-danger/35 bg-state-danger/10 text-state-danger'
    case 'info':
      return 'border-state-info/35 bg-state-info/10 text-state-info'
    case 'neutral':
    default:
      return 'border-border bg-surface-3 text-text-secondary'
  }
}

function StatCardInner({
  label,
  value,
  sublabel,
  trend,
  icon,
  tone = 'neutral',
}: Omit<StatCardProps, 'href' | 'onClick'>) {
  return (
    <div className="relative h-full overflow-hidden rounded-card border border-border-card bg-surface-1 px-4 py-3">
      <span
        aria-hidden="true"
        className={cn('absolute inset-x-0 top-0 h-px', toneStripeClasses(tone))}
      />
      <div className="mb-3 flex items-start justify-between gap-3">
        <p className="text-xs uppercase tracking-[0.08em] text-text-tertiary">{label}</p>
        <span
          className={cn(
            'inline-flex h-8 w-8 items-center justify-center rounded-full border',
            iconToneClasses(tone),
          )}
        >
          {icon}
        </span>
      </div>
      <p className="text-2xl font-semibold leading-none tracking-tight text-text-primary">
        {value}
      </p>
      <p className="mt-2 text-sm text-text-secondary">{sublabel}</p>
      <div className="mt-4 border-t border-border-card pt-2 text-xs text-text-tertiary">
        {trend ?? 'No trend available'}
      </div>
    </div>
  )
}

export function StatCard({ href, onClick, ...props }: StatCardProps) {
  if (href) {
    return (
      <a
        href={href}
        className="block h-full rounded-card focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring/90 focus-visible:ring-offset-2 focus-visible:ring-offset-app"
      >
        <StatCardInner {...props} />
      </a>
    )
  }

  if (onClick) {
    return (
      <button
        type="button"
        onClick={onClick}
        className="block h-full w-full rounded-card text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring/90 focus-visible:ring-offset-2 focus-visible:ring-offset-app"
      >
        <StatCardInner {...props} />
      </button>
    )
  }

  return <StatCardInner {...props} />
}
