import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  checkConflicts,
  createSchedule,
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
  const [saving, setSaving] = useState(false)

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
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-xl rounded-3xl bg-white/10 p-6 shadow-xl">
        <Link className="text-sm text-white/60" to="/calendar">
          Hangout
        </Link>
        <h1 className="mt-6 text-3xl font-semibold">
          {scheduleID ? '编辑日程' : '新建日程'}
        </h1>

        <form className="mt-8 space-y-5" onSubmit={handleSubmit}>
          <label className="block text-sm font-medium text-white/80">
            标题
            <input
              className="mt-2 w-full rounded-2xl border border-white/10 bg-white/10 px-4 py-3 text-white outline-none focus:border-white/40"
              maxLength={80}
              onChange={(event) => setTitle(event.target.value)}
              required
              value={title}
            />
          </label>

          <div className="grid gap-4 sm:grid-cols-2">
            <label className="block text-sm font-medium text-white/80">
              开始时间
              <input
                className="mt-2 w-full rounded-2xl border border-white/10 bg-white/10 px-4 py-3 text-white outline-none focus:border-white/40"
                onChange={(event) => setStartAt(event.target.value)}
                required
                type="datetime-local"
                value={startAt}
              />
            </label>

            <label className="block text-sm font-medium text-white/80">
              结束时间
              <input
                className="mt-2 w-full rounded-2xl border border-white/10 bg-white/10 px-4 py-3 text-white outline-none focus:border-white/40"
                onChange={(event) => setEndAt(event.target.value)}
                type="datetime-local"
                value={endAt}
              />
            </label>
          </div>

          <label className="block text-sm font-medium text-white/80">
            地点
            <input
              className="mt-2 w-full rounded-2xl border border-white/10 bg-white/10 px-4 py-3 text-white outline-none focus:border-white/40"
              onChange={(event) => setLocation(event.target.value)}
              value={location}
            />
          </label>

          <fieldset className="space-y-2">
            <legend className="text-sm font-medium text-white/80">
              可见性
            </legend>
            {visibilityOptions.map((option) => (
              <label
                className="flex items-center gap-3 rounded-2xl border border-white/10 p-3 text-sm"
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
                  <span className="font-medium">{option.label}</span>
                  <span className="block text-white/50">{option.hint}</span>
                </span>
              </label>
            ))}
          </fieldset>

          {conflicts.length > 0 ? (
            <div className="rounded-2xl border border-amber-300/30 bg-amber-300/10 p-4 text-sm text-amber-100">
              <p className="font-medium">
                与 {conflicts.length} 个已有日程冲突（仍可保存）
              </p>
              <ul className="mt-2 space-y-1">
                {conflicts.map((conflict) => (
                  <li key={conflict.id}>{formatConflict(conflict)}</li>
                ))}
              </ul>
            </div>
          ) : null}

          {error ? <p className="text-sm text-red-300">{error}</p> : null}

          <button
            className="w-full rounded-full bg-white px-5 py-3 font-medium text-neutral-950 disabled:cursor-not-allowed disabled:opacity-50"
            disabled={saving}
            type="submit"
          >
            {saving ? '保存中...' : '保存日程'}
          </button>
        </form>
      </section>
    </main>
  )
}
