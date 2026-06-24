package calendar

import (
	"context"
	"testing"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/schedules"
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

func ptr[T any](value T) *T {
	return &value
}

type fakeScheduleReader struct {
	rows []schedules.Schedule
}

func (r fakeScheduleReader) ListByRange(_ context.Context, _ uuid.UUID, _ time.Time, _ time.Time) ([]schedules.Schedule, error) {
	return r.rows, nil
}

func TestCalendarIncludesManualAndEventSchedules(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	eventID := int64(9)
	reader := fakeScheduleReader{rows: []schedules.Schedule{
		{
			ID:         1,
			UserID:     userID,
			Title:      "Personal run",
			StartAt:    mustTime(t, "2026-06-10T08:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-10T09:00:00Z")),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceManual,
		},
		{
			ID:         2,
			UserID:     userID,
			Title:      "Climb event",
			StartAt:    mustTime(t, "2026-06-12T08:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-12T09:00:00Z")),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceEvent,
			EventID:    &eventID,
		},
	}}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "all")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 calendar items, got %d", len(items))
	}

	byID := map[int64]CalendarItem{}
	for _, item := range items {
		byID[item.ID] = item
	}
	if byID[1].Kind != "schedule" || byID[1].Color != "green" {
		t.Fatalf("expected manual schedule to be green schedule, got %+v", byID[1])
	}
	if byID[2].Kind != "event" || byID[2].Color != "blue" {
		t.Fatalf("expected event schedule to be blue event, got %+v", byID[2])
	}
}

func TestCalendarMarksBusyOnlyAsGray(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	reader := fakeScheduleReader{rows: []schedules.Schedule{
		{
			ID:         3,
			UserID:     userID,
			Title:      "Busy block",
			StartAt:    mustTime(t, "2026-06-15T08:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-15T09:00:00Z")),
			Visibility: schedules.VisibilityBusyOnly,
			Source:     schedules.SourceManual,
		},
	}}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "all")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Color != "gray" {
		t.Fatalf("expected busy_only item to be gray, got %q", items[0].Color)
	}
}

func TestCalendarMarksConflictingItems(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	reader := fakeScheduleReader{rows: []schedules.Schedule{
		{
			ID:         4,
			UserID:     userID,
			Title:      "A",
			StartAt:    mustTime(t, "2026-06-20T08:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-20T10:00:00Z")),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceManual,
		},
		{
			ID:         5,
			UserID:     userID,
			Title:      "B",
			StartAt:    mustTime(t, "2026-06-20T09:00:00Z"),
			EndAt:      ptr(mustTime(t, "2026-06-20T11:00:00Z")),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceManual,
		},
	}}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "all")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	for _, item := range items {
		if !item.HasConflict {
			t.Fatalf("expected overlapping item %d to be marked as conflict", item.ID)
		}
	}
}

func TestCalendarAcceptsGroupsFilter(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	eventID := int64(9)
	reader := fakeScheduleReader{rows: []schedules.Schedule{
		{
			ID:         6,
			UserID:     userID,
			Title:      "Personal",
			StartAt:    mustTime(t, "2026-06-10T08:00:00Z"),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceManual,
		},
		{
			ID:         7,
			UserID:     userID,
			Title:      "Group event",
			StartAt:    mustTime(t, "2026-06-12T08:00:00Z"),
			Visibility: schedules.VisibilityPublic,
			Source:     schedules.SourceEvent,
			EventID:    &eventID,
		},
	}}
	service := NewService(reader)

	items, err := service.Month(context.Background(), userID, mustTime(t, "2026-06-01T00:00:00Z"), mustTime(t, "2026-07-01T00:00:00Z"), "groups")
	if err != nil {
		t.Fatalf("calendar month: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected only event items with groups filter, got %d", len(items))
	}
	if items[0].ID != 7 {
		t.Fatalf("expected event item, got %+v", items[0])
	}
}
