package inbox

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	me    = uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	other = uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
)

func fixedNow() time.Time {
	return time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
}

func TestInvitedEventAppearsWithConflictCount(t *testing.T) {
	repo := &fakeRepo{
		invitedEvents: []EventInvite{
			{
				ParticipantID: 1,
				EventID:       10,
				Title:         "Climb",
				StartAt:       fixedNow().Add(2 * time.Hour),
				Location:      strptr("Gym"),
				Status:        "active",
				ConflictCount: 2,
				CreatedAt:     fixedNow(),
			},
		},
	}
	service := NewService(repo, WithNow(fixedNow))

	result, err := service.List(context.Background(), me)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Counts.EventInvites != 1 || result.Counts.Total != 1 {
		t.Fatalf("expected one event invite count, got %+v", result.Counts)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(result.Items))
	}
	item := result.Items[0]
	if item.Kind != KindEventInvite {
		t.Fatalf("expected event_invite kind, got %q", item.Kind)
	}
	if item.Title != "Climb" {
		t.Fatalf("expected title Climb, got %q", item.Title)
	}
	if item.ConflictCount != 2 {
		t.Fatalf("expected conflict count 2, got %d", item.ConflictCount)
	}
	if item.Disabled {
		t.Fatalf("expected active invite to be enabled")
	}
}

func TestPendingFriendRequestAppears(t *testing.T) {
	repo := &fakeRepo{
		friendRequests: []FriendRequest{
			{ID: 5, RequesterID: other, RequesterName: "Bob", CreatedAt: fixedNow()},
		},
	}
	service := NewService(repo, WithNow(fixedNow))

	result, err := service.List(context.Background(), me)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Counts.FriendRequests != 1 {
		t.Fatalf("expected one friend request, got %+v", result.Counts)
	}
	if result.Items[0].Kind != KindFriendRequest || result.Items[0].SourceID != 5 {
		t.Fatalf("unexpected friend request item %+v", result.Items[0])
	}
}

func TestPendingGroupInviteAppears(t *testing.T) {
	repo := &fakeRepo{
		groupInvites: []GroupInvite{
			{ID: 8, GroupID: 3, GroupName: "Climbers", InviterName: "Bob", CreatedAt: fixedNow()},
		},
	}
	service := NewService(repo, WithNow(fixedNow))

	result, err := service.List(context.Background(), me)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Counts.GroupInvites != 1 {
		t.Fatalf("expected one group invite, got %+v", result.Counts)
	}
	if result.Items[0].Kind != KindGroupInvite || result.Items[0].SourceID != 8 {
		t.Fatalf("unexpected group invite item %+v", result.Items[0])
	}
}

func TestCancelledEventInviteIsDisabled(t *testing.T) {
	repo := &fakeRepo{
		invitedEvents: []EventInvite{
			{ParticipantID: 1, EventID: 10, Title: "Climb", StartAt: fixedNow().Add(2 * time.Hour), Status: "cancelled", CreatedAt: fixedNow()},
		},
	}
	service := NewService(repo, WithNow(fixedNow))

	result, err := service.List(context.Background(), me)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	item := result.Items[0]
	if !item.Disabled || item.DisabledReason == "" {
		t.Fatalf("expected cancelled invite disabled with reason, got %+v", item)
	}
}

func TestExpiredEventInviteIsDisabled(t *testing.T) {
	repo := &fakeRepo{
		invitedEvents: []EventInvite{
			{ParticipantID: 1, EventID: 10, Title: "Climb", StartAt: fixedNow().Add(-2 * time.Hour), Status: "active", CreatedAt: fixedNow()},
		},
	}
	service := NewService(repo, WithNow(fixedNow))

	result, err := service.List(context.Background(), me)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	item := result.Items[0]
	if !item.Disabled || item.DisabledReason == "" {
		t.Fatalf("expected expired invite disabled with reason, got %+v", item)
	}
}

func TestItemsSortedByCreatedAtDesc(t *testing.T) {
	repo := &fakeRepo{
		invitedEvents: []EventInvite{
			{ParticipantID: 1, EventID: 10, Title: "Old", StartAt: fixedNow().Add(2 * time.Hour), Status: "active", CreatedAt: fixedNow().Add(-3 * time.Hour)},
		},
		friendRequests: []FriendRequest{
			{ID: 5, RequesterID: other, RequesterName: "Bob", CreatedAt: fixedNow().Add(-1 * time.Hour)},
		},
		groupInvites: []GroupInvite{
			{ID: 8, GroupID: 3, GroupName: "Climbers", InviterName: "Bob", CreatedAt: fixedNow().Add(-2 * time.Hour)},
		},
	}
	service := NewService(repo, WithNow(fixedNow))

	result, err := service.List(context.Background(), me)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result.Items))
	}
	if result.Items[0].Kind != KindFriendRequest || result.Items[2].Kind != KindEventInvite {
		t.Fatalf("expected newest-first ordering, got %+v", result.Items)
	}
	if result.Counts.Total != 3 {
		t.Fatalf("expected total 3, got %d", result.Counts.Total)
	}
}

func strptr(s string) *string {
	return &s
}

type fakeRepo struct {
	invitedEvents  []EventInvite
	friendRequests []FriendRequest
	groupInvites   []GroupInvite
}

func (r *fakeRepo) ListInvitedEvents(_ context.Context, _ uuid.UUID) ([]EventInvite, error) {
	return r.invitedEvents, nil
}

func (r *fakeRepo) ListPendingFriendRequests(_ context.Context, _ uuid.UUID) ([]FriendRequest, error) {
	return r.friendRequests, nil
}

func (r *fakeRepo) ListPendingGroupInvites(_ context.Context, _ uuid.UUID) ([]GroupInvite, error) {
	return r.groupInvites, nil
}
