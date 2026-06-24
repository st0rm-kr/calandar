package http

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/events"
	"github.com/bytedance/calandar/apps/api/internal/friends"
	"github.com/bytedance/calandar/apps/api/internal/groups"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type spyNotifier struct {
	mu      sync.Mutex
	records []notifyRecord
}

type notifyRecord struct {
	UserID uuid.UUID
	Type   string
}

func (s *spyNotifier) Notify(_ context.Context, userID uuid.UUID, typ string, _ map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, notifyRecord{UserID: userID, Type: typ})
	return nil
}

func (s *spyNotifier) has(typ string, userID uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, record := range s.records {
		if record.Type == typ && record.UserID == userID {
			return true
		}
	}
	return false
}

func TestNotificationsFromDomainActions(t *testing.T) {
	requester := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	addressee := uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")

	t.Run("friend request and accept notify", func(t *testing.T) {
		notifier := &spyNotifier{}
		service := friends.NewService(newFakeFriendsRepo(), friends.WithNotifier(notifier))

		request, err := service.SendRequest(context.Background(), requester, addressee)
		if err != nil {
			t.Fatalf("send request: %v", err)
		}
		if !notifier.has("friend_request", addressee) {
			t.Fatalf("expected friend_request notification for addressee")
		}

		if _, err := service.Accept(context.Background(), addressee, request.ID); err != nil {
			t.Fatalf("accept: %v", err)
		}
		if !notifier.has("friend_accepted", requester) {
			t.Fatalf("expected friend_accepted notification for requester")
		}
	})

	t.Run("group invite notifies invitee", func(t *testing.T) {
		notifier := &spyNotifier{}
		service := groups.NewService(newFakeGroupsRepo(), groups.WithNotifier(notifier))

		group, err := service.CreateGroup(context.Background(), requester, groups.CreateGroupInput{Name: "Climbers"})
		if err != nil {
			t.Fatalf("create group: %v", err)
		}
		if _, err := service.Invite(context.Background(), requester, group.ID, addressee); err != nil {
			t.Fatalf("invite: %v", err)
		}
		if !notifier.has("group_invite", addressee) {
			t.Fatalf("expected group_invite notification for invitee")
		}
	})

	t.Run("event invite and cancel notify", func(t *testing.T) {
		notifier := &spyNotifier{}
		repo := newFakeEventsRepo()
		end := time.Now().Add(2 * time.Hour).UTC()
		event := repo.seed(events.Event{
			ID:      1,
			OwnerID: requester,
			Scope:   "personal",
			Title:   "Climb",
			Type:    "climbing",
			StartAt: time.Now().Add(time.Hour).UTC(),
			EndAt:   &end,
			Status:  events.StatusActive,
		})
		service := events.NewService(
			repo,
			events.WithFriendAuthorizer(allowAllFriends{}),
			events.WithNotifier(notifier),
		)

		if _, err := service.Invite(context.Background(), requester, event.ID, []uuid.UUID{addressee}); err != nil {
			t.Fatalf("invite: %v", err)
		}
		if !notifier.has("event_invite", addressee) {
			t.Fatalf("expected event_invite notification for invitee")
		}

		if _, err := service.Cancel(context.Background(), requester, event.ID); err != nil {
			t.Fatalf("cancel: %v", err)
		}
		if !notifier.has("event_cancelled", addressee) {
			t.Fatalf("expected event_cancelled notification for participant")
		}
	})
}

type allowAllFriends struct{}

func (allowAllFriends) AreFriends(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

// --- fake friends repo ---

type fakeFriendsRepo struct {
	rows   map[int64]friends.Friendship
	nextID int64
}

func newFakeFriendsRepo() *fakeFriendsRepo {
	return &fakeFriendsRepo{rows: map[int64]friends.Friendship{}}
}

func (r *fakeFriendsRepo) FindBetween(_ context.Context, a, b uuid.UUID) (friends.Friendship, bool, error) {
	for _, row := range r.rows {
		if (row.RequesterID == a && row.AddresseeID == b) || (row.RequesterID == b && row.AddresseeID == a) {
			return row, true, nil
		}
	}
	return friends.Friendship{}, false, nil
}

func (r *fakeFriendsRepo) Create(_ context.Context, friendship friends.Friendship) (friends.Friendship, error) {
	r.nextID++
	friendship.ID = r.nextID
	r.rows[friendship.ID] = friendship
	return friendship, nil
}

func (r *fakeFriendsRepo) UpdateStatus(_ context.Context, id int64, status string) (friends.Friendship, error) {
	row := r.rows[id]
	row.Status = status
	r.rows[id] = row
	return row, nil
}

func (r *fakeFriendsRepo) FindByID(_ context.Context, id int64) (friends.Friendship, error) {
	row, ok := r.rows[id]
	if !ok {
		return friends.Friendship{}, friends.ErrRequestNotFound
	}
	return row, nil
}

func (r *fakeFriendsRepo) ListAccepted(_ context.Context, _ uuid.UUID) ([]friends.Friend, error) {
	return nil, nil
}

func (r *fakeFriendsRepo) ListRequests(_ context.Context, _ uuid.UUID) ([]friends.Friendship, error) {
	return nil, nil
}

func (r *fakeFriendsRepo) DeleteBetween(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

// --- fake groups repo ---

type fakeGroupsRepo struct {
	groups       map[int64]groups.Group
	members      map[int64]groups.GroupMember
	invites      map[int64]groups.GroupInvite
	nextGroupID  int64
	nextMemberID int64
	nextInviteID int64
}

func newFakeGroupsRepo() *fakeGroupsRepo {
	return &fakeGroupsRepo{
		groups:  map[int64]groups.Group{},
		members: map[int64]groups.GroupMember{},
		invites: map[int64]groups.GroupInvite{},
	}
}

func (r *fakeGroupsRepo) Transaction(ctx context.Context, fn func(groups.Repository) error) error {
	return fn(r)
}

func (r *fakeGroupsRepo) CreateGroup(_ context.Context, group groups.Group) (groups.Group, error) {
	r.nextGroupID++
	group.ID = r.nextGroupID
	r.groups[group.ID] = group
	return group, nil
}

func (r *fakeGroupsRepo) FindGroupByID(_ context.Context, id int64) (groups.Group, error) {
	group, ok := r.groups[id]
	if !ok {
		return groups.Group{}, groups.ErrGroupNotFound
	}
	return group, nil
}

func (r *fakeGroupsRepo) FindGroupByInviteCode(_ context.Context, code string) (groups.Group, error) {
	for _, g := range r.groups {
		if g.InviteCode == code {
			return g, nil
		}
	}
	return groups.Group{}, groups.ErrGroupNotFound
}

func (r *fakeGroupsRepo) UpdateGroup(_ context.Context, id int64, fields map[string]any) (groups.Group, error) {
	group := r.groups[id]
	if status, ok := fields["status"].(string); ok {
		group.Status = status
	}
	r.groups[id] = group
	return group, nil
}

func (r *fakeGroupsRepo) ListGroupsByMember(_ context.Context, _ uuid.UUID) ([]groups.Group, error) {
	return nil, nil
}

func (r *fakeGroupsRepo) FindMember(_ context.Context, groupID int64, userID uuid.UUID) (groups.GroupMember, bool, error) {
	for _, m := range r.members {
		if m.GroupID == groupID && m.UserID == userID {
			return m, true, nil
		}
	}
	return groups.GroupMember{}, false, nil
}

func (r *fakeGroupsRepo) AddMember(_ context.Context, member groups.GroupMember) (groups.GroupMember, error) {
	r.nextMemberID++
	member.ID = r.nextMemberID
	r.members[member.ID] = member
	return member, nil
}

func (r *fakeGroupsRepo) RemoveMember(_ context.Context, _ int64, _ uuid.UUID) error {
	return nil
}

func (r *fakeGroupsRepo) CountMembers(_ context.Context, groupID int64) (int, error) {
	count := 0
	for _, m := range r.members {
		if m.GroupID == groupID {
			count++
		}
	}
	return count, nil
}

func (r *fakeGroupsRepo) ListMembers(_ context.Context, _ int64) ([]groups.Member, error) {
	return nil, nil
}

func (r *fakeGroupsRepo) CreateInvite(_ context.Context, invite groups.GroupInvite) (groups.GroupInvite, error) {
	r.nextInviteID++
	invite.ID = r.nextInviteID
	r.invites[invite.ID] = invite
	return invite, nil
}

func (r *fakeGroupsRepo) FindInviteByID(_ context.Context, id int64) (groups.GroupInvite, error) {
	invite, ok := r.invites[id]
	if !ok {
		return groups.GroupInvite{}, groups.ErrInviteNotFound
	}
	return invite, nil
}

func (r *fakeGroupsRepo) FindInvite(_ context.Context, groupID int64, inviteeID uuid.UUID) (groups.GroupInvite, bool, error) {
	for _, inv := range r.invites {
		if inv.GroupID == groupID && inv.InviteeID == inviteeID {
			return inv, true, nil
		}
	}
	return groups.GroupInvite{}, false, nil
}

func (r *fakeGroupsRepo) UpdateInviteStatus(_ context.Context, id int64, status string) (groups.GroupInvite, error) {
	invite := r.invites[id]
	invite.Status = status
	r.invites[id] = invite
	return invite, nil
}

// --- fake events repo ---

type fakeEventsRepo struct {
	events       map[int64]events.Event
	participants map[string]events.Participant
	nextID       int64
}

func newFakeEventsRepo() *fakeEventsRepo {
	return &fakeEventsRepo{
		events:       map[int64]events.Event{},
		participants: map[string]events.Participant{},
		nextID:       1,
	}
}

func (r *fakeEventsRepo) seed(event events.Event) events.Event {
	r.events[event.ID] = event
	return event
}

func participantKey(eventID int64, userID uuid.UUID) string {
	return fmt.Sprintf("%s:%d", userID.String(), eventID)
}

func (r *fakeEventsRepo) Create(_ context.Context, event events.Event) (events.Event, error) {
	r.nextID++
	event.ID = r.nextID
	r.events[event.ID] = event
	return event, nil
}

func (r *fakeEventsRepo) FindBySlug(_ context.Context, _ string) (events.Event, error) {
	return events.Event{}, gorm.ErrRecordNotFound
}

func (r *fakeEventsRepo) FindByID(_ context.Context, id int64) (events.Event, error) {
	event, ok := r.events[id]
	if !ok {
		return events.Event{}, gorm.ErrRecordNotFound
	}
	return event, nil
}

func (r *fakeEventsRepo) FindByIDForUpdate(ctx context.Context, id int64) (events.Event, error) {
	return r.FindByID(ctx, id)
}

func (r *fakeEventsRepo) CountGoing(_ context.Context, _ int64) (int, error) {
	return 0, nil
}

func (r *fakeEventsRepo) ListMine(_ context.Context, _ uuid.UUID) ([]events.Event, error) {
	return nil, nil
}

func (r *fakeEventsRepo) Transaction(ctx context.Context, fn func(events.EventRepository) error) error {
	return fn(r)
}

func (r *fakeEventsRepo) CountGoingForUpdateExcludingUser(_ context.Context, _ int64, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (r *fakeEventsRepo) UpsertParticipant(_ context.Context, participant events.Participant) (events.Participant, error) {
	r.nextID++
	participant.ID = r.nextID
	r.participants[participantKey(participant.EventID, participant.UserID)] = participant
	return participant, nil
}

func (r *fakeEventsRepo) FindParticipant(_ context.Context, eventID int64, userID uuid.UUID) (events.Participant, bool, error) {
	participant, ok := r.participants[participantKey(eventID, userID)]
	if !ok {
		return events.Participant{}, false, nil
	}
	return participant, true, nil
}

func (r *fakeEventsRepo) ListParticipants(_ context.Context, eventID int64) ([]events.Participant, error) {
	var result []events.Participant
	for _, participant := range r.participants {
		if participant.EventID == eventID {
			result = append(result, participant)
		}
	}
	return result, nil
}

func (r *fakeEventsRepo) UpsertEventSchedule(_ context.Context, _ uuid.UUID, _ events.Event, _ string) error {
	return nil
}

func (r *fakeEventsRepo) DeleteEventSchedule(_ context.Context, _ uuid.UUID, _ int64) error {
	return nil
}

func (r *fakeEventsRepo) DeleteEventSchedulesForEvent(_ context.Context, _ int64) error {
	return nil
}

func (r *fakeEventsRepo) FindConflicts(_ context.Context, _ uuid.UUID, _ time.Time, _ *time.Time) ([]events.ScheduleConflict, error) {
	return nil, nil
}

func (r *fakeEventsRepo) Cancel(_ context.Context, _ uuid.UUID, eventID int64) (events.Event, error) {
	event := r.events[eventID]
	event.Status = events.StatusCancelled
	r.events[eventID] = event
	return event, nil
}
