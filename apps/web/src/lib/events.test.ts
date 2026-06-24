import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { supabase } from './supabase'
import {
  cancelEvent,
  createEvent,
  getEventBySlug,
  listMine,
  rsvpEvent,
} from './events'

vi.mock('./supabase', () => ({
  supabase: {
    auth: {
      getSession: vi.fn(),
    },
  },
}))

describe('events API', () => {
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

  it('creates an event with an authenticated POST request', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: {
            id: 12,
            title: 'Dinner',
            type: 'dinner',
            scope: 'public',
            start_at: '2026-06-24T12:00:00Z',
            end_at: null,
            location: null,
            capacity: null,
            status: 'active',
            share_slug: 'abc123def4',
          },
          error: null,
        }),
      ),
    )

    await createEvent({
      title: 'Dinner',
      type: 'dinner',
      scope: 'public',
      start_at: '2026-06-24T12:00:00Z',
      end_at: null,
      location: null,
      capacity: null,
    })

    expect(fetch).toHaveBeenCalledWith('/api/events', {
      method: 'POST',
      headers: {
        Authorization: 'Bearer test-token',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        title: 'Dinner',
        type: 'dinner',
        scope: 'public',
        start_at: '2026-06-24T12:00:00Z',
        end_at: null,
        location: null,
        capacity: null,
      }),
    })
  })

  it('loads an anonymous event detail by share slug', async () => {
    getSession.mockResolvedValue({
      data: { session: null },
      error: null,
    } as Awaited<ReturnType<typeof supabase.auth.getSession>>)
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: {
            event: {
              id: 12,
              title: 'Dinner',
              type: 'dinner',
              scope: 'public',
              start_at: '2026-06-24T12:00:00Z',
              end_at: null,
              location: null,
              capacity: 8,
              status: 'active',
              share_slug: 'abc123def4',
            },
            going_count: 3,
          },
          error: null,
        }),
      ),
    )

    await expect(getEventBySlug('abc123def4')).resolves.toMatchObject({
      going_count: 3,
      event: { id: 12 },
    })
    expect(fetch).toHaveBeenCalledWith('/api/events/abc123def4', {
      headers: {},
    })
  })

  it('sends RSVP choices and cancels owned events', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: {
            participant: {
              id: 1,
              event_id: 12,
              rsvp: 'going',
              add_to_calendar: true,
            },
            conflicts: [],
          },
          error: null,
        }),
      ),
    )

    await rsvpEvent(12, {
      rsvp: 'going',
      add_to_calendar: true,
      visibility: 'busy',
    })
    await cancelEvent(12)

    expect(fetch).toHaveBeenNthCalledWith(1, '/api/events/12/rsvp', {
      method: 'POST',
      headers: {
        Authorization: 'Bearer test-token',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        rsvp: 'going',
        add_to_calendar: true,
        visibility: 'busy',
      }),
    })
    expect(fetch).toHaveBeenNthCalledWith(2, '/api/events/12/cancel', {
      method: 'POST',
      headers: {
        Authorization: 'Bearer test-token',
        'Content-Type': 'application/json',
      },
    })
  })

  it('lists owned events', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () =>
        Response.json({
          data: [],
          error: null,
        }),
      ),
    )

    await expect(listMine()).resolves.toEqual([])
    expect(fetch).toHaveBeenCalledWith('/api/events/mine', {
      headers: { Authorization: 'Bearer test-token' },
    })
  })
})
