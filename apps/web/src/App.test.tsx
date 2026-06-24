import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { apiGet } from './lib/api'

vi.mock('./lib/api', () => ({
  apiGet: vi.fn(),
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
})
