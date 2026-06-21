import * as React from 'react'
import { Icon } from '../../../dune-ui'
import { ArmorPieceIcon } from './ArmorPieceIcon'
import { ArmorSetIcon } from './ArmorSetIcon'
import { BladeIcon } from './BladeIcon'
import { CompactorIcon } from './CompactorIcon'
import { DewReaperIcon } from './DewReaperIcon'
import { ExtractorIcon } from './ExtractorIcon'
import { FlameIcon } from './FlameIcon'
import { FragmentIcon } from './FragmentIcon'
import { HelmetIcon } from './HelmetIcon'
import { PowerPackIcon } from './PowerPackIcon'
import { RangedIcon } from './RangedIcon'
import { ScannerIcon } from './ScannerIcon'
import { SchematicIcon } from './SchematicIcon'
import { StillsuitIcon } from './StillsuitIcon'
import { SuspensorIcon } from './SuspensorIcon'
import type { RawItem } from './interfaces'
import type { RewardIconProps } from './types'

// ── Classification logic ──────────────────────────────────────────────────────

const CATEGORY_ICONS_FALLBACK: Record<string, string> = {
  level: 'chevrons-up',
  story: 'book-open',
  side_quest: 'map',
  faction: 'landmark',
  exploration: 'compass',
  achievement: 'trophy',
}

const classifyByTemplate = (tpl: string, count: number, className?: string): React.ReactElement => {
  if (count >= 3) return <ArmorSetIcon className={className} />

  const t = tpl.toLowerCase()
  if (t.includes('sword') || t.includes('kindjal') || t.includes('rapier')
    || t.includes('dirk') || t.includes('cutteray')) return <BladeIcon className={className} />
  if (t.includes('flamethrower')) return <FlameIcon className={className} />
  if (t.includes('pistol') || t.includes('smg') || t.includes('longrifle')
    || t.includes('lmg') || t.includes('shotgun') || t.includes('ar_burst')
    || t.includes('uniquear')) return <RangedIcon className={className} />
  if (t.includes('stillsuit')) return <StillsuitIcon className={className} />
  if (t.includes('powerpack')) return <PowerPackIcon className={className} />
  if (t.includes('sandbike') || t.includes('scanner')) return <ScannerIcon className={className} />
  if (t.includes('dewreap')) return <DewReaperIcon className={className} />
  if (t.includes('extractor') || t.includes('bloodsack')) return <ExtractorIcon className={className} />
  if (t.includes('compactor')) return <CompactorIcon className={className} />
  if (t.includes('fragment')) return <FragmentIcon className={className} />
  if (t.includes('suspensor')) return <SuspensorIcon className={className} />
  if (t.includes('helmet') || t.includes('head') || t.includes('mask')
    || t.includes('wrap')) return <HelmetIcon className={className} />
  if (t.includes('top') || t.includes('chest') || t.includes('jacket')
    || t.includes('garb') || t.includes('gloves') || t.includes('gauntlet')
    || t.includes('boots') || t.includes('softstep') || t.includes('feet')
    || t.includes('bottom') || t.includes('legs') || t.includes('pants')
    || t.includes('legging')) return <ArmorPieceIcon className={className} />
  return <SchematicIcon className={className} />
}

/**
 * Renders a hand-drawn SVG icon that reflects the tier's actual reward:
 * weapon shape, stillsuit, scanner, armor set, etc. Falls back to the
 * category icon (Lucide) for tiers with no item rewards (achievements).
 */
export const RewardIcon: React.FC<RewardIconProps> = ({ tier, className }): React.ReactElement => {
  if (!tier.reward_items) {
    return <Icon name={CATEGORY_ICONS_FALLBACK[tier.category] ?? 'circle'} className={className} />
  }

  let items: RawItem[] = []
  try {
    items = JSON.parse(tier.reward_items) as RawItem[]
  }
  catch { /* fall through */ }

  if (items.length === 0) {
    return <Icon name={CATEGORY_ICONS_FALLBACK[tier.category] ?? 'circle'} className={className} />
  }

  return classifyByTemplate(items[0]?.Template ?? '', items.length, className)
}
