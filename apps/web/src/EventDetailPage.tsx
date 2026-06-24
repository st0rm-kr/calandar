import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import type { EventDetail, EventRSVP, RSVPResult } from './lib/events'
import { getEventBySlug, rsvpEvent } from './lib/events'
import type { SessionState } from './hooks/useSession'
import { ConflictBanner } from './components/ConflictBanner'
import { EventTypeBadge } from './components/EventTypeBadge'

const rsvpOptions: Array<{ value: EventRSVP; label: string }> = [
  { value: 'going', label: '我要参加' },
  { value: 'maybe', label: '可能参加' },
  { value: 'not_going', label: '不参加' },
]

function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

type EventDetailPageProps = {
  session?: SessionState
}

export default function EventDetailPage({ session }: EventDetailPageProps) {
  const { slug } = useParams()
  const [detail, setDetail] = useState<EventDetail | null>(null)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [addToCalendar, setAddToCalendar] = useState(true)
  const [visibility, setVisibility] = useState('busy_only')
  const [submittingRSVP, setSubmittingRSVP] = useState<EventRSVP | null>(null)
  const [rsvpResult, setRsvpResult] = useState<RSVPResult | null>(null)

  const isAuthenticated = Boolean(session?.session)

  useEffect(() => {
    if (!slug) {
      return
    }

    let isCurrentRequest = true

    getEventBySlug(slug)
      .then((data) => {
        if (!isCurrentRequest) {
          return
        }
        setDetail(data)
        setError('')
      })
      .catch((err: unknown) => {
        if (!isCurrentRequest) {
          return
        }
        setDetail(null)
        setError(err instanceof Error ? err.message : '无法加载活动')
      })

    return () => {
      isCurrentRequest = false
    }
  }, [slug])

  const isFull = useMemo(() => {
    if (!detail?.event.capacity) {
      return false
    }
    return detail.going_count >= detail.event.capacity
  }, [detail])

  async function submitRSVP(rsvp: EventRSVP) {
    if (!detail) {
      return
    }

    setSubmittingRSVP(rsvp)
    setMessage('')
    setError('')

    try {
      const result = await rsvpEvent(detail.event.id, {
        rsvp,
        add_to_calendar: addToCalendar,
        visibility,
      })
      setRsvpResult(result)
      setMessage('RSVP 已更新')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'RSVP 失败')
    } finally {
      setSubmittingRSVP(null)
    }
  }

  const loadError = slug ? error : '活动链接无效'

  if (loadError && !detail) {
    return (
      <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
        <section className="mx-auto max-w-xl rounded-3xl bg-white/10 p-6 shadow-xl">
          <Link className="text-sm text-white/60" to="/">
            Hangout
          </Link>
          <p className="mt-6 text-white/70">{loadError}</p>
        </section>
      </main>
    )
  }

  if (!detail) {
    return (
      <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
        <section className="mx-auto max-w-xl rounded-3xl bg-white/10 p-6 shadow-xl">
          <p className="text-white/70">正在加载活动...</p>
        </section>
      </main>
    )
  }

  const event = detail.event
  const disabledRSVP = event.status !== 'active'

  return (
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-xl rounded-3xl bg-white/10 p-6 shadow-xl">
        <Link className="text-sm text-white/60" to="/">
          Hangout
        </Link>
        <p className="mt-6 text-sm uppercase tracking-[0.2em] text-white/50">
          Event
        </p>
        <div className="mt-3 flex items-center gap-3">
          <h1 className="text-3xl font-semibold">{event.title}</h1>
          <EventTypeBadge type={event.type} />
        </div>

        <dl className="mt-6 grid gap-4 text-sm text-white/75 sm:grid-cols-2">
          <div>
            <dt className="text-white/40">开始</dt>
            <dd>{formatDateTime(event.start_at)}</dd>
          </div>
          <div>
            <dt className="text-white/40">结束</dt>
            <dd>{event.end_at ? formatDateTime(event.end_at) : '未设置'}</dd>
          </div>
          <div>
            <dt className="text-white/40">地点</dt>
            <dd>{event.location ?? '未设置'}</dd>
          </div>
          <div>
            <dt className="text-white/40">名额</dt>
            <dd>
              {detail.going_count}
              {event.capacity ? ` / ${event.capacity}` : ''} 人已参加
            </dd>
          </div>
          <div>
            <dt className="text-white/40">状态</dt>
            <dd>{event.status}</dd>
          </div>
          <div>
            <dt className="text-white/40">分享链接</dt>
            <dd className="break-all">/events/{event.share_slug}</dd>
          </div>
        </dl>

        {isFull ? (
          <p className="mt-5 rounded-2xl border border-amber-300/30 bg-amber-300/10 p-3 text-sm text-amber-100">
            名额已满，参加按钮暂不可用。
          </p>
        ) : null}

        <div className="mt-8 rounded-3xl border border-white/10 p-4">
          <h2 className="text-lg font-semibold">RSVP</h2>

          {!isAuthenticated ? (
            <div className="mt-4 space-y-3">
              <p className="text-sm text-white/60">登录后即可报名这个活动。</p>
              <Link
                className="inline-block rounded-full bg-white px-5 py-2.5 text-sm font-medium text-neutral-950"
                to={`/login?redirect=${encodeURIComponent(`/e/${event.share_slug}`)}`}
              >
                登录以报名
              </Link>
            </div>
          ) : (
            <>
              <label className="mt-4 flex items-center gap-3 text-sm text-white/75">
                <input
                  checked={addToCalendar}
                  onChange={(event) => setAddToCalendar(event.target.checked)}
                  type="checkbox"
                />
                加入我的日历
              </label>

              <label className="mt-4 block text-sm font-medium text-white/80">
                日历可见性
                <select
                  className="mt-2 w-full rounded-2xl border border-white/10 bg-neutral-900 px-4 py-3 text-white outline-none focus:border-white/40"
                  disabled={!addToCalendar}
                  onChange={(event) => setVisibility(event.target.value)}
                  value={visibility}
                >
                  <option value="public">公开</option>
                  <option value="busy_only">仅显示忙碌</option>
                  <option value="private">私密</option>
                </select>
              </label>

              <div className="mt-5 grid gap-3 sm:grid-cols-3">
                {rsvpOptions.map((option) => {
                  const fullGoingButton = option.value === 'going' && isFull
                  return (
                    <button
                      className="rounded-full border border-white/20 px-4 py-3 text-sm font-medium text-white disabled:cursor-not-allowed disabled:opacity-40"
                      disabled={
                        disabledRSVP || fullGoingButton || submittingRSVP !== null
                      }
                      key={option.value}
                      onClick={() => void submitRSVP(option.value)}
                      type="button"
                    >
                      {submittingRSVP === option.value
                        ? '提交中...'
                        : option.label}
                    </button>
                  )
                })}
              </div>

              {message ? (
                <p className="mt-4 text-sm text-emerald-300">{message}</p>
              ) : null}
              {error ? (
                <p className="mt-4 text-sm text-red-300">{error}</p>
              ) : null}

              {rsvpResult?.conflicts.length ? (
                <div className="mt-5">
                  <ConflictBanner
                    conflicts={rsvpResult.conflicts.map((conflict) => ({
                      id: conflict.id,
                      title: conflict.title,
                    }))}
                  />
                </div>
              ) : null}
            </>
          )}
        </div>
      </section>
    </main>
  )
}
