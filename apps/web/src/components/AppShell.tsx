import type { ReactNode } from 'react'
import { BottomTabs } from './BottomTabs'
import { FloatingActionButton } from './FloatingActionButton'

type AppShellProps = {
  children: ReactNode
  inboxCount?: number
  showTabs?: boolean
  showFab?: boolean
  fabTo?: string
}

export function AppShell({
  children,
  inboxCount = 0,
  showTabs = true,
  showFab = false,
  fabTo = '/events/new',
}: AppShellProps) {
  return (
    <div className="relative min-h-screen overflow-x-hidden bg-canvas text-ink">
      <div className="pointer-events-none fixed -left-28 top-10 h-72 w-72 rounded-full bg-rose/20 blur-3xl" />
      <div className="pointer-events-none fixed -right-24 top-28 h-80 w-80 rounded-full bg-brand/20 blur-3xl" />
      <div className="pointer-events-none fixed bottom-0 left-1/2 h-72 w-96 -translate-x-1/2 rounded-full bg-tangerine/10 blur-3xl" />
      <div
        className={`relative min-h-screen px-5 py-6 ${
          showTabs
            ? 'w-full pl-24 pr-5 lg:pl-32 lg:pr-8'
            : 'mx-auto max-w-2xl'
        }`}
      >
        {children}
      </div>
      {showFab ? <FloatingActionButton to={fabTo} /> : null}
      {showTabs ? <BottomTabs inboxCount={inboxCount} /> : null}
    </div>
  )
}
