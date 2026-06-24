const typeLabels: Record<string, { label: string; className: string }> = {
  climbing: { label: '攀岩', className: 'bg-orange-400/20 text-orange-200' },
  dining: { label: '聚餐', className: 'bg-amber-400/20 text-amber-200' },
  travel: { label: '旅行', className: 'bg-sky-400/20 text-sky-200' },
  gaming: { label: '游戏', className: 'bg-violet-400/20 text-violet-200' },
  other: { label: '其他', className: 'bg-white/15 text-white/70' },
}

type EventTypeBadgeProps = {
  type: string
}

export function EventTypeBadge({ type }: EventTypeBadgeProps) {
  const meta = typeLabels[type] ?? typeLabels.other
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${meta.className}`}
    >
      {meta.label}
    </span>
  )
}
