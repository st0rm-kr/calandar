import type { ReactNode } from 'react'
import { BottomTabs } from './BottomTabs'

type AppShellProps = {
  children: ReactNode
  inboxCount?: number
  showTabs?: boolean
}

export function AppShell({
  children,
  inboxCount = 0,
  showTabs = true,
}: AppShellProps) {
  return (
    <div className="min-h-screen bg-neutral-950 text-white">
      <div className="mx-auto max-w-md px-4 pb-24 pt-4">{children}</div>
      {showTabs ? <BottomTabs inboxCount={inboxCount} /> : null}
    </div>
  )
}
