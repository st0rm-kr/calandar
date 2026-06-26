import { Link } from 'react-router-dom'
import { Avatar } from './Avatar'
import type { Friend } from '../lib/friends'

type FriendInvitePickerProps = {
  friends: Friend[]
  selectedIDs: string[]
  loaded: boolean
  error?: string
  submitting?: boolean
  submitLabel?: string
  showSubmitButton?: boolean
  emptyDescription?: string
  onToggle: (friendID: string) => void
  onSubmit: () => void
}

export function FriendInvitePicker({
  friends,
  selectedIDs,
  loaded,
  error = '',
  submitting = false,
  submitLabel = '发送邀请',
  showSubmitButton = true,
  emptyDescription = '还没有好友，先去好友页添加好友。',
  onToggle,
  onSubmit,
}: FriendInvitePickerProps) {
  const selectedCount = selectedIDs.length

  return (
    <div className="rounded-3xl border border-white/10 bg-black/55 p-4 shadow-[0_0_30px_rgba(0,242,234,0.12)] backdrop-blur-xl">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.2em] text-brand">
            Invite
          </p>
          <h4 className="mt-1 text-lg font-bold">邀请好友</h4>
        </div>
        <span className="rounded-full bg-white/10 px-3 py-1 text-xs font-bold text-muted">
          已选 {selectedCount}
        </span>
      </div>

      {error ? (
        <p className="mt-3 rounded-2xl border border-rose/30 bg-rose-soft px-3 py-2 text-sm font-bold text-rose">
          {error}
        </p>
      ) : null}

      {!loaded ? (
        <p className="mt-4 text-sm font-semibold text-muted">正在加载好友...</p>
      ) : friends.length === 0 ? (
        <div className="mt-4 rounded-2xl bg-white/[0.04] p-4">
          <p className="text-sm font-semibold text-muted">{emptyDescription}</p>
          <Link
            className="mt-3 inline-flex rounded-full bg-white px-4 py-2 text-sm font-bold text-canvas transition-transform duration-200 hover:scale-105"
            to="/friends"
          >
            去好友页添加好友
          </Link>
        </div>
      ) : (
        <>
          <div className="mt-4 grid gap-2">
            {friends.map((friend) => {
              const checked = selectedIDs.includes(friend.user_id)
              return (
                <label
                  className={`flex cursor-pointer items-center justify-between gap-3 rounded-2xl border px-3 py-2 transition-colors duration-200 ${
                    checked
                      ? 'border-brand/50 bg-brand-soft text-brand'
                      : 'border-white/10 bg-white/[0.04] text-ink hover:bg-white/10'
                  }`}
                  key={friend.user_id}
                >
                  <span className="flex items-center gap-3">
                    <Avatar
                      decorative
                      name={friend.display_name}
                      seed={friend.user_id}
                      size="sm"
                      src={friend.avatar_url}
                    />
                    <span>
                      <span className="block text-sm font-bold">
                        {friend.display_name}
                      </span>
                      <span className="block text-xs font-semibold text-muted">
                        {friend.email ?? friend.user_id}
                      </span>
                    </span>
                  </span>
                  <input
                    aria-label={`选择 ${friend.display_name}`}
                    checked={checked}
                    className="h-4 w-4 accent-brand"
                    onChange={() => onToggle(friend.user_id)}
                    type="checkbox"
                  />
                </label>
              )
            })}
          </div>
          {showSubmitButton ? (
            <button
              className="mt-4 w-full cursor-pointer rounded-full bg-gradient-to-r from-tangerine via-rose to-brand px-4 py-2.5 text-sm font-bold text-white shadow-pop transition-transform duration-200 hover:scale-[1.02] disabled:cursor-not-allowed disabled:opacity-50"
              disabled={selectedCount === 0 || submitting}
              onClick={onSubmit}
              type="button"
            >
              {submitting ? '发送中...' : submitLabel}
            </button>
          ) : null}
        </>
      )}
    </div>
  )
}
