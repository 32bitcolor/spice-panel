import * as React from 'react'
import {
  Button, Chip, Modal,
  SearchField, Separator, Spinner, TextField, toast,
} from '../../../ui'
import type { Selection } from 'react-aria-components'
import { useTranslation } from 'react-i18next'
import { useAtom, useAtomValue } from 'jotai'
import { api } from '../../../api/client'
import { ActionBar, DataTable, Icon, LoadingState, NumberInput } from '../../../dune-ui'
import type { Column } from '../../../dune-ui'
import { CategorizedPackPicker } from '../../../components/CategorizedPackPicker'
import { ItemDetailDrawer } from '../../../components/ItemDetailDrawer'
import { ItemOptionRow } from '../../../components/ItemOptionRow'
import { StagedItemCell } from '../../../components/StagedItemCell'
import { packsSyncAtom, itemDataSyncAtom } from '../../../data/store'
import type { GiveItemsModalProps } from './interfaces'
import type { GiveResult, StagedItem } from './types'

export const GiveItemsModal: React.FC<GiveItemsModalProps> = ({ player, open, onClose }) => {
  const { t } = useTranslation()
  const [templates, setTemplates] = React.useState<{ id: string, name: string }[]>([])
  const [loading, setLoading] = React.useState(false)
  const [query, setQuery] = React.useState('')
  const [selected, setSelected] = React.useState('')
  const [qty, setQty] = React.useState(1)
  const [quality, setQuality] = React.useState(0)
  const [staged, setStaged] = React.useState<StagedItem[]>([])
  const [submitting, setSubmitting] = React.useState(false)
  const [result, setResult] = React.useState<GiveResult>(null)
  const [selectedKeys, setSelectedKeys] = React.useState<Selection>(new Set())
  const [detailId, setDetailId] = React.useState<string | null>(null)
  const [packsData] = useAtom(packsSyncAtom)
  const itemData = useAtomValue(itemDataSyncAtom)

  const keyCounter = React.useRef(0)
  const nextKey = () => String(keyCounter.current++)

  React.useEffect(() => {
    if (!open) return
    void Promise.resolve().then(() => {
      setLoading(true)
      setQuery('')
      setSelected('')
      setQty(1)
      setQuality(0)
      setStaged([])
      setResult(null)
      setSelectedKeys(new Set())
    })
      .then(() => api.players.templates())
      .then((tmpls) => { setTemplates(tmpls) })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [open])

  const nameMap = new Map(templates.map((tpl) => [tpl.id, tpl.name]))

  const _giq = query.toLowerCase()
  const filtered = !query
    ? []
    : templates
        .filter((tpl) => tpl.id.toLowerCase().includes(_giq) || tpl.name.toLowerCase().includes(_giq))
        .slice(0, 100)

  const packOptions = Object.entries(packsData.packs).map(([id, pack]) => ({
    id, name: pack.name, category: pack.category, tier: pack.tier,
  }))

  const pick = (tpl: { id: string, name: string }) => {
    setSelected(tpl.id)
    setQuery(tpl.name ? `${tpl.id}  —  ${tpl.name}` : tpl.id)
  }

  const addToStaged = () => {
    if (!selected) {
      toast.warning(t('players.give.selectTemplate'))
      return
    }
    setStaged((prev) => [...prev, { template: selected, qty, quality, _key: nextKey() }])
    setQuery('')
    setSelected('')
    setQty(1)
    setQuality(0)
  }

  const removeFromStaged = (key: string) => {
    setStaged((prev) => prev.filter((it) => it._key !== key))
    setSelectedKeys((prev) => {
      if (prev === 'all') return new Set(staged.filter((it) => it._key !== key).map((it) => it._key))
      const next = new Set(prev as Set<string>)
      next.delete(key)
      return next
    })
  }

  const updateStaged = (key: string, field: 'qty' | 'quality', value: number) => {
    setStaged((prev) => prev.map((item) => item._key === key ? { ...item, [field]: value } : item))
  }

  const selectionCount = selectedKeys === 'all' ? staged.length : (selectedKeys as Set<string>).size

  const handleBulkDelete = () => {
    if (selectedKeys === 'all') {
      setStaged([])
    }
    else {
      const keys = selectedKeys as Set<string>
      setStaged((prev) => prev.filter((it) => !keys.has(it._key)))
    }
    setSelectedKeys(new Set())
  }

  const handleSubmit = async () => {
    if (staged.length === 0) return
    setSubmitting(true)
    try {
      const items = staged.map(({ template, qty, quality }) => ({ template, qty, quality }))
      const res = await api.players.giveItems(player.id, items)
      setResult(res)
      setStaged([])
      setSelectedKeys(new Set())
      if (res.skipped.length === 0) {
        toast.success(t('players.give.gaveItems', { count: res.given.length, player: player.name }))
        onClose()
      }
    }
    catch (e: unknown) {
      toast.danger(e instanceof Error ? e.message : String(e))
    }
    finally {
      setSubmitting(false)
    }
  }

  const columns: Column<'template' | 'qty' | 'quality' | 'actions'>[] = [
    {
      key: 'template',
      isRowHeader: true,
      label: t('players.inventory.columns.template'),
      width: 200,
    },
    {
      key: 'qty',
      label: t('players.give.qty'),
      width: 130,
    },
    {
      key: 'quality',
      label: t('players.give.quality'),
      width: 130,
    },
    {
      key: 'actions',
      label: '',
      width: 88,
    },
  ]

  const renderCell = (item: StagedItem, key: 'template' | 'qty' | 'quality' | 'actions'): React.ReactNode => {
    switch (key) {
      case 'template':
        return (
          <StagedItemCell
            templateId={item.template}
            name={nameMap.get(item.template) || ''}
            entry={itemData.items[item.template] ?? null}
          />
        )
      case 'qty':
        return (
          <NumberInput
            ariaLabel={`${t('players.give.qty')} for ${item.template}`}
            min={1}
            value={item.qty}
            onChange={(v) => updateStaged(item._key, 'qty', v)}
            className="w-full"
          />
        )
      case 'quality':
        return (
          <NumberInput
            ariaLabel={`${t('players.give.quality')} for ${item.template}`}
            min={0}
            value={item.quality}
            onChange={(v) => updateStaged(item._key, 'quality', v)}
            className="w-full"
          />
        )
      case 'actions':
        return (
          <div className="flex items-center gap-1">
            <Button
              size="sm"
              variant="ghost"
              isIconOnly
              onPress={() => setDetailId(item.template)}
              aria-label={t('common.info')}
            >
              <Icon name="info" />
            </Button>
            <Button
              size="sm"
              variant="danger"
              isIconOnly
              onPress={() => removeFromStaged(item._key)}
              aria-label={t('common.remove')}
            >
              <Icon name="trash" />
            </Button>
          </div>
        )
    }
  }

  return (
    <React.Fragment>
      <Modal.Backdrop variant="blur" className="bg-linear-to-t from-(--background)/85 via-(--background)/40 to-transparent" isOpen={open} onOpenChange={(v) => !v && onClose()}>
        <Modal.Container size="cover" scroll="outside">
          <Modal.Dialog className="p-10">
            <Modal.CloseTrigger />
            <Modal.Header>
              <Modal.Heading className="text-accent">
                {t('players.give.modalTitle', { name: player.name })}
              </Modal.Heading>
            </Modal.Header>
            <Modal.Body className="flex flex-col gap-3 h-[80vh] min-h-0">
              {loading
                ? (
                    <LoadingState size="sm" />
                  )
                : (
                    <React.Fragment>
                      {/* Load Pack */}
                      <CategorizedPackPicker
                        packs={packOptions}
                        onSelectPack={(id) => {
                          const pack = packsData.packs[id]
                          if (pack) {
                            setStaged((prev) => [
                              ...prev,
                              ...pack.items.map((item) => ({ ...item, _key: nextKey() })),
                            ])
                          }
                        }}
                        className="w-full shrink-0"
                      />

                      {/* Template + Qty + Quality + Add */}
                      <div className="flex items-end gap-3 shrink-0">
                        <TextField className="flex-1 min-w-0" aria-label={t('players.inventory.columns.template')}>
                          <div className="relative w-full">
                            <SearchField
                              className="w-full"
                              value={query}
                              onChange={(v) => {
                                setQuery(v)
                                setSelected('')
                              }}
                            >
                              <SearchField.Group>
                                <SearchField.SearchIcon />
                                <SearchField.Input placeholder={t('players.give.searchTemplates')} />
                                <SearchField.ClearButton />
                              </SearchField.Group>
                            </SearchField>
                            {filtered.length > 0 && (
                              <div className="absolute z-50 w-full mt-1 rounded-[var(--radius)] border border-border bg-surface overflow-y-auto max-h-52">
                                {filtered.map((tpl) => (
                                  <ItemOptionRow
                                    key={tpl.id}
                                    id={tpl.id}
                                    name={tpl.name}
                                    entry={itemData.items[tpl.id] ?? null}
                                    onPick={() => pick(tpl)}
                                  />
                                ))}
                              </div>
                            )}
                          </div>
                        </TextField>
                        <NumberInput
                          prefix={t('players.give.qty')}
                          ariaLabel={t('players.give.qty')}
                          min={1}
                          value={qty}
                          onChange={setQty}
                          className="w-56 shrink-0"
                        />
                        <NumberInput
                          prefix={t('players.give.quality')}
                          ariaLabel={t('players.give.quality')}
                          min={0}
                          value={quality}
                          onChange={setQuality}
                          className="w-56 shrink-0"
                        />
                        <Button size="sm" onPress={addToStaged} isDisabled={!selected} className="shrink-0">
                          <Icon name="plus" />
                          {' '}
                          {t('players.give.add')}
                        </Button>
                      </div>

                      {/* Quality>0 is a live-state limitation of the game: the item
                        is written to the DB but only materializes after the player
                        relogs, and may land outside their free inventory slots (#207). */}
                      {(quality > 0 || staged.some((s) => s.quality > 0)) && (
                        <div className="shrink-0 flex items-start gap-2 rounded-[var(--radius)] px-3 py-2 bg-surface border border-warning/40 text-xs text-muted">
                          <Icon name="triangle-alert" className="text-warning shrink-0 mt-0.5" />
                          <span>{t('players.give.qualityWarning')}</span>
                        </div>
                      )}

                      {/* Staged items DataGrid */}
                      {staged.length > 0 && (
                        <DataTable
                          aria-label={t('players.give.modalTitle', { name: player.name })}
                          columns={columns}
                          rows={staged}
                          rowId={(item) => item._key}
                          renderCell={renderCell}
                          selectedKeys={selectedKeys}
                          selectionMode="multiple"
                          onSelectionChange={setSelectedKeys}
                          className="flex-1 min-h-0"
                        />
                      )}

                      {result && (
                        <div className="text-xs shrink-0 rounded-[var(--radius)] px-3 py-2 bg-surface border border-border">
                          {result.given.length > 0 && (
                            <div className="text-success">
                              {t('players.give.gave')}
                              {result.given.join(', ')}
                            </div>
                          )}
                          {result.skipped.map((s, i) => (
                            <div key={i} className="text-danger">
                              {t('players.give.skipped', { template: s.template, reason: s.reason })}
                            </div>
                          ))}
                        </div>
                      )}
                    </React.Fragment>
                  )}
            </Modal.Body>
            <Modal.Footer>
              <Button variant="ghost" size="sm" slot="close">{t('common.cancel')}</Button>
              <Button size="sm" onPress={handleSubmit} isDisabled={submitting || staged.length === 0}>
                {submitting ? <Spinner size={16} /> : <Icon name="gift" />}
                {' '}
                {t('players.give.giveCount', { count: staged.length })}
              </Button>
            </Modal.Footer>
            {/* Inside the dialog: outside it, React Aria's modal underlay
              makes the bar inert (no hover, clicks fall through). */}
            <ActionBar aria-label={t('players.give.modalTitle', { name: player.name })} isOpen={selectionCount > 0}>
              <ActionBar.Prefix>
                <Chip size="sm" className="shrink-0 tabular-nums">{selectionCount}</Chip>
              </ActionBar.Prefix>
              <Separator />
              <ActionBar.Content>
                <Button
                  size="sm"
                  variant="ghost"
                  className="text-danger"
                  onPress={handleBulkDelete}
                  aria-label={t('common.deleteSelected')}
                >
                  <Icon name="trash-2" />
                  <span className="action-bar__label">{t('common.deleteSelected')}</span>
                </Button>
              </ActionBar.Content>
              <Separator />
              <ActionBar.Suffix>
                <Button
                  isIconOnly
                  size="sm"
                  variant="ghost"
                  onPress={() => setSelectedKeys(new Set())}
                  aria-label={t('common.clearSelection')}
                >
                  <Icon name="x" />
                </Button>
              </ActionBar.Suffix>
            </ActionBar>
          </Modal.Dialog>
        </Modal.Container>
      </Modal.Backdrop>
      <ItemDetailDrawer
        templateId={detailId}
        name={detailId !== null ? nameMap.get(detailId) : undefined}
        onClose={() => setDetailId(null)}
      />
    </React.Fragment>
  )
}
