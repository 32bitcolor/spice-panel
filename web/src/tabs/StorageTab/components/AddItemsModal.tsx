import * as React from 'react'
import { useTranslation } from 'react-i18next'
import {
  Button, Chip, Modal, SearchField, Separator, Spinner, TextField, toast,
} from '@heroui/react'
import type { Selection } from 'react-aria-components'
import type { DataGridColumn } from '@heroui-pro/react'
import { DataGrid } from '@heroui-pro/react'
import { useAtomValue } from 'jotai'
import { api } from '../../../api/client'
import { itemDataSyncAtom } from '../../../data/store'
import { ActionBar, Icon, LoadingState, NumberInput } from '../../../dune-ui'
import { ItemDetailDrawer } from '../../../components/ItemDetailDrawer'
import { ItemOptionRow } from '../../../components/ItemOptionRow'
import { StagedItemCell } from '../../../components/StagedItemCell'
import type { AddItemsModalProps } from './interfaces'
import type { AddResult, AddStagedItem } from './types'

export const AddItemsModal: React.FC<AddItemsModalProps> = ({
  container, open, onClose, onSuccess, onRefresh,
}) => {
  const { t } = useTranslation()
  const itemData = useAtomValue(itemDataSyncAtom)
  const [templates, setTemplates] = React.useState<{ id: string, name: string }[]>([])
  const [loading, setLoading] = React.useState(false)
  const [query, setQuery] = React.useState('')
  const [selected, setSelected] = React.useState('')
  const [qty, setQty] = React.useState(1)
  const [quality, setQuality] = React.useState(0)
  const [staged, setStaged] = React.useState<AddStagedItem[]>([])
  const [submitting, setSubmitting] = React.useState(false)
  const [result, setResult] = React.useState<AddResult>(null)
  const [selectedKeys, setSelectedKeys] = React.useState<Selection>(new Set())
  const [detailId, setDetailId] = React.useState<string | null>(null)

  const keyCounter = React.useRef(0)
  const nextKey = () => String(keyCounter.current++)

  React.useEffect(() => {
    if (!open) return
    Promise.resolve()
      .then(() => {
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
      .then(setTemplates)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [open])

  const nameMap = new Map(templates.map((tpl) => [tpl.id, tpl.name]))

  const _aimq = query.toLowerCase()
  const filtered = !query
    ? []
    : templates
        .filter((tmpl) => tmpl.id.toLowerCase().includes(_aimq) || tmpl.name.toLowerCase().includes(_aimq))
        .slice(0, 100)

  const pick = (tmpl: { id: string, name: string }) => {
    setSelected(tmpl.id)
    setQuery(tmpl.name ? `${tmpl.id}  —  ${tmpl.name}` : tmpl.id)
  }

  const addToStaged = () => {
    if (!selected) {
      toast.warning(t('storage.addModal.selectTemplate'))
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
      const items = staged.map(({ template, qty: q, quality: ql }) => ({ template, qty: q, quality: ql }))
      const res = await api.storage.giveItems(container.id, items)
      setResult(res)
      setStaged([])
      setSelectedKeys(new Set())
      if (res.skipped.length === 0) onSuccess()
      else if (res.given.length > 0) onRefresh()
    }
    catch (e: unknown) {
      toast.danger(e instanceof Error ? e.message : String(e))
    }
    finally {
      setSubmitting(false)
    }
  }

  const columns: DataGridColumn<AddStagedItem>[] = [
    {
      id: 'template',
      isRowHeader: true,
      header: t('storage.addModal.templateLabel'),
      minWidth: 200,
      allowsResizing: true,
      cell: (item) => (
        <StagedItemCell
          templateId={item.template}
          name={nameMap.get(item.template) || ''}
          entry={itemData.items[item.template] ?? null}
        />
      ),
    },
    {
      id: 'qty',
      header: t('storage.addModal.qtyLabel'),
      minWidth: 130,
      maxWidth: 250,
      allowsResizing: true,
      cell: (item) => (
        <NumberInput
          ariaLabel={t('storage.addModal.qtyColLabel')}
          min={1}
          value={item.qty}
          onChange={(v) => updateStaged(item._key, 'qty', v)}
          className="w-full"
        />
      ),
    },
    {
      id: 'quality',
      header: t('storage.addModal.qualityLabel'),
      minWidth: 130,
      maxWidth: 250,
      allowsResizing: true,
      cell: (item) => (
        <NumberInput
          ariaLabel={t('storage.addModal.qualityColLabel')}
          min={0}
          value={item.quality}
          onChange={(v) => updateStaged(item._key, 'quality', v)}
          className="w-full"
        />
      ),
    },
    {
      id: 'actions',
      header: '',
      width: 88,
      cell: (item) => (
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
            variant="danger-soft"
            isIconOnly
            onPress={() => removeFromStaged(item._key)}
            aria-label={t('common.remove')}
          >
            <Icon name="trash" />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <React.Fragment>
      <Modal.Backdrop variant="blur" className="bg-linear-to-t from-(--background)/85 via-(--background)/40 to-transparent" isOpen={open} onOpenChange={(v) => !v && onClose()}>
        <Modal.Container size="cover" scroll="outside">
          <Modal.Dialog className="p-10">
            <Modal.CloseTrigger />
            <Modal.Header>
              <Modal.Heading className="text-accent">
                {container.name || t('storage.containerTitle', { id: container.id })}
                {' '}
                —
                {' '}
                {t('storage.addItems')}
              </Modal.Heading>
            </Modal.Header>
            <Modal.Body className="flex flex-col gap-3">
              {loading
                ? (
                    <LoadingState size="sm" />
                  )
                : (
                    <React.Fragment>
                      <div className="flex items-end gap-3 shrink-0">
                        <TextField className="flex-1 min-w-0" aria-label={t('storage.addModal.templateLabel')}>
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
                                <SearchField.Input placeholder={t('storage.addModal.searchPlaceholder')} />
                                <SearchField.ClearButton />
                              </SearchField.Group>
                            </SearchField>
                            {filtered.length > 0 && (
                              <div className="absolute z-50 w-full mt-1 rounded-[var(--radius)] border border-border bg-surface overflow-y-auto max-h-52">
                                {filtered.map((tmpl) => (
                                  <ItemOptionRow
                                    key={tmpl.id}
                                    id={tmpl.id}
                                    name={tmpl.name}
                                    entry={itemData.items[tmpl.id] ?? null}
                                    onPick={() => pick(tmpl)}
                                    onDetail={() => setDetailId(tmpl.id)}
                                  />
                                ))}
                              </div>
                            )}
                          </div>
                        </TextField>
                        <NumberInput
                          prefix={t('storage.addModal.qtyLabel')}
                          ariaLabel={t('storage.addModal.qtyLabel')}
                          min={1}
                          value={qty}
                          onChange={setQty}
                          className="w-56 shrink-0"
                        />
                        <NumberInput
                          prefix={t('storage.addModal.qualityLabel')}
                          ariaLabel={t('storage.addModal.qualityLabel')}
                          min={0}
                          value={quality}
                          onChange={setQuality}
                          className="w-56 shrink-0"
                        />
                        <Button size="sm" onPress={addToStaged} isDisabled={!selected} className="shrink-0">
                          <Icon name="plus" />
                          {' '}
                          {t('storage.addModal.add')}
                        </Button>
                      </div>

                      {staged.length > 0 && (
                        <DataGrid
                          aria-label={t('storage.addItems')}
                          columns={columns}
                          data={staged}
                          getRowId={(item) => item._key}
                          selectedKeys={selectedKeys}
                          selectionMode="multiple"
                          showSelectionCheckboxes
                          onSelectionChange={setSelectedKeys}
                          className="flex-1 min-h-0"
                          scrollContainerClassName="h-full overflow-y-auto"
                          allowsColumnResize
                        />
                      )}

                      {result && (
                        <div className="text-xs shrink-0 rounded-[var(--radius)] px-3 py-2 bg-surface border border-border">
                          {result.given.length > 0 && (
                            <div className="text-success">
                              ✓ Added:
                              {result.given.join(', ')}
                            </div>
                          )}
                          {result.skipped.map((s, i) => (
                            <div key={i} className="text-danger">
                              ✕ Skipped
                              {s.template}
                              :
                              {s.reason}
                            </div>
                          ))}
                        </div>
                      )}
                    </React.Fragment>
                  )}
            </Modal.Body>
            <Modal.Footer>
              <Button variant="tertiary" size="sm" slot="close">{t('common.cancel')}</Button>
              <Button size="sm" onPress={handleSubmit} isDisabled={submitting || staged.length === 0}>
                {submitting ? <Spinner size="sm" color="current" /> : <Icon name="plus" />}
                {t('storage.addModal.add')}
                {' '}
                {staged.length}
                {' '}
                Item
                {staged.length !== 1 ? 's' : ''}
              </Button>
            </Modal.Footer>
            {/* Inside the dialog so the ActionBar isn't made inert by React Aria's modal underlay. */}
            <ActionBar aria-label={t('storage.addItems')} isOpen={selectionCount > 0}>
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
