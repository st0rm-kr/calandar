import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
} from './lib/notifications'
import type { Notification, NotificationList } from './lib/notifications'
import { LoadingState } from './components/LoadingState'
import { EmptyState } from './components/EmptyState'

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
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">通知</h1>
        <Link className="text-sm font-semibold text-brand" to="/inbox">
          收件箱
        </Link>
      </div>

      {list && list.unread_count > 0 ? (
        <button
          className="cursor-pointer rounded-full border border-hairline bg-surface px-3 py-1 text-sm font-bold text-muted shadow-card transition-colors duration-200 hover:text-ink"
          onClick={handleReadAll}
          type="button"
        >
          全部已读 ({list.unread_count})
        </button>
      ) : null}

      {error ? <p className="text-sm font-semibold text-rose">{error}</p> : null}

      {!list && !error ? <LoadingState /> : null}

      {list && list.items.length === 0 ? (
        <EmptyState title="还没有通知" description="有新动态时会在这里提醒你。" />
      ) : null}

      <ul className="space-y-2">
        {list?.items.map((notification) => (
          <li key={notification.id}>
            <button
              className={`w-full cursor-pointer rounded-2xl border bg-surface p-4 text-left shadow-card transition-colors duration-200 ${
                notification.read_at
                  ? 'border-hairline'
                  : 'border-brand/30 bg-brand-soft'
              }`}
              onClick={() => handleOpen(notification)}
              type="button"
            >
              <div className="flex items-center justify-between">
                <span className="font-bold">
                  {typeLabels[notification.type] ?? notification.type}
                </span>
                {!notification.read_at ? (
                  <span className="h-2.5 w-2.5 rounded-full bg-brand" />
                ) : null}
              </div>
              <p className="mt-1 text-xs text-muted/70">
                {formatDateTime(notification.created_at)}
              </p>
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
