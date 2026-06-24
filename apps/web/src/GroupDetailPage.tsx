import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { apiGet } from './lib/api'
import {
  dissolveGroup,
  getGroup,
  inviteToGroup,
  leaveGroup,
} from './lib/groups'
import type { GroupDetail } from './lib/groups'
import { searchUsers } from './lib/friends'
import type { UserSearchResult } from './lib/friends'

type Me = { id: string }

export default function GroupDetailPage() {
  const params = useParams()
  const navigate = useNavigate()
  const groupID = Number(params.id)
  const [detail, setDetail] = useState<GroupDetail | null>(null)
  const [meID, setMeID] = useState('')
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<UserSearchResult[]>([])
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  async function refresh() {
    try {
      const data = await getGroup(groupID)
      setDetail(data)
      setError('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '无法加载群组')
    }
  }

  useEffect(() => {
    void Promise.resolve().then(() => {
      if (!Number.isFinite(groupID)) {
        setError('群组不存在')
        return
      }
      void refresh()
      apiGet<Me>('/api/users/me')
        .then((me) => setMeID(me.id))
        .catch(() => setMeID(''))
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [groupID])

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

  async function handleInvite(userID: string) {
    try {
      await inviteToGroup(groupID, userID)
      setMessage('邀请已发送')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '邀请失败')
    }
  }

  async function handleLeave() {
    try {
      await leaveGroup(groupID, meID)
      navigate('/groups')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '退出失败')
    }
  }

  async function handleDissolve() {
    try {
      await dissolveGroup(groupID)
      navigate('/groups')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '解散失败')
    }
  }

  if (!detail) {
    return (
      <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
        <section className="mx-auto max-w-2xl">
          <Link className="text-sm text-white/60" to="/groups">
            返回群组
          </Link>
          <p className="mt-6 text-sm text-white/70">{error || '正在加载...'}</p>
        </section>
      </main>
    )
  }

  const isOwner = detail.group.owner_id === meID

  return (
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-2xl space-y-6">
        <Link className="text-sm text-white/60" to="/groups">
          返回群组
        </Link>

        <div>
          <h1 className="text-3xl font-semibold">{detail.group.name}</h1>
          {detail.group.description ? (
            <p className="mt-1 text-white/60">{detail.group.description}</p>
          ) : null}
          <p className="mt-2 text-sm text-white/40">
            邀请码：<span className="font-mono">{detail.group.invite_code}</span>
          </p>
        </div>

        {error ? <p className="text-sm text-red-300">{error}</p> : null}
        {message ? <p className="text-sm text-emerald-300">{message}</p> : null}

        <Link
          className="inline-block rounded-2xl bg-sky-400 px-4 py-2 text-sm font-medium text-neutral-950"
          to={`/events/new?group_id=${detail.group.id}`}
        >
          在群内发起活动
        </Link>

        <div>
          <h2 className="text-sm uppercase tracking-[0.2em] text-white/50">
            成员
          </h2>
          <ul className="mt-2 space-y-2">
            {detail.members.map((member) => (
              <li
                className="flex items-center justify-between rounded-2xl border border-white/10 p-3 text-sm"
                key={member.user_id}
              >
                <div>
                  <p className="font-medium">{member.display_name}</p>
                  <p className="text-white/50">{member.email ?? member.user_id}</p>
                </div>
                <span className="rounded-full bg-white/10 px-3 py-1 text-xs text-white/60">
                  {member.role === 'owner' ? '群主' : '成员'}
                </span>
              </li>
            ))}
          </ul>
        </div>

        <form className="space-y-3" onSubmit={handleSearch}>
          <h2 className="text-sm uppercase tracking-[0.2em] text-white/50">
            邀请成员
          </h2>
          <div className="flex gap-2">
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
          </div>
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
                  onClick={() => handleInvite(user.id)}
                  type="button"
                >
                  邀请
                </button>
              </li>
            ))}
          </ul>
        ) : null}

        <div className="border-t border-white/10 pt-4">
          {isOwner ? (
            <button
              className="rounded-2xl border border-red-400/40 px-4 py-2 text-sm text-red-300"
              onClick={handleDissolve}
              type="button"
            >
              解散群组
            </button>
          ) : (
            <button
              className="rounded-2xl border border-white/20 px-4 py-2 text-sm text-white/70"
              onClick={handleLeave}
              type="button"
            >
              退出群组
            </button>
          )}
        </div>
      </section>
    </main>
  )
}
