import * as React from 'react'
import { Button, Spinner } from '../ui'
import { Icon } from './Icon'
import type { PageHeaderProps } from './types'

export const PageHeader: React.FC<PageHeaderProps> = ({
  title,
  subtitle,
  onRefresh,
  loading,
  countdown,
  children,
}): React.ReactElement => (
  <div className="mb-1 flex shrink-0 items-start justify-between gap-3 border-b border-border/60 pb-3">
    <div className="min-w-0 flex-1">
      <h2 className="truncate text-base font-semibold text-accent">{title}</h2>
      {renderSubtitle(subtitle)}
    </div>
    {renderActions(onRefresh, loading, countdown, children)}
  </div>
)

const renderSubtitle = (subtitle: React.ReactNode): React.ReactNode => {
  if (subtitle === undefined || subtitle === null) return null
  return <p className="mt-0.5 text-sm text-muted">{subtitle}</p>
}

const renderRefreshBody = (
  loading: boolean | undefined,
  countdown: number | undefined,
): React.ReactNode => {
  if (loading) return <Spinner size={16} />
  return (
    <React.Fragment>
      {renderCountdown(countdown)}
      <Icon name="refresh-cw" />
    </React.Fragment>
  )
}

const renderCountdown = (countdown: number | undefined): React.ReactNode => {
  if (countdown === undefined || countdown === null) return null
  return <span className="w-7 text-right text-xs tabular-nums text-muted/60">{countdown}s</span>
}

const renderActions = (
  onRefresh: (() => void) | undefined,
  loading: boolean | undefined,
  countdown: number | undefined,
  children: React.ReactNode,
): React.ReactNode => {
  if (onRefresh === undefined && !children) return null
  return (
    <div className="flex shrink-0 items-center gap-2">
      {children}
      {renderRefresh(onRefresh, loading, countdown)}
    </div>
  )
}

const renderRefresh = (
  onRefresh: (() => void) | undefined,
  loading: boolean | undefined,
  countdown: number | undefined,
): React.ReactNode => {
  if (onRefresh === undefined) return null
  return (
    <Button size="sm" variant="ghost" onPress={onRefresh} {...(loading !== undefined ? { isDisabled: loading } : {})}>
      {renderRefreshBody(loading, countdown)}
    </Button>
  )
}
