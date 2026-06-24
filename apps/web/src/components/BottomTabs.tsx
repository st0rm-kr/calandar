import { Link, useLocation } from 'react-router-dom'

type Tab = {
  to: string
  label: string
  match: (path: string) => boolean
}

const tabs: Tab[] = [
  { to: '/', label: '日历', match: (path) => path === '/' || path.startsWith('/calendar') },
  { to: '/events', label: '活动', match: (path) => path.startsWith('/event') },
  { to: '/inbox', label: '收件箱', match: (path) => path.startsWith('/inbox') },
  { to: '/friends', label: '好友', match: (path) => path.startsWith('/friend') || path.startsWith('/group') },
  { to: '/me', label: '我的', match: (path) => path.startsWith('/me') || path.startsWith('/profile') },
]

type BottomTabsProps = {
  inboxCount?: number
}

export function BottomTabs({ inboxCount = 0 }: BottomTabsProps) {
  const location = useLocation()

  return (
    <nav className="fixed inset-x-0 bottom-0 z-20 border-t border-white/10 bg-neutral-950/90 backdrop-blur">
      <div className="mx-auto flex max-w-md items-stretch justify-between px-2 pb-[env(safe-area-inset-bottom)]">
        {tabs.map((tab) => {
          const active = tab.match(location.pathname)
          return (
            <Link
              className={`relative flex flex-1 flex-col items-center gap-0.5 py-3 text-xs ${
                active ? 'text-white' : 'text-white/40'
              }`}
              key={tab.to}
              to={tab.to}
            >
              <span>{tab.label}</span>
              {tab.to === '/inbox' && inboxCount > 0 ? (
                <span className="absolute right-1/2 top-1.5 translate-x-3 rounded-full bg-sky-400 px-1.5 text-[10px] font-semibold text-neutral-950">
                  {inboxCount > 99 ? '99+' : inboxCount}
                </span>
              ) : null}
            </Link>
          )
        })}
      </div>
    </nav>
  )
}
