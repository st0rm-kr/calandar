import { Link } from 'react-router-dom'

type FloatingActionButtonProps = {
  to: string
  label?: string
}

export function FloatingActionButton({
  to,
  label = '发起 Hangout',
}: FloatingActionButtonProps) {
  return (
    <div className="pointer-events-none fixed bottom-6 right-6 z-30 flex flex-col items-end gap-3 lg:bottom-8 lg:right-8">
      <Link
        className="pointer-events-auto inline-flex items-center gap-2 rounded-full border border-white/10 bg-black/70 px-4 py-2 text-sm font-bold text-ink shadow-card backdrop-blur-xl transition-colors duration-200 hover:border-brand/50 hover:text-brand"
        to="/schedules/new"
      >
        <svg
          aria-hidden="true"
          className="h-4 w-4"
          fill="none"
          stroke="currentColor"
          strokeWidth={2.4}
          viewBox="0 0 24 24"
        >
          <rect height="16" rx="3" width="18" x="3" y="5" />
          <path d="M3 10h18M8 3v4M16 3v4" strokeLinecap="round" />
        </svg>
        新建日程
      </Link>
      <Link
        aria-label={label}
        className="pointer-events-auto inline-flex items-center gap-2 rounded-full bg-gradient-to-br from-tangerine via-rose to-brand px-5 py-3 font-bold text-white shadow-pop transition-transform duration-200 hover:scale-105"
        to={to}
      >
        <svg
          aria-hidden="true"
          className="h-5 w-5"
          fill="none"
          stroke="currentColor"
          strokeWidth={2.5}
          viewBox="0 0 24 24"
        >
          <path
            d="M12 5v14M5 12h14"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
        <span>{label}</span>
      </Link>
    </div>
  )
}
