import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { apiGet } from './lib/api'
import { getEventBySlug } from './lib/events'

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
    mockedGetEventBySlug.mockResolvedValueOnce({
      event: {
        id: 12,
        title: '周末攀岩',
        type: 'climbing',
        scope: 'personal',
        start_at: '2026-06-24T12:00:00Z',
        end_at: null,
        location: '岩馆',
        capacity: 6,
        status: 'active',
        share_slug: 'abc123def4',
      },
      going_count: 2,
    })

    renderApp('/events/abc123def4')

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: '周末攀岩' })).toBeInTheDocument()
    })
    expect(mockedGetEventBySlug).toHaveBeenCalledWith('abc123def4')
  })
})
