import { supabase } from './supabase'

type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export type InboxCounts = {
  total: number
  event_invites: number
  friend_requests: number
  group_invites: number
}

export type InboxItem = {
  id: string
  kind: 'event_invite' | 'friend_request' | 'group_invite'
  source_id: number
  event_id?: number
  group_id?: number
  title: string
  subtitle: string
  created_at: string
  disabled: boolean
  disabled_reason?: string
  conflict_count?: number
}

export type InboxResult = {
  counts: InboxCounts
  items: InboxItem[]
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

export async function getInbox(): Promise<InboxResult> {
  return apiRequest<InboxResult>('/api/inbox', {
    headers: await authHeaders(),
  })
}
