import { supabase } from './supabase'

export type ApiEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export async function apiGet<T>(path: string): Promise<T> {
  return apiRequest<T>(path)
}

export async function apiPatch<T>(
  path: string,
  body: Record<string, unknown>,
): Promise<T> {
  return apiRequest<T>(path, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

async function apiRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const headers: Record<string, string> = {}
  const {
    data: { session },
  } = await supabase.auth.getSession()
  if (session) {
    headers.Authorization = `Bearer ${session.access_token}`
  }
  for (const [key, value] of Object.entries(init.headers ?? {})) {
    headers[key] = String(value)
  }

  const response = await fetch(path, { ...init, headers })
  const body = (await response.json()) as ApiEnvelope<T>
  if (!response.ok || body.error) {
    throw new Error(body.error?.message ?? `Request failed: ${response.status}`)
  }
  if (body.data === null) {
    throw new Error('Response data is empty')
  }
  return body.data
}
