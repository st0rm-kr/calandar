import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import type { EventDetail, EventRSVP, RSVPResult } from './lib/events'
import { getEventBySlug, rsvpEvent } from './lib/events'
import type { SessionState } from './hooks/useSession'
import { ConflictBanner } from './components/ConflictBanner'
import { EventTypeBadge } from './components/EventTypeBadge'
import { metaForType } from './lib/eventType'

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

function Shell({ children }: { children: React.ReactNode }) {
  return (
    <main className="min-h-screen bg-canvas px-5 py-8 text-ink">
      <section className="mx-auto max-w-xl">{children}</section>
    </main>
  )
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
      <Shell>
        <Link className="text-sm font-semibold text-brand" to="/">
          Hangout
        </Link>
        <div className="mt-6 rounded-4xl border border-hairline bg-surface p-6 shadow-card">
          <p className="text-muted">{loadError}</p>
        </div>
      </Shell>
    )
  }

  if (!detail) {
    return (
      <Shell>
        <div className="mt-6 rounded-4xl border border-hairline bg-surface p-6 shadow-card">
          <p className="text-muted">正在加载活动...</p>
        </div>
      </Shell>
    )
  }

  const event = detail.event
  const disabledRSVP = event.status !== 'active'
  const meta = metaForType(event.type)

  return (
    <Shell>
      <Link className="text-sm font-semibold text-brand" to="/">
        Hangout
      </Link>

      <div className="mt-4 overflow-hidden rounded-4xl border border-hairline bg-surface shadow-soft">
        <div className={`h-24 w-full ${meta.block}`} />
        <div className="p-6">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-3xl font-bold tracking-tight">{event.title}</h1>
            <EventTypeBadge type={event.type} />
          </div>

          <dl className="mt-6 grid gap-4 text-sm sm:grid-cols-2">
            <Detail label="开始" value={formatDateTime(event.start_at)} />
            <Detail
              label="结束"
              value={event.end_at ? formatDateTime(event.end_at) : '未设置'}
            />
            <Detail label="地点" value={event.location ?? '未设置'} />
            <Detail
              label="名额"
              value={`${detail.going_count}${
                event.capacity ? ` / ${event.capacity}` : ''
              } 人已参加`}
            />
          </dl>

          {isFull ? (
            <p className="mt-5 rounded-2xl border border-tangerine/30 bg-tangerine-soft p-3 text-sm font-semibold text-tangerine">
              名额已满，参加按钮暂不可用。
            </p>
          ) : null}
        </div>
      </div>

      <div className="mt-4 rounded-4xl border border-hairline bg-surface p-6 shadow-card">
        <h2 className="text-lg font-bold">报名</h2>

        {!isAuthenticated ? (
          <div className="mt-4 space-y-3">
            <p className="text-sm text-muted">登录后即可报名这个活动。</p>
            <Link
              className="inline-block rounded-full bg-brand px-5 py-2.5 text-sm font-bold text-white transition-colors duration-200 hover:bg-brand-ink"
              to={`/login?redirect=${encodeURIComponent(`/e/${event.share_slug}`)}`}
            >
              登录以报名
            </Link>
          </div>
        ) : (
          <>
            <label className="mt-4 flex items-center gap-3 text-sm font-semibold text-ink">
              <input
                checked={addToCalendar}
                onChange={(event) => setAddToCalendar(event.target.checked)}
                type="checkbox"
              />
              加入我的日历
            </label>

            <label className="mt-4 block text-sm font-semibold text-ink">
              日历可见性
              <select
                className="mt-2 w-full rounded-2xl border border-hairline bg-canvas px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand disabled:opacity-50"
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
                const isPrimary = option.value === 'going'
                return (
                  <button
                    className={`cursor-pointer rounded-full px-4 py-3 text-sm font-bold transition-colors duration-200 disabled:cursor-not-allowed disabled:opacity-40 ${
                      isPrimary
                        ? 'bg-brand text-white hover:bg-brand-ink'
                        : 'border border-hairline text-muted hover:text-ink'
                    }`}
                    disabled={
                      disabledRSVP || fullGoingButton || submittingRSVP !== null
                    }
                    key={option.value}
                    onClick={() => void submitRSVP(option.value)}
                    type="button"
                  >
                    {submittingRSVP === option.value ? '提交中...' : option.label}
                  </button>
                )
              })}
            </div>

            {message ? (
              <p className="mt-4 text-sm font-semibold text-grass">{message}</p>
            ) : null}
            {error ? (
              <p className="mt-4 text-sm font-semibold text-rose">{error}</p>
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
    </Shell>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-muted">{label}</dt>
      <dd className="mt-0.5 font-semibold text-ink">{value}</dd>
    </div>
  )
}
