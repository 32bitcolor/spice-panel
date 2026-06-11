import * as React from 'react'

/**
 * Polls `fn` every `intervalMs` while `active` is true.
 * Returns `countdown` (seconds until next auto-refresh) and a `refresh`
 * function for manual triggers — calling it fires `fn` and resets the timer.
 */
export const useAutoRefresh = (
  fn: () => void,
  intervalMs: number,
): { countdown: number, refresh: () => void } => {
  const fnRef = React.useRef(fn)
  React.useEffect(() => {
    fnRef.current = fn
  })

  const secsTotal = Math.round(intervalMs / 1000)
  const [countdown, setCountdown] = React.useState(secsTotal)

  React.useEffect(() => {
    Promise.resolve().then(() => setCountdown(secsTotal))

    const poll = setInterval(() => {
      fnRef.current()
      setCountdown(secsTotal)
    }, intervalMs)

    const tick = setInterval(() => {
      setCountdown((s) => Math.max(0, s - 1))
    }, 1000)

    return () => {
      clearInterval(poll)
      clearInterval(tick)
    }
  }, [intervalMs, secsTotal])

  const refresh = React.useCallback(() => {
    fnRef.current()
    setCountdown(secsTotal)
  }, [secsTotal])

  return { countdown, refresh }
}
