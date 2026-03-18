import type { ReactNode } from 'react'

interface Props {
  title: string
  description?: string
  children?: ReactNode
  onConfirm: () => void
  onCancel: () => void
  confirmLabel?: string
  confirmDanger?: boolean
  loading?: boolean
}

export function ActionDialog({
  title, description, children,
  onConfirm, onCancel,
  confirmLabel = 'Confirm',
  confirmDanger = false,
  loading = false,
}: Props) {
  return (
    <div className="dialog-backdrop" onClick={onCancel}>
      <div className="dialog" onClick={e => e.stopPropagation()}>
        <div className="dialog-header">
          <h2 className="dialog-title">{title}</h2>
        </div>
        {description && <p className="dialog-description">{description}</p>}
        {children && <div className="dialog-body">{children}</div>}
        <div className="dialog-footer">
          <button className="btn-ghost" onClick={onCancel} disabled={loading}>
            Cancel
          </button>
          <button
            className={confirmDanger ? 'btn-danger' : 'btn-primary'}
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? 'Working…' : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
