type ErrorStateProps = {
  message: string
  onRetry?: () => void
}

export function ErrorState({ message, onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-3xl border border-rose/30 bg-rose-soft px-6 py-8 text-center">
      <p className="text-sm font-semibold text-rose">{message}</p>
      {onRetry ? (
        <button
          className="cursor-pointer rounded-full bg-rose px-4 py-1.5 text-sm font-semibold text-white transition-colors duration-200 hover:bg-rose/90"
          onClick={onRetry}
          type="button"
        >
          重试
        </button>
      ) : null}
    </div>
  )
}
