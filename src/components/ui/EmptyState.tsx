import type { ReactNode } from 'react'

type EmptyStateTone = 'neutral' | 'info' | 'success' | 'warning' | 'critical'

interface EmptyStateProps {
  title: string
  description: string
  icon?: ReactNode
  action?: ReactNode
  tone?: EmptyStateTone
  className?: string
}

function cn(...parts: Array<string | undefined | false>) {
  return parts.filter(Boolean).join(' ')
}

function toneClasses(tone: EmptyStateTone) {
  switch (tone) {
    case 'info':
      return 'border-state-info/35 bg-state-info/5'
    case 'success':
      return 'border-state-success/35 bg-state-success/5'
    case 'warning':
      return 'border-state-warning/35 bg-state-warning/5'
    case 'critical':
      return 'border-state-danger/35 bg-state-danger/5'
    case 'neutral':
    default:
      return 'border-border-muted bg-surface-2'
  }
}

export function EmptyState({
  title,
  description,
  icon,
  action,
  tone = 'neutral',
  className,
}: EmptyStateProps) {
  return (
    <div className={cn('rounded-card border px-4 py-6 text-center', toneClasses(tone), className)}>
      {icon && (
        <div className="mx-auto mb-3 flex h-9 w-9 items-center justify-center rounded-full border border-border bg-surface-3 text-text-secondary">
          {icon}
        </div>
      )}
      <h3 className="text-sm font-medium text-text-primary">{title}</h3>
      <p className="mx-auto mt-2 max-w-[48ch] text-sm text-text-secondary">{description}</p>
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}
