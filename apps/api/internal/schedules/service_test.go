package schedules

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed.UTC()
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func TestManualSchedulesCreateWithoutOverlapReturnsNoConflicts(t *testing.T) {
	repo := newFakeScheduleRepository()
	service := NewService(repo)
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")

	result, err := service.CreateManual(context.Background(), userID, CreateManualInput{
		Title:      "Gym",
		StartAt:    mustTime(t, "2026-06-24T09:00:00Z"),
		EndAt:      ptrTime(mustTime(t, "2026-06-24T10:00:00Z")),
		Visibility: VisibilityBusyOnly,
	})
	if err != nil {
		t.Fatalf("create manual: %v", err)
	}
	if result.Schedule.ID == 0 {
		t.Fatalf("expected schedule to be persisted, got %+v", result.Schedule)
	}
	if result.Schedule.Source != SourceManual {
		t.Fatalf("expected manual source, got %q", result.Schedule.Source)
	}
	if len(result.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %d", len(result.Conflicts))
	}
}

func TestManualSchedulesCreateWithOverlapReturnsConflicts(t *testing.T) {
	repo := newFakeScheduleRepository()
	service := NewService(repo)
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	repo.seed(Schedule{
		UserID:     userID,
		Title:      "Existing",
		StartAt:    mustTime(t, "2026-06-24T08:00:00Z"),
		EndAt:      ptrTime(mustTime(t, "2026-06-24T09:30:00Z")),
		Visibility: VisibilityBusyOnly,
		Source:     SourceManual,
	})

	result, err := service.CreateManual(context.Background(), userID, CreateManualInput{
		Title:      "Overlapping",
		StartAt:    mustTime(t, "2026-06-24T09:00:00Z"),
		EndAt:      ptrTime(mustTime(t, "2026-06-24T10:00:00Z")),
		Visibility: VisibilityBusyOnly,
	})
	if err != nil {
		t.Fatalf("create manual: %v", err)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("expected one conflict, got %d", len(result.Conflicts))
	}
	if result.Conflicts[0].Title != "Existing" {
		t.Fatalf("expected conflict with existing schedule, got %q", result.Conflicts[0].Title)
	}
}

func TestManualSchedulesUpdateCannotModifyAnotherUsersSchedule(t *testing.T) {
	repo := newFakeScheduleRepository()
	service := NewService(repo)
	ownerID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	otherID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	existing := repo.seed(Schedule{
		UserID:     ownerID,
		Title:      "Owner schedule",
		StartAt:    mustTime(t, "2026-06-24T08:00:00Z"),
		Visibility: VisibilityBusyOnly,
		Source:     SourceManual,
	})

	_, err := service.UpdateManual(context.Background(), otherID, existing.ID, UpdateManualInput{
		Title:      "Hijacked",
		StartAt:    mustTime(t, "2026-06-24T08:00:00Z"),
		Visibility: VisibilityBusyOnly,
	})
	if !errors.Is(err, ErrScheduleNotFound) {
		t.Fatalf("expected ErrScheduleNotFound, got %v", err)
	}
}

func TestManualSchedulesDeleteRejectsEventSchedule(t *testing.T) {
	repo := newFakeScheduleRepository()
	service := NewService(repo)
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	eventID := int64(77)
	eventSchedule := repo.seed(Schedule{
		UserID:     userID,
		Title:      "Event schedule",
		StartAt:    mustTime(t, "2026-06-24T08:00:00Z"),
		Visibility: VisibilityBusyOnly,
		Source:     SourceEvent,
		EventID:    &eventID,
	})

	err := service.DeleteManual(context.Background(), userID, eventSchedule.ID)
	if !errors.Is(err, ErrCannotModifyEventSchedule) {
		t.Fatalf("expected ErrCannotModifyEventSchedule, got %v", err)
	}
}

func TestManualSchedulesListByRangeReturnsOverlappingWindow(t *testing.T) {
	repo := newFakeScheduleRepository()
	service := NewService(repo)
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	inRange := repo.seed(Schedule{
		UserID:     userID,
		Title:      "In range",
		StartAt:    mustTime(t, "2026-06-10T08:00:00Z"),
		EndAt:      ptrTime(mustTime(t, "2026-06-10T09:00:00Z")),
		Visibility: VisibilityBusyOnly,
		Source:     SourceManual,
	})
	repo.seed(Schedule{
		UserID:     userID,
		Title:      "Out of range",
		StartAt:    mustTime(t, "2026-07-10T08:00:00Z"),
		Visibility: VisibilityBusyOnly,
		Source:     SourceManual,
	})

	from := mustTime(t, "2026-06-01T00:00:00Z")
	to := mustTime(t, "2026-07-01T00:00:00Z")
	list, err := service.ListByRange(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("list by range: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected one schedule in range, got %d", len(list))
	}
	if list[0].ID != inRange.ID {
		t.Fatalf("expected in-range schedule, got %+v", list[0])
	}
}

func TestManualSchedulesCreateRejectsInvalidVisibility(t *testing.T) {
	repo := newFakeScheduleRepository()
	service := NewService(repo)
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")

	_, err := service.CreateManual(context.Background(), userID, CreateManualInput{
		Title:      "Bad visibility",
		StartAt:    mustTime(t, "2026-06-24T09:00:00Z"),
		Visibility: "everyone",
	})
	if !errors.Is(err, ErrInvalidVisibility) {
		t.Fatalf("expected ErrInvalidVisibility, got %v", err)
	}
}

func TestManualSchedulesCreateRejectsEndBeforeStart(t *testing.T) {
	repo := newFakeScheduleRepository()
	service := NewService(repo)
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")

	_, err := service.CreateManual(context.Background(), userID, CreateManualInput{
		Title:      "Bad time",
		StartAt:    mustTime(t, "2026-06-24T10:00:00Z"),
		EndAt:      ptrTime(mustTime(t, "2026-06-24T09:00:00Z")),
		Visibility: VisibilityBusyOnly,
	})
	if !errors.Is(err, ErrInvalidScheduleTime) {
		t.Fatalf("expected ErrInvalidScheduleTime, got %v", err)
	}
}

type fakeScheduleRepository struct {
	schedules map[int64]Schedule
	nextID    int64
}

func newFakeScheduleRepository() *fakeScheduleRepository {
	return &fakeScheduleRepository{schedules: map[int64]Schedule{}, nextID: 0}
}

func (r *fakeScheduleRepository) seed(schedule Schedule) Schedule {
	r.nextID++
	schedule.ID = r.nextID
	r.schedules[schedule.ID] = schedule
	return schedule
}

func (r *fakeScheduleRepository) ListByRange(_ context.Context, userID uuid.UUID, from, to time.Time) ([]Schedule, error) {
	var list []Schedule
	for _, schedule := range r.schedules {
		if schedule.UserID != userID || schedule.DeletedAt != nil {
			continue
		}
		if schedule.StartAt.Before(to) && from.Before(effectiveEnd(schedule.StartAt, schedule.EndAt)) {
			list = append(list, schedule)
		}
	}
	return list, nil
}

func (r *fakeScheduleRepository) FindByID(_ context.Context, userID uuid.UUID, id int64) (Schedule, error) {
	schedule, ok := r.schedules[id]
	if !ok || schedule.UserID != userID || schedule.DeletedAt != nil {
		return Schedule{}, ErrScheduleNotFound
	}
	return schedule, nil
}

func (r *fakeScheduleRepository) CreateManual(_ context.Context, schedule Schedule) (Schedule, error) {
	r.nextID++
	schedule.ID = r.nextID
	schedule.Source = SourceManual
	r.schedules[schedule.ID] = schedule
	return schedule, nil
}

func (r *fakeScheduleRepository) UpdateManual(_ context.Context, userID uuid.UUID, id int64, patch SchedulePatch) (Schedule, error) {
	schedule, ok := r.schedules[id]
	if !ok || schedule.UserID != userID || schedule.Source != SourceManual || schedule.DeletedAt != nil {
		return Schedule{}, ErrScheduleNotFound
	}
	schedule.Title = patch.Title
	schedule.StartAt = patch.StartAt
	schedule.EndAt = patch.EndAt
	schedule.Location = patch.Location
	schedule.Visibility = patch.Visibility
	r.schedules[id] = schedule
	return schedule, nil
}

func (r *fakeScheduleRepository) DeleteManual(_ context.Context, userID uuid.UUID, id int64) error {
	schedule, ok := r.schedules[id]
	if !ok || schedule.UserID != userID || schedule.Source != SourceManual || schedule.DeletedAt != nil {
		return ErrScheduleNotFound
	}
	delete(r.schedules, id)
	return nil
}

func (r *fakeScheduleRepository) FindConflictsExcluding(_ context.Context, userID uuid.UUID, start time.Time, end *time.Time, excludeID *int64) ([]Conflict, error) {
	newEnd := effectiveEnd(start, end)
	var conflicts []Conflict
	for _, schedule := range r.schedules {
		if schedule.UserID != userID || schedule.DeletedAt != nil {
			continue
		}
		if excludeID != nil && schedule.ID == *excludeID {
			continue
		}
		if schedule.StartAt.Before(newEnd) && start.Before(effectiveEnd(schedule.StartAt, schedule.EndAt)) {
			conflicts = append(conflicts, Conflict{
				ID:         schedule.ID,
				Title:      schedule.Title,
				StartAt:    schedule.StartAt,
				EndAt:      schedule.EndAt,
				Location:   schedule.Location,
				Visibility: schedule.Visibility,
				Source:     schedule.Source,
				EventID:    schedule.EventID,
			})
		}
	}
	return conflicts, nil
}
