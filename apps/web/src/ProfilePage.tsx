import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiGet } from './lib/api'
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
