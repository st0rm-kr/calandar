package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/schedules"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestSchedulesConflictsReturnsOverlaps(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	repo := newFakeScheduleRepo()
	end := mustParse(t, "2026-06-24T09:30:00Z")
	repo.seed(schedules.Schedule{
		UserID:     userID,
		Title:      "Existing",
		StartAt:    mustParse(t, "2026-06-24T08:00:00Z"),
		EndAt:      &end,
		Visibility: schedules.VisibilityBusyOnly,
		Source:     schedules.SourceManual,
	})
	router := NewRouter(Dependencies{
		Config:          configWithJWTSecret(),
		ScheduleService: schedules.NewService(repo),
	})

	rec := doConflicts(t, router, userID, "start=2026-06-24T09:00:00Z&end=2026-06-24T10:00:00Z")
	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	conflicts := decodeConflicts(t, rec)
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %d", len(conflicts))
	}
}

func TestSchedulesConflictsAdjacentIsNotConflict(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	repo := newFakeScheduleRepo()
	end := mustParse(t, "2026-06-24T09:00:00Z")
	repo.seed(schedules.Schedule{
		UserID:     userID,
		Title:      "Earlier",
		StartAt:    mustParse(t, "2026-06-24T08:00:00Z"),
		EndAt:      &end,
		Visibility: schedules.VisibilityBusyOnly,
		Source:     schedules.SourceManual,
	})
	router := NewRouter(Dependencies{
		Config:          configWithJWTSecret(),
		ScheduleService: schedules.NewService(repo),
	})

	rec := doConflicts(t, router, userID, "start=2026-06-24T09:00:00Z&end=2026-06-24T10:00:00Z")
	conflicts := decodeConflicts(t, rec)
	if len(conflicts) != 0 {
		t.Fatalf("expected adjacent schedule not to conflict, got %d", len(conflicts))
	}
}

func TestSchedulesConflictsExcludeIDRemovesEditedSchedule(t *testing.T) {
	userID := uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8")
	repo := newFakeScheduleRepo()
	end := mustParse(t, "2026-06-24T10:00:00Z")
	seeded := repo.seed(schedules.Schedule{
		UserID:     userID,
		Title:      "Self",
		StartAt:    mustParse(t, "2026-06-24T09:00:00Z"),
		EndAt:      &end,
		Visibility: schedules.VisibilityBusyOnly,
		Source:     schedules.SourceManual,
	})
	router := NewRouter(Dependencies{
		Config:          configWithJWTSecret(),
		ScheduleService: schedules.NewService(repo),
	})

	query := "start=2026-06-24T09:00:00Z&end=2026-06-24T10:00:00Z&exclude_id=" + strconv.FormatInt(seeded.ID, 10)
	rec := doConflicts(t, router, userID, query)
	conflicts := decodeConflicts(t, rec)
	if len(conflicts) != 0 {
		t.Fatalf("expected excluded schedule to be removed, got %d", len(conflicts))
	}
}

func doConflicts(t *testing.T, router *gin.Engine, userID uuid.UUID, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(nethttp.MethodGet, "/api/schedules/conflicts?"+query, nil)
	req.Header.Set("Authorization", "Bearer "+signHS256JWT(t, userID))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeConflicts(t *testing.T, rec *httptest.ResponseRecorder) []schedules.Conflict {
	t.Helper()
	var body struct {
		Data  []schedules.Conflict `json:"data"`
		Error any                  `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v; body: %s", err, rec.Body.String())
	}
	return body.Data
}

func mustParse(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed.UTC()
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}

var _ = itoa

type fakeScheduleRepo struct {
	rows   map[int64]schedules.Schedule
	nextID int64
}

func newFakeScheduleRepo() *fakeScheduleRepo {
	return &fakeScheduleRepo{rows: map[int64]schedules.Schedule{}}
}

func (r *fakeScheduleRepo) seed(schedule schedules.Schedule) schedules.Schedule {
	r.nextID++
	schedule.ID = r.nextID
	r.rows[schedule.ID] = schedule
	return schedule
}

func (r *fakeScheduleRepo) ListByRange(_ context.Context, userID uuid.UUID, from, to time.Time) ([]schedules.Schedule, error) {
	var list []schedules.Schedule
	for _, schedule := range r.rows {
		if schedule.UserID != userID {
			continue
		}
		if schedule.StartAt.Before(to) && from.Before(scheduleEnd(schedule)) {
			list = append(list, schedule)
		}
	}
	return list, nil
}

func (r *fakeScheduleRepo) FindByID(_ context.Context, userID uuid.UUID, id int64) (schedules.Schedule, error) {
	schedule, ok := r.rows[id]
	if !ok || schedule.UserID != userID {
		return schedules.Schedule{}, schedules.ErrScheduleNotFound
	}
	return schedule, nil
}

func (r *fakeScheduleRepo) CreateManual(_ context.Context, schedule schedules.Schedule) (schedules.Schedule, error) {
	r.nextID++
	schedule.ID = r.nextID
	r.rows[schedule.ID] = schedule
	return schedule, nil
}

func (r *fakeScheduleRepo) UpdateManual(_ context.Context, userID uuid.UUID, id int64, patch schedules.SchedulePatch) (schedules.Schedule, error) {
	schedule, ok := r.rows[id]
	if !ok || schedule.UserID != userID {
		return schedules.Schedule{}, schedules.ErrScheduleNotFound
	}
	schedule.Title = patch.Title
	schedule.StartAt = patch.StartAt
	schedule.EndAt = patch.EndAt
	schedule.Location = patch.Location
	schedule.Visibility = patch.Visibility
	r.rows[id] = schedule
	return schedule, nil
}

func (r *fakeScheduleRepo) DeleteManual(_ context.Context, userID uuid.UUID, id int64) error {
	schedule, ok := r.rows[id]
	if !ok || schedule.UserID != userID {
		return schedules.ErrScheduleNotFound
	}
	delete(r.rows, id)
	return nil
}

func (r *fakeScheduleRepo) FindConflictsExcluding(_ context.Context, userID uuid.UUID, start time.Time, end *time.Time, excludeID *int64) ([]schedules.Conflict, error) {
	newEnd := start.Add(time.Minute)
	if end != nil {
		newEnd = *end
	}
	var conflicts []schedules.Conflict
	for _, schedule := range r.rows {
		if schedule.UserID != userID {
			continue
		}
		if excludeID != nil && schedule.ID == *excludeID {
			continue
		}
		if schedule.StartAt.Before(newEnd) && start.Before(scheduleEnd(schedule)) {
			conflicts = append(conflicts, schedules.Conflict{
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

func scheduleEnd(schedule schedules.Schedule) time.Time {
	if schedule.EndAt != nil {
		return *schedule.EndAt
	}
	return schedule.StartAt.Add(time.Minute)
}
