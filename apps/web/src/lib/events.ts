import { supabase } from './supabase'

type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export type EventStatus = 'active' | 'cancelled' | 'expired'
export type EventRSVP = 'invited' | 'going' | 'not_going' | 'maybe'

export type Event = {
  id: number
  owner_id?: string
  scope: string
  group_id?: number | null
  title: string
  type: string
  start_at: string
  end_at: string | null
  location: string | null
  capacity: number | null
  status: EventStatus
  share_slug: string
  created_at?: string
  updated_at?: string
}

export type EventDetail = {
  event: Event
  going_count: number
}

export type CreateEventInput = {
  title: string
  type: string
  scope: string
  group_id?: number | null
  start_at: string
  end_at: string | null
  location: string | null
  capacity: number | null
}

export type RSVPInput = {
  rsvp: EventRSVP
  add_to_calendar: boolean
  visibility?: string
}

export type ScheduleConflict = {
  id: number
  title: string
  start_at: string
  end_at: string | null
  location: string | null
  visibility: string
  source: string
  event_id: number | null
}

export type RSVPResult = {
  participant: {
    id: number
    event_id: number
    rsvp: EventRSVP
    add_to_calendar: boolean
  }
  conflicts: ScheduleConflict[]
  going_count: number
}

export type EventParticipant = {
  id: number
  event_id: number
  user_id: string
  rsvp: EventRSVP
  add_to_calendar?: boolean
  created_at?: string
  updated_at?: string
}

async function authHeaders(hasBody = false): Promise<Record<string, string>> {
  const headers: Record<string, string> = {}
  const {
    data: { session },
  } = await supabase.auth.getSession()
  if (session) {
    headers.Authorization = `Bearer ${session.access_token}`
  }
  if (hasBody) {
    headers['Content-Type'] = 'application/json'
  }
  return headers
}

async function apiRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(path, init)
  const body = (await response.json()) as ApiEnvelope<T>
  if (!response.ok || body.error) {
    throw new Error(body.error?.message ?? `Request failed: ${response.status}`)
  }
  if (body.data === null) {
    throw new Error('Response data is empty')
  }
  return body.data
}

export async function createEvent(input: CreateEventInput): Promise<Event> {
  return apiRequest<Event>('/api/events', {
    method: 'POST',
    headers: await authHeaders(true),
    body: JSON.stringify(input),
  })
}

export async function getEventBySlug(slug: string): Promise<EventDetail> {
  return apiRequest<EventDetail>(`/api/events/${encodeURIComponent(slug)}`, {
    headers: await authHeaders(),
  })
}

export async function rsvpEvent(
  eventID: number,
  input: RSVPInput,
): Promise<RSVPResult> {
  return apiRequest<RSVPResult>(`/api/events/${eventID}/rsvp`, {
    method: 'POST',
    headers: await authHeaders(true),
    body: JSON.stringify(input),
  })
}

export async function inviteToEvent(
  eventID: number,
  inviteeIDs: string[],
): Promise<EventParticipant[]> {
  return apiRequest<EventParticipant[]>(`/api/events/${eventID}/invite`, {
    method: 'POST',
    headers: await authHeaders(true),
    body: JSON.stringify({ invitee_ids: inviteeIDs }),
  })
}

export async function cancelEvent(eventID: number): Promise<Event> {
  return apiRequest<Event>(`/api/events/${eventID}/cancel`, {
    method: 'POST',
    headers: await authHeaders(true),
  })
}

export async function listMine(): Promise<Event[]> {
  return apiRequest<Event[]>('/api/events/mine', {
    headers: await authHeaders(),
  })
}
