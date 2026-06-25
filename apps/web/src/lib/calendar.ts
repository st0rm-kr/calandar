import { supabase } from './supabase'

type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export type CalendarColor = 'blue' | 'green' | 'gray'

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
): Promise<CalendarItem[]> {
  const query = new URLSearchParams({ from, to, filter }).toString()
  const response = await fetch(`/api/calendar?${query}`, {
    headers: await authHeaders(),
  })
  const body = (await response.json()) as ApiEnvelope<CalendarItem[]>
  if (!response.ok || body.error) {
    throw new Error(body.error?.message ?? `Request failed: ${response.status}`)
  }
  return body.data ?? []
}
