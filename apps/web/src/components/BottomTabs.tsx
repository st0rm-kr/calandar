import { Link, useLocation } from 'react-router-dom'
import type { ReactNode } from 'react'

type Tab = {
  to: string
  label: string
  icon: ReactNode
  match: (path: string) => boolean
}

const iconProps = {
  className: 'h-6 w-6',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  viewBox: '0 0 24 24',
  'aria-hidden': true,
} as const

const tabs: Tab[] = [
  {
    to: '/',
    label: '首页',
    match: (path) => path === '/' || path.startsWith('/calendar'),
    icon: (
      <svg {...iconProps}>
        <path
          d="M4 9.5 12 4l8 5.5V19a1 1 0 0 1-1 1h-4v-6h-6v6H5a1 1 0 0 1-1-1z"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
  },
  {
    to: '/events',
    label: '活动',
    match: (path) => path.startsWith('/event'),
    icon: (
      <svg {...iconProps}>
        <rect height="16" rx="3" width="18" x="3" y="5" />
        <path d="M3 10h18M8 3v4M16 3v4" strokeLinecap="round" />
      </svg>
    ),
  },
  {
    to: '/inbox',
    label: '收件箱',
    match: (path) => path.startsWith('/inbox'),
    icon: (
      <svg {...iconProps}>
        <path
          d="M4 13h4l1.5 3h5L16 13h4M4 13l2-7h12l2 7v5a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1z"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
  },
  {
    to: '/friends',
    label: '好友',
    match: (path) => path.startsWith('/friend') || path.startsWith('/group'),
    icon: (
      <svg {...iconProps}>
        <circle cx="9" cy="8" r="3" />
        <path d="M4 19c0-2.8 2.2-5 5-5s5 2.2 5 5" strokeLinecap="round" />
        <path d="M16 6.5a3 3 0 0 1 0 5.8M20 19c0-2.2-1.2-4-3-4.7" strokeLinecap="round" />
      </svg>
    ),
  },
  {
    to: '/me',
    label: '我的',
    match: (path) => path.startsWith('/me') || path.startsWith('/profile'),
    icon: (
      <svg {...iconProps}>
        <circle cx="12" cy="8" r="4" />
        <path d="M5 20c0-3.3 3-6 7-6s7 2.7 7 6" strokeLinecap="round" />
      </svg>
    ),
  },
]

type BottomTabsProps = {
  inboxCount?: number
}

export function BottomTabs({ inboxCount = 0 }: BottomTabsProps) {
  const location = useLocation()

  return (
    <nav className="fixed inset-y-0 left-0 z-20 flex w-20 flex-col items-center border-r border-white/10 bg-black/70 px-2 py-5 shadow-tabbar backdrop-blur-xl lg:w-24">
      <Link
        aria-label="Hangout 首页"
        className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-brand via-rose to-tangerine font-display text-xl font-bold text-white shadow-pop"
        to="/"
      >
        H
      </Link>

      <div className="mt-8 flex flex-1 flex-col items-center gap-2">
        {tabs.map((tab) => {
          const active = tab.match(location.pathname)
          return (
            <Link
              className={`relative flex w-full cursor-pointer flex-col items-center gap-1 rounded-2xl px-2 py-3 text-[11px] font-bold transition-colors duration-200 ${
                active
                  ? 'bg-white text-canvas shadow-[0_0_22px_rgba(0,242,234,0.24),0_0_28px_rgba(255,0,80,0.16)]'
                  : 'text-muted hover:bg-white/10 hover:text-ink'
              }`}
              key={tab.to}
              to={tab.to}
            >
              {tab.icon}
              <span>{tab.label}</span>
              {tab.to === '/inbox' && inboxCount > 0 ? (
                <span className="absolute right-2 top-2 inline-flex min-w-4 items-center justify-center rounded-full bg-rose px-1 text-[10px] font-bold text-white">
                  {inboxCount > 99 ? '99+' : inboxCount}
                </span>
              ) : null}
            </Link>
          )
        })}
      </div>

      <p className="hidden rotate-180 text-[10px] font-bold uppercase tracking-[0.2em] text-muted [writing-mode:vertical-rl] lg:block">
        Hangout
      </p>
    </nav>
  )
}
