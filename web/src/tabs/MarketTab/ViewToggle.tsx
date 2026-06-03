import { ToggleButtonGroup, ToggleButton } from '@heroui/react'
import { Icon } from '../../dune-ui'

export type MarketView = 'grid' | 'table'

type Props = {
  view: MarketView
  onChange: (v: MarketView) => void
}

export default function ViewToggle({ view, onChange }: Props) {
  return (
    <ToggleButtonGroup
      selectionMode="single"
      disallowEmptySelection
      selectedKeys={[view]}
      onSelectionChange={(keys) => {
        const next = [...keys][0]
        if (next === 'grid' || next === 'table') onChange(next)
      }}
      className="shrink-0"
    >
      <ToggleButton id="grid" isIconOnly aria-label="Grid view">
        <Icon name="layout-grid" />
      </ToggleButton>
      <ToggleButton id="table" isIconOnly aria-label="Table view">
        <Icon name="list" />
      </ToggleButton>
    </ToggleButtonGroup>
  )
}
