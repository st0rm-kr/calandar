package events

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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

func TestCreateRetriesPostgresUniqueViolationForShareSlug(t *testing.T) {
	repo := newFakeEventRepository()
	repo.createErrors = []error{&pgconn.PgError{
		Code:           "23505",
		ConstraintName: "events_share_slug_key",
	}}
	service := NewService(repo)
	ownerID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	startAt := time.Now().Add(time.Hour).UTC()

	event, err := service.Create(context.Background(), ownerID, CreateEventInput{
		Title:   "Climb",
		Type:    "climbing",
		Scope:   "personal",
		StartAt: startAt,
	})
	if err != nil {
		t.Fatalf("expected retry to create event, got error: %v", err)
	}

	if event.ID == 0 {
		t.Fatalf("expected event to be persisted after retry, got %+v", event)
	}
	if repo.createAttempts != 2 {
		t.Fatalf("expected 2 create attempts, got %d", repo.createAttempts)
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

func TestRSVPGoingInsertsParticipantAndEventSchedule(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	userID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	event := repo.seedEvent(activeEvent())
	otherEventID := int64(99)
	repo.conflicts = []ScheduleConflict{
		{ID: 1, Title: event.Title, EventID: &event.ID},
		{ID: 2, Title: "Existing schedule", EventID: &otherEventID},
	}

	result, err := service.RSVP(context.Background(), userID, event.ID, RSVPInput{
		RSVP:          RSVPGoing,
		AddToCalendar: true,
		Visibility:    "busy_only",
	})
	if err != nil {
		t.Fatalf("rsvp going: %v", err)
	}

	participant := repo.participants[participantKey(event.ID, userID)]
	if participant.RSVP != RSVPGoing || participant.Source != ParticipantSourceSelf {
		t.Fatalf("expected going self participant, got %+v", participant)
	}
	schedule := repo.schedules[participantKey(event.ID, userID)]
	if schedule.EventID != event.ID || schedule.Title != event.Title || schedule.Visibility != "busy_only" {
		t.Fatalf("expected linked event schedule, got %+v", schedule)
	}
	if result.Participant.RSVP != RSVPGoing {
		t.Fatalf("expected result participant going, got %+v", result.Participant)
	}
	if result.GoingCount != 1 {
		t.Fatalf("expected going count 1, got %d", result.GoingCount)
	}
	if len(result.Conflicts) != 1 || result.Conflicts[0].ID != 2 {
		t.Fatalf("expected RSVP result to filter self event conflict, got %+v", result.Conflicts)
	}
}

func TestRepeatedRSVPUpdatesSingleParticipantCount(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	userID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	event := repo.seedEvent(activeEvent())

	result, err := service.RSVP(context.Background(), userID, event.ID, RSVPInput{
		RSVP:          RSVPGoing,
		AddToCalendar: true,
		Visibility:    "busy_only",
	})
	if err != nil {
		t.Fatalf("first rsvp going: %v", err)
	}
	if result.GoingCount != 1 {
		t.Fatalf("expected first going count 1, got %d", result.GoingCount)
	}

	result, err = service.RSVP(context.Background(), userID, event.ID, RSVPInput{
		RSVP:          RSVPGoing,
		AddToCalendar: true,
		Visibility:    "busy_only",
	})
	if err != nil {
		t.Fatalf("second rsvp going: %v", err)
	}
	if result.GoingCount != 1 {
		t.Fatalf("expected repeated going count to stay 1, got %d", result.GoingCount)
	}

	result, err = service.RSVP(context.Background(), userID, event.ID, RSVPInput{
		RSVP:          RSVPNotGoing,
		AddToCalendar: false,
	})
	if err != nil {
		t.Fatalf("rsvp not going: %v", err)
	}
	if result.GoingCount != 0 {
		t.Fatalf("expected not going count 0, got %d", result.GoingCount)
	}

	if len(repo.participants) != 1 {
		t.Fatalf("expected one participant row, got %d", len(repo.participants))
	}
}

func TestRSVPMaybeInsertsParticipantAndEventSchedule(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	userID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	event := repo.seedEvent(activeEvent())

	_, err := service.RSVP(context.Background(), userID, event.ID, RSVPInput{
		RSVP:          RSVPMaybe,
		AddToCalendar: true,
		Visibility:    "public",
	})
	if err != nil {
		t.Fatalf("rsvp maybe: %v", err)
	}

	participant := repo.participants[participantKey(event.ID, userID)]
	if participant.RSVP != RSVPMaybe {
		t.Fatalf("expected maybe participant, got %+v", participant)
	}
	if _, ok := repo.schedules[participantKey(event.ID, userID)]; !ok {
		t.Fatalf("expected event schedule to be linked")
	}
}

func TestRSVPNotGoingDeletesExistingEventSchedule(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	userID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	event := repo.seedEvent(activeEvent())
	repo.schedules[participantKey(event.ID, userID)] = fakeSchedule{EventID: event.ID, UserID: userID}

	_, err := service.RSVP(context.Background(), userID, event.ID, RSVPInput{
		RSVP:          RSVPNotGoing,
		AddToCalendar: true,
	})
	if err != nil {
		t.Fatalf("rsvp not going: %v", err)
	}

	if _, ok := repo.schedules[participantKey(event.ID, userID)]; ok {
		t.Fatalf("expected event schedule to be deleted")
	}
	if participant := repo.participants[participantKey(event.ID, userID)]; participant.RSVP != RSVPNotGoing {
		t.Fatalf("expected not_going participant, got %+v", participant)
	}
}

func TestRSVPAddToCalendarFalseStoresParticipantAndDeletesSchedule(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	userID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	event := repo.seedEvent(activeEvent())
	repo.schedules[participantKey(event.ID, userID)] = fakeSchedule{EventID: event.ID, UserID: userID}

	_, err := service.RSVP(context.Background(), userID, event.ID, RSVPInput{
		RSVP:          RSVPGoing,
		AddToCalendar: false,
	})
	if err != nil {
		t.Fatalf("rsvp without calendar: %v", err)
	}

	if _, ok := repo.schedules[participantKey(event.ID, userID)]; ok {
		t.Fatalf("expected event schedule to be deleted")
	}
	participant := repo.participants[participantKey(event.ID, userID)]
	if participant.RSVP != RSVPGoing || participant.AddToCalendar {
		t.Fatalf("expected going participant without calendar, got %+v", participant)
	}
}

func TestRSVPCapacityFullReturnsDomainError(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	capacity := 1
	event := activeEvent()
	event.Capacity = &capacity
	event = repo.seedEvent(event)
	repo.participants[participantKey(event.ID, uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"))] = Participant{
		EventID: event.ID,
		UserID:  uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"),
		RSVP:    RSVPGoing,
		Source:  ParticipantSourceSelf,
	}

	_, err := service.RSVP(context.Background(), uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"), event.ID, RSVPInput{
		RSVP:          RSVPGoing,
		AddToCalendar: true,
	})

	if !errors.Is(err, ErrCapacityFull) {
		t.Fatalf("expected ErrCapacityFull, got %v", err)
	}
}

func TestRSVPUpdatingExistingParticipantIsIdempotent(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	userID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	event := repo.seedEvent(activeEvent())

	for i := 0; i < 2; i++ {
		_, err := service.RSVP(context.Background(), userID, event.ID, RSVPInput{
			RSVP:          RSVPGoing,
			AddToCalendar: true,
			Visibility:    "busy_only",
		})
		if err != nil {
			t.Fatalf("rsvp attempt %d: %v", i+1, err)
		}
	}

	if len(repo.participants) != 1 {
		t.Fatalf("expected one participant row, got %d", len(repo.participants))
	}
	if len(repo.schedules) != 1 {
		t.Fatalf("expected one schedule row, got %d", len(repo.schedules))
	}
}

func TestRSVPInvitedParticipantChangingToGoingKeepsOneRow(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	userID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	event := repo.seedEvent(activeEvent())
	repo.participants[participantKey(event.ID, userID)] = Participant{
		ID:      44,
		EventID: event.ID,
		UserID:  userID,
		RSVP:    RSVPInvited,
		Source:  ParticipantSourceInvited,
	}

	_, err := service.RSVP(context.Background(), userID, event.ID, RSVPInput{
		RSVP:          RSVPGoing,
		AddToCalendar: true,
	})
	if err != nil {
		t.Fatalf("rsvp invited to going: %v", err)
	}

	if len(repo.participants) != 1 {
		t.Fatalf("expected one participant row, got %d", len(repo.participants))
	}
	participant := repo.participants[participantKey(event.ID, userID)]
	if participant.ID != 44 || participant.RSVP != RSVPGoing || participant.Source != ParticipantSourceSelf {
		t.Fatalf("expected existing participant updated to going self, got %+v", participant)
	}
}

func TestRSVPRejectsUserSubmittedInvited(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	userID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	event := repo.seedEvent(activeEvent())

	_, err := service.RSVP(context.Background(), userID, event.ID, RSVPInput{
		RSVP:          RSVPInvited,
		AddToCalendar: true,
	})

	if !errors.Is(err, ErrInvalidRSVP) {
		t.Fatalf("expected ErrInvalidRSVP, got %v", err)
	}
	if len(repo.participants) != 0 {
		t.Fatalf("expected no participant row for user-submitted invited, got %d", len(repo.participants))
	}
}

func TestCancelDeletesEventSchedules(t *testing.T) {
	repo := newFakeEventRepository()
	service := NewService(repo)
	event := repo.seedEvent(activeEvent())
	ownerID := event.OwnerID
	firstUserID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	secondUserID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	repo.schedules[participantKey(event.ID, firstUserID)] = fakeSchedule{EventID: event.ID, UserID: firstUserID}
	repo.schedules[participantKey(event.ID, secondUserID)] = fakeSchedule{EventID: event.ID, UserID: secondUserID}

	cancelled, err := service.Cancel(context.Background(), ownerID, event.ID)
	if err != nil {
		t.Fatalf("cancel event: %v", err)
	}

	if cancelled.Status != StatusCancelled {
		t.Fatalf("expected cancelled status, got %q", cancelled.Status)
	}
	if len(repo.schedules) != 0 {
		t.Fatalf("expected cancel to delete event schedules, got %d rows", len(repo.schedules))
	}
}

func TestRSVPConcurrentGoingClaimsOnlyLastSeat(t *testing.T) {
	repo := newConcurrentRSVPRepository()
	service := NewService(repo)
	event := activeEvent()
	capacity := 1
	event.Capacity = &capacity
	event = repo.seedEvent(event)
	users := []uuid.UUID{
		uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"),
		uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"),
	}

	var wg sync.WaitGroup
	errs := make([]error, len(users))
	for i, userID := range users {
		wg.Add(1)
		go func(i int, userID uuid.UUID) {
			defer wg.Done()
			_, errs[i] = service.RSVP(context.Background(), userID, event.ID, RSVPInput{
				RSVP:          RSVPGoing,
				AddToCalendar: true,
			})
		}(i, userID)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for concurrent RSVP calls")
	}

	successes := 0
	capacityErrors := 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrCapacityFull):
			capacityErrors++
		default:
			t.Fatalf("unexpected RSVP error: %v", err)
		}
	}
	if successes != 1 || capacityErrors != 1 {
		t.Fatalf("expected one success and one capacity error, got successes=%d capacity_errors=%d errs=%v", successes, capacityErrors, errs)
	}
	if going := repo.goingParticipantCount(event.ID); going != 1 {
		t.Fatalf("expected one going participant after concurrent RSVP, got %d", going)
	}
}

type fakeEventRepository struct {
	events         map[int64]Event
	eventsBySlug   map[string]int64
	participants   map[string]Participant
	schedules      map[string]fakeSchedule
	conflicts      []ScheduleConflict
	goingCounts    map[int64]int
	createErrors   []error
	createAttempts int
	nextID         int64
}

func newFakeEventRepository() *fakeEventRepository {
	return &fakeEventRepository{
		events:       map[int64]Event{},
		eventsBySlug: map[string]int64{},
		participants: map[string]Participant{},
		schedules:    map[string]fakeSchedule{},
		goingCounts:  map[int64]int{},
		nextID:       1,
	}
}

func (r *fakeEventRepository) Create(ctx context.Context, event Event) (Event, error) {
	r.createAttempts++
	if len(r.createErrors) > 0 {
		err := r.createErrors[0]
		r.createErrors = r.createErrors[1:]
		return Event{}, err
	}
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

func (r *fakeEventRepository) FindByIDForUpdate(ctx context.Context, id int64) (Event, error) {
	return r.FindByID(ctx, id)
}

func (r *fakeEventRepository) CountGoing(ctx context.Context, eventID int64) (int, error) {
	count := r.goingCounts[eventID]
	for _, participant := range r.participants {
		if participant.EventID == eventID && participant.RSVP == RSVPGoing && participant.DeletedAt == nil {
			count++
		}
	}
	return count, nil
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

func (r *fakeEventRepository) Transaction(ctx context.Context, fn func(EventRepository) error) error {
	return fn(r)
}

func (r *fakeEventRepository) CountGoingForUpdateExcludingUser(ctx context.Context, eventID int64, userID uuid.UUID) (int, error) {
	count := 0
	for _, participant := range r.participants {
		if participant.EventID == eventID && participant.UserID != userID && participant.RSVP == RSVPGoing && participant.DeletedAt == nil {
			count++
		}
	}
	return count, nil
}

func (r *fakeEventRepository) UpsertParticipant(ctx context.Context, participant Participant) (Participant, error) {
	key := participantKey(participant.EventID, participant.UserID)
	existing, ok := r.participants[key]
	if ok {
		participant.ID = existing.ID
	} else {
		participant.ID = r.nextID
		r.nextID++
	}
	r.participants[key] = participant
	return participant, nil
}

func (r *fakeEventRepository) FindParticipant(ctx context.Context, eventID int64, userID uuid.UUID) (Participant, bool, error) {
	participant, ok := r.participants[participantKey(eventID, userID)]
	if !ok || participant.DeletedAt != nil {
		return Participant{}, false, nil
	}
	return participant, true, nil
}

func (r *fakeEventRepository) ListParticipants(ctx context.Context, eventID int64) ([]Participant, error) {
	var result []Participant
	for _, participant := range r.participants {
		if participant.EventID == eventID && participant.DeletedAt == nil {
			result = append(result, participant)
		}
	}
	return result, nil
}

func (r *fakeEventRepository) UpsertEventSchedule(ctx context.Context, userID uuid.UUID, event Event, visibility string) error {
	key := participantKey(event.ID, userID)
	existing := r.schedules[key]
	if existing.Visibility == "" {
		existing.Visibility = visibility
	}
	existing.EventID = event.ID
	existing.UserID = userID
	existing.Title = event.Title
	r.schedules[key] = existing
	return nil
}

func (r *fakeEventRepository) DeleteEventSchedule(ctx context.Context, userID uuid.UUID, eventID int64) error {
	delete(r.schedules, participantKey(eventID, userID))
	return nil
}

func (r *fakeEventRepository) DeleteEventSchedulesForEvent(ctx context.Context, eventID int64) error {
	for key, schedule := range r.schedules {
		if schedule.EventID == eventID {
			delete(r.schedules, key)
		}
	}
	return nil
}

func (r *fakeEventRepository) FindConflicts(ctx context.Context, userID uuid.UUID, start time.Time, end *time.Time) ([]ScheduleConflict, error) {
	return r.conflicts, nil
}

func (r *fakeEventRepository) Cancel(ctx context.Context, userID uuid.UUID, eventID int64) (Event, error) {
	event, ok := r.events[eventID]
	if !ok {
		return Event{}, gorm.ErrRecordNotFound
	}
	event.Status = StatusCancelled
	r.events[eventID] = event
	return event, nil
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

func activeEvent() Event {
	endAt := time.Now().Add(2 * time.Hour).UTC()
	return Event{
		OwnerID:   uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8"),
		Title:     "Climb",
		Type:      "climbing",
		Scope:     "personal",
		StartAt:   time.Now().Add(time.Hour).UTC(),
		EndAt:     &endAt,
		Status:    StatusActive,
		ShareSlug: "abc123def4",
	}
}

func participantKey(eventID int64, userID uuid.UUID) string {
	return fmt.Sprintf("%s:%d", userID.String(), eventID)
}

type fakeSchedule struct {
	EventID    int64
	UserID     uuid.UUID
	Title      string
	Visibility string
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

type concurrentRSVPRepository struct {
	mu           sync.Mutex
	eventLock    sync.Mutex
	event        Event
	participants map[string]Participant
	nextID       int64
	countArrived chan struct{}
	counted      chan struct{}
}

type concurrentRSVPTransaction struct {
	shared *concurrentRSVPRepository
	locked bool
}

func newConcurrentRSVPRepository() *concurrentRSVPRepository {
	return &concurrentRSVPRepository{
		participants: map[string]Participant{},
		nextID:       1,
		countArrived: make(chan struct{}, 2),
		counted:      make(chan struct{}, 2),
	}
}

func (r *concurrentRSVPRepository) seedEvent(event Event) Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	if event.ID == 0 {
		event.ID = r.nextID
		r.nextID++
	}
	r.event = event
	return event
}

func (r *concurrentRSVPRepository) goingParticipantCount(eventID int64) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.countGoingLocked(eventID, uuid.Nil)
}

func (r *concurrentRSVPRepository) Create(ctx context.Context, event Event) (Event, error) {
	return Event{}, nil
}

func (r *concurrentRSVPRepository) FindBySlug(ctx context.Context, slug string) (Event, error) {
	return Event{}, gorm.ErrRecordNotFound
}

func (r *concurrentRSVPRepository) FindByID(ctx context.Context, id int64) (Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.event.ID != id {
		return Event{}, gorm.ErrRecordNotFound
	}
	return r.event, nil
}

func (r *concurrentRSVPRepository) FindByIDForUpdate(ctx context.Context, id int64) (Event, error) {
	r.eventLock.Lock()
	defer r.eventLock.Unlock()
	return r.FindByID(ctx, id)
}

func (r *concurrentRSVPRepository) CountGoing(ctx context.Context, eventID int64) (int, error) {
	return 0, nil
}

func (r *concurrentRSVPRepository) ListMine(ctx context.Context, ownerID uuid.UUID) ([]Event, error) {
	return nil, nil
}

func (r *concurrentRSVPRepository) Transaction(ctx context.Context, fn func(EventRepository) error) error {
	tx := &concurrentRSVPTransaction{shared: r}
	defer func() {
		if tx.locked {
			r.eventLock.Unlock()
		}
	}()
	return fn(tx)
}

func (r *concurrentRSVPRepository) CountGoingForUpdateExcludingUser(ctx context.Context, eventID int64, userID uuid.UUID) (int, error) {
	if err := waitForConcurrentSnapshot(ctx, r.countArrived); err != nil {
		return 0, err
	}
	r.mu.Lock()
	count := r.countGoingLocked(eventID, userID)
	r.mu.Unlock()
	if err := waitForConcurrentSnapshot(ctx, r.counted); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *concurrentRSVPRepository) UpsertParticipant(ctx context.Context, participant Participant) (Participant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if participant.ID == 0 {
		participant.ID = r.nextID
		r.nextID++
	}
	r.participants[participantKey(participant.EventID, participant.UserID)] = participant
	return participant, nil
}

func (r *concurrentRSVPRepository) FindParticipant(ctx context.Context, eventID int64, userID uuid.UUID) (Participant, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	participant, ok := r.participants[participantKey(eventID, userID)]
	if !ok || participant.DeletedAt != nil {
		return Participant{}, false, nil
	}
	return participant, true, nil
}

func (r *concurrentRSVPRepository) ListParticipants(ctx context.Context, eventID int64) ([]Participant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []Participant
	for _, participant := range r.participants {
		if participant.EventID == eventID && participant.DeletedAt == nil {
			result = append(result, participant)
		}
	}
	return result, nil
}

func (r *concurrentRSVPRepository) UpsertEventSchedule(ctx context.Context, userID uuid.UUID, event Event, visibility string) error {
	return nil
}

func (r *concurrentRSVPRepository) DeleteEventSchedule(ctx context.Context, userID uuid.UUID, eventID int64) error {
	return nil
}

func (r *concurrentRSVPRepository) DeleteEventSchedulesForEvent(ctx context.Context, eventID int64) error {
	return nil
}

func (r *concurrentRSVPRepository) FindConflicts(ctx context.Context, userID uuid.UUID, start time.Time, end *time.Time) ([]ScheduleConflict, error) {
	return nil, nil
}

func (r *concurrentRSVPRepository) Cancel(ctx context.Context, userID uuid.UUID, eventID int64) (Event, error) {
	return Event{}, nil
}

func (r *concurrentRSVPRepository) countGoingLocked(eventID int64, excludeUserID uuid.UUID) int {
	count := 0
	for _, participant := range r.participants {
		if participant.EventID == eventID && participant.UserID != excludeUserID && participant.RSVP == RSVPGoing && participant.DeletedAt == nil {
			count++
		}
	}
	return count
}

func waitForConcurrentSnapshot(ctx context.Context, ch chan struct{}) error {
	ch <- struct{}{}
	deadline := time.After(500 * time.Millisecond)
	for {
		if len(ch) == cap(ch) {
			return nil
		}
		select {
		case <-deadline:
			return context.DeadlineExceeded
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}
}

func (tx *concurrentRSVPTransaction) Create(ctx context.Context, event Event) (Event, error) {
	return tx.shared.Create(ctx, event)
}

func (tx *concurrentRSVPTransaction) FindBySlug(ctx context.Context, slug string) (Event, error) {
	return tx.shared.FindBySlug(ctx, slug)
}

func (tx *concurrentRSVPTransaction) FindByID(ctx context.Context, id int64) (Event, error) {
	return tx.shared.FindByID(ctx, id)
}

func (tx *concurrentRSVPTransaction) FindByIDForUpdate(ctx context.Context, id int64) (Event, error) {
	tx.shared.eventLock.Lock()
	tx.locked = true
	return tx.shared.FindByID(ctx, id)
}

func (tx *concurrentRSVPTransaction) CountGoing(ctx context.Context, eventID int64) (int, error) {
	return tx.shared.CountGoing(ctx, eventID)
}

func (tx *concurrentRSVPTransaction) ListMine(ctx context.Context, ownerID uuid.UUID) ([]Event, error) {
	return tx.shared.ListMine(ctx, ownerID)
}

func (tx *concurrentRSVPTransaction) Transaction(ctx context.Context, fn func(EventRepository) error) error {
	return tx.shared.Transaction(ctx, fn)
}

func (tx *concurrentRSVPTransaction) CountGoingForUpdateExcludingUser(ctx context.Context, eventID int64, userID uuid.UUID) (int, error) {
	if tx.locked {
		tx.shared.mu.Lock()
		defer tx.shared.mu.Unlock()
		return tx.shared.countGoingLocked(eventID, userID), nil
	}
	return tx.shared.CountGoingForUpdateExcludingUser(ctx, eventID, userID)
}

func (tx *concurrentRSVPTransaction) UpsertParticipant(ctx context.Context, participant Participant) (Participant, error) {
	return tx.shared.UpsertParticipant(ctx, participant)
}

func (tx *concurrentRSVPTransaction) FindParticipant(ctx context.Context, eventID int64, userID uuid.UUID) (Participant, bool, error) {
	return tx.shared.FindParticipant(ctx, eventID, userID)
}

func (tx *concurrentRSVPTransaction) ListParticipants(ctx context.Context, eventID int64) ([]Participant, error) {
	return tx.shared.ListParticipants(ctx, eventID)
}

func (tx *concurrentRSVPTransaction) UpsertEventSchedule(ctx context.Context, userID uuid.UUID, event Event, visibility string) error {
	return tx.shared.UpsertEventSchedule(ctx, userID, event, visibility)
}

func (tx *concurrentRSVPTransaction) DeleteEventSchedule(ctx context.Context, userID uuid.UUID, eventID int64) error {
	return tx.shared.DeleteEventSchedule(ctx, userID, eventID)
}

func (tx *concurrentRSVPTransaction) DeleteEventSchedulesForEvent(ctx context.Context, eventID int64) error {
	return tx.shared.DeleteEventSchedulesForEvent(ctx, eventID)
}

func (tx *concurrentRSVPTransaction) FindConflicts(ctx context.Context, userID uuid.UUID, start time.Time, end *time.Time) ([]ScheduleConflict, error) {
	return tx.shared.FindConflicts(ctx, userID, start, end)
}

func (tx *concurrentRSVPTransaction) Cancel(ctx context.Context, userID uuid.UUID, eventID int64) (Event, error) {
	return tx.shared.Cancel(ctx, userID, eventID)
}
