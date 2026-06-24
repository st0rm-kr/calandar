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
    <div className="min-h-screen bg-canvas text-ink">
      <div
        className={`min-h-screen px-5 py-6 ${
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
