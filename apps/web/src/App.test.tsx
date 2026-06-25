import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { apiGet } from './lib/api'
import type { EventDetail } from './lib/events'
import { getEventBySlug, rsvpEvent } from './lib/events'
import { getCalendar } from './lib/calendar'
import { getInbox } from './lib/inbox'
import { supabase } from './lib/supabase'

vi.mock('./lib/api', () => ({
  apiGet: vi.fn(),
}))
vi.mock('./lib/events', () => ({
  getEventBySlug: vi.fn(),
  createEvent: vi.fn(),
  rsvpEvent: vi.fn(),
  cancelEvent: vi.fn(),
  listMine: vi.fn(),
}))
vi.mock('./lib/calendar', () => ({
  getCalendar: vi.fn(),
}))
vi.mock('./lib/inbox', () => ({
  getInbox: vi.fn(),
}))
vi.mock('./lib/schedules', () => ({
  listSchedules: vi.fn(),
  createSchedule: vi.fn(),
  updateSchedule: vi.fn(),
  deleteSchedule: vi.fn(),
  checkConflicts: vi.fn(),
}))
vi.mock('./lib/supabase', () => ({
  supabase: {
    auth: {
      getSession: vi.fn(),
      onAuthStateChange: vi.fn(),
      signInWithPassword: vi.fn(),
      signUp: vi.fn(),
      resetPasswordForEmail: vi.fn(),
      signOut: vi.fn(),
    },
  },
}))

const mockedApiGet = vi.mocked(apiGet)
const mockedGetEventBySlug = vi.mocked(getEventBySlug)
const mockedGetCalendar = vi.mocked(getCalendar)
const mockedGetInbox = vi.mocked(getInbox)
const mockedRsvpEvent = vi.mocked(rsvpEvent)
const mockedGetSession = vi.mocked(supabase.auth.getSession)
const mockedOnAuthStateChange = vi.mocked(supabase.auth.onAuthStateChange)

type FakeSession = { access_token: string }

function setSession(session: FakeSession | null) {
  mockedGetSession.mockResolvedValue({
    data: { session },
    error: null,
  } as Awaited<ReturnType<typeof supabase.auth.getSession>>)
}

function createEventDetail(slug: string, title: string): EventDetail {
  return {
    event: {
      id: 12,
      title,
      type: 'climbing',
      scope: 'personal',
      start_at: '2026-06-24T12:00:00Z',
      end_at: null,
      location: '岩馆',
      capacity: 6,
      status: 'active',
      share_slug: slug,
    },
    going_count: 2,
  }
}

function renderApp(path = '/') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>,
  )
}

describe('App', () => {
  beforeEach(() => {
    mockedApiGet.mockReset()
    mockedGetEventBySlug.mockReset()
    mockedGetCalendar.mockReset()
    mockedGetInbox.mockReset()
    mockedRsvpEvent.mockReset()
    mockedGetSession.mockReset()
    mockedOnAuthStateChange.mockReset()
    mockedOnAuthStateChange.mockReturnValue({
      data: { subscription: { unsubscribe: vi.fn() } },
    } as unknown as ReturnType<typeof supabase.auth.onAuthStateChange>)
    mockedGetInbox.mockResolvedValue({
      counts: {
        total: 0,
        event_invites: 0,
        friend_requests: 0,
        group_invites: 0,
      },
      items: [],
    })
    mockedApiGet.mockResolvedValue({
      id: 'user-1',
      display_name: 'Lily',
      avatar_url: null,
      email: 'lily@example.com',
      status: 'active',
    })
    setSession(null)
  })

  it('redirects unauthenticated users from / to login with redirect', async () => {
    renderApp('/')

    await waitFor(() => {
      expect(
        screen.getByRole('heading', { name: '登录 Hangout' }),
      ).toBeInTheDocument()
    })
  })

  it('does not redirect anonymous users away from a shared event page', async () => {
    mockedGetEventBySlug.mockResolvedValue(
      createEventDetail('abc123def4', '周末攀岩'),
    )

    renderApp('/e/abc123def4')

    await waitFor(() => {
      expect(
        screen.getByRole('heading', { name: '周末攀岩' }),
      ).toBeInTheDocument()
    })
    expect(screen.getByRole('link', { name: '登录以报名' })).toBeInTheDocument()
  })

  it('renders the login route', () => {
    renderApp('/login')

    expect(
      screen.getByRole('heading', { name: '登录 Hangout' }),
    ).toBeInTheDocument()
    expect(screen.getByLabelText('邮箱')).toBeInTheDocument()
    expect(screen.getByLabelText('密码')).toBeInTheDocument()
  })

  it('renders the calendar for authenticated users', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([])

    renderApp('/')

    await waitFor(() => {
      expect(screen.getByText('时间流')).toBeInTheDocument()
    })
  })

  it('loads the event detail route by slug', async () => {
    setSession({ access_token: 'token' })
    mockedGetEventBySlug.mockResolvedValue(
      createEventDetail('abc123def4', '周末攀岩'),
    )

    renderApp('/e/abc123def4')

    await waitFor(() => {
      expect(
        screen.getByRole('heading', { name: '周末攀岩' }),
      ).toBeInTheDocument()
    })
    expect(mockedGetEventBySlug).toHaveBeenCalledWith('abc123def4')
  })

  it('lets authenticated users submit an RSVP', async () => {
    setSession({ access_token: 'token' })
    mockedGetEventBySlug.mockResolvedValue(
      createEventDetail('abc123def4', '周末攀岩'),
    )
    mockedRsvpEvent.mockResolvedValue({
      participant: { id: 1, event_id: 12, rsvp: 'going', add_to_calendar: true },
      conflicts: [],
      going_count: 3,
    })

    renderApp('/e/abc123def4')

    await waitFor(() => {
      expect(screen.getByRole('button', { name: '我要参加' })).toBeEnabled()
    })

    fireEvent.click(screen.getByRole('button', { name: '我要参加' }))

    await waitFor(() => {
      expect(mockedRsvpEvent).toHaveBeenCalledWith(12, {
        rsvp: 'going',
        add_to_calendar: true,
        visibility: 'busy_only',
      })
    })
    expect(await screen.findByText('已加入活动')).toBeInTheDocument()
    expect(screen.getByText('3 / 6 人已参加')).toBeInTheDocument()
  })

  it('renders a 404 for unknown routes', () => {
    renderApp('/does-not-exist')

    expect(screen.getByText('页面不存在')).toBeInTheDocument()
  })
})
