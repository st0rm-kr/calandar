export type EventTypeMeta = {
  label: string
  badge: string
  block: string
  accent: string
}

export const eventTypeMeta: Record<string, EventTypeMeta> = {
  climbing: {
    label: '攀岩',
    badge: 'bg-tangerine-soft text-tangerine',
    block: 'bg-tangerine',
    accent: 'text-tangerine',
  },
  dining: {
    label: '聚餐',
    badge: 'bg-rose-soft text-rose',
    block: 'bg-rose',
    accent: 'text-rose',
  },
  travel: {
    label: '旅行',
    badge: 'bg-brand-soft text-brand',
    block: 'bg-brand',
    accent: 'text-brand',
  },
  gaming: {
    label: '游戏',
    badge: 'bg-grape-soft text-grape',
    block: 'bg-grape',
    accent: 'text-grape',
  },
  other: {
    label: '其他',
    badge: 'bg-grass-soft text-grass',
    block: 'bg-grass',
    accent: 'text-grass',
  },
}

export function metaForType(type: string): EventTypeMeta {
  return eventTypeMeta[type] ?? eventTypeMeta.other
}
