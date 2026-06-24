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
import { Avatar } from './components/Avatar'

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
      <div>
        <Link className="text-sm font-semibold text-brand" to="/groups">
          ‹ 返回群组
        </Link>
        <p className="mt-6 text-sm font-semibold text-muted">
          {error || '正在加载...'}
        </p>
      </div>
    )
  }

  const isOwner = detail.group.owner_id === meID

  return (
    <div className="space-y-5">
      <Link className="text-sm font-semibold text-brand" to="/groups">
        ‹ 返回群组
      </Link>

      <div className="rounded-4xl border border-hairline bg-surface p-6 shadow-card">
        <div className="flex items-center gap-3">
          <Avatar name={detail.group.name} seed={`group-${detail.group.id}`} size="lg" />
          <div>
            <h1 className="text-2xl font-bold tracking-tight">
              {detail.group.name}
            </h1>
            {detail.group.description ? (
              <p className="text-muted">{detail.group.description}</p>
            ) : null}
          </div>
        </div>
        <p className="mt-4 text-sm text-muted">
          邀请码：
          <span className="font-mono font-bold text-ink">
            {detail.group.invite_code}
          </span>
        </p>
      </div>

      {error ? <p className="text-sm font-semibold text-rose">{error}</p> : null}
      {message ? (
        <p className="text-sm font-semibold text-grass">{message}</p>
      ) : null}

      <Link
        className="inline-block rounded-full bg-brand px-4 py-2.5 text-sm font-bold text-white transition-colors duration-200 hover:bg-brand-ink"
        to={`/events/new?group_id=${detail.group.id}`}
      >
        在群内发起活动
      </Link>

      <div>
        <h2 className="text-sm font-bold uppercase tracking-[0.2em] text-muted">
          成员
        </h2>
        <ul className="mt-2 space-y-2">
          {detail.members.map((member) => (
            <li
              className="flex items-center justify-between rounded-2xl border border-hairline bg-surface p-3 text-sm shadow-card"
              key={member.user_id}
            >
              <div className="flex items-center gap-3">
                <Avatar
                  name={member.display_name}
                  seed={member.user_id}
                  src={member.avatar_url}
                />
                <div>
                  <p className="font-bold">{member.display_name}</p>
                  <p className="text-muted">{member.email ?? member.user_id}</p>
                </div>
              </div>
              <span className="rounded-full bg-canvas px-3 py-1 text-xs font-bold text-muted">
                {member.role === 'owner' ? '群主' : '成员'}
              </span>
            </li>
          ))}
        </ul>
      </div>

      <form className="space-y-3" onSubmit={handleSearch}>
        <h2 className="text-sm font-bold uppercase tracking-[0.2em] text-muted">
          邀请成员
        </h2>
        <div className="flex gap-2">
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
        </div>
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
                onClick={() => handleInvite(user.id)}
                type="button"
              >
                邀请
              </button>
            </li>
          ))}
        </ul>
      ) : null}

      <div className="border-t border-hairline pt-4">
        {isOwner ? (
          <button
            className="cursor-pointer rounded-2xl border border-rose/40 px-4 py-2 text-sm font-bold text-rose transition-colors duration-200 hover:bg-rose-soft"
            onClick={handleDissolve}
            type="button"
          >
            解散群组
          </button>
        ) : (
          <button
            className="cursor-pointer rounded-2xl border border-hairline px-4 py-2 text-sm font-bold text-muted transition-colors duration-200 hover:text-ink"
            onClick={handleLeave}
            type="button"
          >
            退出群组
          </button>
        )}
      </div>
    </div>
  )
}
