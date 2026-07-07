import * as React from 'react'
import { cn } from './lib/cn'

export interface SpinnerProps {
  size?: number
  className?: string
  label?: string
}

export const Spinner: React.FC<SpinnerProps> = ({
  size = 20,
  className,
  label = 'Loading',
}): React.ReactElement => (
  <svg
    role="progressbar"
    aria-label={label}
    width={size}
    height={size}
    viewBox="0 0 24 24"
    className={cn('animate-spin text-accent motion-reduce:animate-none', className)}
    fill="none"
  >
    <circle cx="12" cy="12" r="9" stroke="currentColor" strokeOpacity="0.2" strokeWidth="2.5" />
    <path
      d="M21 12a9 9 0 0 0-9-9"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
    />
  </svg>
)
