import { Link } from 'react-router-dom'
import { listMine } from './lib/events'
import type { Event } from './lib/events'
import { useAsync } from './hooks/useAsync'
import { Card } from './components/Card'
import { EmptyState } from './components/EmptyState'
import { ErrorState } from './components/ErrorState'
import { LoadingState } from './components/LoadingState'
import { EventTypeBadge } from './components/EventTypeBadge'

const statusLabels: Record<string, string> = {
  active: '进行中',
  cancelled: '已取消',
  expired: '已结束',
}

function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

export default function EventsPage() {
  const { data, error, loading, reload } = useAsync<Event[]>(() => listMine(), [])

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">我的活动</h1>
        <Link
          className="rounded-full bg-white px-4 py-2 text-sm font-medium text-neutral-950"
          to="/events/new"
        >
          + 活动
        </Link>
      </div>

      {loading ? <LoadingState /> : null}
      {error ? <ErrorState message={error} onRetry={reload} /> : null}

      {data && data.length === 0 ? (
        <EmptyState
          title="还没有活动"
          description="创建你的第一个活动，邀请朋友一起来。"
          action={
            <Link
              className="rounded-full bg-white px-4 py-2 text-sm font-medium text-neutral-950"
              to="/events/new"
            >
              创建活动
            </Link>
          }
        />
      ) : null}

      <ul className="space-y-3">
        {data?.map((event) => (
          <li key={event.id}>
            <Link to={`/e/${event.share_slug}`}>
              <Card className="transition hover:bg-white/10">
                <div className="flex items-center justify-between gap-2">
                  <span className="font-medium">{event.title}</span>
                  <EventTypeBadge type={event.type} />
                </div>
                <p className="mt-1 text-sm text-white/50">
                  {formatDateTime(event.start_at)}
                  {event.location ? ` · ${event.location}` : ''}
                </p>
                <p className="mt-1 text-xs text-white/30">
                  {statusLabels[event.status] ?? event.status}
                </p>
              </Card>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  )
}
