import { useEffect, useState } from 'react'
import type { Session } from '@supabase/supabase-js'
import { supabase } from '../lib/supabase'
import { apiGet } from '../lib/api'

export type SessionProfile = {
  id: string
  display_name: string
  avatar_url: string | null
  email: string | null
  status: string
}

export type SessionState = {
  session: Session | null
  profile: SessionProfile | null
  loading: boolean
  signOut: () => Promise<void>
}

export function useSession(): SessionState {
  const [session, setSession] = useState<Session | null>(null)
  const [profile, setProfile] = useState<SessionProfile | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let active = true

    supabase.auth
      .getSession()
      .then(({ data }) => {
        if (!active) {
          return
        }
        setSession(data.session)
        setLoading(false)
      })
      .catch(() => {
        if (active) {
          setLoading(false)
        }
      })

    const { data: subscription } = supabase.auth.onAuthStateChange(
      (_event, nextSession) => {
        if (!active) {
          return
        }
        setSession(nextSession)
        setLoading(false)
      },
    )

    return () => {
      active = false
      subscription.subscription.unsubscribe()
    }
  }, [])

  useEffect(() => {
    let active = true
    if (!session) {
      Promise.resolve().then(() => {
        if (active) {
          setProfile(null)
        }
      })
      return () => {
        active = false
      }
    }
    apiGet<SessionProfile>('/api/users/me')
      .then((data) => {
        if (active) {
          setProfile(data)
        }
      })
      .catch(() => {
        if (active) {
          setProfile(null)
        }
      })
    return () => {
      active = false
    }
  }, [session])

  async function signOut() {
    await supabase.auth.signOut()
    setSession(null)
    setProfile(null)
  }

  return { session, profile, loading, signOut }
}
