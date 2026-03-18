import type { ReactNode } from 'react'
import { ArrowRightIcon } from './icons'

export interface QuickActionItem {
  id: string
  label: string
  description: string
  onClick: () => void
  disabled?: boolean
  icon?: ReactNode
}

interface QuickActionCardProps {
  actions: QuickActionItem[]
}

export function QuickActionCard({ actions }: QuickActionCardProps) {
  return (
    <ul className="space-y-2">
      {actions.map((action) => (
        <li key={action.id}>
          <button
            type="button"
            onClick={action.onClick}
            disabled={action.disabled}
            className="group flex w-full items-center justify-between rounded-card border border-border-muted bg-surface-2 px-3 py-2 text-left transition hover:border-border hover:bg-surface-3 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <div className="flex min-w-0 items-start gap-2.5">
              {action.icon ? (
                <span className="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-border bg-surface-3 text-text-secondary">
                  {action.icon}
                </span>
              ) : null}
              <div className="min-w-0">
                <p className="truncate text-sm font-medium text-text-primary">{action.label}</p>
                <p className="mt-0.5 text-xs text-text-secondary">{action.description}</p>
              </div>
            </div>
            <ArrowRightIcon className="h-4 w-4 shrink-0 text-text-tertiary transition group-hover:translate-x-0.5 group-hover:text-text-secondary" />
          </button>
        </li>
      ))}
    </ul>
  )
}
