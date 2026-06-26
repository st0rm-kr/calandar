import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { addDays, format, startOfDay } from 'date-fns'
import { deleteSchedule, listSchedules } from './lib/schedules'
import type { Schedule } from './lib/schedules'
import { EmptyState } from './components/EmptyState'
import { ErrorState } from './components/ErrorState'
import { LoadingState } from './components/LoadingState'

function formatScheduleTime(schedule: Schedule): string {
  const start = format(new Date(schedule.start_at), 'M月d日 HH:mm')
  if (!schedule.end_at) {
    return start
  }
  return `${start} - ${format(new Date(schedule.end_at), 'HH:mm')}`
}

export default function SchedulesPage() {
  const [schedules, setSchedules] = useState<Schedule[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [deleting, setDeleting] = useState<Set<number>>(() => new Set())

  const range = useMemo(() => {
    const today = startOfDay(new Date())
    return {
      from: addDays(today, -90).toISOString(),
      to: addDays(today, 90).toISOString(),
    }
  }, [])

  async function refresh() {
    setLoading(true)
    try {
      const rows = await listSchedules(range.from, range.to)
      setSchedules(rows.filter((schedule) => schedule.source === 'manual'))
      setError('')
    } catch (err) {
      setSchedules([])
      setError(err instanceof Error ? err.message : '无法加载日程')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void Promise.resolve().then(() => refresh())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [range.from, range.to])

  async function handleDelete(schedule: Schedule) {
    if (!window.confirm('确认删除这个日程吗？')) {
      return
    }
    setDeleting((current) => new Set(current).add(schedule.id))
    try {
      await deleteSchedule(schedule.id)
      setSchedules((current) =>
        current.filter((candidate) => candidate.id !== schedule.id),
      )
      setMessage('日程已删除')
      setError('')
    } catch (err) {
      setMessage('')
      setError(err instanceof Error ? err.message : '删除日程失败')
    } finally {
      setDeleting((current) => {
        const next = new Set(current)
        next.delete(schedule.id)
        return next
      })
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-bold uppercase tracking-[0.2em] text-brand">
            Schedules
          </p>
          <h1 className="text-2xl font-bold tracking-tight">日程管理</h1>
          <p className="mt-1 text-sm font-semibold text-muted">
            管理你的手动日程；活动日程仍在活动页处理。
          </p>
        </div>
        <Link
          className="rounded-full bg-gradient-to-r from-tangerine via-rose to-brand px-4 py-2 text-sm font-bold text-white shadow-pop transition-transform duration-200 hover:scale-105"
          to="/schedules/new"
        >
          新建日程
        </Link>
      </div>

      {message ? (
        <p className="rounded-2xl border border-grass/30 bg-grass-soft px-4 py-3 text-sm font-bold text-grass">
          {message}
        </p>
      ) : null}
      {error ? <ErrorState message={error} onRetry={refresh} /> : null}
      {loading ? <LoadingState /> : null}

      {!loading && schedules.length === 0 && !error ? (
        <EmptyState
          action={
            <Link
              className="rounded-full bg-brand px-5 py-2.5 text-sm font-bold text-white transition-colors duration-200 hover:bg-brand-ink"
              to="/schedules/new"
            >
              新建日程
            </Link>
          }
          description="新建一个手动日程，它会出现在首页 Feed 里。"
          title="还没有手动日程"
        />
      ) : null}

      <ul className="space-y-3">
        {schedules.map((schedule) => (
          <li
            className="rounded-3xl border border-white/10 bg-black/40 p-4 shadow-card backdrop-blur"
            key={schedule.id}
          >
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p className="text-lg font-bold">{schedule.title}</p>
                <p className="mt-1 text-sm font-semibold text-muted">
                  {formatScheduleTime(schedule)}
                  {schedule.location ? ` · ${schedule.location}` : ''}
                </p>
                <p className="mt-2 text-xs font-bold uppercase tracking-[0.18em] text-muted">
                  {schedule.visibility}
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <Link
                  aria-label={`编辑 ${schedule.title}`}
                  className="rounded-full border border-white/10 px-4 py-2 text-sm font-bold text-muted transition-colors duration-200 hover:border-brand/50 hover:text-brand"
                  to={`/schedules/${schedule.id}`}
                >
                  编辑
                </Link>
                <button
                  aria-label={`删除 ${schedule.title}`}
                  className="cursor-pointer rounded-full bg-rose px-4 py-2 text-sm font-bold text-white transition-colors duration-200 hover:bg-rose/90 disabled:cursor-not-allowed disabled:opacity-50"
                  disabled={deleting.has(schedule.id)}
                  onClick={() => handleDelete(schedule)}
                  type="button"
                >
                  {deleting.has(schedule.id) ? '删除中...' : '删除'}
                </button>
              </div>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
