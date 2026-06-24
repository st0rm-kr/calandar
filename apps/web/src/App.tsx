import { useEffect, useState } from 'react'
import { Link, Route, Routes } from 'react-router-dom'
import { apiGet } from './lib/api'
import LoginPage from './LoginPage'
import ProfilePage from './ProfilePage'
import './index.css'

type Health = { status: string }

function HomePage() {
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
        <nav className="mt-6 flex gap-3 text-sm">
          <Link className="rounded-full bg-white px-4 py-2 font-medium text-neutral-950" to="/login">
            登录
          </Link>
          <Link className="rounded-full border border-white/30 px-4 py-2 text-white" to="/profile">
            个人页
          </Link>
        </nav>
      </section>
    </main>
  )
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/profile" element={<ProfilePage />} />
    </Routes>
  )
}
