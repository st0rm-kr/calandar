import { useEffect, useMemo, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  addMonths,
  eachDayOfInterval,
  endOfMonth,
  endOfWeek,
  format,
  isSameDay,
  isSameMonth,
  isToday,
  isTomorrow,
  startOfMonth,
  startOfWeek,
} from 'date-fns'
import { getCalendar } from './lib/calendar'
import type { CalendarFilter, CalendarItem } from './lib/calendar'
import { rsvpEvent } from './lib/events'
import { EmptyState } from './components/EmptyState'
import { ErrorState } from './components/ErrorState'
import { LoadingState } from './components/LoadingState'

const weekdayLabels = ['日', '一', '二', '三', '四', '五', '六']

const filterOptions: Array<{ value: CalendarFilter; label: string }> = [
  { value: 'all', label: '全部' },
  { value: 'groups', label: '我的群' },
  { value: 'friends', label: '好友' },
]

const colorBlock: Record<string, string> = {
  blue: 'bg-brand',
  green: 'bg-tangerine',
  gray: 'bg-rose',
}

const colorSoft: Record<string, string> = {
  blue: 'bg-brand-soft text-brand',
  green: 'bg-tangerine-soft text-tangerine',
  gray: 'bg-rose-soft text-rose',
}

const colorHero: Record<string, string> = {
  blue: 'bg-gradient-to-br from-brand via-rose to-tangerine',
  green: 'bg-gradient-to-br from-tangerine via-rose to-grass',
  gray: 'bg-gradient-to-br from-rose via-grape to-brand',
}

function dayKey(date: Date): string {
  return format(date, 'yyyy-MM-dd')
}

function relativeTime(value: string): string {
  const date = new Date(value)
  const hm = format(date, 'HH:mm')
  if (isToday(date)) {
    return `${date.getHours() >= 18 ? '今晚' : '今天'} ${hm}`
  }
  if (isTomorrow(date)) {
    return `明天 ${hm}`
  }
  const diffDays = Math.round(
    (date.getTime() - new Date().setHours(0, 0, 0, 0)) / 86_400_000,
  )
  if (diffDays > 0 && diffDays < 7) {
    return `周${weekdayLabels[date.getDay()]} ${hm}`
  }
  return `${format(date, 'M月d日')} ${hm}`
}

function dayHeading(date: Date): string {
  const base = `${format(date, 'M月d日')} 周${weekdayLabels[date.getDay()]}`
  if (isToday(date)) {
    return `${base} · 今天`
  }
  if (isTomorrow(date)) {
    return `${base} · 明天`
  }
  return base
}

function prefersReducedMotion(): boolean {
  return (
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}

export default function CalendarPage() {
  const [month, setMonth] = useState(() => startOfMonth(new Date()))
  const [selectedDay, setSelectedDay] = useState(() => new Date())
  const [filter, setFilter] = useState<CalendarFilter>('all')
  const [items, setItems] = useState<CalendarItem[]>([])
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [loading, setLoading] = useState(true)
  const [nonce, setNonce] = useState(0)
  const [submitting, setSubmitting] = useState<Set<number>>(new Set())
  const [flashKey, setFlashKey] = useState<string | null>(null)

  const timelineRefs = useRef(new Map<string, HTMLButtonElement>())

  const gridStart = useMemo(
    () => startOfWeek(startOfMonth(month), { weekStartsOn: 0 }),
    [month],
  )
  const gridEnd = useMemo(
    () => endOfWeek(endOfMonth(month), { weekStartsOn: 0 }),
    [month],
  )
  const days = useMemo(
    () => eachDayOfInterval({ start: gridStart, end: gridEnd }),
    [gridStart, gridEnd],
  )

  useEffect(() => {
    let active = true
    Promise.resolve().then(() => {
      if (active) {
        setLoading(true)
      }
    })
    getCalendar(gridStart.toISOString(), gridEnd.toISOString(), filter)
      .then((data) => {
        if (!active) {
          return
        }
        setItems(data)
        setError('')
      })
      .catch((err: unknown) => {
        if (!active) {
          return
        }
        setItems([])
        setError(err instanceof Error ? err.message : '无法加载日历')
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
  }, [gridStart, gridEnd, filter, nonce])

  const itemsByDay = useMemo(() => {
    const map = new Map<string, CalendarItem[]>()
    for (const item of items) {
      const key = dayKey(new Date(item.start_at))
      const list = map.get(key) ?? []
      list.push(item)
      map.set(key, list)
    }
    for (const list of map.values()) {
      list.sort(
        (a, b) =>
          new Date(a.start_at).getTime() - new Date(b.start_at).getTime(),
      )
    }
    return map
  }, [items])

  const timelineDays = useMemo(
    () =>
      days
        .filter((day) => isSameMonth(day, month))
        .map((day) => {
          const key = dayKey(day)
          return {
            key,
            date: day,
            items: itemsByDay.get(key) ?? [],
          }
        }),
    [days, itemsByDay, month],
  )

  const selectedKey = dayKey(selectedDay)
  const selectedItems = itemsByDay.get(selectedKey) ?? []

  function scrollTimelineToDay(key: string) {
    const node = timelineRefs.current.get(key)
    if (!node) {
      return
    }
    node.scrollIntoView({
      behavior: prefersReducedMotion() ? 'auto' : 'smooth',
      block: 'nearest',
      inline: 'center',
    })
  }

  function handleSelectDay(day: Date) {
    const key = dayKey(day)
    setSelectedDay(day)
    setFlashKey(key)
    window.setTimeout(() => setFlashKey(null), 1200)
    window.requestAnimationFrame(() => scrollTimelineToDay(key))
  }

  function shiftMonth(direction: 1 | -1) {
    setMonth((prev) => {
      const next = addMonths(prev, direction)
      setSelectedDay(next)
      window.requestAnimationFrame(() => scrollTimelineToDay(dayKey(next)))
      return next
    })
  }

  async function quickRSVP(item: CalendarItem, going: boolean) {
    if (!item.event_id) {
      return
    }
    const eventID = item.event_id
    setSubmitting((prev) => new Set(prev).add(eventID))
    try {
      await rsvpEvent(eventID, {
        rsvp: going ? 'going' : 'not_going',
        add_to_calendar: going,
        visibility: 'busy_only',
      })
      setError('')
      setNotice(going ? '已加入活动，已同步到你的日历' : '已标记为不参加')
      setNonce((value) => value + 1)
    } catch (err) {
      setNotice('')
      setError(err instanceof Error ? err.message : 'RSVP 失败')
    } finally {
      setSubmitting((prev) => {
        const next = new Set(prev)
        next.delete(eventID)
        return next
      })
    }
  }

  return (
    <div className="space-y-6">
      <section className="min-h-[58vh] rounded-4xl border border-white/10 bg-surface/75 p-5 shadow-card backdrop-blur-xl lg:p-8">
        <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
          <div>
            <p className="text-sm font-bold uppercase tracking-[0.2em] text-brand">
              Calendar
            </p>
            <h1 className="neon-text mt-1 text-3xl font-bold tracking-tight lg:text-5xl">
              {format(month, 'yyyy 年 M 月')}
            </h1>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {filterOptions.map((option) => (
              <button
                className={`cursor-pointer rounded-full px-4 py-2 text-sm font-bold transition-colors duration-200 ${
                  filter === option.value
                    ? 'bg-gradient-to-r from-brand to-rose text-white shadow-[0_0_24px_rgba(0,242,234,0.25)]'
                    : 'bg-white/5 text-muted hover:bg-white/10 hover:text-ink'
                }`}
                key={option.value}
                onClick={() => setFilter(option.value)}
                type="button"
              >
                {option.label}
              </button>
            ))}
            <span className="mx-1 h-8 w-px bg-hairline" />
            <button
              aria-label="上个月"
              className="flex h-10 w-10 cursor-pointer items-center justify-center rounded-full bg-white/5 text-muted transition-colors duration-200 hover:bg-white/10 hover:text-ink"
              onClick={() => shiftMonth(-1)}
              type="button"
            >
              <svg
                className="h-5 w-5"
                fill="none"
                stroke="currentColor"
                strokeWidth={2}
                viewBox="0 0 24 24"
              >
                <path
                  d="M15 6l-6 6 6 6"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </button>
            <button
              aria-label="下个月"
              className="flex h-10 w-10 cursor-pointer items-center justify-center rounded-full bg-white/5 text-muted transition-colors duration-200 hover:bg-white/10 hover:text-ink"
              onClick={() => shiftMonth(1)}
              type="button"
            >
              <svg
                className="h-5 w-5"
                fill="none"
                stroke="currentColor"
                strokeWidth={2}
                viewBox="0 0 24 24"
              >
                <path
                  d="M9 6l6 6-6 6"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </button>
          </div>
        </div>

        {error ? (
          <div className="mt-5">
            <ErrorState message={error} onRetry={() => setNonce((v) => v + 1)} />
          </div>
        ) : null}
        {notice ? (
          <p className="mt-5 rounded-2xl border border-grass/30 bg-grass-soft px-4 py-3 text-sm font-bold text-grass">
            {notice}
          </p>
        ) : null}

        <div className="mt-6 grid grid-cols-7 gap-2 text-center text-xs font-bold uppercase tracking-[0.12em] text-muted lg:gap-3">
          {weekdayLabels.map((label) => (
            <div key={label}>{label}</div>
          ))}
        </div>

        <div className="mt-2 grid grid-cols-7 gap-2 lg:gap-3">
          {days.map((day) => {
            const key = dayKey(day)
            const dayItems = itemsByDay.get(key) ?? []
            const inMonth = isSameMonth(day, month)
            const selected = isSameDay(day, selectedDay)
            const today = isToday(day)
            const colors = Array.from(
              new Set(dayItems.map((item) => item.color)),
            )
            return (
              <button
                className={`min-h-20 cursor-pointer rounded-3xl border p-3 text-left transition-all duration-200 lg:min-h-28 xl:min-h-32 ${
                  selected
                    ? 'border-transparent bg-gradient-to-br from-rose via-tangerine to-brand text-white shadow-pop'
                    : inMonth
                      ? 'border-white/10 bg-white/[0.045] hover:-translate-y-0.5 hover:border-brand/40 hover:bg-white/10 hover:shadow-card'
                      : 'border-transparent bg-transparent text-muted/40'
                }`}
                key={key}
                onClick={() => handleSelectDay(day)}
                type="button"
              >
                <div className="flex items-start justify-between gap-2">
                  <span
                    className={`font-display text-lg font-bold lg:text-2xl ${
                      today && !selected ? 'text-brand' : ''
                    }`}
                  >
                    {format(day, 'd')}
                  </span>
                  {dayItems.length > 0 ? (
                    <span
                      className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${
                        selected
                          ? 'bg-white/20 text-white'
                          : 'bg-white/10 text-muted'
                      }`}
                    >
                      {dayItems.length}
                    </span>
                  ) : null}
                </div>
                {dayItems.length > 0 ? (
                  <div className="mt-5 flex flex-wrap gap-1.5">
                    {colors.map((color) => (
                      <span
                        className={`h-2.5 w-8 rounded-full ${
                          colorBlock[color] ?? 'bg-brand'
                        } ${selected ? 'ring-2 ring-white/70' : ''}`}
                        key={color}
                      />
                    ))}
                  </div>
                ) : null}
              </button>
            )
          })}
        </div>
      </section>

      <section className="space-y-4">
        <div className="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-sm font-bold uppercase tracking-[0.2em] text-rose">
              Feed
            </p>
            <h2 className="neon-text text-2xl font-bold tracking-tight">
              时间流
            </h2>
          </div>
          <p className="text-sm font-semibold text-muted">
            点击日期后，活动会从对应时间点向下展开。
          </p>
        </div>

        {loading ? <LoadingState /> : null}

        <div
          className={`overflow-hidden rounded-4xl border border-white/10 bg-surface/80 shadow-card backdrop-blur-xl transition-shadow duration-500 ${
            flashKey === selectedKey
              ? 'ring-2 ring-brand ring-offset-4 ring-offset-canvas'
              : ''
          }`}
        >
          <div className="overflow-x-auto">
            <div className="relative flex min-w-max items-start gap-3 px-8 pb-4 pt-8">
              <div className="absolute left-8 right-8 top-[4.35rem] h-1 rounded-full bg-gradient-to-r from-rose via-tangerine to-brand opacity-70" />
              {timelineDays.map((entry) => {
                const selected = entry.key === selectedKey
                const hasItems = entry.items.length > 0
                const colors = Array.from(
                  new Set(entry.items.map((item) => item.color)),
                )
                return (
                  <button
                    className="relative z-10 flex w-28 shrink-0 cursor-pointer flex-col items-center gap-3 text-center"
                    key={entry.key}
                    onClick={() => handleSelectDay(entry.date)}
                    ref={(node) => {
                      if (node) {
                        timelineRefs.current.set(entry.key, node)
                      } else {
                        timelineRefs.current.delete(entry.key)
                      }
                    }}
                    type="button"
                  >
                    <span
                      className={`text-xs font-bold ${
                        selected ? 'text-tangerine' : 'text-muted'
                      }`}
                    >
                      {format(entry.date, 'M/d')}
                    </span>
                    <span
                      className={`flex h-11 w-11 items-center justify-center rounded-full border-4 transition-all duration-200 ${
                        selected
                          ? 'scale-110 border-tangerine bg-gradient-to-br from-rose via-tangerine to-brand text-white shadow-pop'
                          : hasItems
                            ? 'border-white/20 bg-black shadow-[0_0_18px_rgba(0,242,234,0.2)]'
                            : 'border-white/10 bg-white/10'
                      }`}
                    >
                      {hasItems ? (
                        <span className="flex -space-x-1">
                          {colors.slice(0, 3).map((color) => (
                            <span
                              className={`h-3 w-3 rounded-full ${
                                selected
                                  ? 'bg-white'
                                  : colorBlock[color] ?? 'bg-brand'
                              }`}
                              key={color}
                            />
                          ))}
                        </span>
                      ) : null}
                    </span>
                    <span
                      className={`text-xs font-bold ${
                        hasItems ? 'text-ink' : 'text-muted/60'
                      }`}
                    >
                      周{weekdayLabels[entry.date.getDay()]}
                    </span>
                    {selected ? (
                      <span className="h-8 w-px bg-gradient-to-b from-tangerine to-transparent" />
                    ) : (
                      <span className="h-8" />
                    )}
                  </button>
                )
              })}
            </div>
          </div>

          <div className="border-t border-white/10 bg-black/25 px-5 pb-5 pt-4 lg:px-6 lg:pb-6">
            <div
              className="timeline-reveal"
              key={selectedKey}
            >
              <div className="mb-5 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p className="text-sm font-bold uppercase tracking-[0.2em] text-tangerine">
                    Node detail
                  </p>
                  <h3 className="text-2xl font-bold tracking-tight">
                    {dayHeading(selectedDay)}
                  </h3>
                </div>
                <span className="rounded-full bg-white/10 px-3 py-1 text-sm font-bold text-muted">
                  {selectedItems.length} 项
                </span>
              </div>

              {selectedItems.length === 0 ? (
                <EmptyState
                  title="这一天还没有安排"
                  description="可以从右下角发起一个新的 Hangout。"
                  action={
                    <Link
                      className="rounded-full bg-gradient-to-r from-tangerine via-rose to-brand px-5 py-2.5 text-sm font-bold text-white shadow-pop transition-transform duration-200 hover:scale-105"
                      to="/events/new"
                    >
                      发起 Hangout
                    </Link>
                  }
                />
              ) : (
                <div className="grid gap-4 xl:grid-cols-2">
                  {selectedItems.map((item) =>
                    item.kind === 'event' ? (
                      <HangoutCard
                        item={item}
                        key={`event-${item.id}`}
                        onRSVP={quickRSVP}
                        submitting={
                          item.event_id ? submitting.has(item.event_id) : false
                        }
                      />
                    ) : (
                      <ScheduleCard item={item} key={`schedule-${item.id}`} />
                    ),
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      </section>
    </div>
  )
}

type HangoutCardProps = {
  item: CalendarItem
  submitting: boolean
  onRSVP: (item: CalendarItem, going: boolean) => void
}

function HangoutCard({ item, submitting, onRSVP }: HangoutCardProps) {
  const block = colorHero[item.color] ?? colorHero.blue
  const soft = colorSoft[item.color] ?? colorSoft.blue
  return (
    <article className="overflow-hidden rounded-3xl border border-white/10 bg-black/45 shadow-card backdrop-blur">
      <div className={`relative h-40 ${block} p-5 text-white`}>
        <div className="absolute right-5 top-5 h-16 w-16 rounded-3xl bg-white/20 blur-[1px]" />
        <div className="absolute bottom-5 right-16 h-10 w-24 rounded-full bg-white/15" />
        <div className="absolute -bottom-8 -left-6 h-24 w-24 rounded-full bg-black/20 blur-xl" />
        <p className="text-xs font-bold uppercase tracking-[0.2em] text-white/70">
          Hangout
        </p>
        <h3 className="relative mt-8 max-w-sm font-display text-2xl font-bold leading-tight">
          {item.title}
        </h3>
      </div>
      <div className="p-5">
        <p className="text-sm font-bold text-muted">
          {relativeTime(item.start_at)}
          {item.location ? ` · ${item.location}` : ''}
        </p>
        {item.has_conflict ? (
          <span className="mt-3 inline-flex rounded-full bg-tangerine-soft px-3 py-1 text-xs font-bold text-tangerine">
            时间冲突
          </span>
        ) : null}
        <div className="mt-5 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <p className="rounded-full bg-white/10 px-3 py-2 text-xs font-bold text-muted">
            真实报名人数在活动详情中展示
          </p>
          <div className="flex items-center gap-2">
            <button
              className="inline-flex cursor-pointer items-center gap-2 rounded-full bg-gradient-to-r from-tangerine via-rose to-brand px-4 py-2 text-sm font-bold text-white shadow-pop transition-transform duration-200 hover:scale-105 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={submitting}
              onClick={() => onRSVP(item, true)}
              type="button"
            >
              <svg
                aria-hidden="true"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth={2.4}
                viewBox="0 0 24 24"
              >
                <path
                  d="M12 5v14M5 12h14"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
              加入
            </button>
            <button
              className={`cursor-pointer rounded-full px-4 py-2 text-sm font-bold transition-colors duration-200 disabled:cursor-not-allowed disabled:opacity-50 ${soft}`}
              disabled={submitting}
              onClick={() => onRSVP(item, false)}
              type="button"
            >
              下次一定
            </button>
          </div>
        </div>
      </div>
    </article>
  )
}

function ScheduleCard({ item }: { item: CalendarItem }) {
  return (
    <article className="overflow-hidden rounded-3xl border border-white/10 bg-black/45 shadow-card backdrop-blur">
      <div className="h-3 bg-gradient-to-r from-tangerine via-grass to-brand" />
      <div className="p-5">
        <p className="text-xs font-bold uppercase tracking-[0.2em] text-muted">
          Schedule
        </p>
        <h3 className="mt-3 text-xl font-bold">{item.title}</h3>
        <p className="mt-2 text-sm font-bold text-muted">
          {relativeTime(item.start_at)}
          {item.location ? ` · ${item.location}` : ''}
        </p>
      </div>
    </article>
  )
}
