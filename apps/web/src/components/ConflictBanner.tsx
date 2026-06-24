type ConflictItem = {
  id: number
  title: string
}

type ConflictBannerProps = {
  conflicts: ConflictItem[]
}

export function ConflictBanner({ conflicts }: ConflictBannerProps) {
  if (conflicts.length === 0) {
    return null
  }
  return (
    <div className="rounded-2xl border border-tangerine/30 bg-tangerine-soft p-3 text-sm text-tangerine">
      <p className="font-bold">与你已有 {conflicts.length} 个日程时间冲突</p>
      <ul className="mt-1 space-y-0.5 text-tangerine/80">
        {conflicts.map((conflict) => (
          <li key={conflict.id}>· {conflict.title}</li>
        ))}
      </ul>
    </div>
  )
}
