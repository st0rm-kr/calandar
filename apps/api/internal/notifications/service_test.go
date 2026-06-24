package notifications

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	userA = uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	userB = uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
)

func TestCreateStoresPayload(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	if err := service.Create(context.Background(), userA, TypeFriendRequest, Payload{"request_id": 7}); err != nil {
		t.Fatalf("create: %v", err)
	}

	if len(repo.rows) != 1 {
		t.Fatalf("expected one notification, got %d", len(repo.rows))
	}
	for _, row := range repo.rows {
		if row.Type != TypeFriendRequest {
			t.Fatalf("expected friend_request type, got %q", row.Type)
		}
		if row.Payload["request_id"] != 7 {
			t.Fatalf("expected payload request_id 7, got %+v", row.Payload)
		}
	}
}

func TestListReturnsUnreadCountAndLatestFirst(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	older := repo.seed(Notification{UserID: userA, Type: TypeFriendRequest, CreatedAt: time.Now().Add(-time.Hour)})
	newer := repo.seed(Notification{UserID: userA, Type: TypeEventInvite, CreatedAt: time.Now()})
	readAt := time.Now()
	repo.rows[older.ID] = Notification{ID: older.ID, UserID: userA, Type: TypeFriendRequest, CreatedAt: older.CreatedAt, ReadAt: &readAt}

	result, err := service.List(context.Background(), userA)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.UnreadCount != 1 {
		t.Fatalf("expected unread count 1, got %d", result.UnreadCount)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].ID != newer.ID {
		t.Fatalf("expected newest item first, got %d", result.Items[0].ID)
	}
}

func TestReadMarksOnlyOwnNotification(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	own := repo.seed(Notification{UserID: userA, Type: TypeFriendRequest, CreatedAt: time.Now()})
	other := repo.seed(Notification{UserID: userB, Type: TypeFriendRequest, CreatedAt: time.Now()})

	if err := service.MarkRead(context.Background(), userA, own.ID); err != nil {
		t.Fatalf("mark read own: %v", err)
	}
	if repo.rows[own.ID].ReadAt == nil {
		t.Fatalf("expected own notification marked read")
	}

	if err := service.MarkRead(context.Background(), userA, other.ID); err == nil {
		t.Fatalf("expected error marking another user's notification")
	}
	if repo.rows[other.ID].ReadAt != nil {
		t.Fatalf("expected other user's notification untouched")
	}
}

func TestReadAllMarksAllUnread(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo)

	first := repo.seed(Notification{UserID: userA, Type: TypeFriendRequest, CreatedAt: time.Now()})
	second := repo.seed(Notification{UserID: userA, Type: TypeEventInvite, CreatedAt: time.Now()})
	otherUser := repo.seed(Notification{UserID: userB, Type: TypeEventInvite, CreatedAt: time.Now()})

	if err := service.MarkAllRead(context.Background(), userA); err != nil {
		t.Fatalf("mark all read: %v", err)
	}
	if repo.rows[first.ID].ReadAt == nil || repo.rows[second.ID].ReadAt == nil {
		t.Fatalf("expected both of userA's notifications read")
	}
	if repo.rows[otherUser.ID].ReadAt != nil {
		t.Fatalf("expected userB's notification untouched")
	}
}

type fakeRepo struct {
	rows   map[int64]Notification
	nextID int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{rows: map[int64]Notification{}}
}

func (r *fakeRepo) seed(n Notification) Notification {
	r.nextID++
	n.ID = r.nextID
	r.rows[n.ID] = n
	return n
}

func (r *fakeRepo) Create(_ context.Context, n Notification) error {
	r.nextID++
	n.ID = r.nextID
	r.rows[n.ID] = n
	return nil
}

func (r *fakeRepo) List(_ context.Context, userID uuid.UUID) ([]Notification, error) {
	var items []Notification
	for _, row := range r.rows {
		if row.UserID == userID && row.DeletedAt == nil {
			items = append(items, row)
		}
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].CreatedAt.After(items[i].CreatedAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items, nil
}

func (r *fakeRepo) CountUnread(_ context.Context, userID uuid.UUID) (int, error) {
	count := 0
	for _, row := range r.rows {
		if row.UserID == userID && row.ReadAt == nil && row.DeletedAt == nil {
			count++
		}
	}
	return count, nil
}

func (r *fakeRepo) MarkRead(_ context.Context, userID uuid.UUID, id int64) (bool, error) {
	row, ok := r.rows[id]
	if !ok || row.UserID != userID || row.DeletedAt != nil {
		return false, nil
	}
	now := time.Now()
	row.ReadAt = &now
	r.rows[id] = row
	return true, nil
}

func (r *fakeRepo) MarkAllRead(_ context.Context, userID uuid.UUID) error {
	now := time.Now()
	for id, row := range r.rows {
		if row.UserID == userID && row.ReadAt == nil && row.DeletedAt == nil {
			row.ReadAt = &now
			r.rows[id] = row
		}
	}
	return nil
}
