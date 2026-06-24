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
    <div className="rounded-2xl border border-amber-300/40 bg-amber-400/10 p-3 text-sm text-amber-100">
      <p className="font-medium">
        与你已有 {conflicts.length} 个日程时间冲突
      </p>
      <ul className="mt-1 space-y-0.5 text-amber-100/80">
        {conflicts.map((conflict) => (
          <li key={conflict.id}>· {conflict.title}</li>
        ))}
      </ul>
    </div>
  )
}
