import * as React from 'react'
import { useAtom, useSetAtom } from 'jotai'
import { toast } from '../../ui'
import { useTranslation } from 'react-i18next'
import { api } from '../../api/client'
import type { SyncStatus } from '../../api/client'
import {
  settingsOpenAtom,
  syncStepAtom,
  updateApplyingAtom,
  updateCheckingAtom,
  updateErrorAtom,
  updateInfoAtom,
  updatePhaseAtom,
  updatePromptOpenAtom,
} from '../../atoms/app'

const sleep = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms))

const UPDATE_CACHE_KEY = 'dune_update_cache'
const UPDATE_CACHE_TTL_MS = 60 * 60 * 1000

export interface AppUpdate {
  checkUpdate: () => Promise<void>
  applyUpdate: (force?: boolean) => Promise<void>
  syncUpstream: () => Promise<void>
}

// Owns the update check/apply/poll-and-reload flow. State lives in atoms so the
// navbar release widget and the Settings/prompt modals all observe it.
export const useAppUpdate = (): AppUpdate => {
  const { t } = useTranslation()
  const [, setUpdateInfo] = useAtom(updateInfoAtom)
  const setChecking = useSetAtom(updateCheckingAtom)
  const setApplying = useSetAtom(updateApplyingAtom)
  const setPhase = useSetAtom(updatePhaseAtom)
  const setError = useSetAtom(updateErrorAtom)
  const setSettingsOpen = useSetAtom(settingsOpenAtom)
  const setPromptOpen = useSetAtom(updatePromptOpenAtom)
  const setSyncStep = useSetAtom(syncStepAtom)

  // Check for a newer release via the backend — cached in localStorage for 1 hour
  // to avoid hammering GitHub's unauthenticated API rate limit during dev HMR cycles.
  React.useEffect(() => {
    try {
      const cached = localStorage.getItem(UPDATE_CACHE_KEY)
      if (cached) {
        const { ts, data } = JSON.parse(cached)
        if (Date.now() - ts < UPDATE_CACHE_TTL_MS) {
          Promise.resolve().then(() => setUpdateInfo(data))
          return
        }
      }
    }
    catch { /* ignore corrupt cache */ }
    api.update.check().then((data) => {
      setUpdateInfo(data)
      try {
        localStorage.setItem(UPDATE_CACHE_KEY, JSON.stringify({ ts: Date.now(), data }))
      }
      catch { /* ignore */ }
    }).catch(() => {})
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const checkUpdate = async (): Promise<void> => {
    setChecking(true)
    try {
      const data = await api.update.check()
      setUpdateInfo(data)
      try {
        localStorage.setItem(UPDATE_CACHE_KEY, JSON.stringify({ ts: Date.now(), data }))
      }
      catch { /* ignore */ }
    }
    catch {
      // silently ignore — user can retry
    }
    finally {
      setChecking(false)
    }
  }

  const applyUpdate = async (force = false): Promise<void> => {
    setApplying(true)
    setPhase('downloading')
    setError(undefined)
    setPromptOpen(false)
    setSettingsOpen(false)
    try {
      const result = await api.update.apply(force)
      if (!result.updated) {
        toast.info(result.message)
        setApplying(false)
        return
      }
      localStorage.removeItem(UPDATE_CACHE_KEY)
      setUpdateInfo(null)

      // Binary swapped; server is restarting in ~500ms.
      setPhase('verifying')
      await sleep(400)
      setPhase('extracting')
      await sleep(300)
      setPhase('restarting')

      // Poll /status until the server comes back up (max 90s).
      const started = Date.now()
      const TIMEOUT_MS = 90_000
      const LONG_WAIT_MS = 20_000
      await sleep(2000)
      setPhase('waiting')

      let back = false
      while (Date.now() - started < TIMEOUT_MS) {
        if (Date.now() - started > LONG_WAIT_MS) {
          setPhase('waitingLong')
        }
        try {
          await api.status()
          back = true
          break
        }
        catch {
          await sleep(2000)
        }
      }

      if (back) {
        setPhase('ready')
        await sleep(800)
        window.location.reload()
      }
      else {
        // Timed out but keep polling in background; page will reload on next success.
        setPhase('waitingLong')
        const keepPolling = async () => {
          while (true) {
            await sleep(3000)
            try {
              await api.status()
              window.location.reload()
              return
            }
            catch { /* keep trying */ }
          }
        }
        void keepPolling()
      }
    }
    catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setError(t('app.updateFailed', { message: msg }))
      setPhase('error')
    }
  }

  // Wait for the backend to come back after a re-exec, then reload the page.
  const waitForServerAndReload = async (): Promise<void> => {
    const started = Date.now()
    const TIMEOUT_MS = 120_000
    await sleep(2000)
    setPhase('waiting')
    while (Date.now() - started < TIMEOUT_MS) {
      if (Date.now() - started > 20_000) setPhase('waitingLong')
      try {
        await api.status()
        setPhase('ready')
        await sleep(800)
        window.location.reload()
        return
      }
      catch {
        await sleep(2000)
      }
    }
    const keepPolling = async (): Promise<void> => {
      for (;;) {
        await sleep(3000)
        try {
          await api.status()
          window.location.reload()
          return
        }
        catch { /* keep trying */ }
      }
    }
    void keepPolling()
  }

  // Fork-safe upstream sync: merge upstream (backend + safe frontend), keep the
  // spice-panel UI, rebuild + swap + re-exec on the host. Polls progress, then
  // reuses the wait-and-reload once the server restarts into the new binary.
  const syncUpstream = async (): Promise<void> => {
    setApplying(true)
    setError(undefined)
    setSyncStep('Starting…')
    try {
      await api.update.sync()
      let st: SyncStatus | null = null
      const started = Date.now()
      const RUN_TIMEOUT_MS = 6 * 60_000
      while (Date.now() - started < RUN_TIMEOUT_MS) {
        await sleep(1500)
        try {
          st = await api.update.syncStatus()
        }
        catch {
          // Connection dropped — the server likely re-exec'd on success.
          st = null
          break
        }
        setSyncStep(st.step ? `${st.step} · ${st.message}` : st.message)
        if (st.done) break
      }

      if (st?.error) {
        setSyncStep(null)
        setApplying(false)
        toast.danger(st.error)
        return
      }
      if (st?.no_op) {
        setSyncStep(null)
        setApplying(false)
        toast.info(st.message || t('app.upToDate', 'Already up to date'))
        return
      }

      // Success (done, no error) or the connection dropped on re-exec → the
      // server is restarting into the freshly-built binary.
      localStorage.removeItem(UPDATE_CACHE_KEY)
      setUpdateInfo(null)
      setSyncStep(null)
      setSettingsOpen(false)
      setPhase('restarting')
      await waitForServerAndReload()
    }
    catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      setSyncStep(null)
      setApplying(false)
      setError(t('app.updateFailed', { message: msg }))
      toast.danger(msg)
    }
  }

  return { checkUpdate, applyUpdate, syncUpstream }
}
