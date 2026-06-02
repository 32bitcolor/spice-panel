import { useTranslation } from 'react-i18next'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import type { SessionRecord } from '../../../api/client'
import { SectionLabel } from '../../../dune-ui'

interface Props {
  data: SessionRecord[]
}

type DayBucket = { date: string, minutes: number }

function aggregate(records: SessionRecord[]): DayBucket[] {
  const map = new Map<string, number>()
  for (const r of records) {
    const day = r.started_at.slice(0, 10)
    map.set(day, (map.get(day) ?? 0) + Math.round(r.duration_secs / 60))
  }
  return Array.from(map.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, minutes]) => ({ date, minutes }))
}

function fmtDate(d: string): string {
  return new Date(d + 'T12:00:00Z').toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

export function SessionChart({ data }: Props) {
  const { t } = useTranslation()
  const buckets = aggregate(data)

  if (buckets.length === 0) {
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
      <div className="mt-3 h-40">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={buckets} margin={{ top: 4, right: 8, left: 8, bottom: 0 }}>
            <XAxis
              dataKey="date"
              tickFormatter={fmtDate}
              tick={{ fontSize: 11, fill: 'var(--muted)' }}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              unit="m"
              tick={{ fontSize: 11, fill: 'var(--muted)' }}
              tickLine={false}
              axisLine={false}
              width={36}
            />
            <Tooltip
              formatter={(val) => [`${val as number}m`, t('players.detail.playtime')]}
              labelFormatter={(d) => fmtDate(String(d))}
              contentStyle={{
                background: 'var(--surface)',
                border: '1px solid var(--border)',
                borderRadius: 'var(--radius)',
                fontSize: 12,
              }}
            />
            <Bar dataKey="minutes" fill="var(--accent)" radius={[3, 3, 0, 0]} maxBarSize={32} />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}
