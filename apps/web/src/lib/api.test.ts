import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiGet } from './api'

describe('apiGet', () => {
  afterEach(() => {
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
    expect(fetch).toHaveBeenCalledWith('/api/health')
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
