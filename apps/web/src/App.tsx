import { useEffect, useState } from 'react'
import { Link, Route, Routes } from 'react-router-dom'
import { apiGet } from './lib/api'
import CalendarPage from './CalendarPage'
import EventDetailPage from './EventDetailPage'
import FriendsPage from './FriendsPage'
import GroupDetailPage from './GroupDetailPage'
import GroupsPage from './GroupsPage'
import InboxPage from './InboxPage'
import LoginPage from './LoginPage'
import NewEventPage from './NewEventPage'
import NotificationsPage from './NotificationsPage'
import ProfilePage from './ProfilePage'
import ScheduleEditorPage from './ScheduleEditorPage'
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
        <nav className="mt-6 flex flex-wrap gap-3 text-sm">
          <Link className="rounded-full bg-white px-4 py-2 font-medium text-neutral-950" to="/calendar">
            日历
          </Link>
          <Link className="rounded-full border border-white/30 px-4 py-2 text-white" to="/login">
            登录
          </Link>
          <Link className="rounded-full border border-white/30 px-4 py-2 text-white" to="/events/new">
            创建活动
          </Link>
          <Link className="rounded-full border border-white/30 px-4 py-2 text-white" to="/friends">
            好友
          </Link>
          <Link className="rounded-full border border-white/30 px-4 py-2 text-white" to="/groups">
            群组
          </Link>
          <Link className="rounded-full border border-white/30 px-4 py-2 text-white" to="/inbox">
            收件箱
          </Link>
          <Link className="rounded-full border border-white/30 px-4 py-2 text-white" to="/notifications">
            通知
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
      <Route path="/calendar" element={<CalendarPage />} />
      <Route path="/schedules/new" element={<ScheduleEditorPage />} />
      <Route path="/schedules/:id" element={<ScheduleEditorPage />} />
      <Route path="/events/new" element={<NewEventPage />} />
      <Route path="/events/:slug" element={<EventDetailPage />} />
      <Route path="/friends" element={<FriendsPage />} />
      <Route path="/groups" element={<GroupsPage />} />
      <Route path="/groups/:id" element={<GroupDetailPage />} />
      <Route path="/inbox" element={<InboxPage />} />
      <Route path="/notifications" element={<NotificationsPage />} />
    </Routes>
  )
}
