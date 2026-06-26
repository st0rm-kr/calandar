import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { getCalendar, listCalendarSubscriptions } from './calendar'
import { supabase } from './supabase'

vi.mock('./supabase', () => ({
  supabase: {
    auth: {
      getSession: vi.fn(),
    },
  },
}))

describe('calendar API', () => {
  const getSession = vi.mocked(supabase.auth.getSession)

  beforeEach(() => {
    getSession.mockResolvedValue({
      data: {
        session: {
          access_token: 'test-token',
        },
      },
      error: null,
    } as Awaited<ReturnType<typeof supabase.auth.getSession>>)
  })

  afterEach(() => {
    vi.clearAllMocks()
    vi.unstubAllGlobals()
  })

  it('loads calendar subscriptions with auth headers', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: [
            {
              id: 'self:schedules',
              type: 'self_schedules',
              label: '自身日程',
              color: 'green',
              enabled: true,
            },
          ],
          error: null,
        }),
      ),
    )

    await expect(listCalendarSubscriptions()).resolves.toHaveLength(1)
    expect(fetch).toHaveBeenCalledWith('/api/calendar/subscriptions', {
      headers: { Authorization: 'Bearer test-token' },
    })
  })

  it('filters calendar items by selected source ids', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: [],
          error: null,
        }),
      ),
    )

    await getCalendar(
      '2026-06-01T00:00:00Z',
      '2026-07-01T00:00:00Z',
      'all',
      ['self:schedules', 'friend:aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa:schedules'],
    )

    expect(fetch).toHaveBeenCalledWith(
      '/api/calendar?from=2026-06-01T00%3A00%3A00Z&to=2026-07-01T00%3A00%3A00Z&filter=all&sources=self%3Aschedules%2Cfriend%3Aaaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa%3Aschedules',
      {
        headers: { Authorization: 'Bearer test-token' },
      },
    )
  })

  it('sends an empty sources query when all subscriptions are unchecked', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ data: [], error: null })),
    )

    await getCalendar(
      '2026-06-01T00:00:00Z',
      '2026-07-01T00:00:00Z',
      'all',
      [],
    )

    expect(fetch).toHaveBeenCalledWith(
      '/api/calendar?from=2026-06-01T00%3A00%3A00Z&to=2026-07-01T00%3A00%3A00Z&filter=all&sources=',
      {
        headers: { Authorization: 'Bearer test-token' },
      },
    )
  })
})
