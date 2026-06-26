import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiGet, apiPatch } from './lib/api'
import { Avatar } from './components/Avatar'

type Profile = {
  id: string
  display_name: string
  avatar_url: string | null
  email: string | null
  status: string
}

export default function ProfilePage() {
  const [profile, setProfile] = useState<Profile | null>(null)
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    apiGet<Profile>('/api/users/me')
      .then((data) => {
        setProfile(data)
        setDisplayName(data.display_name)
        setError('')
      })
      .catch((err: unknown) => {
        setProfile(null)
        setError(err instanceof Error ? err.message : '无法加载个人资料')
      })
  }, [])

  async function saveDisplayName() {
    if (!profile) {
      return
    }
    setSaving(true)
    setError('')
    setNotice('')
    try {
      const nextProfile = await apiPatch<Profile>('/api/users/me', {
        display_name: displayName,
        avatar_url: profile.avatar_url,
      })
      setProfile(nextProfile)
      setDisplayName(nextProfile.display_name)
      setNotice('用户名已更新')
    } catch (err) {
      setError(err instanceof Error ? err.message : '用户名更新失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <Link className="text-sm font-semibold text-brand" to="/me">
        ‹ Hangout
      </Link>
      {profile ? (
        <div className="mt-4 rounded-4xl border border-hairline bg-surface p-6 shadow-card">
          <Avatar
            name={profile.display_name}
            seed={profile.id}
            size="lg"
            src={profile.avatar_url}
          />
          <h1 className="mt-4 text-2xl font-bold tracking-tight">
            {profile.display_name}
          </h1>
          <div className="mt-6 rounded-3xl border border-hairline bg-canvas/60 p-4">
            <label className="text-sm font-bold text-muted" htmlFor="display-name">
              用户名
            </label>
            <div className="mt-2 flex flex-col gap-2 sm:flex-row">
              <input
                className="min-h-11 flex-1 rounded-2xl border border-hairline bg-surface px-4 text-sm font-semibold text-ink outline-none transition-colors duration-200 focus:border-brand"
                id="display-name"
                onChange={(event) => setDisplayName(event.target.value)}
                value={displayName}
              />
              <button
                className="cursor-pointer rounded-2xl bg-gradient-to-r from-brand to-rose px-5 py-2 text-sm font-bold text-white shadow-pop transition-transform duration-200 hover:scale-105 disabled:cursor-not-allowed disabled:opacity-50"
                disabled={saving}
                onClick={() => {
                  void saveDisplayName()
                }}
                type="button"
              >
                {saving ? '保存中...' : '保存用户名'}
              </button>
            </div>
            {notice ? (
              <p className="mt-3 text-sm font-bold text-grass">{notice}</p>
            ) : null}
            {error ? (
              <p className="mt-3 text-sm font-bold text-rose">{error}</p>
            ) : null}
          </div>
          <dl className="mt-6 space-y-3 text-sm">
            <Field label="邮箱" value={profile.email ?? '未设置'} />
            <Field label="状态" value={profile.status} />
            <Field label="用户 ID" value={profile.id} mono />
          </dl>
        </div>
      ) : (
        <p className="mt-6 font-semibold text-muted">
          {error || '正在加载个人资料...'}
        </p>
      )}
    </div>
  )
}

function Field({
  label,
  value,
  mono = false,
}: {
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div>
      <dt className="text-muted">{label}</dt>
      <dd className={`mt-0.5 font-semibold text-ink ${mono ? 'break-all font-mono' : ''}`}>
        {value}
      </dd>
    </div>
  )
}
