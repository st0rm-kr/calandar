import { supabase } from './supabase'

type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export type Notification = {
  id: number
  user_id: string
  type: string
  payload: Record<string, unknown>
  read_at: string | null
  created_at: string
}

export type NotificationList = {
  unread_count: number
  items: Notification[]
}

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

export async function listNotifications(): Promise<NotificationList> {
  return apiRequest<NotificationList>('/api/notifications', {
    headers: await authHeaders(),
  })
}

export async function markNotificationRead(id: number): Promise<void> {
  await apiRequest<unknown>(`/api/notifications/${id}/read`, {
    method: 'POST',
    headers: await authHeaders(),
  })
}

export async function markAllNotificationsRead(): Promise<void> {
  await apiRequest<unknown>('/api/notifications/read-all', {
    method: 'POST',
    headers: await authHeaders(),
  })
}
