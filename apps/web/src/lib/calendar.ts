import { supabase } from './supabase'

type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export type CalendarColor = 'blue' | 'green' | 'gray'

export type CalendarSourceType =
  | 'self_schedules'
  | 'self_events'
  | 'friend_schedules'
  | 'group_events'

export type CalendarSubscription = {
  id: string
  type: CalendarSourceType
  label: string
  color: CalendarColor
  enabled: boolean
}

export type CalendarItem = {
  id: number
  kind: 'schedule' | 'event'
  title: string
  start_at: string
  end_at: string | null
  location: string | null
  visibility: string
  color: CalendarColor
  has_conflict: boolean
  event_id: number | null
  viewer_rsvp: 'invited' | 'going' | 'not_going' | 'maybe' | null
  going_count: number
  source_id: string
  source_type: CalendarSourceType
  source_label: string
  source_color: CalendarColor
  participants_preview: Array<{
    id: string
    display_name: string
    avatar_url: string | null
  }> | null
}

export type CalendarFilter = 'all' | 'groups' | 'friends'

async function authHeaders(): Promise<Record<string, string>> {
  const headers: Record<string, string> = {}
  const {
    data: { session },
  } = await supabase.auth.getSession()
  if (session) {
    headers.Authorization = `Bearer ${session.access_token}`
  }
  return headers
}

export async function getCalendar(
  from: string,
  to: string,
  filter: CalendarFilter = 'all',
  sources?: string[],
): Promise<CalendarItem[]> {
  const query = new URLSearchParams({ from, to, filter })
  if (sources) {
    query.set('sources', sources.join(','))
  }
  const response = await fetch(`/api/calendar?${query}`, {
    headers: await authHeaders(),
  })
  const body = (await response.json()) as ApiEnvelope<CalendarItem[]>
  if (!response.ok || body.error) {
    throw new Error(body.error?.message ?? `Request failed: ${response.status}`)
  }
  return body.data ?? []
}

export async function listCalendarSubscriptions(): Promise<
  CalendarSubscription[]
> {
  const response = await fetch('/api/calendar/subscriptions', {
    headers: await authHeaders(),
  })
  const body = (await response.json()) as ApiEnvelope<CalendarSubscription[]>
  if (!response.ok || body.error) {
    throw new Error(body.error?.message ?? `Request failed: ${response.status}`)
  }
  return body.data ?? []
}
