import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { supabase } from './supabase'
import {
  checkConflicts,
  createSchedule,
  deleteSchedule,
  listSchedules,
  updateSchedule,
} from './schedules'

vi.mock('./supabase', () => ({
  supabase: {
    auth: {
      getSession: vi.fn(),
    },
  },
}))

describe('schedules API', () => {
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

  it('lists schedules within a range with auth header', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ data: [], error: null })),
    )

    await listSchedules('2026-06-01T00:00:00Z', '2026-07-01T00:00:00Z')

    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/schedules?'),
      {
        headers: { Authorization: 'Bearer test-token' },
      },
    )
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('from=2026-06-01'),
      expect.anything(),
    )
  })

  it('creates a schedule with a POST request', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: {
            schedule: {
              id: 1,
              title: 'Gym',
              start_at: '2026-06-24T09:00:00Z',
              end_at: '2026-06-24T10:00:00Z',
              location: null,
              visibility: 'busy_only',
              source: 'manual',
              event_id: null,
            },
            conflicts: [],
          },
          error: null,
        }),
      ),
    )

    const result = await createSchedule({
      title: 'Gym',
      start_at: '2026-06-24T09:00:00Z',
      end_at: '2026-06-24T10:00:00Z',
      location: null,
      visibility: 'busy_only',
    })

    expect(result.schedule.id).toBe(1)
    expect(fetch).toHaveBeenCalledWith('/api/schedules', {
      method: 'POST',
      headers: {
        Authorization: 'Bearer test-token',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        title: 'Gym',
        start_at: '2026-06-24T09:00:00Z',
        end_at: '2026-06-24T10:00:00Z',
        location: null,
        visibility: 'busy_only',
      }),
    })
  })

  it('updates a schedule with a PATCH request', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: {
            schedule: {
              id: 5,
              title: 'Updated',
              start_at: '2026-06-24T09:00:00Z',
              end_at: null,
              location: null,
              visibility: 'public',
              source: 'manual',
              event_id: null,
            },
            conflicts: [],
          },
          error: null,
        }),
      ),
    )

    await updateSchedule(5, {
      title: 'Updated',
      start_at: '2026-06-24T09:00:00Z',
      end_at: null,
      location: null,
      visibility: 'public',
    })

    expect(fetch).toHaveBeenCalledWith(
      '/api/schedules/5',
      expect.objectContaining({ method: 'PATCH' }),
    )
  })

  it('deletes a schedule with a DELETE request', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ data: { deleted: true }, error: null })),
    )

    await deleteSchedule(7)

    expect(fetch).toHaveBeenCalledWith(
      '/api/schedules/7',
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('requests conflicts with start, end and exclude_id', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ data: [], error: null })),
    )

    await checkConflicts('2026-06-24T09:00:00Z', '2026-06-24T10:00:00Z', 42)

    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/schedules/conflicts?'),
      expect.anything(),
    )
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('exclude_id=42'),
      expect.anything(),
    )
  })
})
