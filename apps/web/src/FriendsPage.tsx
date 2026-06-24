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
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-2xl space-y-6">
        <div className="flex items-center justify-between">
          <Link className="text-sm text-white/60" to="/">
            Hangout
          </Link>
          <Link className="text-sm text-white/60" to="/groups">
            我的群组
          </Link>
        </div>

        <h1 className="text-3xl font-semibold">好友</h1>

        {error ? <p className="text-sm text-red-300">{error}</p> : null}
        {message ? <p className="text-sm text-emerald-300">{message}</p> : null}

        <form className="flex gap-2" onSubmit={handleSearch}>
          <input
            className="flex-1 rounded-2xl bg-white/10 px-4 py-2 text-sm outline-none"
            onChange={(event) => setQuery(event.target.value)}
            placeholder="按昵称或邮箱搜索"
            value={query}
          />
          <button
            className="rounded-2xl bg-white px-4 py-2 text-sm font-medium text-neutral-950"
            type="submit"
          >
            搜索
          </button>
        </form>

        {results.length > 0 ? (
          <ul className="space-y-2">
            {results.map((user) => (
              <li
                className="flex items-center justify-between rounded-2xl border border-white/10 p-3 text-sm"
                key={user.id}
              >
                <div>
                  <p className="font-medium">{user.display_name}</p>
                  <p className="text-white/50">{user.email ?? user.id}</p>
                </div>
                <button
                  className="rounded-full bg-sky-400 px-3 py-1 text-neutral-950"
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
            <h2 className="text-sm uppercase tracking-[0.2em] text-white/50">
              收到的申请
            </h2>
            <ul className="mt-2 space-y-2">
              {incoming.map((request) => (
                <li
                  className="flex items-center justify-between rounded-2xl border border-white/10 p-3 text-sm"
                  key={request.id}
                >
                  <span className="break-all text-white/70">
                    {request.requester_id}
                  </span>
                  <div className="flex gap-2">
                    <button
                      className="rounded-full bg-emerald-400 px-3 py-1 text-neutral-950"
                      onClick={() => handleAccept(request.id)}
                      type="button"
                    >
                      接受
                    </button>
                    <button
                      className="rounded-full border border-white/20 px-3 py-1"
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
            <h2 className="text-sm uppercase tracking-[0.2em] text-white/50">
              已发送
            </h2>
            <ul className="mt-2 space-y-2">
              {outgoing.map((request) => (
                <li
                  className="rounded-2xl border border-white/10 p-3 text-sm text-white/60"
                  key={request.id}
                >
                  {request.addressee_id} · 等待确认
                </li>
              ))}
            </ul>
          </div>
        ) : null}

        <div>
          <h2 className="text-sm uppercase tracking-[0.2em] text-white/50">
            好友列表
          </h2>
          {!loaded ? (
            <LoadingState />
          ) : friends.length === 0 ? (
            <p className="mt-2 text-sm text-white/60">还没有好友。</p>
          ) : (
            <ul className="mt-2 space-y-2">
              {friends.map((friend) => (
                <li
                  className="flex items-center justify-between rounded-2xl border border-white/10 p-3 text-sm"
                  key={friend.user_id}
                >
                  <div>
                    <p className="font-medium">{friend.display_name}</p>
                    <p className="text-white/50">{friend.email ?? friend.user_id}</p>
                  </div>
                  <button
                    className="rounded-full border border-white/20 px-3 py-1 text-white/70"
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
      </section>
    </main>
  )
}
