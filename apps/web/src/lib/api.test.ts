import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { apiGet } from './api'
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
