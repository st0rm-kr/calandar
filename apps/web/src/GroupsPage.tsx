import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { createGroup, joinGroup, listGroups } from './lib/groups'
import type { Group } from './lib/groups'
import { LoadingState } from './components/LoadingState'
import { EmptyState } from './components/EmptyState'
import { Avatar } from './components/Avatar'

export default function GroupsPage() {
  const [groups, setGroups] = useState<Group[]>([])
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [inviteCode, setInviteCode] = useState('')
  const [error, setError] = useState('')
  const [loaded, setLoaded] = useState(false)

  async function refresh() {
    try {
      setGroups(await listGroups())
      setError('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '无法加载群组')
    } finally {
      setLoaded(true)
    }
  }

  useEffect(() => {
    void Promise.resolve().then(() => refresh())
  }, [])

  async function handleCreate(event: React.FormEvent) {
    event.preventDefault()
    if (!name.trim()) {
      return
    }
    try {
      await createGroup(name.trim(), description.trim() || null)
      setName('')
      setDescription('')
      await refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '创建失败')
    }
  }

  async function handleJoin(event: React.FormEvent) {
    event.preventDefault()
    if (!inviteCode.trim()) {
      return
    }
    try {
      await joinGroup(inviteCode.trim())
      setInviteCode('')
      await refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '加入失败')
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">群组</h1>
        <Link className="text-sm font-semibold text-brand" to="/friends">
          我的好友
        </Link>
      </div>

      {error ? <p className="text-sm font-semibold text-rose">{error}</p> : null}

      <form
        className="space-y-3 rounded-3xl border border-hairline bg-surface p-4 shadow-card"
        onSubmit={handleCreate}
      >
        <h2 className="text-sm font-bold uppercase tracking-[0.2em] text-muted">
          创建群组
        </h2>
        <input
          className="w-full rounded-2xl border border-hairline bg-canvas px-4 py-2.5 text-sm outline-none transition-colors duration-200 focus:border-brand"
          onChange={(event) => setName(event.target.value)}
          placeholder="群名称"
          value={name}
        />
        <input
          className="w-full rounded-2xl border border-hairline bg-canvas px-4 py-2.5 text-sm outline-none transition-colors duration-200 focus:border-brand"
          onChange={(event) => setDescription(event.target.value)}
          placeholder="群简介（可选）"
          value={description}
        />
        <button
          className="cursor-pointer rounded-2xl bg-brand px-4 py-2.5 text-sm font-bold text-white transition-colors duration-200 hover:bg-brand-ink"
          type="submit"
        >
          创建
        </button>
      </form>

      <form
        className="flex gap-2 rounded-3xl border border-hairline bg-surface p-4 shadow-card"
        onSubmit={handleJoin}
      >
        <input
          className="flex-1 rounded-2xl border border-hairline bg-canvas px-4 py-2.5 text-sm outline-none transition-colors duration-200 focus:border-brand"
          onChange={(event) => setInviteCode(event.target.value)}
          placeholder="输入邀请码加入群组"
          value={inviteCode}
        />
        <button
          className="cursor-pointer rounded-2xl bg-grass px-4 py-2.5 text-sm font-bold text-white transition-colors duration-200 hover:bg-grass/90"
          type="submit"
        >
          加入
        </button>
      </form>

      <div>
        <h2 className="text-sm font-bold uppercase tracking-[0.2em] text-muted">
          我的群组
        </h2>
        {!loaded ? (
          <LoadingState />
        ) : groups.length === 0 ? (
          <div className="mt-2">
            <EmptyState
              title="还没有加入任何群组"
              description="创建群组或用邀请码加入。"
            />
          </div>
        ) : (
          <ul className="mt-2 space-y-2">
            {groups.map((group) => (
              <li key={group.id}>
                <Link
                  className="flex items-center justify-between rounded-2xl border border-hairline bg-surface p-4 text-sm shadow-card transition-colors duration-200 hover:bg-canvas"
                  to={`/groups/${group.id}`}
                >
                  <div className="flex items-center gap-3">
                    <Avatar name={group.name} seed={`group-${group.id}`} />
                    <div>
                      <p className="font-bold">{group.name}</p>
                      {group.description ? (
                        <p className="text-muted">{group.description}</p>
                      ) : null}
                    </div>
                  </div>
                  <span className="text-muted">›</span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
