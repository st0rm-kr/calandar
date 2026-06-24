import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { createEvent } from './lib/events'

const eventTypes = [
  { value: 'climbing', label: '攀岩' },
  { value: 'dining', label: '聚餐' },
  { value: 'travel', label: '旅行' },
  { value: 'gaming', label: '游戏' },
  { value: 'other', label: '其他' },
]

function toISODateTime(value: string): string {
  return new Date(value).toISOString()
}

export default function NewEventPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const groupIDParam = searchParams.get('group_id')
  const groupID = groupIDParam ? Number(groupIDParam) : null
  const [title, setTitle] = useState('')
  const [scope, setScope] = useState(groupID ? 'group' : 'personal')
  const [type, setType] = useState('climbing')
  const [startAt, setStartAt] = useState('')
  const [endAt, setEndAt] = useState('')
  const [location, setLocation] = useState('')
  const [capacity, setCapacity] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError('')

    try {
      const created = await createEvent({
        title,
        scope,
        type,
        group_id: scope === 'group' ? groupID : null,
        start_at: toISODateTime(startAt),
        end_at: endAt ? toISODateTime(endAt) : null,
        location: location.trim() || null,
        capacity: capacity ? Number(capacity) : null,
      })
      navigate(`/events/${created.share_slug}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建活动失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-xl rounded-3xl bg-white/10 p-6 shadow-xl">
        <Link className="text-sm text-white/60" to="/">
          Hangout
        </Link>
        <h1 className="mt-6 text-3xl font-semibold">创建活动</h1>
        <p className="mt-3 text-sm text-white/60">
          填写活动信息后会生成一个可分享详情页。
        </p>

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
              范围
              <select
                className="mt-2 w-full rounded-2xl border border-white/10 bg-neutral-900 px-4 py-3 text-white outline-none focus:border-white/40"
                onChange={(event) => setScope(event.target.value)}
                value={scope}
              >
                <option value="personal">个人</option>
                <option value="group">群组</option>
              </select>
            </label>

            <label className="block text-sm font-medium text-white/80">
              类型
              <select
                className="mt-2 w-full rounded-2xl border border-white/10 bg-neutral-900 px-4 py-3 text-white outline-none focus:border-white/40"
                onChange={(event) => setType(event.target.value)}
                value={type}
              >
                {eventTypes.map((eventType) => (
                  <option key={eventType.value} value={eventType.value}>
                    {eventType.label}
                  </option>
                ))}
              </select>
            </label>
          </div>

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

          <label className="block text-sm font-medium text-white/80">
            人数上限
            <input
              className="mt-2 w-full rounded-2xl border border-white/10 bg-white/10 px-4 py-3 text-white outline-none focus:border-white/40"
              min={1}
              onChange={(event) => setCapacity(event.target.value)}
              type="number"
              value={capacity}
            />
          </label>

          {error ? <p className="text-sm text-red-300">{error}</p> : null}

          <button
            className="w-full rounded-full bg-white px-5 py-3 font-medium text-neutral-950 disabled:cursor-not-allowed disabled:opacity-50"
            disabled={submitting}
            type="submit"
          >
            {submitting ? '创建中...' : '创建活动'}
          </button>
        </form>
      </section>
    </main>
  )
}
