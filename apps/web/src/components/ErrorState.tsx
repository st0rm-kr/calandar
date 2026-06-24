type ErrorStateProps = {
  message: string
  onRetry?: () => void
}

export function ErrorState({ message, onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-3xl border border-red-400/30 bg-red-500/10 px-6 py-8 text-center">
      <p className="text-sm text-red-200">{message}</p>
      {onRetry ? (
        <button
          className="rounded-full border border-white/20 px-4 py-1.5 text-sm text-white/80"
          onClick={onRetry}
          type="button"
        >
          重试
        </button>
      ) : null}
    </div>
  )
}
