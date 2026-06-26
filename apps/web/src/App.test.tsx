import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { apiGet, apiPatch } from './lib/api'
import type { EventDetail } from './lib/events'
import {
  createEvent,
  getEventBySlug,
  inviteToEvent,
  listMine,
  rsvpEvent,
} from './lib/events'
import { getCalendar, listCalendarSubscriptions } from './lib/calendar'
import type { CalendarItem } from './lib/calendar'
import {
  listFriendRequests,
  listFriends,
} from './lib/friends'
import { getInbox } from './lib/inbox'
import {
  checkConflicts,
  deleteSchedule,
  getSchedule,
  listSchedules,
} from './lib/schedules'
import { supabase } from './lib/supabase'

vi.mock('./lib/api', () => ({
  apiGet: vi.fn(),
  apiPatch: vi.fn(),
}))
vi.mock('./lib/events', () => ({
  getEventBySlug: vi.fn(),
  createEvent: vi.fn(),
  inviteToEvent: vi.fn(),
  rsvpEvent: vi.fn(),
  cancelEvent: vi.fn(),
  listMine: vi.fn(),
}))
vi.mock('./lib/calendar', () => ({
  getCalendar: vi.fn(),
  listCalendarSubscriptions: vi.fn(),
}))
vi.mock('./lib/inbox', () => ({
  getInbox: vi.fn(),
}))
vi.mock('./lib/friends', () => ({
  acceptFriendRequest: vi.fn(),
  deleteFriend: vi.fn(),
  listFriendRequests: vi.fn(),
  listFriends: vi.fn(),
  rejectFriendRequest: vi.fn(),
  searchUsers: vi.fn(),
  sendFriendRequest: vi.fn(),
}))
vi.mock('./lib/groups', () => ({
  acceptGroupInvite: vi.fn(),
  rejectGroupInvite: vi.fn(),
}))
vi.mock('./lib/schedules', () => ({
  listSchedules: vi.fn(),
  getSchedule: vi.fn(),
  createSchedule: vi.fn(),
  updateSchedule: vi.fn(),
  deleteSchedule: vi.fn(),
  checkConflicts: vi.fn(),
}))
vi.mock('./lib/supabase', () => ({
  supabase: {
    auth: {
      getSession: vi.fn(),
      onAuthStateChange: vi.fn(),
      signInWithPassword: vi.fn(),
      signUp: vi.fn(),
      resetPasswordForEmail: vi.fn(),
      signOut: vi.fn(),
    },
  },
}))

const mockedApiGet = vi.mocked(apiGet)
const mockedApiPatch = vi.mocked(apiPatch)
const mockedGetEventBySlug = vi.mocked(getEventBySlug)
const mockedCreateEvent = vi.mocked(createEvent)
const mockedInviteToEvent = vi.mocked(inviteToEvent)
const mockedListMine = vi.mocked(listMine)
const mockedGetCalendar = vi.mocked(getCalendar)
const mockedListCalendarSubscriptions = vi.mocked(listCalendarSubscriptions)
const mockedGetInbox = vi.mocked(getInbox)
const mockedRsvpEvent = vi.mocked(rsvpEvent)
const mockedListFriends = vi.mocked(listFriends)
const mockedListFriendRequests = vi.mocked(listFriendRequests)
const mockedListSchedules = vi.mocked(listSchedules)
const mockedGetSchedule = vi.mocked(getSchedule)
const mockedDeleteSchedule = vi.mocked(deleteSchedule)
const mockedCheckConflicts = vi.mocked(checkConflicts)
const mockedGetSession = vi.mocked(supabase.auth.getSession)
const mockedOnAuthStateChange = vi.mocked(supabase.auth.onAuthStateChange)

type FakeSession = { access_token: string }

function setSession(session: FakeSession | null) {
  mockedGetSession.mockResolvedValue({
    data: { session },
    error: null,
  } as Awaited<ReturnType<typeof supabase.auth.getSession>>)
}

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

function createHomeEvent(overrides: Partial<CalendarItem> = {}): CalendarItem {
  const start = new Date()
  start.setHours(20, 0, 0, 0)
  return {
    id: 21,
    kind: 'event',
    title: '周末攀岩',
    start_at: start.toISOString(),
    end_at: null,
    location: '岩馆',
    visibility: 'public',
    color: 'blue',
    has_conflict: false,
    event_id: 12,
    viewer_rsvp: null,
    going_count: 2,
    source_id: 'self:events',
    source_type: 'self_events',
    source_label: '自身活动',
    source_color: 'blue',
    participants_preview: [
      { id: 'user-2', display_name: 'Alex', avatar_url: null },
      { id: 'user-3', display_name: 'Mia', avatar_url: null },
    ],
    ...overrides,
  }
}

function createHomeSchedule(overrides: Partial<CalendarItem> = {}): CalendarItem {
  const start = new Date()
  start.setHours(10, 0, 0, 0)
  return {
    id: 31,
    kind: 'schedule',
    title: '专注写方案',
    start_at: start.toISOString(),
    end_at: null,
    location: '公司',
    visibility: 'busy_only',
    color: 'green',
    has_conflict: false,
    event_id: null,
    viewer_rsvp: null,
    going_count: 0,
    source_id: 'self:schedules',
    source_type: 'self_schedules',
    source_label: '自身日程',
    source_color: 'green',
    participants_preview: null,
    ...overrides,
  }
}

function renderApp(path = '/') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>,
  )
}

describe('App', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    mockedApiGet.mockReset()
    mockedApiPatch.mockReset()
    mockedGetEventBySlug.mockReset()
    mockedCreateEvent.mockReset()
    mockedInviteToEvent.mockReset()
    mockedListMine.mockReset()
    mockedGetCalendar.mockReset()
    mockedListCalendarSubscriptions.mockReset()
    mockedGetInbox.mockReset()
    mockedRsvpEvent.mockReset()
    mockedListFriends.mockReset()
    mockedListFriendRequests.mockReset()
    mockedListSchedules.mockReset()
    mockedGetSchedule.mockReset()
    mockedDeleteSchedule.mockReset()
    mockedCheckConflicts.mockReset()
    mockedGetSession.mockReset()
    mockedOnAuthStateChange.mockReset()
    mockedOnAuthStateChange.mockReturnValue({
      data: { subscription: { unsubscribe: vi.fn() } },
    } as unknown as ReturnType<typeof supabase.auth.onAuthStateChange>)
    mockedGetInbox.mockResolvedValue({
      counts: {
        total: 0,
        event_invites: 0,
        friend_requests: 0,
        group_invites: 0,
      },
      items: [],
    })
    mockedApiGet.mockResolvedValue({
      id: 'user-1',
      display_name: 'Lily',
      avatar_url: null,
      email: 'lily@example.com',
      status: 'active',
    })
    mockedApiPatch.mockResolvedValue({
      id: 'user-1',
      display_name: 'Lily Chen',
      avatar_url: null,
      email: 'lily@example.com',
      status: 'active',
    })
    mockedListFriends.mockResolvedValue([])
    mockedListCalendarSubscriptions.mockResolvedValue([
      {
        id: 'self:schedules',
        type: 'self_schedules',
        label: '自身日程',
        color: 'green',
        enabled: true,
      },
      {
        id: 'self:events',
        type: 'self_events',
        label: '自身活动',
        color: 'blue',
        enabled: true,
      },
    ])
    mockedListFriendRequests.mockResolvedValue([])
    mockedListMine.mockResolvedValue([])
    mockedListSchedules.mockResolvedValue([])
    mockedDeleteSchedule.mockResolvedValue(undefined)
    mockedCheckConflicts.mockResolvedValue([])
    setSession(null)
  })

  it('redirects unauthenticated users from / to login with redirect', async () => {
    renderApp('/')

    await waitFor(() => {
      expect(
        screen.getByRole('heading', { name: '登录 Hangout' }),
      ).toBeInTheDocument()
    })
  })

  it('does not redirect anonymous users away from a shared event page', async () => {
    mockedGetEventBySlug.mockResolvedValue(
      createEventDetail('abc123def4', '周末攀岩'),
    )

    renderApp('/e/abc123def4')

    await waitFor(() => {
      expect(
        screen.getByRole('heading', { name: '周末攀岩' }),
      ).toBeInTheDocument()
    })
    expect(screen.getByRole('link', { name: '登录以报名' })).toBeInTheDocument()
  })

  it('renders the login route', () => {
    renderApp('/login')

    expect(
      screen.getByRole('heading', { name: '登录 Hangout' }),
    ).toBeInTheDocument()
    expect(screen.getByLabelText('邮箱')).toBeInTheDocument()
    expect(screen.getByLabelText('密码')).toBeInTheDocument()
  })

  it('renders the calendar for authenticated users', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([])

    renderApp('/')

    await waitFor(() => {
      expect(screen.getByText('时间流')).toBeInTheDocument()
    })
    expect(screen.getByRole('link', { name: '新建日程' })).toHaveAttribute(
      'href',
      '/schedules/new',
    )
    expect(screen.getByRole('link', { name: '发起 Hangout' })).toHaveAttribute(
      'href',
      '/events/new',
    )
    expect(screen.queryByLabelText('发起 Hangout')).not.toBeInTheDocument()
  })

  it('replaces coarse calendar filters with calendar view switching', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([createHomeEvent(), createHomeSchedule()])

    renderApp('/')

    expect(
      await screen.findByRole('button', { name: '月视图' }),
    ).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: '周视图' })).toHaveAttribute(
      'aria-pressed',
      'false',
    )
    expect(screen.queryByRole('button', { name: '全部' })).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: '我的群' }),
    ).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '好友' })).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: '周视图' }))

    expect(await screen.findByText('周视图时间表')).toBeInTheDocument()
    expect(screen.getByText('07:00')).toBeInTheDocument()
    expect(screen.getAllByText('自身活动').length).toBeGreaterThan(0)

    fireEvent.click(screen.getByRole('button', { name: '列表' }))

    expect(await screen.findByText('接下来 14 天')).toBeInTheDocument()
    expect(screen.getAllByText('周末攀岩').length).toBeGreaterThan(0)

    fireEvent.click(screen.getByRole('button', { name: '空闲' }))

    expect(await screen.findByText('可约时间')).toBeInTheDocument()
    expect(screen.getAllByText('忙碌时段').length).toBeGreaterThan(0)
    expect(screen.getAllByText(/可约/).length).toBeGreaterThan(0)
  })

  it('loads the calendar view from the URL query', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([createHomeEvent()])

    renderApp('/?view=week')

    expect(await screen.findByText('周视图时间表')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '周视图' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
  })

  it.each([
    ['list', '接下来 14 天', '列表'],
    ['availability', '可约时间', '空闲'],
  ] as const)('loads the %s calendar view from the URL query', async (view, text, button) => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([createHomeEvent(), createHomeSchedule()])

    renderApp(`/?view=${view}`)

    expect(await screen.findByText(text)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: button })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
  })

  it('falls back to month view for invalid calendar view query', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([])

    renderApp('/?view=unknown')

    expect(
      await screen.findByRole('button', { name: '月视图' }),
    ).toHaveAttribute('aria-pressed', 'true')
  })

  it('filters the calendar by checked subscription sources', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([])
    mockedListCalendarSubscriptions.mockResolvedValue([
      {
        id: 'self:schedules',
        type: 'self_schedules',
        label: '自身日程',
        color: 'green',
        enabled: true,
      },
      {
        id: 'friend:aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa:schedules',
        type: 'friend_schedules',
        label: 'Alex 日程',
        color: 'gray',
        enabled: true,
      },
    ])

    renderApp('/')

    expect(
      await screen.findByRole('heading', { name: '我订阅的' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('checkbox', { name: '自身日程' })).toBeChecked()
    expect(
      await screen.findByRole('checkbox', { name: 'Alex 日程' }),
    ).toBeChecked()

    await waitFor(() => {
      expect(mockedGetCalendar).toHaveBeenLastCalledWith(
        expect.any(String),
        expect.any(String),
        'all',
        [
          'self:schedules',
          'friend:aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa:schedules',
        ],
      )
    })

    fireEvent.click(screen.getByRole('checkbox', { name: 'Alex 日程' }))

    await waitFor(() => {
      expect(mockedGetCalendar).toHaveBeenLastCalledWith(
        expect.any(String),
        expect.any(String),
        'all',
        ['self:schedules'],
      )
    })
  })

  it('extends the feed timeline range forward on demand', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([])

    renderApp('/')

    await screen.findByText('时间流')
    const timeline = screen.getByLabelText('无限日期时间轴')
    expect(timeline).toHaveClass('scrollbar-none')
    timeline.scrollLeft = 0

    fireEvent.wheel(timeline, { deltaY: 120 })

    expect(timeline.scrollLeft).toBe(120)

    fireEvent.click(screen.getByRole('button', { name: '向后延展时间流' }))

    await waitFor(() => {
      expect(mockedGetCalendar.mock.calls.length).toBeGreaterThan(1)
    })
  })

  it('renders selected day items as vertical timeline previews before expansion', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([createHomeEvent()])

    renderApp('/')

    expect(
      await screen.findByRole('button', { name: '展开日程 周末攀岩' }),
    ).toBeInTheDocument()
    expect(screen.getByText('20:00')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '加入' })).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: '展开日程 周末攀岩' }))

    expect(await screen.findByRole('button', { name: '加入' })).toBeInTheDocument()
  })

  it('cancels manual schedules from the expanded feed card', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([createHomeSchedule()])
    vi.stubGlobal('confirm', vi.fn(() => true))

    renderApp('/')

    fireEvent.click(
      await screen.findByRole('button', { name: '展开日程 专注写方案' }),
    )
    fireEvent.click(screen.getByRole('button', { name: '取消日程' }))

    await waitFor(() => {
      expect(mockedDeleteSchedule).toHaveBeenCalledWith(31)
    })
    expect(await screen.findByText('日程已取消')).toBeInTheDocument()
    expect(screen.queryByText('专注写方案')).not.toBeInTheDocument()
  })

  it('updates the home event card after joining', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([createHomeEvent()])
    mockedRsvpEvent.mockResolvedValue({
      participant: { id: 1, event_id: 12, rsvp: 'going', add_to_calendar: true },
      conflicts: [],
      going_count: 3,
    })

    renderApp('/')

    fireEvent.click(
      await screen.findByRole('button', { name: '展开日程 周末攀岩' }),
    )

    await waitFor(() => {
      expect(screen.getByRole('button', { name: '加入' })).toBeEnabled()
    })

    fireEvent.click(screen.getByRole('button', { name: '加入' }))

    expect(await screen.findByText('已加入活动，已同步到你的日历')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '取消参加' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '邀请好友' })).toBeInTheDocument()
    expect(screen.getByText('3 人已参加')).toBeInTheDocument()
  })

  it('invites selected friends from a joined home event card', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([
      createHomeEvent({ viewer_rsvp: 'going' }),
    ])
    mockedListFriends.mockResolvedValue([
      {
        user_id: 'friend-1',
        display_name: 'Alex',
        avatar_url: null,
        email: 'alex@example.com',
      },
      {
        user_id: 'friend-2',
        display_name: 'Mia',
        avatar_url: null,
        email: 'mia@example.com',
      },
    ])
    mockedInviteToEvent.mockResolvedValue([
      { id: 1, event_id: 12, user_id: 'friend-1', rsvp: 'invited' },
      { id: 2, event_id: 12, user_id: 'friend-2', rsvp: 'invited' },
    ])

    renderApp('/')

    fireEvent.click(
      await screen.findByRole('button', { name: '展开日程 周末攀岩' }),
    )
    fireEvent.click(await screen.findByRole('button', { name: '邀请好友' }))
    fireEvent.click(await screen.findByLabelText('选择 Alex'))
    fireEvent.click(screen.getByLabelText('选择 Mia'))
    fireEvent.click(screen.getByRole('button', { name: '发送邀请' }))

    await waitFor(() => {
      expect(mockedInviteToEvent).toHaveBeenCalledWith(12, [
        'friend-1',
        'friend-2',
      ])
    })
    expect(await screen.findByText('已邀请 2 位好友')).toBeInTheDocument()
  })

  it('collapses dismissed home event cards and expands them on hover', async () => {
    setSession({ access_token: 'token' })
    mockedGetCalendar.mockResolvedValue([createHomeEvent()])
    mockedRsvpEvent.mockResolvedValue({
      participant: {
        id: 1,
        event_id: 12,
        rsvp: 'not_going',
        add_to_calendar: false,
      },
      conflicts: [],
      going_count: 2,
    })

    renderApp('/')

    fireEvent.click(
      await screen.findByRole('button', { name: '展开日程 周末攀岩' }),
    )

    await waitFor(() => {
      expect(screen.getByRole('button', { name: '下次一定' })).toBeEnabled()
    })

    fireEvent.click(screen.getByRole('button', { name: '下次一定' }))

    const collapsed = await screen.findByLabelText(
      '已收纳活动 周末攀岩，悬停后展开',
    )
    expect(screen.getByText('已收纳')).toBeInTheDocument()

    fireEvent.mouseEnter(collapsed)

    expect(
      await screen.findByRole('button', { name: '重新加入' }),
    ).toBeInTheDocument()
  })

  it('loads the event detail route by slug', async () => {
    setSession({ access_token: 'token' })
    mockedGetEventBySlug.mockResolvedValue(
      createEventDetail('abc123def4', '周末攀岩'),
    )

    renderApp('/e/abc123def4')

    await waitFor(() => {
      expect(
        screen.getByRole('heading', { name: '周末攀岩' }),
      ).toBeInTheDocument()
    })
    expect(mockedGetEventBySlug).toHaveBeenCalledWith('abc123def4')
  })

  it('lets authenticated users submit an RSVP', async () => {
    setSession({ access_token: 'token' })
    mockedGetEventBySlug.mockResolvedValue(
      createEventDetail('abc123def4', '周末攀岩'),
    )
    mockedRsvpEvent.mockResolvedValue({
      participant: { id: 1, event_id: 12, rsvp: 'going', add_to_calendar: true },
      conflicts: [],
      going_count: 3,
    })

    renderApp('/e/abc123def4')

    await waitFor(() => {
      expect(screen.getByRole('button', { name: '我要参加' })).toBeEnabled()
    })

    fireEvent.click(screen.getByRole('button', { name: '我要参加' }))

    await waitFor(() => {
      expect(mockedRsvpEvent).toHaveBeenCalledWith(12, {
        rsvp: 'going',
        add_to_calendar: true,
        visibility: 'busy_only',
      })
    })
    expect(await screen.findByText('已加入活动')).toBeInTheDocument()
    expect(screen.getByText('3 / 6 人已参加')).toBeInTheDocument()
  })

  it('manages manual schedules from the schedule tab', async () => {
    setSession({ access_token: 'token' })
    mockedListSchedules.mockResolvedValue([
      {
        id: 31,
        title: '专注写方案',
        start_at: '2026-06-24T10:00:00Z',
        end_at: null,
        location: '公司',
        visibility: 'busy_only',
        source: 'manual',
        event_id: null,
      },
    ])
    vi.stubGlobal('confirm', vi.fn(() => true))

    renderApp('/schedules')

    expect(
      await screen.findByRole('heading', { name: '日程管理' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('link', { name: '日程' })).toHaveAttribute(
      'href',
      '/schedules',
    )
    expect(await screen.findByText('专注写方案')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: '编辑 专注写方案' })).toHaveAttribute(
      'href',
      '/schedules/31',
    )

    fireEvent.click(screen.getByRole('button', { name: '删除 专注写方案' }))

    await waitFor(() => {
      expect(mockedDeleteSchedule).toHaveBeenCalledWith(31)
    })
  })

  it('loads an existing manual schedule into the editor', async () => {
    setSession({ access_token: 'token' })
    mockedGetSchedule.mockResolvedValue({
      id: 31,
      title: '专注写方案',
      start_at: '2026-06-24T10:00:00Z',
      end_at: null,
      location: '公司',
      visibility: 'busy_only',
      source: 'manual',
      event_id: null,
    })

    renderApp('/schedules/31')

    expect(await screen.findByDisplayValue('专注写方案')).toBeInTheDocument()
    expect(screen.getByDisplayValue('公司')).toBeInTheDocument()
  })

  it('edits the display name from the profile page', async () => {
    setSession({ access_token: 'token' })

    renderApp('/profile')

    const nameInput = await screen.findByLabelText('用户名')
    fireEvent.change(nameInput, { target: { value: 'Lily Chen' } })
    fireEvent.click(screen.getByRole('button', { name: '保存用户名' }))

    await waitFor(() => {
      expect(mockedApiPatch).toHaveBeenCalledWith('/api/users/me', {
        display_name: 'Lily Chen',
        avatar_url: null,
      })
    })
    expect(await screen.findByText('用户名已更新')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Lily Chen' })).toBeInTheDocument()
  })

  it('invites friends to a selected activity from the friends page', async () => {
    setSession({ access_token: 'token' })
    mockedListFriends.mockResolvedValue([
      {
        user_id: 'friend-1',
        display_name: 'Alex',
        avatar_url: null,
        email: 'alex@example.com',
      },
    ])
    mockedListMine.mockResolvedValue([
      {
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
    ])
    mockedInviteToEvent.mockResolvedValue([
      { id: 1, event_id: 12, user_id: 'friend-1', rsvp: 'invited' },
    ])

    renderApp('/friends')

    await screen.findByRole('heading', { name: '邀请好友参加活动' })
    await screen.findByRole('option', { name: '周末攀岩' })
    fireEvent.change(screen.getByLabelText('选择活动'), {
      target: { value: '12' },
    })
    fireEvent.click(screen.getByLabelText('选择 Alex'))
    fireEvent.click(screen.getByRole('button', { name: '发送活动邀请' }))

    await waitFor(() => {
      expect(mockedInviteToEvent).toHaveBeenCalledWith(12, ['friend-1'])
    })
    expect(screen.getByText('已邀请 1 位好友参加活动')).toBeInTheDocument()
  })

  it('invites selected friends after creating an activity', async () => {
    setSession({ access_token: 'token' })
    mockedListFriends.mockResolvedValue([
      {
        user_id: 'friend-1',
        display_name: 'Alex',
        avatar_url: null,
        email: 'alex@example.com',
      },
    ])
    mockedCreateEvent.mockResolvedValue({
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
    })
    mockedInviteToEvent.mockResolvedValue([
      { id: 1, event_id: 12, user_id: 'friend-1', rsvp: 'invited' },
    ])
    mockedGetEventBySlug.mockResolvedValue(
      createEventDetail('abc123def4', '周末攀岩'),
    )

    renderApp('/events/new')

    await screen.findByRole('heading', { name: '创建活动' })
    fireEvent.change(screen.getByLabelText('标题'), {
      target: { value: '周末攀岩' },
    })
    fireEvent.change(screen.getByLabelText('开始时间'), {
      target: { value: '2026-06-24T20:00' },
    })
    fireEvent.click(await screen.findByLabelText('选择 Alex'))
    fireEvent.click(screen.getByRole('button', { name: '创建活动' }))

    await waitFor(() => {
      expect(mockedCreateEvent).toHaveBeenCalled()
      expect(mockedInviteToEvent).toHaveBeenCalledWith(12, ['friend-1'])
    })
  })

  it('refreshes the bottom inbox badge after handling an inbox item', async () => {
    setSession({ access_token: 'token' })
    let inboxTotal = 1
    mockedGetInbox.mockImplementation(async () => ({
      counts: {
        total: inboxTotal,
        event_invites: inboxTotal,
        friend_requests: 0,
        group_invites: 0,
      },
      items:
        inboxTotal > 0
          ? [
              {
                id: 'event-invite-12',
                kind: 'event_invite',
                source_id: 1,
                event_id: 12,
                title: '周末攀岩',
                subtitle: 'Alex 邀请你参加',
                created_at: '2026-06-24T12:00:00Z',
                disabled: false,
              },
            ]
          : [],
    }))
    mockedRsvpEvent.mockImplementation(async () => {
      inboxTotal = 0
      return {
        participant: {
          id: 1,
          event_id: 12,
          rsvp: 'going',
          add_to_calendar: true,
        },
        conflicts: [],
        going_count: 3,
      }
    })

    renderApp('/inbox')

    expect(await screen.findByLabelText('收件箱未读 1')).toBeInTheDocument()
    fireEvent.click(await screen.findByRole('button', { name: '参加' }))

    await waitFor(() => {
      expect(screen.queryByLabelText('收件箱未读 1')).not.toBeInTheDocument()
    })
    expect(screen.getByText('收件箱已清空')).toBeInTheDocument()
  })

  it('renders a 404 for unknown routes', () => {
    renderApp('/does-not-exist')

    expect(screen.getByText('页面不存在')).toBeInTheDocument()
  })
})
