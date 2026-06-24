import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  addMonths,
  eachDayOfInterval,
  endOfMonth,
  endOfWeek,
  format,
  isSameDay,
  isSameMonth,
  startOfMonth,
  startOfWeek,
} from 'date-fns'
import { getCalendar } from './lib/calendar'
import type { CalendarFilter, CalendarItem } from './lib/calendar'

const weekdayLabels = ['日', '一', '二', '三', '四', '五', '六']

const filterOptions: Array<{ value: CalendarFilter; label: string }> = [
  { value: 'all', label: '全部' },
  { value: 'groups', label: '我的群' },
  { value: 'friends', label: '好友' },
]

const colorDotClass: Record<string, string> = {
  blue: 'bg-sky-400',
  green: 'bg-emerald-400',
  gray: 'bg-neutral-400',
}

function formatTime(value: string): string {
  return new Intl.DateTimeFormat('zh-CN', { timeStyle: 'short' }).format(
    new Date(value),
  )
}

export default function CalendarPage() {
  const [month, setMonth] = useState(() => startOfMonth(new Date()))
  const [filter, setFilter] = useState<CalendarFilter>('all')
  const [items, setItems] = useState<CalendarItem[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [selectedDay, setSelectedDay] = useState<Date | null>(null)

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
  }, [gridStart, gridEnd, filter])

  const itemsByDay = useMemo(() => {
    const map = new Map<string, CalendarItem[]>()
    for (const item of items) {
      const key = format(new Date(item.start_at), 'yyyy-MM-dd')
      const list = map.get(key) ?? []
      list.push(item)
      map.set(key, list)
    }
    return map
  }, [items])

  const selectedItems = selectedDay
    ? (itemsByDay.get(format(selectedDay, 'yyyy-MM-dd')) ?? [])
    : []

  return (
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-2xl rounded-3xl bg-white/10 p-6 shadow-xl">
        <div className="flex items-center justify-between">
          <Link className="text-sm text-white/60" to="/">
            Hangout
          </Link>
          <div className="flex items-center gap-2 text-sm">
            <button
              className="rounded-full border border-white/20 px-3 py-1"
              onClick={() => setMonth((prev) => addMonths(prev, -1))}
              type="button"
            >
              上个月
            </button>
            <span className="min-w-24 text-center font-medium">
              {format(month, 'yyyy 年 M 月')}
            </span>
            <button
              className="rounded-full border border-white/20 px-3 py-1"
              onClick={() => setMonth((prev) => addMonths(prev, 1))}
              type="button"
            >
              下个月
            </button>
          </div>
        </div>

        <div className="mt-5 flex gap-2 text-sm">
          {filterOptions.map((option) => (
            <button
              className={`rounded-full px-3 py-1 ${
                filter === option.value
                  ? 'bg-white text-neutral-950'
                  : 'border border-white/20 text-white/80'
              }`}
              key={option.value}
              onClick={() => setFilter(option.value)}
              type="button"
            >
              {option.label}
            </button>
          ))}
        </div>

        <div className="mt-4 flex gap-4 text-xs text-white/60">
          <span className="flex items-center gap-1">
            <span className="h-2 w-2 rounded-full bg-sky-400" /> 群活动
          </span>
          <span className="flex items-center gap-1">
            <span className="h-2 w-2 rounded-full bg-emerald-400" /> 个人日程
          </span>
          <span className="flex items-center gap-1">
            <span className="h-2 w-2 rounded-full bg-neutral-400" /> 忙碌
          </span>
        </div>

        {error ? <p className="mt-4 text-sm text-red-300">{error}</p> : null}
        {loading ? (
          <p className="mt-4 text-sm text-white/60">正在加载...</p>
        ) : null}

        <div className="mt-5 grid grid-cols-7 gap-1 text-center text-xs text-white/40">
          {weekdayLabels.map((label) => (
            <div key={label}>{label}</div>
          ))}
        </div>

        <div className="mt-1 grid grid-cols-7 gap-1">
          {days.map((day) => {
            const key = format(day, 'yyyy-MM-dd')
            const dayItems = itemsByDay.get(key) ?? []
            const inMonth = isSameMonth(day, month)
            const isToday = isSameDay(day, new Date())
            const hasConflict = dayItems.some((item) => item.has_conflict)
            const colors = Array.from(
              new Set(dayItems.map((item) => item.color)),
            )
            return (
              <button
                className={`relative flex h-16 flex-col items-center rounded-2xl border p-1 text-sm ${
                  inMonth
                    ? 'border-white/10 text-white'
                    : 'border-transparent text-white/30'
                } ${isToday ? 'bg-white/10' : ''}`}
                key={key}
                onClick={() => setSelectedDay(day)}
                type="button"
              >
                <span>{format(day, 'd')}</span>
                {hasConflict ? (
                  <span className="absolute right-1 top-1 text-[10px] text-amber-300">
                    !
                  </span>
                ) : null}
                <span className="mt-auto flex gap-0.5 pb-1">
                  {colors.map((color) => (
                    <span
                      className={`h-1.5 w-1.5 rounded-full ${
                        colorDotClass[color] ?? 'bg-white'
                      }`}
                      key={color}
                    />
                  ))}
                </span>
              </button>
            )
          })}
        </div>
      </section>

      {selectedDay ? (
        <div className="fixed inset-x-0 bottom-0 mx-auto max-w-2xl rounded-t-3xl bg-neutral-900 p-6 shadow-2xl">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">
              {format(selectedDay, 'M 月 d 日')}
            </h2>
            <button
              className="text-sm text-white/60"
              onClick={() => setSelectedDay(null)}
              type="button"
            >
              关闭
            </button>
          </div>
          {selectedItems.length === 0 ? (
            <p className="mt-4 text-sm text-white/60">这天还没有安排。</p>
          ) : (
            <ul className="mt-4 space-y-2">
              {selectedItems.map((item) => (
                <li
                  className="rounded-2xl border border-white/10 p-3 text-sm"
                  key={`${item.kind}-${item.id}`}
                >
                  <div className="flex items-center gap-2">
                    <span
                      className={`h-2 w-2 rounded-full ${
                        colorDotClass[item.color] ?? 'bg-white'
                      }`}
                    />
                    <span className="font-medium">{item.title}</span>
                    {item.has_conflict ? (
                      <span className="rounded-full bg-amber-300/20 px-2 py-0.5 text-[10px] text-amber-200">
                        冲突
                      </span>
                    ) : null}
                  </div>
                  <p className="mt-1 text-white/60">
                    {formatTime(item.start_at)}
                    {item.end_at ? ` - ${formatTime(item.end_at)}` : ''}
                    {item.location ? ` · ${item.location}` : ''}
                  </p>
                </li>
              ))}
            </ul>
          )}
        </div>
      ) : null}

      <div className="fixed bottom-6 right-6 flex flex-col gap-3">
        <Link
          className="rounded-full bg-white px-5 py-3 text-sm font-medium text-neutral-950 shadow-lg"
          to="/schedules/new"
        >
          + 日程
        </Link>
        <Link
          className="rounded-full bg-sky-400 px-5 py-3 text-sm font-medium text-neutral-950 shadow-lg"
          to="/events/new"
        >
          + 活动
        </Link>
      </div>
    </main>
  )
}
