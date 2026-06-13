import * as React from 'react'
import { useTranslation } from 'react-i18next'
import {
  Button,
  Label,
  Modal,
  Spinner,
  TextField,
  toast,
} from '@heroui/react'
import { EmptyState } from '@heroui-pro/react'
import { Icon as IconifyIcon } from '@iconify/react'
import { api } from '../api/client'
import type { BlueprintRow, Player } from '../api/client'
import { DataTable, Dropzone, Icon, PageHeader, type Column } from '../dune-ui'
import { usePermissions } from '../hooks/usePermissions'
import { PlayerSearchField } from '../components/PlayerSearchField'
import type { BlueprintsTabKey, BlueprintsTabProps, ImportModalProps } from './types'

export const BlueprintsTab: React.FC<BlueprintsTabProps> = ({ isSignedIn = true }) => {
  const { t } = useTranslation()
  const { can } = usePermissions()
  const canWorldWrite = can('world:write')
  const canExportData = can('data:export')
  const [blueprints, setBlueprints] = React.useState<BlueprintRow[]>([])
  const [loading, setLoading] = React.useState(false)
  const [showImport, setShowImport] = React.useState(false)

  const COLUMNS: Column<BlueprintsTabKey>[] = [
    { key: 'id', label: t('blueprints.columns.id'), width: 80 },
    { key: 'owner_name', label: t('blueprints.columns.owner'), minWidth: 140 },
    { key: 'name', label: t('blueprints.columns.name'), minWidth: 200 },
    { key: 'item_id', label: t('blueprints.columns.itemId'), minWidth: 200 },
    { key: 'pieces', label: t('blueprints.columns.pieces'), width: 100 },
    { key: 'placeables', label: t('blueprints.columns.placeables'), width: 110 },
    { key: 'actions', label: '', width: 110, sortable: false },
  ]

  const load = React.useCallback(() => {
    Promise.resolve()
      .then(() => setLoading(true))
      .then(() => api.blueprints.list())
      .then(setBlueprints)
      .catch((e: unknown) => toast.danger(t('blueprints.failedToLoad', { message: e instanceof Error ? e.message : String(e) })))
      .finally(() => setLoading(false))
  }, [t])

  React.useEffect(() => {
    load()
  }, [load])

  return (
    <div className="flex flex-col h-full gap-3 min-h-0">
      {!isSignedIn && (
        <div className="shrink-0 rounded-[var(--radius)] px-4 py-2 text-xs font-medium bg-danger/10 border border-danger/40 text-danger flex items-center gap-2">
          <Icon name="triangle-alert" />
          <span>
            A
            {' '}
            <strong>{t('blueprints.layoutAccountStrong')}</strong>
            {' '}
            account is required to export or import blueprints. Sign in using the button
            in the top right.
          </span>
        </div>
      )}

      <PageHeader
        title={t('blueprints.title', { count: blueprints.length })}
        subtitle={t('blueprints.subtitle')}
      >
        <Button size="sm" variant="ghost" onPress={load} isDisabled={loading}>
          {loading
            ? (
                <Spinner size="sm" color="current" />
              )
            : (
                <>
                  <Icon name="refresh-cw" />
                  {' '}
                  {t('common.refresh')}
                </>
              )}
        </Button>
        {canWorldWrite && (
          <Button size="sm" onPress={() => setShowImport(true)} isDisabled={!isSignedIn}>
            <Icon name="upload" />
            {' '}
            {t('blueprints.importBlueprint')}
          </Button>
        )}
      </PageHeader>

      <DataTable<BlueprintRow, BlueprintsTabKey>
        aria-label={t('blueprints.ariaLabel')}
        className="min-h-0 max-h-full"
        columns={canExportData ? COLUMNS : COLUMNS.filter((c) => c.key !== 'actions')}
        rows={blueprints}
        loading={loading}
        rowId={(b) => String(b.id)}
        initialSort={{ column: 'id', direction: 'ascending' }}
        sortValue={(b, k) => (k === 'actions' ? '' : (b as unknown as Record<string, string | number>)[k])}
        emptyState={(
          <EmptyState size="sm">
            <EmptyState.Header>
              <EmptyState.Media variant="icon">
                <IconifyIcon icon="gravity-ui:box" className="size-5" />
              </EmptyState.Media>
              <EmptyState.Title>{t('blueprints.noBlueprintsFound')}</EmptyState.Title>
            </EmptyState.Header>
          </EmptyState>
        )}
        renderCell={(b, key) => {
          switch (key) {
            case 'id':
              return <span className="font-mono text-muted">{b.id}</span>
            case 'owner_name':
              return b.owner_name
            case 'name':
              return b.name || <span className="text-muted">—</span>
            case 'item_id':
              return <span className="font-mono text-muted">{b.item_id}</span>
            case 'pieces':
              return <span className="text-muted">{b.pieces}</span>
            case 'placeables':
              return <span className="text-muted">{b.placeables}</span>
            case 'actions':
              return isSignedIn
                ? (
                    <a
                      href={api.blueprints.exportUrl(b.id)}
                      download={b.name ? `${b.name.replace(/[/\\:*?"<>|]/g, '_')}.json` : `blueprint_${b.id}.json`}
                    >
                      <Button size="sm" variant="outline" className="w-full">
                        <Icon name="download" />
                        {' '}
                        {t('common.export')}
                      </Button>
                    </a>
                  )
                : (
                    <Button size="sm" variant="outline" className="w-full" isDisabled>
                      <Icon name="download" />
                      {' '}
                      {t('common.export')}
                    </Button>
                  )
          }
        }}
      />

      {canWorldWrite && (
        <ImportModal
          open={showImport}
          onClose={() => setShowImport(false)}
          onSuccess={() => {
            setShowImport(false)
            load()
          }}
        />
      )}
    </div>
  )
}

const ImportModal: React.FC<ImportModalProps> = ({ open, onClose, onSuccess }) => {
  const { t } = useTranslation()
  const [file, setFile] = React.useState<File | null>(null)
  const [selectedPlayer, setSelectedPlayer] = React.useState<Player | null>(null)
  const [submitting, setSubmitting] = React.useState(false)

  React.useEffect(() => {
    if (!open) return
    Promise.resolve().then(() => {
      setFile(null)
      setSelectedPlayer(null)
    })
  }, [open])

  const handleSubmit = async () => {
    if (!file) {
      toast.warning(t('blueprints.selectFile'))
      return
    }
    if (!selectedPlayer) {
      toast.warning(t('blueprints.selectPlayer'))
      return
    }
    setSubmitting(true)
    try {
      const res = await api.blueprints.import(file, selectedPlayer.id)
      if (res.ok) {
        toast.success(t('blueprints.importSuccess'))
        onSuccess()
      }
      else {
        toast.danger(t('blueprints.importFailed', { message: res.error ?? 'unknown error' }))
      }
    }
    catch (e: unknown) {
      toast.danger(t('blueprints.importFailed', { message: e instanceof Error ? e.message : String(e) }))
    }
    finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal.Backdrop variant="blur" className="bg-linear-to-t from-(--background)/85 via-(--background)/40 to-transparent" isOpen={open} onOpenChange={(v) => !v && onClose()}>
      <Modal.Container>
        <Modal.Dialog className="p-10 !overflow-visible">
          <Modal.CloseTrigger />
          <Modal.Header>
            <Modal.Heading className="text-accent">{t('blueprints.importModal.title')}</Modal.Heading>
          </Modal.Header>
          <Modal.Body className="flex flex-col gap-4">
            <TextField>
              <Label>{t('blueprints.importModal.blueprintFile')}</Label>
              <Dropzone
                accept=".json"
                file={file}
                onSelect={setFile}
                prompt={t('blueprints.importModal.dropzone')}
              />
            </TextField>

            <TextField>
              <Label>{t('blueprints.importModal.playerLabel')}</Label>
              <PlayerSearchField
                ariaLabel={t('blueprints.importModal.playerLabel')}
                placeholder={t('blueprints.importModal.playerPlaceholder')}
                onSelect={setSelectedPlayer}
                onClear={() => setSelectedPlayer(null)}
                className="w-full"
              />
            </TextField>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="tertiary" slot="close">
              {t('common.cancel')}
            </Button>
            <Button onPress={handleSubmit} isDisabled={submitting || !file || !selectedPlayer}>
              {submitting ? <Spinner size="sm" color="current" /> : <Icon name="upload" />}
              {t('blueprints.importModal.import')}
            </Button>
          </Modal.Footer>
        </Modal.Dialog>
      </Modal.Container>
    </Modal.Backdrop>
  )
}
