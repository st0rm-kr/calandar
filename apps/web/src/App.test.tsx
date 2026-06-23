import { render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { apiGet } from './lib/api'

vi.mock('./lib/api', () => ({
  apiGet: vi.fn(),
}))

const mockedApiGet = vi.mocked(apiGet)

describe('App', () => {
  beforeEach(() => {
    mockedApiGet.mockReset()
  })

  it('renders the health status returned by the API', async () => {
    mockedApiGet.mockResolvedValueOnce({ status: 'ok' })

    render(<App />)

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

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('API status: unavailable')).toBeInTheDocument()
    })
  })
})
