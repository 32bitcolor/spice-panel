import { useState, useEffect, useCallback } from 'react'
import { Button, Spinner, Tabs } from '@heroui/react'
import { api } from '../../../api/client'
import type { BotStatus, BotConfig } from '../../../api/client'
import { Icon } from '../../../dune-ui'
import BotStatusCard from './BotStatusCard'
import BotActions from './BotActions'
import BotLogViewer from './BotLogViewer'
import BotConfigEditor from './BotConfigEditor'

export default function BotControlPanel() {
  const [status, setStatus] = useState<BotStatus | null>(null)
  const [config, setConfig] = useState<BotConfig | null>(null)
  const [statusLoading, setStatusLoading] = useState(false)
  const [configLoading, setConfigLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const loadStatus = useCallback(async () => {
    setStatusLoading(true)
    try {
      setStatus(await api.marketBot.status())
      setError(null)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setStatusLoading(false)
    }
  }, [])

  const loadConfig = useCallback(async () => {
    setConfigLoading(true)
    try {
      setConfig(await api.marketBot.config())
    } catch {
      // config load failure is non-fatal — status still shows
    } finally {
      setConfigLoading(false)
    }
  }, [])

  useEffect(() => {
    loadStatus()
    loadConfig()
  }, [loadStatus, loadConfig])

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold text-muted uppercase tracking-wider">Bot Control — Revy</span>
        <Button size="sm" variant="ghost" onPress={() => { loadStatus(); loadConfig() }} isDisabled={statusLoading}>
          {statusLoading ? <Spinner size="sm" color="current" /> : <Icon name="refresh-cw" />}
        </Button>
      </div>

      {error ? (
        <p className="text-xs text-danger">{error}</p>
      ) : status ? (
        <>
          <div className="flex flex-wrap items-center gap-4 justify-between">
            <BotStatusCard status={status} />
            <BotActions status={status} onRefresh={loadStatus} />
          </div>

          <Tabs className="mt-1">
            <Tabs.ListContainer>
              <Tabs.List aria-label="Bot sections">
                <Tabs.Tab id="config">Config<Tabs.Indicator /></Tabs.Tab>
                <Tabs.Tab id="logs">Logs<Tabs.Indicator /></Tabs.Tab>
              </Tabs.List>
            </Tabs.ListContainer>
            <Tabs.Panel id="config" className="pt-3">
              {configLoading ? (
                <div className="flex justify-center py-6"><Spinner size="sm" /></div>
              ) : config ? (
                <BotConfigEditor config={config} onSaved={setConfig} />
              ) : (
                <p className="text-xs text-muted">Config unavailable.</p>
              )}
            </Tabs.Panel>
            <Tabs.Panel id="logs" className="pt-3 h-64">
              <BotLogViewer />
            </Tabs.Panel>
          </Tabs>
        </>
      ) : statusLoading ? (
        <div className="flex justify-center py-6"><Spinner size="sm" /></div>
      ) : null}
    </div>
  )
}
