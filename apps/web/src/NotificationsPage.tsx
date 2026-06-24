import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
} from './lib/notifications'
import type { Notification, NotificationList } from './lib/notifications'

const typeLabels: Record<string, string> = {
  friend_request: '好友申请',
  friend_accepted: '好友已通过',
  event_invite: '活动邀请',
  event_changed: '活动变更',
  event_cancelled: '活动取消',
  group_invite: '群组邀请',
}

function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

function targetPath(notification: Notification): string {
  const payload = notification.payload
  switch (notification.type) {
    case 'friend_request':
    case 'friend_accepted':
      return '/friends'
    case 'group_invite':
      return typeof payload.group_id === 'number'
        ? `/groups/${payload.group_id}`
        : '/groups'
    case 'event_invite':
    case 'event_changed':
    case 'event_cancelled':
      return '/inbox'
    default:
      return '/inbox'
  }
}

export default function NotificationsPage() {
  const navigate = useNavigate()
  const [list, setList] = useState<NotificationList | null>(null)
  const [error, setError] = useState('')

  async function refresh() {
    try {
      setList(await listNotifications())
      setError('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '无法加载通知')
    }
  }

  useEffect(() => {
    void Promise.resolve().then(() => refresh())
  }, [])

  async function handleOpen(notification: Notification) {
    try {
      if (!notification.read_at) {
        await markNotificationRead(notification.id)
      }
    } catch {
      // navigation should not be blocked by a read failure
    }
    navigate(targetPath(notification))
  }

  async function handleReadAll() {
    try {
      await markAllNotificationsRead()
      await refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '操作失败')
    }
  }

  return (
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-2xl space-y-6">
        <div className="flex items-center justify-between">
          <Link className="text-sm text-white/60" to="/">
            Hangout
          </Link>
          <Link className="text-sm text-white/60" to="/inbox">
            收件箱
          </Link>
        </div>

        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-semibold">通知</h1>
          {list && list.unread_count > 0 ? (
            <button
              className="rounded-full border border-white/20 px-3 py-1 text-sm"
              onClick={handleReadAll}
              type="button"
            >
              全部已读 ({list.unread_count})
            </button>
          ) : null}
        </div>

        {error ? <p className="text-sm text-red-300">{error}</p> : null}

        {list && list.items.length === 0 ? (
          <p className="text-sm text-white/60">还没有通知。</p>
        ) : null}

        <ul className="space-y-2">
          {list?.items.map((notification) => (
            <li key={notification.id}>
              <button
                className={`w-full rounded-2xl border p-4 text-left ${
                  notification.read_at
                    ? 'border-white/5 text-white/50'
                    : 'border-white/20 bg-white/5'
                }`}
                onClick={() => handleOpen(notification)}
                type="button"
              >
                <div className="flex items-center justify-between">
                  <span className="font-medium">
                    {typeLabels[notification.type] ?? notification.type}
                  </span>
                  {!notification.read_at ? (
                    <span className="h-2 w-2 rounded-full bg-sky-400" />
                  ) : null}
                </div>
                <p className="mt-1 text-xs text-white/30">
                  {formatDateTime(notification.created_at)}
                </p>
              </button>
            </li>
          ))}
        </ul>
      </section>
    </main>
  )
}
