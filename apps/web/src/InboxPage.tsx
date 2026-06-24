import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getInbox } from './lib/inbox'
import type { InboxItem, InboxResult } from './lib/inbox'
import { acceptFriendRequest, rejectFriendRequest } from './lib/friends'
import { acceptGroupInvite, rejectGroupInvite } from './lib/groups'
import { rsvpEvent } from './lib/events'
import type { EventRSVP } from './lib/events'

const rsvpOptions: Array<{ value: EventRSVP; label: string }> = [
  { value: 'going', label: '参加' },
  { value: 'maybe', label: '可能' },
  { value: 'not_going', label: '不参加' },
]

function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

export default function InboxPage() {
  const [result, setResult] = useState<InboxResult | null>(null)
  const [error, setError] = useState('')
  const [addToCalendar, setAddToCalendar] = useState(true)
  const [visibility, setVisibility] = useState('busy_only')

  async function refresh() {
    try {
      setResult(await getInbox())
      setError('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '无法加载收件箱')
    }
  }

  useEffect(() => {
    void Promise.resolve().then(() => refresh())
  }, [])

  async function handleRSVP(eventID: number, rsvp: EventRSVP) {
    try {
      await rsvpEvent(eventID, { rsvp, add_to_calendar: addToCalendar, visibility })
      await refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '操作失败')
    }
  }

  async function handleFriend(requestID: number, accept: boolean) {
    try {
      if (accept) {
        await acceptFriendRequest(requestID)
      } else {
        await rejectFriendRequest(requestID)
      }
      await refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '操作失败')
    }
  }

  async function handleGroup(inviteID: number, accept: boolean) {
    try {
      if (accept) {
        await acceptGroupInvite(inviteID)
      } else {
        await rejectGroupInvite(inviteID)
      }
      await refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '操作失败')
    }
  }

  function renderActions(item: InboxItem) {
    if (item.disabled) {
      return (
        <span className="text-xs text-white/40">{item.disabled_reason}</span>
      )
    }
    if (item.kind === 'event_invite' && item.event_id) {
      const eventID = item.event_id
      return (
        <div className="flex flex-wrap gap-2">
          {rsvpOptions.map((option) => (
            <button
              className="rounded-full bg-white/10 px-3 py-1 text-xs"
              key={option.value}
              onClick={() => handleRSVP(eventID, option.value)}
              type="button"
            >
              {option.label}
            </button>
          ))}
        </div>
      )
    }
    if (item.kind === 'friend_request') {
      return (
        <div className="flex gap-2">
          <button
            className="rounded-full bg-emerald-400 px-3 py-1 text-xs text-neutral-950"
            onClick={() => handleFriend(item.source_id, true)}
            type="button"
          >
            接受
          </button>
          <button
            className="rounded-full border border-white/20 px-3 py-1 text-xs"
            onClick={() => handleFriend(item.source_id, false)}
            type="button"
          >
            拒绝
          </button>
        </div>
      )
    }
    if (item.kind === 'group_invite') {
      return (
        <div className="flex gap-2">
          <button
            className="rounded-full bg-emerald-400 px-3 py-1 text-xs text-neutral-950"
            onClick={() => handleGroup(item.source_id, true)}
            type="button"
          >
            接受
          </button>
          <button
            className="rounded-full border border-white/20 px-3 py-1 text-xs"
            onClick={() => handleGroup(item.source_id, false)}
            type="button"
          >
            拒绝
          </button>
        </div>
      )
    }
    return null
  }

  return (
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-2xl space-y-6">
        <div className="flex items-center justify-between">
          <Link className="text-sm text-white/60" to="/">
            Hangout
          </Link>
          <Link className="text-sm text-white/60" to="/notifications">
            通知
          </Link>
        </div>

        <h1 className="text-3xl font-semibold">收件箱</h1>

        {result ? (
          <p className="text-sm text-white/50">
            待处理 {result.counts.total} 项 · 活动 {result.counts.event_invites} ·
            好友 {result.counts.friend_requests} · 群组{' '}
            {result.counts.group_invites}
          </p>
        ) : null}

        {error ? <p className="text-sm text-red-300">{error}</p> : null}

        <div className="flex flex-wrap items-center gap-4 rounded-2xl bg-white/5 p-3 text-sm">
          <label className="flex items-center gap-2">
            <input
              checked={addToCalendar}
              onChange={(event) => setAddToCalendar(event.target.checked)}
              type="checkbox"
            />
            加入我的日历
          </label>
          <label className="flex items-center gap-2">
            可见性
            <select
              className="rounded-xl bg-neutral-900 px-2 py-1"
              onChange={(event) => setVisibility(event.target.value)}
              value={visibility}
            >
              <option value="busy_only">仅忙碌</option>
              <option value="public">公开</option>
              <option value="private">私密</option>
            </select>
          </label>
        </div>

        {result && result.items.length === 0 ? (
          <p className="text-sm text-white/60">收件箱已清空。</p>
        ) : null}

        <ul className="space-y-2">
          {result?.items.map((item) => (
            <li
              className={`rounded-2xl border p-4 ${
                item.disabled
                  ? 'border-white/5 text-white/40'
                  : 'border-white/10'
              } ${
                item.conflict_count && item.conflict_count > 0
                  ? 'ring-1 ring-red-400/50'
                  : ''
              }`}
              key={item.id}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-medium">{item.title}</p>
                  <p className="text-sm text-white/50">{item.subtitle}</p>
                  {item.conflict_count && item.conflict_count > 0 ? (
                    <p className="mt-1 text-xs text-red-300">
                      与你已有 {item.conflict_count} 个日程冲突
                    </p>
                  ) : null}
                  <p className="mt-1 text-xs text-white/30">
                    {formatDateTime(item.created_at)}
                  </p>
                </div>
                <div>{renderActions(item)}</div>
              </div>
            </li>
          ))}
        </ul>
      </section>
    </main>
  )
}
