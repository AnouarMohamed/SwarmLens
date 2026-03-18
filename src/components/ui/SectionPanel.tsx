import type { ReactNode } from 'react'

interface SectionPanelProps {
  title: string
  subtitle?: string
  action?: ReactNode
  children: ReactNode
  className?: string
}

function cn(...parts: Array<string | undefined | false>) {
  return parts.filter(Boolean).join(' ')
}

export function SectionPanel({ title, subtitle, action, children, className }: SectionPanelProps) {
  return (
    <section
      className={cn('rounded-panel border border-border-muted bg-surface-1 px-5 py-4', className)}
      aria-label={title}
    >
      <header className="mb-4 flex items-start justify-between gap-3">
        <div className="space-y-1">
          <h2 className="text-sm font-semibold tracking-wide text-text-primary">{title}</h2>
          {subtitle && <p className="text-xs text-text-secondary">{subtitle}</p>}
        </div>
        {action && <div className="shrink-0">{action}</div>}
      </header>
      {children}
    </section>
  )
}
