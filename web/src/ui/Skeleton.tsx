import * as React from 'react'
import { cn } from './lib/cn'

export interface SkeletonProps {
  className?: string
  /** Convenience: render N stacked lines. */
  lines?: number
}

const bar = 'bg-[linear-gradient(90deg,var(--raised),var(--ridge),var(--raised))] bg-[length:200%_100%] motion-safe:animate-[hud-shimmer_1.4s_linear_infinite]'

export const Skeleton: React.FC<SkeletonProps> = ({ className, lines }): React.ReactElement => {
  if (lines === undefined) {
    return <div className={cn('h-3', bar, className)} />
  }
  return (
    <div className={cn('flex flex-col gap-2.5', className)}>
      {Array.from({ length: lines }, (_, i) => (
        <div
          key={i}
          className={cn('h-3', bar)}
          style={{ width: `${70 + ((i * 37) % 30)}%` }}
        />
      ))}
    </div>
  )
}
