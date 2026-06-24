import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { createGroup, joinGroup, listGroups } from './lib/groups'
import type { Group } from './lib/groups'

export default function GroupsPage() {
  const [groups, setGroups] = useState<Group[]>([])
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [inviteCode, setInviteCode] = useState('')
  const [error, setError] = useState('')

  async function refresh() {
    try {
      setGroups(await listGroups())
      setError('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '无法加载群组')
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
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-2xl space-y-6">
        <div className="flex items-center justify-between">
          <Link className="text-sm text-white/60" to="/">
            Hangout
          </Link>
          <Link className="text-sm text-white/60" to="/friends">
            我的好友
          </Link>
        </div>

        <h1 className="text-3xl font-semibold">群组</h1>

        {error ? <p className="text-sm text-red-300">{error}</p> : null}

        <form
          className="space-y-3 rounded-3xl bg-white/5 p-4"
          onSubmit={handleCreate}
        >
          <h2 className="text-sm uppercase tracking-[0.2em] text-white/50">
            创建群组
          </h2>
          <input
            className="w-full rounded-2xl bg-white/10 px-4 py-2 text-sm outline-none"
            onChange={(event) => setName(event.target.value)}
            placeholder="群名称"
            value={name}
          />
          <input
            className="w-full rounded-2xl bg-white/10 px-4 py-2 text-sm outline-none"
            onChange={(event) => setDescription(event.target.value)}
            placeholder="群简介（可选）"
            value={description}
          />
          <button
            className="rounded-2xl bg-white px-4 py-2 text-sm font-medium text-neutral-950"
            type="submit"
          >
            创建
          </button>
        </form>

        <form
          className="flex gap-2 rounded-3xl bg-white/5 p-4"
          onSubmit={handleJoin}
        >
          <input
            className="flex-1 rounded-2xl bg-white/10 px-4 py-2 text-sm outline-none"
            onChange={(event) => setInviteCode(event.target.value)}
            placeholder="输入邀请码加入群组"
            value={inviteCode}
          />
          <button
            className="rounded-2xl bg-sky-400 px-4 py-2 text-sm font-medium text-neutral-950"
            type="submit"
          >
            加入
          </button>
        </form>

        <div>
          <h2 className="text-sm uppercase tracking-[0.2em] text-white/50">
            我的群组
          </h2>
          {groups.length === 0 ? (
            <p className="mt-2 text-sm text-white/60">还没有加入任何群组。</p>
          ) : (
            <ul className="mt-2 space-y-2">
              {groups.map((group) => (
                <li key={group.id}>
                  <Link
                    className="flex items-center justify-between rounded-2xl border border-white/10 p-3 text-sm hover:bg-white/5"
                    to={`/groups/${group.id}`}
                  >
                    <div>
                      <p className="font-medium">{group.name}</p>
                      {group.description ? (
                        <p className="text-white/50">{group.description}</p>
                      ) : null}
                    </div>
                    <span className="text-white/40">查看</span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>
    </main>
  )
}
