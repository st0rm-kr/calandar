package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestCreateGeneratesUniqueShareSlug(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	ownerID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	startAt := time.Now().Add(time.Hour).UTC()

	first, err := service.Create(context.Background(), ownerID, CreateEventInput{
		Title:   "Climb",
		Type:    "climbing",
		Scope:   "personal",
		StartAt: startAt,
	})
	if err != nil {
		t.Fatalf("create first event: %v", err)
	}
	second, err := service.Create(context.Background(), ownerID, CreateEventInput{
		Title:   "Dinner",
		Type:    "dining",
		Scope:   "personal",
		StartAt: startAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("create second event: %v", err)
	}

	if first.ShareSlug == "" || second.ShareSlug == "" {
		t.Fatalf("expected share slugs to be generated, got %q and %q", first.ShareSlug, second.ShareSlug)
	}
	if len(first.ShareSlug) != 10 || len(second.ShareSlug) != 10 {
		t.Fatalf("expected 10 character slugs, got %q and %q", first.ShareSlug, second.ShareSlug)
	}
	if first.ShareSlug == second.ShareSlug {
		t.Fatalf("expected unique slugs, got %q", first.ShareSlug)
	}
}

func TestCreateRejectsEndAtBeforeOrEqualStartAt(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	ownerID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	startAt := time.Now().Add(time.Hour).UTC()

	_, err := service.Create(context.Background(), ownerID, CreateEventInput{
		Title:   "Climb",
		Type:    "climbing",
		Scope:   "personal",
		StartAt: startAt,
		EndAt:   ptrTime(startAt),
	})

	if !errors.Is(err, ErrInvalidEventTime) {
		t.Fatalf("expected ErrInvalidEventTime, got %v", err)
	}
	if len(repo.events) != 0 {
		t.Fatalf("expected no event to be persisted, got %d", len(repo.events))
	}
}

func TestGetDetailReturnsParticipantCountsWithoutAuth(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	startAt := time.Now().Add(time.Hour).UTC()
	event := repo.seedEvent(Event{
		OwnerID:   uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8"),
		Title:     "Climb",
		Type:      "climbing",
		Scope:     "personal",
		StartAt:   startAt,
		Status:    StatusActive,
		ShareSlug: "abc123def4",
	})
	repo.goingCounts[event.ID] = 3

	detail, err := service.GetDetail(context.Background(), "abc123def4")
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}

	if detail.Event.ID != event.ID {
		t.Fatalf("expected event %d, got %d", event.ID, detail.Event.ID)
	}
	if detail.GoingCount != 3 {
		t.Fatalf("expected going count 3, got %d", detail.GoingCount)
	}
}

func TestGetDetailDerivesExpiredStatusFromEndOrStartAt(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	now := time.Now().UTC()
	endedAt := now.Add(-time.Hour)
	event := repo.seedEvent(Event{
		OwnerID:   uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8"),
		Title:     "Dinner",
		Type:      "dining",
		Scope:     "personal",
		StartAt:   now.Add(-2 * time.Hour),
		EndAt:     &endedAt,
		Status:    StatusActive,
		ShareSlug: "expired001",
	})

	detail, err := service.GetDetail(context.Background(), "expired001")
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}

	if detail.Event.ID != event.ID {
		t.Fatalf("expected event %d, got %d", event.ID, detail.Event.ID)
	}
	if detail.Event.Status != StatusExpired {
		t.Fatalf("expected derived status %q, got %q", StatusExpired, detail.Event.Status)
	}
}

type fakeEventRepository struct {
	events       map[int64]Event
	eventsBySlug map[string]int64
	goingCounts  map[int64]int
	nextID       int64
}

func newFakeEventRepository() *fakeEventRepository {
	return &fakeEventRepository{
		events:       map[int64]Event{},
		eventsBySlug: map[string]int64{},
		goingCounts:  map[int64]int{},
		nextID:       1,
	}
}

func (r *fakeEventRepository) Create(ctx context.Context, event Event) (Event, error) {
	if _, ok := r.eventsBySlug[event.ShareSlug]; ok {
		return Event{}, ErrShareSlugConflict
	}
	event.ID = r.nextID
	r.nextID++
	r.events[event.ID] = event
	r.eventsBySlug[event.ShareSlug] = event.ID
	return event, nil
}

func (r *fakeEventRepository) FindBySlug(ctx context.Context, slug string) (Event, error) {
	id, ok := r.eventsBySlug[slug]
	if !ok {
		return Event{}, gorm.ErrRecordNotFound
	}
	return r.events[id], nil
}

func (r *fakeEventRepository) FindByID(ctx context.Context, id int64) (Event, error) {
	event, ok := r.events[id]
	if !ok {
		return Event{}, gorm.ErrRecordNotFound
	}
	return event, nil
}

func (r *fakeEventRepository) CountGoing(ctx context.Context, eventID int64) (int, error) {
	return r.goingCounts[eventID], nil
}

func (r *fakeEventRepository) ListMine(ctx context.Context, ownerID uuid.UUID) ([]Event, error) {
	var mine []Event
	for _, event := range r.events {
		if event.OwnerID == ownerID {
			mine = append(mine, event)
		}
	}
	return mine, nil
}

func (r *fakeEventRepository) seedEvent(event Event) Event {
	if event.ID == 0 {
		event.ID = r.nextID
		r.nextID++
	}
	r.events[event.ID] = event
	r.eventsBySlug[event.ShareSlug] = event.ID
	return event
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
