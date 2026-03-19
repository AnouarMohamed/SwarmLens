import type { SVGProps } from 'react'

type IconProps = SVGProps<SVGSVGElement>

function BaseIcon({ children, className, ...props }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.8}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className={className}
      {...props}
    >
      {children}
    </svg>
  )
}

export function OverviewIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M4 4h6v6H4zM14 4h6v10h-6zM4 14h6v6H4zM14 18h6" />
    </BaseIcon>
  )
}

export function DiagnosticsIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <circle cx="12" cy="12" r="8" />
      <path d="M12 8v4l2 2M7 12h2M15 12h2" />
    </BaseIcon>
  )
}

export function IncidentIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M12 4l8 14H4z" />
      <path d="M12 10v4M12 18h.01" />
    </BaseIcon>
  )
}

export function AuditIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M5 5h14M5 10h14M5 15h10M5 20h10" />
    </BaseIcon>
  )
}

export function StackIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M12 4l8 4-8 4-8-4 8-4zM4 12l8 4 8-4M4 16l8 4 8-4" />
    </BaseIcon>
  )
}

export function ServiceIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <rect x="4" y="5" width="16" height="14" rx="2" />
      <path d="M8 9h8M8 13h8" />
    </BaseIcon>
  )
}

export function TaskIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M5 7h14M5 12h14M5 17h10" />
      <path d="m15 16 2 2 3-4" />
    </BaseIcon>
  )
}

export function NodeIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <rect x="4" y="5" width="16" height="6" rx="1.5" />
      <rect x="4" y="13" width="16" height="6" rx="1.5" />
    </BaseIcon>
  )
}

export function NetworkIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <circle cx="6" cy="12" r="2" />
      <circle cx="18" cy="7" r="2" />
      <circle cx="18" cy="17" r="2" />
      <path d="M8 12h6M16.5 8.5l-4 2.5M16.5 15.5l-4-2.5" />
    </BaseIcon>
  )
}

export function VolumeIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <ellipse cx="12" cy="7" rx="7" ry="3" />
      <path d="M5 7v7c0 1.7 3.1 3 7 3s7-1.3 7-3V7" />
    </BaseIcon>
  )
}

export function SecretIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M7 12a5 5 0 1 1 4.2 4.9L8 20H5v-3l2.3-2.3" />
      <circle cx="14" cy="9" r="1.2" />
    </BaseIcon>
  )
}

export function ShieldIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M12 3 5 6v5c0 4.5 2.7 8 7 10 4.3-2 7-5.5 7-10V6l-7-3z" />
      <path d="m9 12 2 2 4-4" />
    </BaseIcon>
  )
}

export function ServerIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <rect x="4" y="5" width="16" height="5" rx="1.2" />
      <rect x="4" y="14" width="16" height="5" rx="1.2" />
      <path d="M7 7.5h.01M7 16.5h.01" />
    </BaseIcon>
  )
}

export function ActivityIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M3 12h4l2-4 4 8 2-4h6" />
    </BaseIcon>
  )
}

export function AlertIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M12 4 3.5 19h17L12 4z" />
      <path d="M12 10v4M12 17h.01" />
    </BaseIcon>
  )
}

export function CheckCircleIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <circle cx="12" cy="12" r="9" />
      <path d="m8.5 12 2.3 2.4 4.7-4.8" />
    </BaseIcon>
  )
}

export function WarningIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M12 3.5 2.5 20h19L12 3.5z" />
      <path d="M12 9v4.5M12 17h.01" />
    </BaseIcon>
  )
}

export function InfoIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 10v6M12 7h.01" />
    </BaseIcon>
  )
}

export function DisconnectIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M3 3 21 21" />
      <path d="M8 12a4 4 0 0 1 4-4M16 12a4 4 0 0 1-1.2 2.8M5 16a9.5 9.5 0 0 1 14 0" />
    </BaseIcon>
  )
}

export function ClockIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </BaseIcon>
  )
}

export function RefreshIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M20 12a8 8 0 0 0-14-5.3M4 12a8 8 0 0 0 14 5.3" />
      <path d="M6 3v4h4M18 21v-4h-4" />
    </BaseIcon>
  )
}

export function MoreIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <circle cx="5" cy="12" r="1.5" />
      <circle cx="12" cy="12" r="1.5" />
      <circle cx="19" cy="12" r="1.5" />
    </BaseIcon>
  )
}

export function MenuIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M4 7h16M4 12h16M4 17h16" />
    </BaseIcon>
  )
}

export function ChevronDownIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="m6 9 6 6 6-6" />
    </BaseIcon>
  )
}

export function ArrowRightIcon(props: IconProps) {
  return (
    <BaseIcon {...props}>
      <path d="M5 12h14M13 6l6 6-6 6" />
    </BaseIcon>
  )
}
