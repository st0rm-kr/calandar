import { supabase } from './supabase'

type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export type Friend = {
  user_id: string
  display_name: string
  avatar_url: string | null
  email: string | null
}

export type FriendRequest = {
  id: number
  requester_id: string
  addressee_id: string
  status: string
  direction: 'incoming' | 'outgoing'
}

export type UserSearchResult = {
  id: string
  display_name: string
  avatar_url: string | null
  email: string | null
  status: string
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

export async function listFriends(): Promise<Friend[]> {
  return apiRequest<Friend[]>('/api/friends', {
    headers: await authHeaders(),
  })
}

export async function listFriendRequests(): Promise<FriendRequest[]> {
  return apiRequest<FriendRequest[]>('/api/friends/requests', {
    headers: await authHeaders(),
  })
}

export async function searchUsers(query: string): Promise<UserSearchResult[]> {
  const params = new URLSearchParams({ q: query }).toString()
  return apiRequest<UserSearchResult[]>(`/api/users/search?${params}`, {
    headers: await authHeaders(),
  })
}

export async function sendFriendRequest(addresseeID: string): Promise<void> {
  await apiRequest<unknown>('/api/friends/requests', {
    method: 'POST',
    headers: await authHeaders(true),
    body: JSON.stringify({ addressee_id: addresseeID }),
  })
}

export async function acceptFriendRequest(requestID: number): Promise<void> {
  await apiRequest<unknown>(`/api/friends/requests/${requestID}/accept`, {
    method: 'POST',
    headers: await authHeaders(true),
  })
}

export async function rejectFriendRequest(requestID: number): Promise<void> {
  await apiRequest<unknown>(`/api/friends/requests/${requestID}/reject`, {
    method: 'POST',
    headers: await authHeaders(true),
  })
}

export async function deleteFriend(userID: string): Promise<void> {
  await apiRequest<unknown>(`/api/friends/${userID}`, {
    method: 'DELETE',
    headers: await authHeaders(),
  })
}
