import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import {
  addDays,
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
  startOfDay,
  startOfWeek,
} from 'date-fns'
import type { UIEvent } from 'react'
import { getCalendar, listCalendarSubscriptions } from './lib/calendar'
import type {
  CalendarItem,
  CalendarSourceType,
  CalendarSubscription,
} from './lib/calendar'
import { bindExclusiveHorizontalWheel } from './lib/exclusiveWheel'
import { inviteToEvent, rsvpEvent } from './lib/events'
import { listFriends } from './lib/friends'
import type { Friend } from './lib/friends'
import { deleteSchedule } from './lib/schedules'
import { AvatarStack } from './components/Avatar'
import type { AvatarPerson } from './components/Avatar'
import { EmptyState } from './components/EmptyState'
import { ErrorState } from './components/ErrorState'
import { FriendInvitePicker } from './components/FriendInvitePicker'
import { LoadingState } from './components/LoadingState'

const weekdayLabels = ['日', '一', '二', '三', '四', '五', '六']

type CalendarView = 'month' | 'week' | 'list' | 'availability'

const calendarViewOptions: Array<{ value: CalendarView; label: string }> = [
  { value: 'month', label: '月视图' },
  { value: 'week', label: '周视图' },
  { value: 'list', label: '列表' },
  { value: 'availability', label: '空闲' },
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

const sourceTypeText: Record<CalendarSourceType, string> = {
  self_schedules: '个人日程',
  self_events: '个人活动',
  friend_schedules: '好友日程',
  group_events: '群组活动',
}

const defaultSubscriptions: CalendarSubscription[] = [
  {
    id: 'self:schedules',
    type: 'self_schedules',
    label: '自身日程',
    color: 'green',
    enabled: true,
  },
  {
    id: 'self:events',
    type: 'self_events',
    label: '自身活动',
    color: 'blue',
    enabled: true,
  },
]

const dayStartHour = 7
const dayEndHour = 24
const defaultAvailabilityStart = 18 * 60
const defaultAvailabilityEnd = 23 * 60

function parseCalendarView(value: string | null): CalendarView {
  switch (value) {
    case 'week':
    case 'list':
    case 'availability':
      return value
    default:
      return 'month'
  }
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

function timelineItemKey(item: CalendarItem): string {
  return `${item.kind}-${item.id}`
}

function timeOfDay(value: string): string {
  return format(new Date(value), 'HH:mm')
}

function itemEndDate(item: CalendarItem): Date {
  if (item.end_at) {
    return new Date(item.end_at)
  }
  return new Date(new Date(item.start_at).getTime() + 60 * 60 * 1000)
}

function minutesOfDay(date: Date): number {
  return date.getHours() * 60 + date.getMinutes()
}

function minutesLabel(minutes: number): string {
  const hour = Math.floor(minutes / 60)
  const minute = minutes % 60
  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

function prefersReducedMotion(): boolean {
  return (
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}

export default function CalendarPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const calendarView = parseCalendarView(searchParams.get('view'))
  const [month, setMonth] = useState(() => startOfMonth(new Date()))
  const [selectedDay, setSelectedDay] = useState(() => new Date())
  const [timelineStart, setTimelineStart] = useState(() =>
    startOfDay(addDays(new Date(), -30)),
  )
  const [timelineEnd, setTimelineEnd] = useState(() =>
    startOfDay(addDays(new Date(), 60)),
  )
  const [subscriptions, setSubscriptions] = useState<CalendarSubscription[]>(
    defaultSubscriptions,
  )
  const [subscriptionsLoaded, setSubscriptionsLoaded] = useState(true)
  const [subscriptionError, setSubscriptionError] = useState('')
  const [selectedSourceIDs, setSelectedSourceIDs] = useState<Set<string>>(
    () =>
      new Set(
        defaultSubscriptions
          .filter((source) => source.enabled)
          .map((source) => source.id),
      ),
  )
  const [items, setItems] = useState<CalendarItem[]>([])
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [loading, setLoading] = useState(true)
  const [nonce, setNonce] = useState(0)
  const [submitting, setSubmitting] = useState<Set<number>>(new Set())
  const [deletingSchedules, setDeletingSchedules] = useState<Set<number>>(
    () => new Set(),
  )
  const [cardMessages, setCardMessages] = useState<Record<number, string>>({})
  const [expandedDismissed, setExpandedDismissed] = useState<Set<number>>(
    () => new Set(),
  )
  const [friends, setFriends] = useState<Friend[]>([])
  const [friendsLoaded, setFriendsLoaded] = useState(false)
  const [friendsError, setFriendsError] = useState('')
  const [invitePanelEventID, setInvitePanelEventID] = useState<number | null>(
    null,
  )
  const [inviteSelections, setInviteSelections] = useState<
    Record<number, string[]>
  >({})
  const [inviteErrors, setInviteErrors] = useState<Record<number, string>>({})
  const [inviteSubmitting, setInviteSubmitting] = useState<Set<number>>(
    () => new Set(),
  )
  const [flashKey, setFlashKey] = useState<string | null>(null)
  const [expandedItemKey, setExpandedItemKey] = useState<string | null>(null)

  const timelineRefs = useRef(new Map<string, HTMLButtonElement>())
  const timelineScrollerRef = useRef<HTMLDivElement | null>(null)

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

  const fetchStart = useMemo(
    () =>
      gridStart.getTime() < timelineStart.getTime() ? gridStart : timelineStart,
    [gridStart, timelineStart],
  )
  const fetchEnd = useMemo(() => {
    const timelineFetchEnd = addDays(timelineEnd, 1)
    return gridEnd.getTime() > timelineFetchEnd.getTime()
      ? gridEnd
      : timelineFetchEnd
  }, [gridEnd, timelineEnd])
  const selectedSourceIDsKey = useMemo(
    () =>
      subscriptions
        .filter((source) => selectedSourceIDs.has(source.id))
        .map((source) => source.id)
        .join(','),
    [selectedSourceIDs, subscriptions],
  )

  useEffect(() => {
    let active = true
    if (!subscriptionsLoaded) {
      return () => {
        active = false
      }
    }
    const sources =
      selectedSourceIDsKey.length > 0 ? selectedSourceIDsKey.split(',') : []
    Promise.resolve().then(() => {
      if (active) {
        setLoading(true)
      }
    })
    getCalendar(fetchStart.toISOString(), fetchEnd.toISOString(), 'all', sources)
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
  }, [fetchStart, fetchEnd, nonce, selectedSourceIDsKey, subscriptionsLoaded])

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
      eachDayOfInterval({ start: timelineStart, end: timelineEnd }).map((day) => {
        const key = dayKey(day)
        return {
          key,
          date: day,
          items: itemsByDay.get(key) ?? [],
        }
      }),
    [itemsByDay, timelineEnd, timelineStart],
  )

  const selectedKey = dayKey(selectedDay)
  const selectedItems = itemsByDay.get(selectedKey) ?? []
  const selectedWeekDays = useMemo(() => {
    const start = startOfWeek(selectedDay, { weekStartsOn: 0 })
    return eachDayOfInterval({ start, end: addDays(start, 6) })
  }, [selectedDay])

  useEffect(() => {
    let active = true
    listCalendarSubscriptions()
      .then((data) => {
        if (!active) {
          return
        }
        setSubscriptions(data)
        setSelectedSourceIDs(
          new Set(data.filter((source) => source.enabled).map((source) => source.id)),
        )
        setSubscriptionError('')
      })
      .catch((err: unknown) => {
        if (!active) {
          return
        }
        setSubscriptions([])
        setSelectedSourceIDs(new Set())
        setSubscriptionError(
          err instanceof Error ? err.message : '无法加载订阅列表',
        )
      })
      .finally(() => {
        if (active) {
          setSubscriptionsLoaded(true)
        }
      })
    return () => {
      active = false
    }
  }, [nonce])

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

  const extendTimeline = useCallback((direction: -1 | 1) => {
    if (direction < 0) {
      setTimelineStart((current) => startOfDay(addDays(current, -30)))
    } else {
      setTimelineEnd((current) => startOfDay(addDays(current, 30)))
    }
  }, [])

  const handleTimelineBoundary = useCallback(
    (node: HTMLElement) => {
    if (node.scrollLeft < 48) {
      extendTimeline(-1)
    }
    if (node.scrollWidth - node.clientWidth - node.scrollLeft < 48) {
      extendTimeline(1)
    }
    },
    [extendTimeline],
  )

  const handleTimelineScroll = useCallback(
    (event: UIEvent<HTMLDivElement>) => handleTimelineBoundary(event.currentTarget),
    [handleTimelineBoundary],
  )

  useEffect(() => {
    const node = timelineScrollerRef.current
    if (!node) {
      return
    }
    return bindExclusiveHorizontalWheel(node, handleTimelineBoundary)
  }, [handleTimelineBoundary])

  function ensureTimelineIncludes(day: Date) {
    const target = startOfDay(day)
    if (target.getTime() < timelineStart.getTime()) {
      setTimelineStart(startOfDay(addDays(target, -30)))
    }
    if (target.getTime() > timelineEnd.getTime()) {
      setTimelineEnd(startOfDay(addDays(target, 60)))
    }
  }

  function handleSelectDay(day: Date) {
    const key = dayKey(day)
    setSelectedDay(day)
    setExpandedItemKey(null)
    ensureTimelineIncludes(day)
    setFlashKey(key)
    window.setTimeout(() => setFlashKey(null), 1200)
    window.requestAnimationFrame(() => scrollTimelineToDay(key))
  }

  function changeCalendarView(nextView: CalendarView) {
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      if (nextView === 'month') {
        next.delete('view')
      } else {
        next.set('view', nextView)
      }
      return next
    })
  }

  function focusItem(item: CalendarItem) {
    const date = new Date(item.start_at)
    handleSelectDay(date)
    setExpandedItemKey(timelineItemKey(item))
  }

  function shiftMonth(direction: 1 | -1) {
    setMonth((prev) => {
      const next = addMonths(prev, direction)
      setSelectedDay(next)
      setExpandedItemKey(null)
      setTimelineStart(startOfDay(addDays(next, -30)))
      setTimelineEnd(startOfDay(addDays(next, 60)))
      window.requestAnimationFrame(() => scrollTimelineToDay(dayKey(next)))
      return next
    })
  }

  function shiftCalendarWindow(direction: 1 | -1) {
    if (calendarView === 'month') {
      shiftMonth(direction)
      return
    }
    const daysToShift = calendarView === 'list' ? 14 : 7
    const next = addDays(selectedDay, direction * daysToShift)
    setSelectedDay(next)
    setExpandedItemKey(null)
    setMonth(startOfMonth(next))
    ensureTimelineIncludes(next)
    window.requestAnimationFrame(() => scrollTimelineToDay(dayKey(next)))
  }

  async function quickRSVP(item: CalendarItem, going: boolean) {
    if (!item.event_id) {
      return
    }
    const eventID = item.event_id
    setSubmitting((prev) => new Set(prev).add(eventID))
    try {
      const result = await rsvpEvent(eventID, {
        rsvp: going ? 'going' : 'not_going',
        add_to_calendar: going,
        visibility: 'busy_only',
      })
      setItems((current) =>
        current.map((candidate) =>
          candidate.event_id === eventID
            ? {
                ...candidate,
                viewer_rsvp: result.participant.rsvp,
                going_count: result.going_count,
              }
            : candidate,
        ),
      )
      setCardMessages((current) => ({
        ...current,
        [eventID]: going
          ? '已加入活动，已同步到你的日历'
          : '已收纳，悬停卡片可重新加入',
      }))
      setExpandedDismissed((current) => {
        const next = new Set(current)
        if (going) {
          next.delete(eventID)
        } else {
          next.delete(eventID)
        }
        return next
      })
      setError('')
      setNotice('')
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

  function setDismissedCardExpanded(eventID: number, expanded: boolean) {
    setExpandedDismissed((current) => {
      const next = new Set(current)
      if (expanded) {
        next.add(eventID)
      } else {
        next.delete(eventID)
      }
      return next
    })
  }

  async function ensureFriendsLoaded() {
    if (friendsLoaded) {
      return
    }
    setFriendsError('')
    try {
      setFriends(await listFriends())
    } catch (err) {
      setFriends([])
      setFriendsError(err instanceof Error ? err.message : '无法加载好友')
    } finally {
      setFriendsLoaded(true)
    }
  }

  function handleInviteEntrance(item: CalendarItem) {
    if (!item.event_id) {
      return
    }
    setInvitePanelEventID(item.event_id)
    void ensureFriendsLoaded()
  }

  function toggleInvitee(eventID: number, friendID: string) {
    setInviteSelections((current) => {
      const selected = current[eventID] ?? []
      const next = selected.includes(friendID)
        ? selected.filter((id) => id !== friendID)
        : [...selected, friendID]
      return { ...current, [eventID]: next }
    })
  }

  async function sendInvites(eventID: number) {
    const inviteeIDs = inviteSelections[eventID] ?? []
    if (inviteeIDs.length === 0) {
      return
    }
    setInviteSubmitting((current) => new Set(current).add(eventID))
    setInviteErrors((current) => ({ ...current, [eventID]: '' }))
    try {
      const invited = await inviteToEvent(eventID, inviteeIDs)
      const count = invited.length || inviteeIDs.length
      setCardMessages((current) => ({
        ...current,
        [eventID]: `已邀请 ${count} 位好友`,
      }))
      setInviteSelections((current) => ({ ...current, [eventID]: [] }))
      setInvitePanelEventID(null)
    } catch (err) {
      setInviteErrors((current) => ({
        ...current,
        [eventID]: err instanceof Error ? err.message : '邀请失败',
      }))
    } finally {
      setInviteSubmitting((current) => {
        const next = new Set(current)
        next.delete(eventID)
        return next
      })
    }
  }

  async function deleteManualSchedule(item: CalendarItem) {
    if (item.kind !== 'schedule') {
      return
    }
    if (!window.confirm('确认取消这个日程吗？')) {
      return
    }
    setDeletingSchedules((current) => new Set(current).add(item.id))
    try {
      await deleteSchedule(item.id)
      setItems((current) =>
        current.filter(
          (candidate) =>
            !(candidate.kind === 'schedule' && candidate.id === item.id),
        ),
      )
      setExpandedItemKey(null)
      setNotice('日程已取消')
      setError('')
    } catch (err) {
      setNotice('')
      setError(err instanceof Error ? err.message : '取消日程失败')
    } finally {
      setDeletingSchedules((current) => {
        const next = new Set(current)
        next.delete(item.id)
        return next
      })
    }
  }

  function toggleSubscriptionSource(sourceID: string) {
    setSelectedSourceIDs((current) => {
      const next = new Set(current)
      if (next.has(sourceID)) {
        next.delete(sourceID)
      } else {
        next.add(sourceID)
      }
      return next
    })
  }

  return (
    <div className="grid gap-6 xl:grid-cols-[17rem_minmax(0,1fr)]">
      <SubscriptionSidebar
        error={subscriptionError}
        loaded={subscriptionsLoaded}
        onToggle={toggleSubscriptionSource}
        selectedIDs={selectedSourceIDs}
        subscriptions={subscriptions}
      />
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
            <CalendarViewSwitcher
              onChange={changeCalendarView}
              value={calendarView}
            />
            <span className="mx-1 h-8 w-px bg-hairline" />
            <button
              aria-label="上个月"
              className="flex h-10 w-10 cursor-pointer items-center justify-center rounded-full bg-white/5 text-muted transition-colors duration-200 hover:bg-white/10 hover:text-ink"
              onClick={() => shiftCalendarWindow(-1)}
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
              onClick={() => shiftCalendarWindow(1)}
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

        {calendarView === 'month' ? (
          <MonthView
            days={days}
            itemsByDay={itemsByDay}
            month={month}
            onSelectDay={handleSelectDay}
            selectedDay={selectedDay}
          />
        ) : null}
        {calendarView === 'week' ? (
          <WeekView
            days={selectedWeekDays}
            itemsByDay={itemsByDay}
            onFocusItem={focusItem}
            onSelectDay={handleSelectDay}
          />
        ) : null}
        {calendarView === 'list' ? (
          <AgendaView items={items} onFocusItem={focusItem} />
        ) : null}
        {calendarView === 'availability' ? (
          <AvailabilityView days={selectedWeekDays} itemsByDay={itemsByDay} />
        ) : null}
      </section>

      <section className="space-y-4">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p className="text-sm font-bold uppercase tracking-[0.2em] text-rose">
              Feed
            </p>
            <h2 className="neon-text text-2xl font-bold tracking-tight">
              时间流
            </h2>
            <p className="mt-1 text-sm font-semibold text-muted">
              横向滚动选择日期，选中日期会向下展开当天时间轴。
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Link
              className="rounded-full border border-white/10 bg-black/55 px-4 py-2 text-sm font-bold text-ink shadow-card backdrop-blur-xl transition-colors duration-200 hover:border-brand/50 hover:text-brand"
              to="/schedules/new"
            >
              新建日程
            </Link>
            <Link
              className="rounded-full bg-gradient-to-r from-tangerine via-rose to-brand px-4 py-2 text-sm font-bold text-white shadow-pop transition-transform duration-200 hover:scale-105"
              to="/events/new"
            >
              发起 Hangout
            </Link>
          </div>
        </div>

        {loading ? <LoadingState /> : null}

        <div
          className={`overflow-hidden rounded-4xl border border-white/10 bg-surface/80 shadow-card backdrop-blur-xl transition-shadow duration-500 ${
            flashKey === selectedKey
              ? 'ring-2 ring-brand ring-offset-4 ring-offset-canvas'
              : ''
          }`}
        >
            <div
              aria-label="无限日期时间轴"
              className="scrollbar-none overflow-x-auto"
              onScroll={handleTimelineScroll}
              ref={timelineScrollerRef}
            >
            <div className="relative flex min-w-max items-start gap-3 px-8 pb-4 pt-8">
              <div className="absolute left-8 right-8 top-[4.35rem] h-1 rounded-full bg-gradient-to-r from-rose via-tangerine to-brand opacity-70" />
                <button
                  aria-label="向前延展时间流"
                  className="relative z-10 mt-[2.15rem] flex h-11 w-11 shrink-0 cursor-pointer items-center justify-center rounded-full border border-white/10 bg-black/70 text-muted transition-colors duration-200 hover:border-brand/50 hover:text-brand"
                  onClick={() => extendTimeline(-1)}
                  type="button"
                >
                  <svg
                    aria-hidden="true"
                    className="h-5 w-5"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth={2.4}
                    viewBox="0 0 24 24"
                  >
                    <path
                      d="M15 6l-6 6 6 6"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </button>
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
                <button
                  aria-label="向后延展时间流"
                  className="relative z-10 mt-[2.15rem] flex h-11 w-11 shrink-0 cursor-pointer items-center justify-center rounded-full border border-white/10 bg-black/70 text-muted transition-colors duration-200 hover:border-brand/50 hover:text-brand"
                  onClick={() => extendTimeline(1)}
                  type="button"
                >
                  <svg
                    aria-hidden="true"
                    className="h-5 w-5"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth={2.4}
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
                    description="可以从上方操作区发起一个新的 Hangout。"
                  action={
                    <Link
                      className="rounded-full bg-gradient-to-r from-tangerine via-rose to-brand px-5 py-2.5 text-sm font-bold text-white shadow-pop transition-transform duration-200 hover:scale-105"
                      to="/events/new"
                    >
                        创建 Hangout
                    </Link>
                  }
                />
              ) : (
                  <div className="relative">
                    <div className="absolute bottom-0 left-[4.5rem] top-0 w-px bg-gradient-to-b from-tangerine via-brand to-transparent" />
                    <div className="space-y-4">
                      {selectedItems.map((item) => {
                        const key = timelineItemKey(item)
                        const expanded = expandedItemKey === key
                        return (
                          <div
                            className="grid grid-cols-[4.5rem_minmax(0,1fr)] gap-4"
                            key={key}
                          >
                            <div className="pt-4 text-right font-display text-sm font-bold text-tangerine">
                              {timeOfDay(item.start_at)}
                            </div>
                            <div className="relative pl-6">
                              <span className="absolute left-[-0.45rem] top-5 h-4 w-4 rounded-full border-2 border-tangerine bg-canvas shadow-[0_0_18px_rgba(255,184,0,0.35)]" />
                              {expanded ? (
                                item.kind === 'event' ? (
                                  <HangoutCard
                                    dismissedExpanded={
                                      item.event_id
                                        ? expandedDismissed.has(item.event_id)
                                        : false
                                    }
                                    friends={friends}
                                    friendsLoaded={friendsLoaded}
                                    inviteError={
                                      item.event_id
                                        ? inviteErrors[item.event_id] ||
                                          friendsError
                                        : friendsError
                                    }
                                    inviteOpen={
                                      item.event_id === invitePanelEventID
                                    }
                                    inviteSelectedIDs={
                                      item.event_id
                                        ? inviteSelections[item.event_id] ?? []
                                        : []
                                    }
                                    inviteSubmitting={
                                      item.event_id
                                        ? inviteSubmitting.has(item.event_id)
                                        : false
                                    }
                                    item={item}
                                    message={
                                      item.event_id
                                        ? cardMessages[item.event_id]
                                        : undefined
                                    }
                                    onDismissedHoverChange={
                                      setDismissedCardExpanded
                                    }
                                    onInvite={handleInviteEntrance}
                                    onInviteClose={() =>
                                      setInvitePanelEventID(null)
                                    }
                                    onInviteSubmit={sendInvites}
                                    onInviteeToggle={toggleInvitee}
                                    onRSVP={quickRSVP}
                                    submitting={
                                      item.event_id
                                        ? submitting.has(item.event_id)
                                        : false
                                    }
                                  />
                                ) : (
                                  <ScheduleCard
                                    deleting={deletingSchedules.has(item.id)}
                                    item={item}
                                    onDelete={deleteManualSchedule}
                                  />
                                )
                              ) : (
                                <TimelinePreview
                                  item={item}
                                  onExpand={() => setExpandedItemKey(key)}
                                />
                              )}
                            </div>
                          </div>
                        )
                      })}
                    </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </section>
      </div>
    </div>
  )
}

function CalendarViewSwitcher({
  value,
  onChange,
}: {
  value: CalendarView
  onChange: (view: CalendarView) => void
}) {
  return (
    <div
      aria-label="日历视图"
      className="flex flex-wrap items-center gap-2"
      role="group"
    >
      {calendarViewOptions.map((option) => {
        const selected = value === option.value
        return (
          <button
            aria-pressed={selected}
            className={`cursor-pointer rounded-full px-4 py-2 text-sm font-bold transition-colors duration-200 ${
              selected
                ? 'bg-gradient-to-r from-brand to-rose text-white shadow-[0_0_24px_rgba(0,242,234,0.25)]'
                : 'bg-white/5 text-muted hover:bg-white/10 hover:text-ink'
            }`}
            key={option.value}
            onClick={() => onChange(option.value)}
            type="button"
          >
            {option.label}
          </button>
        )
      })}
    </div>
  )
}

function MonthView({
  days,
  month,
  selectedDay,
  itemsByDay,
  onSelectDay,
}: {
  days: Date[]
  month: Date
  selectedDay: Date
  itemsByDay: Map<string, CalendarItem[]>
  onSelectDay: (day: Date) => void
}) {
  return (
    <>
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
          const colors = Array.from(new Set(dayItems.map((item) => item.color)))
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
              onClick={() => onSelectDay(day)}
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
                      selected ? 'bg-white/20 text-white' : 'bg-white/10 text-muted'
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
    </>
  )
}

function WeekView({
  days,
  itemsByDay,
  onSelectDay,
  onFocusItem,
}: {
  days: Date[]
  itemsByDay: Map<string, CalendarItem[]>
  onSelectDay: (day: Date) => void
  onFocusItem: (item: CalendarItem) => void
}) {
  const hours = Array.from(
    { length: dayEndHour - dayStartHour + 1 },
    (_, index) => dayStartHour + index,
  )

  return (
    <div className="mt-6 rounded-4xl border border-white/10 bg-black/25 p-4">
      <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.2em] text-muted">
            Week
          </p>
          <h2 className="text-xl font-bold">周视图时间表</h2>
        </div>
        <p className="text-sm font-semibold text-muted">
          7 天时间轴，点击日程会在下方 Feed 聚焦。
        </p>
      </div>

      <div className="mt-5 grid grid-cols-[4rem_repeat(7,minmax(7rem,1fr))] gap-2 overflow-x-auto pb-2">
        <div className="pt-12">
          {hours.map((hour) => (
            <div className="h-14 text-right text-xs font-bold text-muted" key={hour}>
              {String(hour).padStart(2, '0')}:00
            </div>
          ))}
        </div>
        {days.map((day) => {
          const key = dayKey(day)
          const dayItems = itemsByDay.get(key) ?? []
          return (
            <div className="min-w-28" key={key}>
              <button
                className="mb-3 w-full cursor-pointer rounded-2xl bg-white/5 px-3 py-2 text-left transition-colors duration-200 hover:bg-white/10"
                onClick={() => onSelectDay(day)}
                type="button"
              >
                <span className="block text-xs font-bold text-muted">
                  周{weekdayLabels[day.getDay()]}
                </span>
                <span className="block font-display text-lg font-bold">
                  {format(day, 'M/d')}
                </span>
              </button>
              <div className="space-y-2">
                {dayItems.length === 0 ? (
                  <p className="rounded-2xl border border-dashed border-white/10 px-3 py-4 text-xs font-semibold text-muted">
                    暂无安排
                  </p>
                ) : null}
                {dayItems.map((item) => (
                  <button
                    className="w-full cursor-pointer rounded-2xl border border-white/10 bg-white/[0.05] p-3 text-left transition-colors duration-200 hover:border-brand/40 hover:bg-white/10"
                    key={timelineItemKey(item)}
                    onClick={() => onFocusItem(item)}
                    type="button"
                  >
                    <span className="text-xs font-bold text-tangerine">
                      {timeOfDay(item.start_at)}
                    </span>
                    <span className="mt-1 block text-sm font-bold">{item.title}</span>
                    <span className="mt-1 block text-xs font-semibold text-muted">
                      {item.source_label}
                    </span>
                  </button>
                ))}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function AgendaView({
  items,
  onFocusItem,
}: {
  items: CalendarItem[]
  onFocusItem: (item: CalendarItem) => void
}) {
  const sortedItems = [...items].sort(
    (a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime(),
  )
  const groups = new Map<string, CalendarItem[]>()
  for (const item of sortedItems) {
    const key = dayKey(new Date(item.start_at))
    groups.set(key, [...(groups.get(key) ?? []), item])
  }

  return (
    <div className="mt-6 rounded-4xl border border-white/10 bg-black/25 p-4">
      <p className="text-xs font-bold uppercase tracking-[0.2em] text-muted">
        Agenda
      </p>
      <h2 className="text-xl font-bold">接下来 14 天</h2>
      <div className="mt-4 space-y-4">
        {sortedItems.length === 0 ? (
          <EmptyState title="接下来两周暂无安排" />
        ) : null}
        {Array.from(groups.entries()).map(([key, groupItems]) => (
          <section
            className="rounded-3xl border border-white/10 bg-white/[0.04] p-4"
            key={key}
          >
            <h3 className="font-bold">{dayHeading(new Date(`${key}T00:00:00`))}</h3>
            <div className="mt-3 space-y-2">
              {groupItems.map((item) => (
                <button
                  className="flex w-full cursor-pointer flex-wrap items-center justify-between gap-3 rounded-2xl bg-black/35 px-4 py-3 text-left transition-colors duration-200 hover:bg-black/55"
                  key={timelineItemKey(item)}
                  onClick={() => onFocusItem(item)}
                  type="button"
                >
                  <span>
                    <span className="block text-xs font-bold text-tangerine">
                      {timeOfDay(item.start_at)}
                    </span>
                    <span className="block text-sm font-bold">{item.title}</span>
                  </span>
                  <span className="rounded-full bg-white/10 px-3 py-1 text-xs font-bold text-muted">
                    {item.source_label}
                  </span>
                </button>
              ))}
            </div>
          </section>
        ))}
      </div>
    </div>
  )
}

function AvailabilityView({
  days,
  itemsByDay,
}: {
  days: Date[]
  itemsByDay: Map<string, CalendarItem[]>
}) {
  return (
    <div className="mt-6 rounded-4xl border border-white/10 bg-black/25 p-4">
      <p className="text-xs font-bold uppercase tracking-[0.2em] text-muted">
        Availability
      </p>
      <h2 className="text-xl font-bold">可约时间</h2>
      <p className="mt-1 text-sm font-semibold text-muted">
        默认按 18:00-23:00 扣除忙碌时段，好友忙碌信息保持脱敏。
      </p>
      <div className="mt-4 grid gap-3 lg:grid-cols-2">
        {days.map((day) => {
          const items = itemsByDay.get(dayKey(day)) ?? []
          const busyBlocks = items.map((item) => ({
            item,
            start: minutesOfDay(new Date(item.start_at)),
            end: minutesOfDay(itemEndDate(item)),
          }))
          const availableWindows = availabilityWindows(busyBlocks)
          return (
            <section
              className="rounded-3xl border border-white/10 bg-white/[0.04] p-4"
              key={dayKey(day)}
            >
              <h3 className="font-bold">{dayHeading(day)}</h3>
              <p className="mt-3 text-xs font-bold uppercase tracking-[0.18em] text-muted">
                忙碌时段
              </p>
              <div className="mt-2 space-y-2">
                {busyBlocks.length === 0 ? (
                  <p className="rounded-2xl bg-black/25 px-3 py-2 text-sm font-semibold text-muted">
                    暂无忙碌块
                  </p>
                ) : null}
                {busyBlocks.map(({ item, start, end }) => (
                  <div
                    className="rounded-2xl bg-black/35 px-3 py-2 text-sm"
                    key={timelineItemKey(item)}
                  >
                    <span className="font-bold">
                      {minutesLabel(start)}-{minutesLabel(end)}
                    </span>
                    <span className="ml-2 text-muted">{item.source_label}</span>
                  </div>
                ))}
              </div>
              <p className="mt-4 text-xs font-bold uppercase tracking-[0.18em] text-muted">
                可约窗口
              </p>
              <div className="mt-2 flex flex-wrap gap-2">
                {availableWindows.length === 0 ? (
                  <span className="rounded-full bg-rose-soft px-3 py-1 text-xs font-bold text-rose">
                    暂无可约窗口
                  </span>
                ) : null}
                {availableWindows.map((window) => (
                  <span
                    className="rounded-full bg-grass-soft px-3 py-1 text-xs font-bold text-grass"
                    key={`${window.start}-${window.end}`}
                  >
                    {minutesLabel(window.start)}-{minutesLabel(window.end)} 可约
                  </span>
                ))}
              </div>
            </section>
          )
        })}
      </div>
    </div>
  )
}

function availabilityWindows(
  busyBlocks: Array<{ start: number; end: number }>,
): Array<{ start: number; end: number }> {
  const normalized = busyBlocks
    .map((block) => ({
      start: Math.max(defaultAvailabilityStart, block.start),
      end: Math.min(defaultAvailabilityEnd, block.end),
    }))
    .filter((block) => block.start < block.end)
    .sort((a, b) => a.start - b.start)

  const windows: Array<{ start: number; end: number }> = []
  let cursor = defaultAvailabilityStart
  for (const block of normalized) {
    if (cursor < block.start) {
      windows.push({ start: cursor, end: block.start })
    }
    cursor = Math.max(cursor, block.end)
  }
  if (cursor < defaultAvailabilityEnd) {
    windows.push({ start: cursor, end: defaultAvailabilityEnd })
  }
  return windows
}

function SubscriptionSidebar({
  subscriptions,
  selectedIDs,
  loaded,
  error,
  onToggle,
}: {
  subscriptions: CalendarSubscription[]
  selectedIDs: Set<string>
  loaded: boolean
  error: string
  onToggle: (sourceID: string) => void
}) {
  return (
    <aside className="h-fit rounded-4xl border border-white/10 bg-surface/75 p-5 shadow-card backdrop-blur-xl xl:sticky xl:top-6">
      <p className="text-xs font-bold uppercase tracking-[0.2em] text-brand">
        Calendar sources
      </p>
      <h2 className="mt-1 text-2xl font-bold tracking-tight">我订阅的</h2>
      <p className="mt-2 text-sm font-semibold text-muted">
        勾选后，日历和 Feed 只显示对应来源。
      </p>

      {error ? (
        <p className="mt-4 rounded-2xl border border-rose/30 bg-rose-soft px-3 py-2 text-sm font-bold text-rose">
          {error}
        </p>
      ) : null}

      {!loaded ? (
        <p className="mt-4 rounded-2xl bg-white/5 px-3 py-2 text-sm font-bold text-muted">
          订阅加载中...
        </p>
      ) : null}

      {loaded && subscriptions.length === 0 && !error ? (
        <p className="mt-4 rounded-2xl bg-white/5 px-3 py-2 text-sm font-bold text-muted">
          暂无可订阅来源
        </p>
      ) : null}

      <div className="mt-4 space-y-2">
        {subscriptions.map((source) => {
          const checked = selectedIDs.has(source.id)
          return (
            <label
              className={`flex cursor-pointer items-center gap-3 rounded-3xl border px-3 py-3 transition-colors duration-200 ${
                checked
                  ? 'border-brand/40 bg-brand-soft/50'
                  : 'border-white/10 bg-white/[0.04] hover:bg-white/[0.07]'
              }`}
              key={source.id}
            >
              <input
                aria-label={source.label}
                checked={checked}
                className="h-4 w-4 accent-brand"
                onChange={() => onToggle(source.id)}
                type="checkbox"
              />
              <span
                aria-hidden="true"
                className={`h-3 w-3 shrink-0 rounded-full ${
                  colorBlock[source.color] ?? 'bg-brand'
                }`}
              />
              <span className="min-w-0">
                <span className="block truncate text-sm font-bold text-ink">
                  {source.label}
                </span>
                <span className="block text-xs font-semibold text-muted">
                  {sourceTypeText[source.type]}
                </span>
              </span>
            </label>
          )
        })}
      </div>

      {loaded && subscriptions.length > 0 ? (
        <p className="mt-4 rounded-full bg-white/5 px-3 py-2 text-xs font-bold text-muted">
          已显示 {selectedIDs.size} / {subscriptions.length} 个订阅
        </p>
      ) : null}
    </aside>
  )
}

type HangoutCardProps = {
  item: CalendarItem
  submitting: boolean
  dismissedExpanded: boolean
  message?: string
  friends: Friend[]
  friendsLoaded: boolean
  inviteOpen: boolean
  inviteSelectedIDs: string[]
  inviteSubmitting: boolean
  inviteError: string
  onDismissedHoverChange: (eventID: number, expanded: boolean) => void
  onInvite: (item: CalendarItem) => void
  onInviteClose: () => void
  onInviteeToggle: (eventID: number, friendID: string) => void
  onInviteSubmit: (eventID: number) => void
  onRSVP: (item: CalendarItem, going: boolean) => void
}

function TimelinePreview({
  item,
  onExpand,
}: {
  item: CalendarItem
  onExpand: () => void
}) {
  const soft = colorSoft[item.color] ?? colorSoft.blue
  return (
    <button
      aria-label={`展开日程 ${item.title}`}
      className="group w-full cursor-pointer rounded-3xl border border-white/10 bg-black/45 p-4 text-left shadow-card backdrop-blur transition-all duration-200 motion-safe:hover:-translate-y-1 motion-safe:hover:scale-[1.015] hover:border-brand/40 hover:bg-black/60 hover:shadow-[0_0_28px_rgba(0,242,234,0.14)]"
      onClick={onExpand}
      type="button"
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.2em] text-muted">
            {item.kind === 'event' ? 'Hangout' : 'Schedule'}
          </p>
          <h4 className="mt-1 text-base font-bold text-ink transition-colors duration-200 group-hover:text-brand">
            {item.title}
          </h4>
          <p className="mt-1 text-xs font-semibold text-muted">
            {item.location ?? '未设置地点'}
          </p>
        </div>
        <span className={`rounded-full px-3 py-1 text-xs font-bold ${soft}`}>
          {item.kind === 'event'
            ? `${item.going_count} 人参加`
            : item.visibility}
        </span>
      </div>
      <p className="mt-3 text-xs font-semibold text-muted">
        点击展开完整卡片
      </p>
    </button>
  )
}

function HangoutCard({
  item,
  submitting,
  dismissedExpanded,
  message,
  friends,
  friendsLoaded,
  inviteOpen,
  inviteSelectedIDs,
  inviteSubmitting,
  inviteError,
  onDismissedHoverChange,
  onInvite,
  onInviteClose,
  onInviteeToggle,
  onInviteSubmit,
  onRSVP,
}: HangoutCardProps) {
  const block = colorHero[item.color] ?? colorHero.blue
  const soft = colorSoft[item.color] ?? colorSoft.blue
  const isGoing = item.viewer_rsvp === 'going'
  const isDismissed = item.viewer_rsvp === 'not_going'
  const isExpanded = !isDismissed || dismissedExpanded
  const eventID = item.event_id
  const people = participantPeople(item)

  if (isDismissed && !isExpanded) {
    return (
      <article
        aria-label={`已收纳活动 ${item.title}，悬停后展开`}
        className="group cursor-pointer rounded-3xl border border-white/10 bg-black/35 p-4 shadow-card backdrop-blur transition-all duration-200 hover:border-brand/40 hover:bg-black/55"
        onFocus={() => eventID && onDismissedHoverChange(eventID, true)}
        onMouseEnter={() => eventID && onDismissedHoverChange(eventID, true)}
        tabIndex={0}
      >
        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-xs font-bold uppercase tracking-[0.2em] text-muted">
              已收纳
            </p>
            <h3 className="mt-1 max-w-xs truncate text-lg font-bold">
              {item.title}
            </h3>
          </div>
          <p className="rounded-full bg-white/10 px-3 py-2 text-xs font-bold text-muted">
            {item.going_count} 人已参加
          </p>
        </div>
      </article>
    )
  }

  return (
    <article
      className={`overflow-hidden rounded-3xl border border-white/10 bg-black/45 shadow-card backdrop-blur ${
        isDismissed ? 'ring-1 ring-brand/30' : ''
      }`}
      onMouseLeave={() =>
        isDismissed && eventID && onDismissedHoverChange(eventID, false)
      }
    >
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
        <div className="mt-4 flex flex-wrap items-center justify-between gap-3 rounded-3xl bg-white/[0.04] p-3">
          <div>
            <p className="text-sm font-bold text-ink">
              {item.going_count} 人已参加
            </p>
            <p className="text-xs font-semibold text-muted">
              {isGoing
                ? '你已加入'
                : isDismissed
                  ? '已收纳，悬停可重新加入'
                  : '看看谁也想去'}
            </p>
          </div>
          {people.length > 0 ? (
            <AvatarStack people={people} size="sm" />
          ) : (
            <span className="rounded-full bg-white/10 px-3 py-1 text-xs font-semibold text-muted">
              等待第一个参与者
            </span>
          )}
        </div>
        {message ? (
          <p className="mt-4 rounded-2xl border border-grass/30 bg-grass-soft px-4 py-3 text-sm font-bold text-grass">
            {message}
          </p>
        ) : null}
        <div className="mt-5 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <p className="rounded-full bg-white/10 px-3 py-2 text-xs font-bold text-muted">
            {isGoing ? '报名已确认' : isDismissed ? '已放到一边' : '可快速 RSVP'}
          </p>
          <div className="flex items-center gap-2">
            {isGoing ? (
              <>
                <button
                  className={`cursor-pointer rounded-full px-4 py-2 text-sm font-bold transition-colors duration-200 disabled:cursor-not-allowed disabled:opacity-50 ${soft}`}
                  disabled={submitting}
                  onClick={() => onRSVP(item, false)}
                  type="button"
                >
                  取消参加
                </button>
                <button
                  className="cursor-pointer rounded-full bg-white px-4 py-2 text-sm font-bold text-canvas transition-transform duration-200 hover:scale-105"
                  onClick={() => onInvite(item)}
                  type="button"
                >
                  邀请好友
                </button>
              </>
            ) : (
              <>
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
                  {isDismissed ? '重新加入' : '加入'}
                </button>
                <button
                  className={`cursor-pointer rounded-full px-4 py-2 text-sm font-bold transition-colors duration-200 disabled:cursor-not-allowed disabled:opacity-50 ${soft}`}
                  disabled={submitting}
                  onClick={() => onRSVP(item, false)}
                  type="button"
                >
                  下次一定
                </button>
              </>
            )}
          </div>
        </div>
        {inviteOpen && eventID ? (
          <div className="mt-4">
            <div className="mb-2 flex justify-end">
              <button
                className="cursor-pointer rounded-full px-3 py-1 text-xs font-bold text-muted transition-colors duration-200 hover:bg-white/10 hover:text-ink"
                onClick={onInviteClose}
                type="button"
              >
                收起
              </button>
            </div>
            <FriendInvitePicker
              emptyDescription="还没有好友，先去好友页添加好友，再回到这张活动卡片发邀请。"
              error={inviteError}
              friends={friends}
              loaded={friendsLoaded}
              onSubmit={() => onInviteSubmit(eventID)}
              onToggle={(friendID) => onInviteeToggle(eventID, friendID)}
              selectedIDs={inviteSelectedIDs}
              submitting={inviteSubmitting}
            />
          </div>
        ) : null}
      </div>
    </article>
  )
}

function participantPeople(item: CalendarItem): AvatarPerson[] {
  const preview = (item.participants_preview ?? []).map((person) => ({
    id: person.id,
    name: person.display_name,
    src: person.avatar_url,
  }))
  if (preview.length === 0 && item.viewer_rsvp === 'going') {
    return [{ id: 'viewer', name: '你' }]
  }
  return preview
}

function ScheduleCard({
  item,
  deleting = false,
  onDelete,
}: {
  item: CalendarItem
  deleting?: boolean
  onDelete?: (item: CalendarItem) => void
}) {
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
        {onDelete ? (
          <div className="mt-5 flex flex-wrap items-center gap-2">
            <Link
              className="cursor-pointer rounded-full border border-white/10 px-4 py-2 text-sm font-bold text-muted transition-colors duration-200 hover:border-brand/50 hover:text-brand"
              to={`/schedules/${item.id}`}
            >
              编辑日程
            </Link>
            <button
              className="cursor-pointer rounded-full bg-rose px-4 py-2 text-sm font-bold text-white transition-colors duration-200 hover:bg-rose/90 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={deleting}
              onClick={() => onDelete(item)}
              type="button"
            >
              {deleting ? '取消中...' : '取消日程'}
            </button>
          </div>
        ) : null}
      </div>
    </article>
  )
}
