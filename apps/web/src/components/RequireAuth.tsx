import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import type { SessionState } from '../hooks/useSession'
import { LoadingState } from './LoadingState'

type RequireAuthProps = {
  session: SessionState
  children: ReactNode
}

export function RequireAuth({ session, children }: RequireAuthProps) {
  const location = useLocation()

  if (session.loading) {
    return (
      <div className="min-h-screen bg-canvas text-ink">
        <LoadingState />
      </div>
    )
  }

  if (!session.session) {
    const redirect = encodeURIComponent(
      location.pathname + location.search,
    )
    return <Navigate replace to={`/login?redirect=${redirect}`} />
  }

  return <>{children}</>
}
