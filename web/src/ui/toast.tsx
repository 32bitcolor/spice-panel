import * as React from 'react'
import {
  UNSTABLE_ToastQueue as ToastQueue,
  UNSTABLE_ToastRegion as AriaToastRegion,
  UNSTABLE_Toast as AriaToast,
  UNSTABLE_ToastContent as AriaToastContent,
  Button as AriaButton,
  Text,
} from 'react-aria-components'
import { cn } from './lib/cn'

export type ToastKind = 'danger' | 'success' | 'warning' | 'info'

export interface ToastContent {
  title: string
  kind: ToastKind
}

/** Module-level queue so `toast.*` can be called imperatively from anywhere. */
export const toastQueue = new ToastQueue<ToastContent>({ maxVisibleToasts: 5 })

const add = (title: string, kind: ToastKind, timeout?: number): void => {
  toastQueue.add({ title, kind }, timeout === undefined ? {} : { timeout })
}

/**
 * Imperative toast API, signature-compatible with the call sites migrated from
 * HeroUI: `toast.danger(msg)`, `toast.success(msg)`, etc. Danger toasts persist
 * until dismissed; the rest auto-expire.
 */
export const toast = {
  danger: (title: string): void => add(title, 'danger'),
  success: (title: string): void => add(title, 'success', 5000),
  warning: (title: string): void => add(title, 'warning', 6000),
  info: (title: string): void => add(title, 'info', 5000),
}

const kindAccent: Record<ToastKind, string> = {
  danger: 'text-danger',
  success: 'text-success',
  warning: 'text-warning',
  info: 'text-accent',
}

const kindStripe: Record<ToastKind, string> = {
  danger: 'bg-danger',
  success: 'bg-success',
  warning: 'bg-warning',
  info: 'bg-accent',
}

/** Mount once near the app root (replaces HeroUI's `<Toast.Provider />`). */
export const ToastRegion: React.FC = (): React.ReactElement => (
  <AriaToastRegion
    queue={toastQueue}
    className="fixed bottom-4 right-4 z-[1000] flex flex-col gap-2.5 outline-none"
  >
    {({ toast: t }) => (
      <AriaToast
        toast={t}
        className="hud-panel flex min-w-[280px] max-w-[380px] items-start gap-3 py-3 pl-4 pr-3 data-[focus-visible]:hud-glow"
      >
        <span className={cn('mt-0.5 h-full w-0.5 self-stretch', kindStripe[t.content.kind])} />
        <AriaToastContent className="flex-1">
          <Text
            slot="title"
            className={cn('text-[13px] font-medium', kindAccent[t.content.kind])}
          >
            {t.content.title}
          </Text>
        </AriaToastContent>
        <AriaButton
          slot="close"
          aria-label="Dismiss"
          className="grid h-5 w-5 place-items-center text-muted outline-none transition hover:text-foreground data-[focus-visible]:hud-glow"
        >
          <svg viewBox="0 0 16 16" className="h-3.5 w-3.5" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <path d="m4 4 8 8M12 4l-8 8" />
          </svg>
        </AriaButton>
      </AriaToast>
    )}
  </AriaToastRegion>
)
