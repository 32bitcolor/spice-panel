import * as React from 'react'

export interface BrandProps {
  className?: string
  /** Show the SPICE-PANEL wordmark next to the mark. */
  wordmark?: boolean
  /** Mark size in px. */
  markSize?: number
}

/**
 * spice-panel brand: a chamfered HUD mark (the Fremen "eyes of Ibad" spice
 * waves) in spice-cyan + the wordmark. DOM/SVG so it inherits the theme tokens
 * and stays crisp at any size — no baked image asset.
 */
export const Brand: React.FC<BrandProps> = ({
  className,
  wordmark = true,
  markSize = 30,
}): React.ReactElement => (
  <span className={joinClass('inline-flex items-center gap-2.5', className)}>
    <span
      style={{ width: markSize, height: markSize }}
      className="grid shrink-0 place-items-center bg-accent/12 text-accent shadow-[inset_0_0_0_1px_color-mix(in_srgb,var(--accent)_45%,transparent),0_0_14px_-4px_var(--accent)] [clip-path:polygon(6px_0,100%_0,100%_calc(100%-6px),calc(100%-6px)_100%,0_100%,0_6px)]"
    >
      <svg
        viewBox="0 0 24 24"
        width={Math.round(markSize * 0.62)}
        height={Math.round(markSize * 0.62)}
        fill="none"
        stroke="currentColor"
        strokeWidth="1.7"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M2 15 Q6 9 12 12 Q18 15 22 9" />
        <path d="M2 19 Q6 13 12 16 Q18 19 22 13" />
        <circle cx="12" cy="6" r="1.5" fill="currentColor" stroke="none" />
      </svg>
    </span>
    {renderWordmark(wordmark)}
  </span>
)

const renderWordmark = (show: boolean): React.ReactNode => {
  if (!show) return null
  return (
    <span className="font-display text-lg font-bold uppercase leading-none tracking-[0.28em] text-foreground">
      <span className="text-accent">SPICE</span>
      <span className="text-muted">·</span>
      PANEL
    </span>
  )
}

const joinClass = (base: string, extra: string | undefined): string =>
  extra === undefined ? base : `${base} ${extra}`
