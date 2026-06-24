import { supabase } from './supabase'

type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export type Group = {
  id: number
  name: string
  description: string | null
  owner_id: string
  invite_code: string
  status: string
}

export type GroupMember = {
  user_id: string
  display_name: string
  avatar_url: string | null
  email: string | null
  role: string
}

export type GroupDetail = {
  group: Group
  members: GroupMember[]
}

export type GroupInvite = {
  id: number
  group_id: number
  inviter_id: string
  invitee_id: string
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

export async function listGroups(): Promise<Group[]> {
  return apiRequest<Group[]>('/api/groups', {
    headers: await authHeaders(),
  })
}

export async function createGroup(
  name: string,
  description: string | null,
): Promise<Group> {
  return apiRequest<Group>('/api/groups', {
    method: 'POST',
    headers: await authHeaders(true),
    body: JSON.stringify({ name, description }),
  })
}

export async function getGroup(groupID: number): Promise<GroupDetail> {
  return apiRequest<GroupDetail>(`/api/groups/${groupID}`, {
    headers: await authHeaders(),
  })
}

export async function joinGroup(inviteCode: string): Promise<Group> {
  return apiRequest<Group>('/api/groups/join', {
    method: 'POST',
    headers: await authHeaders(true),
    body: JSON.stringify({ invite_code: inviteCode }),
  })
}

export async function inviteToGroup(
  groupID: number,
  inviteeID: string,
): Promise<GroupInvite> {
  return apiRequest<GroupInvite>('/api/group-invites', {
    method: 'POST',
    headers: await authHeaders(true),
    body: JSON.stringify({ group_id: groupID, invitee_id: inviteeID }),
  })
}

export async function acceptGroupInvite(inviteID: number): Promise<void> {
  await apiRequest<unknown>(`/api/group-invites/${inviteID}/accept`, {
    method: 'POST',
    headers: await authHeaders(true),
  })
}

export async function rejectGroupInvite(inviteID: number): Promise<void> {
  await apiRequest<unknown>(`/api/group-invites/${inviteID}/reject`, {
    method: 'POST',
    headers: await authHeaders(true),
  })
}

export async function leaveGroup(
  groupID: number,
  userID: string,
): Promise<void> {
  await apiRequest<unknown>(`/api/groups/${groupID}/members/${userID}`, {
    method: 'DELETE',
    headers: await authHeaders(),
  })
}

export async function dissolveGroup(groupID: number): Promise<void> {
  await apiRequest<unknown>(`/api/groups/${groupID}`, {
    method: 'DELETE',
    headers: await authHeaders(),
  })
}
