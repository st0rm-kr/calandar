import { metaForType } from '../lib/eventType'

type EventTypeBadgeProps = {
  type: string
}

export function EventTypeBadge({ type }: EventTypeBadgeProps) {
  const meta = metaForType(type)
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-bold ${meta.badge}`}
    >
      {meta.label}
    </span>
  )
}
