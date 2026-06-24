import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes, useNavigate } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import EventDetailPage from './EventDetailPage'
import { apiGet } from './lib/api'
import type { EventDetail } from './lib/events'
import { getEventBySlug } from './lib/events'
import { getCalendar } from './lib/calendar'
import { checkConflicts } from './lib/schedules'

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
const mockedCheckConflicts = vi.mocked(checkConflicts)

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

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((promiseResolve) => {
    resolve = promiseResolve
  })
  return { promise, resolve }
}

function renderApp(path = '/') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>,
  )
}

function EventRouteHarness() {
  const navigate = useNavigate()

  return (
    <>
      <button onClick={() => navigate('/events/second')} type="button">
        切换活动
      </button>
      <Routes>
        <Route path="/events/:slug" element={<EventDetailPage />} />
      </Routes>
    </>
  )
}

describe('App', () => {
  beforeEach(() => {
    mockedApiGet.mockReset()
    mockedGetEventBySlug.mockReset()
  })

  it('renders the health status returned by the API', async () => {
    mockedApiGet.mockResolvedValueOnce({ status: 'ok' })

    renderApp()

    expect(screen.getByText('Hangout')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: '社交日历' })).toBeInTheDocument()
    expect(screen.getByText('API status: checking')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText('API status: ok')).toBeInTheDocument()
    })
    expect(mockedApiGet).toHaveBeenCalledWith('/api/health')
  })

  it('renders unavailable when the health check fails', async () => {
    mockedApiGet.mockRejectedValueOnce(new Error('offline'))

    renderApp()

    await waitFor(() => {
      expect(screen.getByText('API status: unavailable')).toBeInTheDocument()
    })
  })

  it('renders the login route', () => {
    renderApp('/login')

    expect(screen.getByRole('heading', { name: '登录 Hangout' })).toBeInTheDocument()
    expect(screen.getByLabelText('邮箱')).toBeInTheDocument()
    expect(screen.getByLabelText('密码')).toBeInTheDocument()
  })

  it('loads the current user profile route', async () => {
    mockedApiGet.mockResolvedValueOnce({
      id: 'user-1',
      display_name: 'Lily',
      avatar_url: null,
      email: 'lily@example.com',
      status: 'active',
    })

    renderApp('/profile')

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Lily' })).toBeInTheDocument()
    })
    expect(mockedApiGet).toHaveBeenCalledWith('/api/users/me')
  })

  it('renders the new event route', () => {
    renderApp('/events/new')

    expect(screen.getByRole('heading', { name: '创建活动' })).toBeInTheDocument()
    expect(screen.getByLabelText('标题')).toBeInTheDocument()
    expect(screen.getByLabelText('开始时间')).toBeInTheDocument()
  })

  it('loads the event detail route by slug', async () => {
    mockedGetEventBySlug.mockResolvedValueOnce(
      createEventDetail('abc123def4', '周末攀岩'),
    )

    renderApp('/events/abc123def4')

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: '周末攀岩' })).toBeInTheDocument()
    })
    expect(mockedGetEventBySlug).toHaveBeenCalledWith('abc123def4')
  })

  it('ignores stale event detail requests after route changes', async () => {
    const firstRequest = deferred<EventDetail>()
    const secondRequest = deferred<EventDetail>()
    mockedGetEventBySlug
      .mockReturnValueOnce(firstRequest.promise)
      .mockReturnValueOnce(secondRequest.promise)

    render(
      <MemoryRouter initialEntries={['/events/first']}>
        <EventRouteHarness />
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(mockedGetEventBySlug).toHaveBeenCalledWith('first')
    })

    fireEvent.click(screen.getByRole('button', { name: '切换活动' }))

    await waitFor(() => {
      expect(mockedGetEventBySlug).toHaveBeenCalledWith('second')
    })

    secondRequest.resolve(createEventDetail('second', '新的活动'))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: '新的活动' })).toBeInTheDocument()
    })

    firstRequest.resolve(createEventDetail('first', '过期活动'))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: '新的活动' })).toBeInTheDocument()
    })
    expect(screen.queryByRole('heading', { name: '过期活动' })).not.toBeInTheDocument()
  })

  it('renders the calendar route with items', async () => {
    mockedGetCalendar.mockResolvedValueOnce([
      {
        id: 1,
        kind: 'schedule',
        title: '晨跑',
        start_at: '2026-06-15T01:00:00Z',
        end_at: '2026-06-15T02:00:00Z',
        location: null,
        visibility: 'public',
        color: 'green',
        has_conflict: false,
        event_id: null,
      },
    ])

    renderApp('/calendar')

    await waitFor(() => {
      expect(mockedGetCalendar).toHaveBeenCalled()
    })
    expect(screen.getByText('群活动')).toBeInTheDocument()
    expect(screen.getByText('个人日程')).toBeInTheDocument()
  })

  it('renders the new schedule editor route', async () => {
    mockedCheckConflicts.mockResolvedValue([])

    renderApp('/schedules/new')

    expect(screen.getByRole('heading', { name: '新建日程' })).toBeInTheDocument()
    expect(screen.getByLabelText('标题')).toBeInTheDocument()
    expect(screen.getByLabelText('开始时间')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: '保存日程' })).toBeEnabled()
    })
  })
})
