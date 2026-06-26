import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { apiGet, apiPatch } from './api'
import { supabase } from './supabase'

vi.mock('./supabase', () => ({
  supabase: {
    auth: {
      getSession: vi.fn(),
    },
  },
}))

describe('apiGet', () => {
  const getSession = vi.mocked(supabase.auth.getSession)

  beforeEach(() => {
    getSession.mockResolvedValue({
      data: { session: null },
      error: null,
    } as Awaited<ReturnType<typeof supabase.auth.getSession>>)
  })

  afterEach(() => {
    vi.clearAllMocks()
    vi.unstubAllGlobals()
  })

  it('returns data from a successful API envelope', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: { status: 'ok' },
          error: null,
        }),
      ),
    )

    await expect(apiGet<{ status: string }>('/api/health')).resolves.toEqual({
      status: 'ok',
    })
    expect(fetch).toHaveBeenCalledWith('/api/health', { headers: {} })
  })

  it('adds bearer authorization when a Supabase session exists', async () => {
    getSession.mockResolvedValue({
      data: {
        session: {
          access_token: 'test-token',
        },
      },
      error: null,
    } as Awaited<ReturnType<typeof supabase.auth.getSession>>)
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: { status: 'ok' },
          error: null,
        }),
      ),
    )

    await apiGet<{ status: string }>('/api/health')

    expect(fetch).toHaveBeenCalledWith('/api/health', {
      headers: { Authorization: 'Bearer test-token' },
    })
  })

  it('patches JSON data with authenticated headers', async () => {
    getSession.mockResolvedValue({
      data: {
        session: {
          access_token: 'test-token',
        },
      },
      error: null,
    } as Awaited<ReturnType<typeof supabase.auth.getSession>>)
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: { display_name: 'Lily Chen' },
          error: null,
        }),
      ),
    )

    await expect(
      apiPatch<{ display_name: string }>('/api/users/me', {
        display_name: 'Lily Chen',
      }),
    ).resolves.toEqual({ display_name: 'Lily Chen' })
    expect(fetch).toHaveBeenCalledWith('/api/users/me', {
      method: 'PATCH',
      headers: {
        Authorization: 'Bearer test-token',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ display_name: 'Lily Chen' }),
    })
  })

  it('throws the API error message when the envelope contains an error', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json(
          {
            data: null,
            error: { code: 'unavailable', message: 'service unavailable' },
          },
          { status: 503 },
        ),
      ),
    )

    await expect(apiGet('/api/health')).rejects.toThrow('service unavailable')
  })

  it('throws when the envelope has no data', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: null,
          error: null,
        }),
      ),
    )

    await expect(apiGet('/api/health')).rejects.toThrow('Response data is empty')
  })
})
