import * as React from 'react'
import { cn } from './lib/cn'

export interface AvatarProps {
  /** Display name — initials are derived from it when no image/fallback is given. */
  name?: string
  src?: string
  alt?: string
  /** Rendered when there is no image (e.g. an icon). Overrides initials. */
  fallback?: React.ReactNode
  size?: number | 'sm' | 'md' | 'lg'
  className?: string
}

const SIZES = { sm: 24, md: 32, lg: 40 }
const resolveSize = (size: AvatarProps['size']): number =>
  typeof size === 'number' ? size : SIZES[size ?? 'md']

const initials = (name: string): string =>
  name
    .split(/[\s_]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')

const CLIP = '[clip-path:polygon(4px_0,100%_0,100%_calc(100%-4px),calc(100%-4px)_100%,0_100%,0_4px)]'

export interface AvatarRootProps extends AvatarProps {
  children?: React.ReactNode
}

const AvatarRoot: React.FC<AvatarRootProps> = ({
  name,
  src,
  alt,
  fallback,
  size,
  className,
  children,
}): React.ReactElement => {
  const px = resolveSize(size)
  const style: React.CSSProperties = { width: px, height: px }

  // Compound mode: <Avatar><Avatar.Image/><Avatar.Fallback/></Avatar>
  if (children !== undefined) {
    return (
      <span
        aria-label={name ?? alt}
        style={{ ...style, fontSize: Math.round(px * 0.42) }}
        className={cn(
          'relative grid shrink-0 place-items-center overflow-hidden bg-[linear-gradient(135deg,var(--spice-hi),var(--ember))] font-mono font-bold text-[color:var(--void)]',
          CLIP,
          className,
        )}
      >
        {children}
      </span>
    )
  }

  if (src !== undefined) {
    return (
      <img
        src={src}
        alt={alt ?? name ?? ''}
        style={style}
        className={cn('shrink-0 object-cover', CLIP, className)}
      />
    )
  }

  return (
    <span
      aria-label={name ?? alt}
      style={{ ...style, fontSize: Math.round(px * 0.42) }}
      className={cn(
        'grid shrink-0 place-items-center bg-[linear-gradient(135deg,var(--spice-hi),var(--ember))] font-mono font-bold text-[color:var(--void)]',
        CLIP,
        className,
      )}
    >
      {fallback ?? (name === undefined ? null : initials(name))}
    </span>
  )
}

/* ── Compound slots (HeroUI-compatible) ───────────────────────────────────── */

const Image: React.FC<{ src?: string, alt?: string, className?: string }> = ({
  src,
  alt,
  className,
}): React.ReactElement | null => {
  if (src === undefined) return null
  return (
    <img src={src} alt={alt ?? ''} className={cn('absolute inset-0 h-full w-full object-cover', className)} />
  )
}

const Fallback: React.FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <span {...props} className={cn('grid h-full w-full place-items-center', className)}>
    {children}
  </span>
)

export const Avatar = Object.assign(AvatarRoot, { Image, Fallback })
