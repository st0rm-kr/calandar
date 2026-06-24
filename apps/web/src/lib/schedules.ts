import { supabase } from './supabase'

type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export type ScheduleVisibility = 'public' | 'busy_only' | 'private'

export type Schedule = {
  id: number
  user_id?: string
  title: string
  start_at: string
  end_at: string | null
  location: string | null
  visibility: ScheduleVisibility
  source: 'manual' | 'event'
  event_id: number | null
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

export type ScheduleResult = {
  schedule: Schedule
  conflicts: ScheduleConflict[]
}

export type ScheduleInput = {
  title: string
  start_at: string
  end_at: string | null
  location: string | null
  visibility: ScheduleVisibility
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

async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
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

export async function listSchedules(
  from: string,
  to: string,
): Promise<Schedule[]> {
  const query = new URLSearchParams({ from, to }).toString()
  return apiRequest<Schedule[]>(`/api/schedules?${query}`, {
    headers: await authHeaders(),
  })
}

export async function createSchedule(
  input: ScheduleInput,
): Promise<ScheduleResult> {
  return apiRequest<ScheduleResult>('/api/schedules', {
    method: 'POST',
    headers: await authHeaders(true),
    body: JSON.stringify(input),
  })
}

export async function updateSchedule(
  scheduleID: number,
  input: ScheduleInput,
): Promise<ScheduleResult> {
  return apiRequest<ScheduleResult>(`/api/schedules/${scheduleID}`, {
    method: 'PATCH',
    headers: await authHeaders(true),
    body: JSON.stringify(input),
  })
}

export async function deleteSchedule(scheduleID: number): Promise<void> {
  await apiRequest<{ deleted: boolean }>(`/api/schedules/${scheduleID}`, {
    method: 'DELETE',
    headers: await authHeaders(),
  })
}

export async function checkConflicts(
  start: string,
  end: string | null,
  excludeID?: number,
): Promise<ScheduleConflict[]> {
  const params = new URLSearchParams({ start })
  if (end) {
    params.set('end', end)
  }
  if (excludeID) {
    params.set('exclude_id', String(excludeID))
  }
  return apiRequest<ScheduleConflict[]>(
    `/api/schedules/conflicts?${params.toString()}`,
    {
      headers: await authHeaders(),
    },
  )
}
