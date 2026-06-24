import type { ReactNode } from 'react'

type EmptyStateProps = {
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center gap-2 rounded-3xl border border-dashed border-white/15 px-6 py-10 text-center">
      <p className="text-base font-medium text-white/80">{title}</p>
      {description ? (
        <p className="text-sm text-white/50">{description}</p>
      ) : null}
      {action ? <div className="mt-2">{action}</div> : null}
    </div>
  )
}
