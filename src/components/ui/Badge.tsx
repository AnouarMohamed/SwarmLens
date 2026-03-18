import type { ReactNode } from 'react'
import type { Severity } from '../../types'
import { CheckCircleIcon, DisconnectIcon, InfoIcon, WarningIcon } from './icons'

export type StatusBadgeVariant =
  | 'healthy'
  | 'warning'
  | 'critical'
  | 'neutral'
  | 'disconnected'
  | 'info'

interface StatusBadgeProps {
  variant?: StatusBadgeVariant
  status?: string
  label?: string
  size?: 'sm' | 'md'
  className?: string
}

interface SeverityBadgeProps {
  severity: Severity
  size?: 'sm' | 'md'
  className?: string
}

interface BadgePresentation {
  variant: StatusBadgeVariant
  label: string
}

function cn(...parts: Array<string | undefined | false>) {
  return parts.filter(Boolean).join(' ')
}

function normalizeStatus(status?: string): BadgePresentation {
  const normalized = status?.toLowerCase() ?? 'unknown'

  if (['healthy', 'running', 'ready', 'active', 'reachable', 'success'].includes(normalized)) {
    return { variant: 'healthy', label: status ?? 'Healthy' }
  }

  if (['critical', 'failed', 'down', 'error', 'unreachable', 'rejected'].includes(normalized)) {
    return { variant: 'critical', label: status ?? 'Critical' }
  }

  if (['warning', 'warn', 'drain', 'pause', 'paused', 'degraded'].includes(normalized)) {
    return { variant: 'warning', label: status ?? 'Warning' }
  }

  if (['disconnected', 'offline'].includes(normalized)) {
    return { variant: 'disconnected', label: status ?? 'Disconnected' }
  }

  if (['info', 'pending', 'connecting', 'investigating', 'mitigating'].includes(normalized)) {
    return { variant: 'info', label: status ?? 'Info' }
  }

  return { variant: 'neutral', label: status ?? 'Unknown' }
}

function variantTheme(variant: StatusBadgeVariant): { icon: ReactNode; classes: string } {
  switch (variant) {
    case 'healthy':
      return {
        icon: <CheckCircleIcon className="h-3.5 w-3.5" />,
        classes: 'border-state-success/40 bg-state-success/10 text-state-success',
      }
    case 'warning':
      return {
        icon: <WarningIcon className="h-3.5 w-3.5" />,
        classes: 'border-state-warning/40 bg-state-warning/10 text-state-warning',
      }
    case 'critical':
      return {
        icon: <WarningIcon className="h-3.5 w-3.5" />,
        classes: 'border-state-danger/40 bg-state-danger/10 text-state-danger',
      }
    case 'disconnected':
      return {
        icon: <DisconnectIcon className="h-3.5 w-3.5" />,
        classes: 'border-state-danger/35 bg-surface-2 text-state-danger',
      }
    case 'info':
      return {
        icon: <InfoIcon className="h-3.5 w-3.5" />,
        classes: 'border-state-info/40 bg-state-info/10 text-state-info',
      }
    case 'neutral':
    default:
      return {
        icon: <InfoIcon className="h-3.5 w-3.5" />,
        classes: 'border-border bg-surface-2 text-text-secondary',
      }
  }
}

export function StatusBadge({ variant, status, label, size = 'sm', className }: StatusBadgeProps) {
  const resolved = variant
    ? { variant, label: label ?? status ?? variant }
    : normalizeStatus(status ?? label)
  const { icon, classes } = variantTheme(resolved.variant)
  const sizeClasses = size === 'md' ? 'h-7 px-2.5 text-xs' : 'h-6 px-2 text-[11px]'

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-pill border font-medium uppercase tracking-[0.08em]',
        sizeClasses,
        classes,
        className,
      )}
    >
      <span aria-hidden="true">{icon}</span>
      <span>{resolved.label}</span>
    </span>
  )
}

export function SeverityBadge({ severity, size = 'sm', className }: SeverityBadgeProps) {
  const severityMap: Record<Severity, StatusBadgeVariant> = {
    critical: 'critical',
    high: 'warning',
    medium: 'warning',
    low: 'healthy',
    info: 'info',
  }

  return (
    <StatusBadge
      variant={severityMap[severity]}
      label={severity}
      size={size}
      className={className}
    />
  )
}
