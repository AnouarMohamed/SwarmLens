import type { ReactNode } from 'react'
import { StatusBadge, type StatusBadgeVariant } from './Badge'
import { MoreIcon } from './icons'

export interface PageHeaderBadge {
  id: string
  label: string
  variant: StatusBadgeVariant
}

export interface PageHeaderAction {
  id: string
  label: string
  onClick: () => void
  variant?: 'primary' | 'secondary' | 'ghost'
  disabled?: boolean
  icon?: ReactNode
  ariaLabel?: string
}

export interface PageHeaderMetaItem {
  id: string
  label: string
  value: string
}

interface PageHeaderProps {
  title: string
  subtitle: string
  badges?: PageHeaderBadge[]
  actions?: PageHeaderAction[]
  meta?: PageHeaderMetaItem[]
  className?: string
}

function cn(...parts: Array<string | undefined | false>) {
  return parts.filter(Boolean).join(' ')
}

function actionClasses(variant: NonNullable<PageHeaderAction['variant']>) {
  switch (variant) {
    case 'primary':
      return 'border-state-info/50 bg-state-info/20 text-state-info hover:bg-state-info/30'
    case 'ghost':
      return 'border-border-muted bg-transparent text-text-secondary hover:border-border hover:bg-surface-3 hover:text-text-primary'
    case 'secondary':
    default:
      return 'border-border-muted bg-surface-2 text-text-primary hover:border-border hover:bg-surface-3'
  }
}

export function PageHeader({
  title,
  subtitle,
  badges = [],
  actions = [],
  meta = [],
  className,
}: PageHeaderProps) {
  return (
    <header
      className={cn(
        'flex flex-col gap-4 rounded-panel border border-border-muted bg-surface-1 px-5 py-4 lg:flex-row lg:items-start lg:justify-between',
        className,
      )}
    >
      <div className="min-w-0 space-y-2">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-text-primary">{title}</h1>
          <p className="mt-1 text-sm text-text-secondary">{subtitle}</p>
        </div>

        {(badges.length > 0 || meta.length > 0) && (
          <div className="flex flex-wrap items-center gap-2">
            {badges.map((badge) => (
              <StatusBadge key={badge.id} variant={badge.variant} label={badge.label} size="md" />
            ))}
            {meta.map((item) => (
              <div
                key={item.id}
                className="inline-flex items-center gap-1 rounded-pill border border-border-muted bg-surface-2 px-3 py-1 text-xs text-text-secondary"
              >
                <span className="uppercase tracking-[0.08em] text-text-tertiary">{item.label}</span>
                <span className="font-mono text-text-primary">{item.value}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      {actions.length > 0 && (
        <div className="flex flex-wrap items-center gap-2 lg:justify-end">
          {actions.map((action) => (
            <button
              key={action.id}
              type="button"
              onClick={action.onClick}
              disabled={action.disabled}
              aria-label={action.ariaLabel ?? action.label}
              className={cn(
                'inline-flex h-10 items-center gap-2 rounded-card border px-4 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-45',
                actionClasses(action.variant ?? 'secondary'),
              )}
            >
              {action.icon}
              <span>{action.label}</span>
            </button>
          ))}
          <button
            type="button"
            aria-label="More actions"
            className="inline-flex h-10 w-10 items-center justify-center rounded-card border border-border-muted bg-surface-2 text-text-secondary transition hover:border-border hover:bg-surface-3 hover:text-text-primary"
          >
            <MoreIcon className="h-4 w-4" />
          </button>
        </div>
      )}
    </header>
  )
}
