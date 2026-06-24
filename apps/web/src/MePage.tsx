import { Link } from 'react-router-dom'
import type { SessionState } from './hooks/useSession'
import { Card } from './components/Card'

type MePageProps = {
  session: SessionState
}

const links = [
  { to: '/profile', label: '个人资料' },
  { to: '/friends', label: '好友' },
  { to: '/groups', label: '群组' },
  { to: '/notifications', label: '通知' },
]

export default function MePage({ session }: MePageProps) {
  const { profile, signOut } = session

  return (
    <div className="space-y-5">
      <h1 className="text-2xl font-semibold tracking-tight">我的</h1>

      <Card>
        <div className="flex items-center gap-4">
          {profile?.avatar_url ? (
            <img
              alt=""
              className="h-14 w-14 rounded-full object-cover"
              src={profile.avatar_url}
            />
          ) : (
            <div className="flex h-14 w-14 items-center justify-center rounded-full bg-white/10 text-lg">
              {profile?.display_name?.slice(0, 1) ?? '?'}
            </div>
          )}
          <div>
            <p className="text-lg font-medium">
              {profile?.display_name ?? '未登录'}
            </p>
            <p className="text-sm text-white/50">{profile?.email ?? ''}</p>
          </div>
        </div>
      </Card>

      <div className="space-y-2">
        {links.map((link) => (
          <Link
            className="flex items-center justify-between rounded-2xl border border-white/10 px-4 py-3 text-sm"
            key={link.to}
            to={link.to}
          >
            <span>{link.label}</span>
            <span className="text-white/30">›</span>
          </Link>
        ))}
      </div>

      <button
        className="w-full rounded-2xl border border-white/15 px-4 py-3 text-sm text-white/70"
        onClick={() => {
          void signOut()
        }}
        type="button"
      >
        退出登录
      </button>
    </div>
  )
}
