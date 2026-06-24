type LoadingStateProps = {
  label?: string
}

export function LoadingState({ label = '加载中…' }: LoadingStateProps) {
  return (
    <div className="flex items-center justify-center gap-2 py-12 text-sm font-semibold text-muted">
      <span className="h-2.5 w-2.5 animate-pulse rounded-full bg-brand" />
      {label}
    </div>
  )
}
