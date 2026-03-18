import type { ReactNode } from 'react'
import { StatusBadge } from './Badge'

export interface AttentionItem {
  id: string
  serviceName: string
  replicas: string
  state: 'healthy' | 'warning' | 'critical' | 'neutral'
  restartsOrErrors: string
  action?: ReactNode
}

interface AttentionListProps {
  items: AttentionItem[]
}

export function AttentionList({ items }: AttentionListProps) {
  return (
    <div className="overflow-hidden rounded-card border border-border-muted">
      <table className="w-full min-w-[560px] border-collapse text-sm">
        <thead className="bg-surface-2">
          <tr className="text-left text-xs uppercase tracking-[0.08em] text-text-tertiary">
            <th className="px-3 py-2 font-medium">Service</th>
            <th className="px-3 py-2 font-medium">Replicas</th>
            <th className="px-3 py-2 font-medium">State</th>
            <th className="px-3 py-2 font-medium">Restarts / errors</th>
            <th className="px-3 py-2 font-medium">Action</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id} className="border-t border-border-muted bg-surface-1">
              <td className="px-3 py-2.5 font-mono text-xs text-text-secondary">
                {item.serviceName}
              </td>
              <td className="px-3 py-2.5 text-text-primary">{item.replicas}</td>
              <td className="px-3 py-2.5">
                <StatusBadge variant={item.state} label={item.state} />
              </td>
              <td className="px-3 py-2.5 text-text-secondary">{item.restartsOrErrors}</td>
              <td className="px-3 py-2.5 text-right">{item.action}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
