import * as React from 'react'
import { useTranslation } from 'react-i18next'
import { Button, Spinner, toast } from '@heroui/react'
import { Icon } from '../dune-ui'
import { SettingsConfigForm } from './SettingsConfigForm'
import { DeleteServerModal } from './DeleteServerModal'
import { useActiveServer } from '../context/useActiveServer'

export interface ManageServerPageProps {
  serverId: string
  /** Return to the dashboard. */
  onBack: () => void
  /** Whether the session may delete/control servers. */
  canControl: boolean
}

// Full-page per-server settings: rename + control/SSH/DB/broker/advanced +
// delete. Reuses SettingsConfigForm in per-server scope targeted at serverId.
export const ManageServerPage: React.FC<ManageServerPageProps> = ({ serverId, onBack, canControl }) => {
  const { t } = useTranslation()
  const { servers, removeServer } = useActiveServer()
  const saveRef = React.useRef<(() => Promise<void>) | null>(null)
  const [saving, setSaving] = React.useState(false)
  const [deleteOpen, setDeleteOpen] = React.useState(false)
  const [deleting, setDeleting] = React.useState(false)

  const serverName = servers.find((s) => s.id === serverId)?.name ?? serverId

  const handleSave = () => {
    void saveRef.current?.()
  }

  return (
    <main className="flex-1 flex flex-col overflow-hidden min-h-0">
      {/* Header */}
      <div className="shrink-0 flex items-center justify-between gap-3 px-4 py-3 border-b border-border">
        <div className="flex items-center gap-2 min-w-0">
          <Button size="sm" variant="ghost" isIconOnly aria-label={t('common.back', 'Back')} onPress={onBack}>
            <Icon name="arrow-left" />
          </Button>
          <span className="text-sm text-muted">{t('manage.title', 'Manage server')}</span>
          <span className="text-sm font-semibold text-foreground truncate">{serverName}</span>
        </div>
        <Button size="sm" onPress={handleSave} isDisabled={saving}>
          {saving ? <Spinner size="sm" color="current" /> : t('common.save', 'Save')}
        </Button>
      </div>

      {/* Per-server settings form (targeted at serverId) */}
      <div className="flex-1 overflow-y-auto p-4 flex flex-col min-h-0">
        <SettingsConfigForm
          key={serverId}
          serverId={serverId}
          saveRef={saveRef}
          onSavingChange={setSaving}
          onRequestDeleteServer={canControl ? () => setDeleteOpen(true) : undefined}
        />
      </div>

      <DeleteServerModal
        open={deleteOpen}
        serverName={serverName}
        busy={deleting}
        onConfirm={() => {
          setDeleting(true)
          removeServer(serverId)
            .then(() => {
              setDeleteOpen(false)
              onBack()
            })
            .catch((e: unknown) => {
              toast.danger(`${t('servers.removeFailed', 'Remove failed')}: ${e instanceof Error ? e.message : String(e)}`)
            })
            .finally(() => setDeleting(false))
        }}
        onCancel={() => setDeleteOpen(false)}
      />
    </main>
  )
}
