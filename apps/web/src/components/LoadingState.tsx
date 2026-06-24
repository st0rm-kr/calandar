type LoadingStateProps = {
  label?: string
}

export function LoadingState({ label = '正在加载...' }: LoadingStateProps) {
  return (
    <div className="flex items-center justify-center gap-2 py-10 text-sm text-white/50">
      <span className="h-2 w-2 animate-pulse rounded-full bg-white/60" />
      {label}
    </div>
  )
}
