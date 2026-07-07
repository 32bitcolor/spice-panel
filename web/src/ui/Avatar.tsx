import * as React from 'react'
import { cn } from './lib/cn'

export interface AvatarProps {
  /** Display name — initials are derived from it when no image is given. */
  name: string
  src?: string
  size?: number
  className?: string
}

const initials = (name: string): string =>
  name
    .split(/[\s_]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')

export const Avatar: React.FC<AvatarProps> = ({
  name,
  src,
  size = 26,
  className,
}): React.ReactElement => {
  const style: React.CSSProperties = { width: size, height: size }
  const clip =
    '[clip-path:polygon(4px_0,100%_0,100%_calc(100%-4px),calc(100%-4px)_100%,0_100%,0_4px)]'

  if (src !== undefined) {
    return (
      <img
        src={src}
        alt={name}
        style={style}
        className={cn('shrink-0 object-cover', clip, className)}
      />
    )
  }

  return (
    <span
      aria-label={name}
      style={{ ...style, fontSize: Math.round(size * 0.42) }}
      className={cn(
        'grid shrink-0 place-items-center bg-[linear-gradient(135deg,var(--spice-hi),var(--ember))] font-mono font-bold text-[color:var(--void)]',
        clip,
        className,
      )}
    >
      {initials(name)}
    </span>
  )
}
