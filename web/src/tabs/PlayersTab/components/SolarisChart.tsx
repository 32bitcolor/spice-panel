import type React from 'react'
import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import {
  LineChart, Line, XAxis, YAxis, Tooltip, Legend, ResponsiveContainer,
} from 'recharts'
import type { StatSnapshot } from '../../../api/client'
import { SectionLabel } from '../../../dune-ui'

interface SolarisChartProps {
  data: StatSnapshot[]
}

interface SolarisPoint {
  time: string
  balance: number
  cum_earned: number
  cum_spent: number
}

function fmtSolaris(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return String(n)
}

function fmtTime(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

export const SolarisChart: React.FC<SolarisChartProps> = ({ data }) => {
  const { t } = useTranslation()
  const [hidden, setHidden] = useState<Set<string>>(new Set())

  const LINES: { key: keyof SolarisPoint, label: string, color: string }[] = [
    { key: 'balance', label: t('players.detail.solarisBalance'), color: 'var(--accent)' },
    { key: 'cum_earned', label: t('players.detail.earned'), color: '#52c080' },
    { key: 'cum_spent', label: t('players.detail.spent'), color: '#e05252' },
  ]

  const toggle = (key: string) => {
    setHidden((prev) => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  const points = useMemo<SolarisPoint[]>(() => {
    const snaps = data.filter((s) => s.solaris_balance != null)
    if (snaps.length === 0) return []
    let cumEarned = 0
    let cumSpent = 0
    return snaps.map((s, i) => {
      const balance = s.solaris_balance as number
      if (i > 0) {
        const prev = snaps[i - 1].solaris_balance as number
        const delta = balance - prev
        if (delta > 0) cumEarned += delta
        else if (delta < 0) cumSpent += -delta
      }
      return { time: s.snapped_at, balance, cum_earned: cumEarned, cum_spent: cumSpent }
    })
  }, [data])

  if (points.length === 0) {
    return (
      <div>
        <SectionLabel>{t('players.detail.solarisHistory')}</SectionLabel>
        <p className="text-muted text-sm mt-2">
          {t('players.detail.solarisHistoryEmpty')}
        </p>
      </div>
    )
  }

  const visibleLines = LINES.filter((l) => !hidden.has(l.key))

  return (
    <div>
      <SectionLabel>{t('players.detail.solarisHistory')}</SectionLabel>
      <div className="mt-3 h-56">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={points} margin={{ top: 4, right: 8, left: 8, bottom: 0 }}>
            <XAxis
              dataKey="time"
              tickFormatter={fmtTime}
              tick={{ fontSize: 11, fill: 'var(--muted)' }}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              tickFormatter={fmtSolaris}
              tick={{ fontSize: 11, fill: 'var(--muted)' }}
              tickLine={false}
              axisLine={false}
              width={48}
            />
            <Tooltip
              formatter={(val, name) => [fmtSolaris(val as number), String(name)]}
              labelFormatter={(d) => fmtTime(String(d))}
              contentStyle={{
                background: 'var(--surface)',
                border: '1px solid var(--border)',
                borderRadius: 'var(--radius)',
                fontSize: 12,
              }}
            />
            <Legend
              onClick={(e) => toggle(e.dataKey as string)}
              formatter={(value, entry) => (
                <span style={{ color: hidden.has((entry as { dataKey: string }).dataKey) ? 'var(--muted)' : undefined }}>
                  {value}
                </span>
              )}
            />
            {visibleLines.map((l) => (
              <Line
                key={String(l.key)}
                type="monotone"
                dataKey={l.key}
                name={l.label}
                stroke={l.color}
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 4 }}
                connectNulls
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}
