import { Link } from 'react-router-dom'
import type { SessionState } from './hooks/useSession'
import { Avatar } from './components/Avatar'

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
      <h1 className="text-2xl font-bold tracking-tight">我的</h1>

      <div className="rounded-4xl border border-hairline bg-surface p-6 shadow-card">
        <div className="flex items-center gap-4">
          <Avatar
            name={profile?.display_name ?? '未登录'}
            seed={profile?.id ?? 'me'}
            size="lg"
            src={profile?.avatar_url}
          />
          <div>
            <p className="text-xl font-bold">
              {profile?.display_name ?? '未登录'}
            </p>
            <p className="text-sm text-muted">{profile?.email ?? ''}</p>
          </div>
        </div>
      </div>

      <div className="space-y-2">
        {links.map((link) => (
          <Link
            className="flex cursor-pointer items-center justify-between rounded-2xl border border-hairline bg-surface px-5 py-4 text-sm font-semibold shadow-card transition-colors duration-200 hover:bg-canvas"
            key={link.to}
            to={link.to}
          >
            <span>{link.label}</span>
            <span className="text-muted">›</span>
          </Link>
        ))}
      </div>

      <button
        className="w-full cursor-pointer rounded-2xl border border-hairline bg-surface px-5 py-4 text-sm font-bold text-rose shadow-card transition-colors duration-200 hover:bg-rose-soft"
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
