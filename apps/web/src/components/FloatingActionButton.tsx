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
    <div className="pointer-events-none fixed bottom-6 right-6 z-30 lg:bottom-8 lg:right-8">
      <Link
        aria-label={label}
        className="pointer-events-auto flex h-14 w-14 items-center justify-center rounded-full bg-brand text-white shadow-pop transition-colors duration-200 hover:bg-brand-ink lg:h-16 lg:w-16"
        to={to}
      >
        <svg
          aria-hidden="true"
          className="h-8 w-8"
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
      </Link>
    </div>
  )
}
