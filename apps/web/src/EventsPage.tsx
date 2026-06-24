import { Link } from 'react-router-dom'
import { listMine } from './lib/events'
import type { Event } from './lib/events'
import { useAsync } from './hooks/useAsync'
import { EmptyState } from './components/EmptyState'
import { ErrorState } from './components/ErrorState'
import { LoadingState } from './components/LoadingState'
import { EventTypeBadge } from './components/EventTypeBadge'
import { metaForType } from './lib/eventType'

const statusLabels: Record<string, string> = {
  active: '进行中',
  cancelled: '已取消',
  expired: '已结束',
}

const statusClass: Record<string, string> = {
  active: 'bg-grass-soft text-grass',
  cancelled: 'bg-rose-soft text-rose',
  expired: 'bg-canvas text-muted',
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
        <h1 className="text-2xl font-bold tracking-tight">我的活动</h1>
        <Link
          className="rounded-full bg-brand px-4 py-2 text-sm font-bold text-white transition-colors duration-200 hover:bg-brand-ink"
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
              className="rounded-full bg-brand px-5 py-2.5 text-sm font-bold text-white transition-colors duration-200 hover:bg-brand-ink"
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
            <Link className="block" to={`/e/${event.share_slug}`}>
              <div className="overflow-hidden rounded-3xl border border-hairline bg-surface shadow-card transition-shadow duration-200 hover:shadow-soft">
                <div className={`h-1.5 w-full ${metaForType(event.type).block}`} />
                <div className="p-5">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-lg font-bold leading-tight">
                      {event.title}
                    </span>
                    <EventTypeBadge type={event.type} />
                  </div>
                  <p className="mt-1 text-sm font-semibold text-muted">
                    {formatDateTime(event.start_at)}
                    {event.location ? ` · ${event.location}` : ''}
                  </p>
                  <span
                    className={`mt-3 inline-flex rounded-full px-2.5 py-0.5 text-xs font-bold ${
                      statusClass[event.status] ?? 'bg-canvas text-muted'
                    }`}
                  >
                    {statusLabels[event.status] ?? event.status}
                  </span>
                </div>
              </div>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  )
}
