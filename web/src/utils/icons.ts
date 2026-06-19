// VITE_ICON_BASE_URL: base URL for item icons served from R2 (e.g. https://icons.example.com).
// When set, icons load from <base>/<template_id>.webp.
// When unset, no icon URL is produced and components fall back to category placeholders.
const ICON_BASE = ((import.meta.env.VITE_CDN_BASE_URL as string) ?? 'https://assets.dune.layout.tools')?.replace(
  /\/$/,
  '',
)

export function iconUrl(templateId: string, variant: 'detail' | 'thumb' = 'detail'): string | null {
  if (!ICON_BASE) return null
  return `${ICON_BASE}/${variant}/${templateId}.webp`
}

// Backgrounds sampled from in-game screenshots, darkened ~25% (HDR boosts brightness).
// All backgrounds use linear gradients — the game renders them universally.
// More-specific category prefixes MUST appear before their parent.

const G = (a: string, b: string, dir = 'to bottom') => `linear-gradient(${dir}, ${a}, ${b})`

// Named unique set wearables: purple → black (sampled Image #18, user confirmed "starts black at bottom")
export const BG_NAMED_SET = G('#3a1e58 0%, #12083a 55%', '#060308 100%')

// Mementos: dark green → black
const BG_MEMENTO = G('#122814', '#05080302')

// Regular / common fall-through: near-black → black (sampled Image #22)
const BG_DEFAULT = G('#0e0e0a', '#050504')

const CATEGORY_COLORS: Record<string, string> = {
  // Stillsuits — rust/red → black (sampled Image #19: CHOAM Stillsuit Gloves)
  'items/garment/stillsuits': G('#501a0a', '#120504'),

  // Social + utility wearables — near-black → black (sampled Image #22: Caladan Casual Tunic)
  'items/garment/socialwearables': G('#0e0e0a', '#050504'),
  'items/garment/utilitywearables': G('#0e0e0a', '#050504'),

  // Heavy + light armor — dark green → black (sampled Image #21: CHOAM Heavy Chestplate)
  'items/garment/heavyarmor': G('#122810', '#050804'),
  'items/garment/lightarmor': G('#122810', '#050804'),

  // Garment fallback → dark green → black
  'items/garment': G('#122810', '#050804'),

  // Augments — dark blue → black (tech items, same family as components)
  'items/augment': G('#0e1828', '#050608'),

  // Components — dark blue → black (sampled Image #24: Advanced Servoks)
  'items/misc/components': G('#0e1828', '#050608'),

  // Misc fallback — near-black → black
  'items/misc': G('#0e0e0a', '#050504'),

  // Weapons — near-black → black (named sets handled separately via templateId)
  'items/weapons': G('#0e0e0c', '#050508'),

  // Utility — near-black → black
  'items/utility': G('#0e0e0a', '#050504'),

  // Vehicles — near-black → black
  'items/vehicles': G('#0e0e0c', '#050508'),
}

// Export for MarketGrid (named sets use the gradient, schematics add a grid on top)
export const BG_PURPLE = BG_NAMED_SET

export function categoryColor(category: string, rarity?: string, templateId?: string): string {
  // Memento overrides everything
  if (rarity === 'memento') return BG_MEMENTO

  // Named unique set wearables — template ID contains "Unique", rarity is rare
  // (404 items: Pincushion, Idaho's Charge, Syndicate set, Aren's set, etc.)
  if (rarity === 'rare' && templateId && /Unique/.test(templateId)) return BG_NAMED_SET

  // Category determines the specific background
  for (const [prefix, color] of Object.entries(CATEGORY_COLORS)) {
    if (category.startsWith(prefix)) return color
  }
  return BG_DEFAULT
}

const QUALITY_LABELS = ['Standard', 'Refined', 'Superior', 'Masterwork', 'Pristine', 'Flawless']

export function qualityLabel(q: number): string {
  return QUALITY_LABELS[q] ?? `Q${q}`
}
