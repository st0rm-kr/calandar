import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiGet } from './lib/api'

type Profile = {
  id: string
  display_name: string
  avatar_url: string | null
  email: string | null
  status: string
}

export default function ProfilePage() {
  const [profile, setProfile] = useState<Profile | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    apiGet<Profile>('/api/users/me')
      .then((data) => {
        setProfile(data)
        setError('')
      })
      .catch((err: unknown) => {
        setProfile(null)
        setError(err instanceof Error ? err.message : '无法加载个人资料')
      })
  }, [])

  return (
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-md rounded-3xl bg-white/10 p-6 shadow-xl">
        <Link className="text-sm text-white/60" to="/">
          Hangout
        </Link>
        <p className="mt-6 text-sm uppercase tracking-[0.2em] text-white/50">Profile</p>
        {profile ? (
          <div className="mt-4">
            {profile.avatar_url ? (
              <img alt="" className="mb-4 h-20 w-20 rounded-full object-cover" src={profile.avatar_url} />
            ) : null}
            <h1 className="text-3xl font-semibold">{profile.display_name}</h1>
            <dl className="mt-6 space-y-3 text-sm text-white/70">
              <div>
                <dt className="text-white/40">邮箱</dt>
                <dd>{profile.email ?? '未设置'}</dd>
              </div>
              <div>
                <dt className="text-white/40">状态</dt>
                <dd>{profile.status}</dd>
              </div>
              <div>
                <dt className="text-white/40">用户 ID</dt>
                <dd className="break-all">{profile.id}</dd>
              </div>
            </dl>
          </div>
        ) : (
          <p className="mt-4 text-white/70">{error || '正在加载个人资料...'}</p>
        )}
      </section>
    </main>
  )
}
