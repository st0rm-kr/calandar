import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  acceptFriendRequest,
  deleteFriend,
  listFriendRequests,
  listFriends,
  rejectFriendRequest,
  searchUsers,
  sendFriendRequest,
} from './lib/friends'
import type { Friend, FriendRequest, UserSearchResult } from './lib/friends'
import { inviteToEvent, listMine } from './lib/events'
import type { Event } from './lib/events'
import { LoadingState } from './components/LoadingState'
import { EmptyState } from './components/EmptyState'
import { Avatar } from './components/Avatar'
import { FriendInvitePicker } from './components/FriendInvitePicker'

export default function FriendsPage() {
  const [friends, setFriends] = useState<Friend[]>([])
  const [requests, setRequests] = useState<FriendRequest[]>([])
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<UserSearchResult[]>([])
  const [events, setEvents] = useState<Event[]>([])
  const [selectedEventID, setSelectedEventID] = useState('')
  const [selectedInviteeIDs, setSelectedInviteeIDs] = useState<string[]>([])
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [loaded, setLoaded] = useState(false)
  const [inviteSubmitting, setInviteSubmitting] = useState(false)

  async function refresh() {
    try {
      const [friendList, requestList, eventList] = await Promise.all([
        listFriends(),
        listFriendRequests(),
        listMine(),
      ])
      setFriends(friendList)
      setRequests(requestList)
      setEvents(eventList.filter((event) => event.status === 'active'))
      setError('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '无法加载好友')
    } finally {
      setLoaded(true)
    }
  }

  useEffect(() => {
    void Promise.resolve().then(() => refresh())
  }, [])

  async function handleSearch(event: React.FormEvent) {
    event.preventDefault()
    if (!query.trim()) {
      return
    }
    try {
      setResults(await searchUsers(query.trim()))
      setError('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '搜索失败')
    }
  }

  async function handleSend(userID: string) {
    try {
      await sendFriendRequest(userID)
      setMessage('好友申请已发送')
      await refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '发送失败')
    }
  }

  async function handleAccept(requestID: number) {
    await acceptFriendRequest(requestID)
    await refresh()
  }

  async function handleReject(requestID: number) {
    await rejectFriendRequest(requestID)
    await refresh()
  }

  async function handleDelete(userID: string) {
    await deleteFriend(userID)
    await refresh()
  }

  function toggleInvitee(friendID: string) {
    setSelectedInviteeIDs((current) =>
      current.includes(friendID)
        ? current.filter((id) => id !== friendID)
        : [...current, friendID],
    )
  }

  async function handleInviteFriends() {
    if (!selectedEventID) {
      setError('请先选择一个活动')
      return
    }
    if (selectedInviteeIDs.length === 0) {
      setError('请至少选择一位好友')
      return
    }
    setInviteSubmitting(true)
    try {
      const invited = await inviteToEvent(
        Number(selectedEventID),
        selectedInviteeIDs,
      )
      const count = invited.length || selectedInviteeIDs.length
      setMessage(`已邀请 ${count} 位好友参加活动`)
      setSelectedInviteeIDs([])
      setError('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '邀请失败')
    } finally {
      setInviteSubmitting(false)
    }
  }

  const incoming = requests.filter((request) => request.direction === 'incoming')
  const outgoing = requests.filter((request) => request.direction === 'outgoing')

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">好友</h1>
        <Link className="text-sm font-semibold text-brand" to="/groups">
          我的群组
        </Link>
      </div>

      {error ? <p className="text-sm font-semibold text-rose">{error}</p> : null}
      {message ? (
        <p className="text-sm font-semibold text-grass">{message}</p>
      ) : null}

      <section className="rounded-3xl border border-white/10 bg-black/35 p-4 shadow-card backdrop-blur">
        <div className="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-xs font-bold uppercase tracking-[0.2em] text-brand">
              Bulk invite
            </p>
            <h2 className="mt-1 text-xl font-bold">邀请好友参加活动</h2>
          </div>
          <Link className="text-sm font-semibold text-brand" to="/events/new">
            创建活动
          </Link>
        </div>

        <label className="mt-4 block text-sm font-semibold text-ink">
          选择活动
          <select
            className="mt-2 w-full rounded-2xl border border-hairline bg-surface px-4 py-3 text-ink outline-none transition-colors duration-200 focus:border-brand"
            onChange={(event) => setSelectedEventID(event.target.value)}
            value={selectedEventID}
          >
            <option value="">先选择一个活动</option>
            {events.map((event) => (
              <option key={event.id} value={event.id}>
                {event.title}
              </option>
            ))}
          </select>
        </label>

        {loaded && events.length === 0 ? (
          <p className="mt-3 text-sm font-semibold text-muted">
            还没有可邀请的活动，先创建一个 Hangout。
          </p>
        ) : null}

        <div className="mt-4">
          <FriendInvitePicker
            emptyDescription="还没有好友，先搜索昵称或邮箱添加好友。"
            friends={friends}
            loaded={loaded}
            onSubmit={handleInviteFriends}
            onToggle={toggleInvitee}
            selectedIDs={selectedInviteeIDs}
            submitLabel="发送活动邀请"
            submitting={inviteSubmitting}
          />
        </div>
      </section>

      <form className="flex gap-2" onSubmit={handleSearch}>
        <input
          className="flex-1 rounded-2xl border border-hairline bg-surface px-4 py-2.5 text-sm outline-none transition-colors duration-200 focus:border-brand"
          onChange={(event) => setQuery(event.target.value)}
          placeholder="按昵称或邮箱搜索"
          value={query}
        />
        <button
          className="cursor-pointer rounded-2xl bg-brand px-4 py-2.5 text-sm font-bold text-white transition-colors duration-200 hover:bg-brand-ink"
          type="submit"
        >
          搜索
        </button>
      </form>

      {results.length > 0 ? (
        <ul className="space-y-2">
          {results.map((user) => (
            <li
              className="flex items-center justify-between rounded-2xl border border-hairline bg-surface p-3 text-sm shadow-card"
              key={user.id}
            >
              <div className="flex items-center gap-3">
                <Avatar name={user.display_name} seed={user.id} src={user.avatar_url} />
                <div>
                  <p className="font-bold">{user.display_name}</p>
                  <p className="text-muted">{user.email ?? user.id}</p>
                </div>
              </div>
              <button
                className="cursor-pointer rounded-full bg-brand px-3 py-1 font-bold text-white transition-colors duration-200 hover:bg-brand-ink"
                onClick={() => handleSend(user.id)}
                type="button"
              >
                添加
              </button>
            </li>
          ))}
        </ul>
      ) : null}

      {incoming.length > 0 ? (
        <div>
          <h2 className="text-sm font-bold uppercase tracking-[0.2em] text-muted">
            收到的申请
          </h2>
          <ul className="mt-2 space-y-2">
            {incoming.map((request) => (
              <li
                className="flex items-center justify-between rounded-2xl border border-hairline bg-surface p-3 text-sm shadow-card"
                key={request.id}
              >
                <span className="break-all font-semibold text-ink">
                  {request.requester_id}
                </span>
                <div className="flex gap-2">
                  <button
                    className="cursor-pointer rounded-full bg-grass px-3 py-1 font-bold text-white transition-colors duration-200 hover:bg-grass/90"
                    onClick={() => handleAccept(request.id)}
                    type="button"
                  >
                    接受
                  </button>
                  <button
                    className="cursor-pointer rounded-full border border-hairline px-3 py-1 font-bold text-muted transition-colors duration-200 hover:text-ink"
                    onClick={() => handleReject(request.id)}
                    type="button"
                  >
                    拒绝
                  </button>
                </div>
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      {outgoing.length > 0 ? (
        <div>
          <h2 className="text-sm font-bold uppercase tracking-[0.2em] text-muted">
            已发送
          </h2>
          <ul className="mt-2 space-y-2">
            {outgoing.map((request) => (
              <li
                className="rounded-2xl border border-hairline bg-surface p-3 text-sm font-semibold text-muted shadow-card"
                key={request.id}
              >
                {request.addressee_id} · 等待确认
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      <div>
        <h2 className="text-sm font-bold uppercase tracking-[0.2em] text-muted">
          好友列表
        </h2>
        {!loaded ? (
          <LoadingState />
        ) : friends.length === 0 ? (
          <div className="mt-2">
            <EmptyState title="还没有好友" description="搜索昵称或邮箱，添加第一个好友。" />
          </div>
        ) : (
          <ul className="mt-2 space-y-2">
            {friends.map((friend) => (
              <li
                className="flex items-center justify-between rounded-2xl border border-hairline bg-surface p-3 text-sm shadow-card"
                key={friend.user_id}
              >
                <div className="flex items-center gap-3">
                  <Avatar
                    name={friend.display_name}
                    seed={friend.user_id}
                    src={friend.avatar_url}
                  />
                  <div>
                    <p className="font-bold">{friend.display_name}</p>
                    <p className="text-muted">{friend.email ?? friend.user_id}</p>
                  </div>
                </div>
                <button
                  className="cursor-pointer rounded-full border border-hairline px-3 py-1 font-bold text-muted transition-colors duration-200 hover:text-rose"
                  onClick={() => handleDelete(friend.user_id)}
                  type="button"
                >
                  删除
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
