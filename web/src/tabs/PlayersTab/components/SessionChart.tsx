import * as React from 'react'
import { useTranslation } from 'react-i18next'
import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip } from 'recharts'
import type { SessionRecord } from '../../../api/client'
import { SectionLabel } from '../../../dune-ui'
import type { SessionChartProps } from './interfaces'
import type { DayBucket } from './types'

const WINDOW_DAYS = 14

const todayUTC = (): string => {
  return new Date().toISOString().slice(0, 10)
}

const aggregate = (records: SessionRecord[]): DayBucket[] => {
  const minutesByDay = new Map<string, number>()
  for (const r of records) {
    const day = r.started_at.slice(0, 10)
    minutesByDay.set(day, (minutesByDay.get(day) ?? 0) + Math.round(r.duration_secs / 60))
  }

  const buckets: DayBucket[] = []
  const today = todayUTC()
  for (let i = WINDOW_DAYS - 1; i >= 0; i--) {
    const d = new Date(today + 'T12:00:00Z')
    d.setUTCDate(d.getUTCDate() - i)
    const date = d.toISOString().slice(0, 10)
    buckets.push({ date, minutes: minutesByDay.get(date) ?? 0 })
  }
  return buckets
}

const fmtDate = (d: string): string => {
  return new Date(d + 'T12:00:00Z').toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

export const SessionChart: React.FC<SessionChartProps> = ({ data }) => {
  const { t } = useTranslation()
  const buckets = aggregate(data)

  if (data.length === 0) {
    return (
      <div>
        <SectionLabel>{t('players.detail.sessionHistory')}</SectionLabel>
        <p className="text-muted text-sm mt-2">
          {t('players.detail.sessionHistoryEmpty')}
        </p>
      </div>
    )
  }

  return (
    <div>
      <SectionLabel>{t('players.detail.sessionHistory')}</SectionLabel>
      <ResponsiveContainer width="100%" height={140}>
        <BarChart data={buckets} margin={{ top: 4, right: 8, bottom: 0, left: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
          <XAxis
            dataKey="date"
            tickFormatter={fmtDate}
            tickMargin={8}
            tick={{ fill: 'var(--muted)', fontSize: 11 }}
            tickLine={false}
            axisLine={{ stroke: 'var(--border)' }}
          />
          <YAxis
            width={36}
            tickFormatter={(v: number) => `${v}m`}
            tick={{ fill: 'var(--muted)', fontSize: 11 }}
            tickLine={false}
            axisLine={{ stroke: 'var(--border)' }}
          />
          <Bar
            dataKey="minutes"
            fill="var(--accent)"
            radius={[3, 3, 0, 0]}
            maxBarSize={32}
            name={t('players.detail.playtime')}
          />
          <Tooltip
            cursor={{ fill: 'var(--surface-secondary)' }}
            contentStyle={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 3, fontSize: 12 }}
            labelStyle={{ color: 'var(--muted)' }}
            labelFormatter={(d) => fmtDate(String(d))}
            formatter={(v) => [`${v}m`, t('players.detail.playtime')]}
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
