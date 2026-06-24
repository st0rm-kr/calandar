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
import { LoadingState } from './components/LoadingState'
import { EmptyState } from './components/EmptyState'
import { Avatar } from './components/Avatar'

export default function FriendsPage() {
  const [friends, setFriends] = useState<Friend[]>([])
  const [requests, setRequests] = useState<FriendRequest[]>([])
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<UserSearchResult[]>([])
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [loaded, setLoaded] = useState(false)

  async function refresh() {
    try {
      const [friendList, requestList] = await Promise.all([
        listFriends(),
        listFriendRequests(),
      ])
      setFriends(friendList)
      setRequests(requestList)
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
