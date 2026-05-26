import { useState, useEffect, useRef, useCallback } from 'react'
import { Button, Checkbox } from '@heroui/react'
import { getWsBase } from '../../../api/client'
import { Icon } from '../../../dune-ui'

export default function BotLogViewer() {
  const [connected, setConnected] = useState(false)
  const [lines, setLines] = useState<string[]>([])
  const [autoScroll, setAutoScroll] = useState(true)
  const wsRef = useRef<WebSocket | null>(null)
  const bufRef = useRef<string[]>([])
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const containerRef = useRef<HTMLPreElement | null>(null)

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [lines, autoScroll])

  const startFlush = useCallback(() => {
    if (timerRef.current) return
    timerRef.current = setInterval(() => {
      if (bufRef.current.length > 0) {
        setLines(prev => {
          const combined = [...prev, ...bufRef.current]
          bufRef.current = []
          return combined.length > 5000 ? combined.slice(combined.length - 5000) : combined
        })
      }
    }, 200)
  }, [])

  const stopFlush = useCallback(() => {
    if (timerRef.current) { clearInterval(timerRef.current); timerRef.current = null }
  }, [])

  const connect = useCallback(() => {
    if (wsRef.current) { wsRef.current.close(); wsRef.current = null }
    stopFlush()
    bufRef.current = []
    setLines([])
    setConnected(false)

    const ws = new WebSocket(`${getWsBase()}/market-bot/logs`)
    wsRef.current = ws
    ws.onopen = () => { setConnected(true); startFlush() }
    ws.onmessage = (e: MessageEvent) => { bufRef.current.push(e.data as string) }
    ws.onerror = () => {}
    ws.onclose = () => {
      setConnected(false)
      stopFlush()
      if (bufRef.current.length > 0) {
        setLines(prev => [...prev, ...bufRef.current])
        bufRef.current = []
      }
    }
  }, [startFlush, stopFlush])

  const disconnect = useCallback(() => {
    if (wsRef.current) { wsRef.current.close(); wsRef.current = null }
    stopFlush()
    setConnected(false)
  }, [stopFlush])

  useEffect(() => () => { disconnect() }, [disconnect])

  return (
    <div className="flex flex-col gap-2 h-full min-h-0">
      <div className="flex items-center gap-2 shrink-0">
        <span className={`text-xs font-mono ${connected ? 'text-success' : 'text-muted'}`}>
          {connected ? '● connected' : '○ disconnected'}
        </span>
        <div className="flex-1" />
        <Checkbox isSelected={autoScroll} onChange={setAutoScroll}>Auto-scroll</Checkbox>
        {!connected ? (
          <Button size="sm" variant="outline" onPress={connect}>
            <Icon name="play" /> Connect
          </Button>
        ) : (
          <Button size="sm" variant="danger-soft" onPress={disconnect}>
            <Icon name="square" /> Stop
          </Button>
        )}
        {lines.length > 0 && (
          <Button size="sm" variant="ghost" onPress={() => { setLines([]); bufRef.current = [] }}>
            <Icon name="trash-2" /> Clear
          </Button>
        )}
      </div>
      <pre
        ref={containerRef}
        className="flex-1 overflow-auto p-3 text-xs font-mono m-0 whitespace-pre-wrap break-all rounded-[var(--radius)] border border-border/60 bg-background text-success min-h-[120px]"
      >
        {lines.length === 0
          ? (connected ? 'Waiting for log lines...' : 'Press Connect to stream bot logs.')
          : lines.join('\n')}
      </pre>
    </div>
  )
}
