import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { createEvent, inviteToEvent } from './lib/events'
import { listFriends } from './lib/friends'
import type { Friend } from './lib/friends'
import { FriendInvitePicker } from './components/FriendInvitePicker'

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
  const [friends, setFriends] = useState<Friend[]>([])
  const [friendsLoaded, setFriendsLoaded] = useState(false)
  const [selectedInviteeIDs, setSelectedInviteeIDs] = useState<string[]>([])
  const [friendsError, setFriendsError] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    let active = true
    listFriends()
      .then((friendList) => {
        if (!active) {
          return
        }
        setFriends(friendList)
        setFriendsError('')
      })
      .catch((err: unknown) => {
        if (!active) {
          return
        }
        setFriends([])
        setFriendsError(err instanceof Error ? err.message : '无法加载好友')
      })
      .finally(() => {
        if (active) {
          setFriendsLoaded(true)
        }
      })
    return () => {
      active = false
    }
  }, [])

  function toggleInvitee(friendID: string) {
    setSelectedInviteeIDs((current) =>
      current.includes(friendID)
        ? current.filter((id) => id !== friendID)
        : [...current, friendID],
    )
  }

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
      if (selectedInviteeIDs.length > 0) {
        await inviteToEvent(created.id, selectedInviteeIDs)
      }
      navigate(`/events/${created.share_slug}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建活动失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <Link className="text-sm font-semibold text-brand" to="/">
        ‹ Hangout
      </Link>
      <h1 className="mt-4 text-2xl font-bold tracking-tight">创建活动</h1>
      <p className="mt-1 text-sm text-muted">
        填写活动信息后会生成一个可分享详情页。
      </p>

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
            范围
            <select
              className="mt-2 w-full rounded-2xl border border-hairline bg-surface px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
              onChange={(event) => setScope(event.target.value)}
              value={scope}
            >
              <option value="personal">个人</option>
              <option value="group">群组</option>
            </select>
          </label>

          <label className="block text-sm font-semibold text-ink">
            类型
            <select
              className="mt-2 w-full rounded-2xl border border-hairline bg-surface px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
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

        <label className="block text-sm font-semibold text-ink">
          人数上限
          <input
            className="mt-2 w-full rounded-2xl border border-hairline bg-surface px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
            min={1}
            onChange={(event) => setCapacity(event.target.value)}
            type="number"
            value={capacity}
          />
        </label>

        <section className="rounded-3xl border border-white/10 bg-black/35 p-4 shadow-card backdrop-blur">
          <div className="flex items-start justify-between gap-3">
            <div>
              <p className="text-xs font-bold uppercase tracking-[0.2em] text-brand">
                Optional
              </p>
              <h2 className="mt-1 text-lg font-bold">顺便邀请好友</h2>
            </div>
            <span className="rounded-full bg-white/10 px-3 py-1 text-xs font-bold text-muted">
              创建后发送
            </span>
          </div>
          <p className="mt-2 text-sm font-semibold text-muted">
            提交时会先创建活动，再向选中的好友发送邀请。
          </p>
          <div className="mt-4">
            <FriendInvitePicker
              emptyDescription="还没有好友，创建活动后可以先去好友页添加好友。"
              error={friendsError}
              friends={friends}
              loaded={friendsLoaded}
              onSubmit={() => undefined}
              onToggle={toggleInvitee}
              selectedIDs={selectedInviteeIDs}
              showSubmitButton={false}
            />
          </div>
        </section>

        {error ? (
          <p className="text-sm font-semibold text-rose">{error}</p>
        ) : null}

        <button
          className="w-full cursor-pointer rounded-full bg-brand px-5 py-3 font-bold text-white transition-colors duration-200 hover:bg-brand-ink disabled:cursor-not-allowed disabled:opacity-50"
          disabled={submitting}
          type="submit"
        >
          {submitting ? '创建中...' : '创建活动'}
        </button>
      </form>
    </div>
  )
}
