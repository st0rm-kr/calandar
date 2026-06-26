import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  checkConflicts,
  createSchedule,
  getSchedule,
  updateSchedule,
} from './lib/schedules'
import type { ScheduleConflict, ScheduleVisibility } from './lib/schedules'

const visibilityOptions: Array<{
  value: ScheduleVisibility
  label: string
  hint: string
}> = [
  { value: 'public', label: '公开', hint: '好友可见标题与详情' },
  { value: 'busy_only', label: '仅显示忙碌', hint: '好友只看到忙碌时段' },
  { value: 'private', label: '私密', hint: '仅自己可见' },
]

function toISO(value: string): string {
  return new Date(value).toISOString()
}

function toDateTimeLocal(value: string | null): string {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  const pad = (input: number) => String(input).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function formatConflict(conflict: ScheduleConflict): string {
  const start = new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(new Date(conflict.start_at))
  return `${conflict.title} · ${start}`
}

export default function ScheduleEditorPage() {
  const navigate = useNavigate()
  const { id } = useParams()
  const scheduleID = id ? Number(id) : undefined

  const [title, setTitle] = useState('')
  const [startAt, setStartAt] = useState('')
  const [endAt, setEndAt] = useState('')
  const [location, setLocation] = useState('')
  const [visibility, setVisibility] = useState<ScheduleVisibility>('busy_only')
  const [conflicts, setConflicts] = useState<ScheduleConflict[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(Boolean(scheduleID))
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let active = true
    if (!scheduleID) {
      return () => {
        active = false
      }
    }
    Promise.resolve()
      .then(async () => {
        if (active) {
          setLoading(true)
        }
        return getSchedule(scheduleID)
      })
      .then((schedule) => {
        if (!active) {
          return
        }
        setTitle(schedule.title)
        setStartAt(toDateTimeLocal(schedule.start_at))
        setEndAt(toDateTimeLocal(schedule.end_at))
        setLocation(schedule.location ?? '')
        setVisibility(schedule.visibility)
        setError('')
      })
      .catch((err: unknown) => {
        if (active) {
          setError(err instanceof Error ? err.message : '无法加载日程')
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
  }, [scheduleID])

  useEffect(() => {
    let active = true

    if (startAt === '') {
      Promise.resolve().then(() => {
        if (active) {
          setConflicts([])
        }
      })
      return () => {
        active = false
      }
    }

    const startISO = toISO(startAt)
    const endISO = endAt ? toISO(endAt) : null
    checkConflicts(startISO, endISO, scheduleID)
      .then((data) => {
        if (active) {
          setConflicts(data)
        }
      })
      .catch(() => {
        if (active) {
          setConflicts([])
        }
      })

    return () => {
      active = false
    }
  }, [startAt, endAt, scheduleID])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSaving(true)
    setError('')

    const input = {
      title,
      start_at: toISO(startAt),
      end_at: endAt ? toISO(endAt) : null,
      location: location.trim() || null,
      visibility,
    }

    try {
      if (scheduleID) {
        await updateSchedule(scheduleID, input)
      } else {
        await createSchedule(input)
      }
      navigate('/calendar')
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存日程失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <Link className="text-sm font-semibold text-brand" to="/calendar">
        ‹ Hangout
      </Link>
      <h1 className="mt-4 text-2xl font-bold tracking-tight">
        {scheduleID ? '编辑日程' : '新建日程'}
      </h1>

      {loading ? (
        <p className="mt-6 text-sm font-semibold text-muted">正在加载日程...</p>
      ) : null}

      <form className="mt-6 space-y-5" onSubmit={handleSubmit}>
        <label className="block text-sm font-semibold text-ink">
          标题
          <input
            className="mt-2 w-full rounded-2xl border border-hairline bg-surface px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
            maxLength={80}
            onChange={(event) => setTitle(event.target.value)}
            required
            value={title}
          />
        </label>

        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block text-sm font-semibold text-ink">
            开始时间
            <input
              className="mt-2 w-full rounded-2xl border border-hairline bg-surface px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
              onChange={(event) => setStartAt(event.target.value)}
              required
              type="datetime-local"
              value={startAt}
            />
          </label>

          <label className="block text-sm font-semibold text-ink">
            结束时间
            <input
              className="mt-2 w-full rounded-2xl border border-hairline bg-surface px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
              onChange={(event) => setEndAt(event.target.value)}
              type="datetime-local"
              value={endAt}
            />
          </label>
        </div>

        <label className="block text-sm font-semibold text-ink">
          地点
          <input
            className="mt-2 w-full rounded-2xl border border-hairline bg-surface px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
            onChange={(event) => setLocation(event.target.value)}
            value={location}
          />
        </label>

        <fieldset className="space-y-2">
          <legend className="text-sm font-semibold text-ink">可见性</legend>
          {visibilityOptions.map((option) => (
            <label
              className={`flex cursor-pointer items-center gap-3 rounded-2xl border bg-surface p-3 text-sm transition-colors duration-200 ${
                visibility === option.value
                  ? 'border-brand'
                  : 'border-hairline'
              }`}
              key={option.value}
            >
              <input
                checked={visibility === option.value}
                name="visibility"
                onChange={() => setVisibility(option.value)}
                type="radio"
                value={option.value}
              />
              <span>
                <span className="font-bold">{option.label}</span>
                <span className="block text-muted">{option.hint}</span>
              </span>
            </label>
          ))}
        </fieldset>

        {conflicts.length > 0 ? (
          <div className="rounded-2xl border border-tangerine/30 bg-tangerine-soft p-4 text-sm text-tangerine">
            <p className="font-bold">
              与 {conflicts.length} 个已有日程冲突（仍可保存）
            </p>
            <ul className="mt-2 space-y-1">
              {conflicts.map((conflict) => (
                <li key={conflict.id}>{formatConflict(conflict)}</li>
              ))}
            </ul>
          </div>
        ) : null}

        {error ? (
          <p className="text-sm font-semibold text-rose">{error}</p>
        ) : null}

        <button
          className="w-full cursor-pointer rounded-full bg-brand px-5 py-3 font-bold text-white transition-colors duration-200 hover:bg-brand-ink disabled:cursor-not-allowed disabled:opacity-50"
          disabled={saving}
          type="submit"
        >
          {saving ? '保存中...' : '保存日程'}
        </button>
      </form>
    </div>
  )
}
