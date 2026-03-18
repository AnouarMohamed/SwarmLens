interface SkeletonProps {
  className?: string
}

function cn(...parts: Array<string | undefined | false>) {
  return parts.filter(Boolean).join(' ')
}

export function Skeleton({ className }: SkeletonProps) {
  return (
    <div
      aria-hidden="true"
      className={cn(
        'animate-shimmer rounded-card bg-[linear-gradient(110deg,rgba(154,170,189,0.07)_8%,rgba(154,170,189,0.14)_18%,rgba(154,170,189,0.07)_33%)] bg-[length:200%_100%]',
        className,
      )}
    />
  )
}
