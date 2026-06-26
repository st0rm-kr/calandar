import { useEffect } from 'react'
import { Route, Routes } from 'react-router-dom'
import CalendarPage from './CalendarPage'
import EventDetailPage from './EventDetailPage'
import EventsPage from './EventsPage'
import FriendsPage from './FriendsPage'
import GroupDetailPage from './GroupDetailPage'
import GroupsPage from './GroupsPage'
import InboxPage from './InboxPage'
import LoginPage from './LoginPage'
import MePage from './MePage'
import NewEventPage from './NewEventPage'
import NotFoundPage from './NotFoundPage'
import NotificationsPage from './NotificationsPage'
import ProfilePage from './ProfilePage'
import ScheduleEditorPage from './ScheduleEditorPage'
import SchedulesPage from './SchedulesPage'
import { AppShell } from './components/AppShell'
import { RequireAuth } from './components/RequireAuth'
import { useSession } from './hooks/useSession'
import { getInbox } from './lib/inbox'
import { subscribeInboxRefresh } from './lib/inboxRefresh'
import { useAsync } from './hooks/useAsync'
import './index.css'

function AuthedShell({
  session,
  children,
  showTabs = true,
  showFab = false,
}: {
  session: ReturnType<typeof useSession>
  children: React.ReactNode
  showTabs?: boolean
  showFab?: boolean
}) {
  const inbox = useAsync(() => getInbox(), [session.session?.access_token])

  useEffect(() => subscribeInboxRefresh(inbox.reload), [inbox.reload])

  return (
    <RequireAuth session={session}>
      <AppShell
        inboxCount={inbox.data?.counts.total ?? 0}
        showFab={showFab}
        showTabs={showTabs}
      >
        {children}
      </AppShell>
    </RequireAuth>
  )
}

export default function App() {
  const session = useSession()

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/e/:slug" element={<EventDetailPage session={session} />} />
      <Route
        path="/"
        element={
          <AuthedShell session={session}>
            <CalendarPage />
          </AuthedShell>
        }
      />
      <Route
        path="/calendar"
        element={
          <AuthedShell session={session}>
            <CalendarPage />
          </AuthedShell>
        }
      />
      <Route
        path="/events"
        element={
          <AuthedShell session={session}>
            <EventsPage />
          </AuthedShell>
        }
      />
      <Route
        path="/events/new"
        element={
          <AuthedShell session={session} showTabs={false}>
            <NewEventPage />
          </AuthedShell>
        }
      />
      <Route
        path="/events/:slug"
        element={<EventDetailPage session={session} />}
      />
      <Route
        path="/schedules"
        element={
          <AuthedShell session={session}>
            <SchedulesPage />
          </AuthedShell>
        }
      />
      <Route
        path="/schedules/new"
        element={
          <AuthedShell session={session} showTabs={false}>
            <ScheduleEditorPage />
          </AuthedShell>
        }
      />
      <Route
        path="/schedules/:id"
        element={
          <AuthedShell session={session} showTabs={false}>
            <ScheduleEditorPage />
          </AuthedShell>
        }
      />
      <Route
        path="/inbox"
        element={
          <AuthedShell session={session}>
            <InboxPage />
          </AuthedShell>
        }
      />
      <Route
        path="/friends"
        element={
          <AuthedShell session={session}>
            <FriendsPage />
          </AuthedShell>
        }
      />
      <Route
        path="/groups"
        element={
          <AuthedShell session={session}>
            <GroupsPage />
          </AuthedShell>
        }
      />
      <Route
        path="/groups/:id"
        element={
          <AuthedShell session={session} showTabs={false}>
            <GroupDetailPage />
          </AuthedShell>
        }
      />
      <Route
        path="/notifications"
        element={
          <AuthedShell session={session} showTabs={false}>
            <NotificationsPage />
          </AuthedShell>
        }
      />
      <Route
        path="/profile"
        element={
          <AuthedShell session={session} showTabs={false}>
            <ProfilePage />
          </AuthedShell>
        }
      />
      <Route
        path="/me"
        element={
          <AuthedShell session={session}>
            <MePage session={session} />
          </AuthedShell>
        }
      />
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
