import { useEffect, useState } from 'react'
import { apiGet } from './lib/api'
import './index.css'

type Health = { status: string }

export default function App() {
  const [status, setStatus] = useState('checking')

  useEffect(() => {
    apiGet<Health>('/api/health')
      .then((data) => setStatus(data.status))
      .catch(() => setStatus('unavailable'))
  }, [])

  return (
    <main className="min-h-screen bg-neutral-950 px-5 py-8 text-white">
      <section className="mx-auto max-w-md rounded-3xl bg-white/10 p-6 shadow-xl">
        <p className="text-sm text-white/60">Hangout</p>
        <h1 className="mt-3 text-3xl font-semibold">社交日历</h1>
        <p className="mt-4 text-white/70">API status: {status}</p>
      </section>
    </main>
  )
}
