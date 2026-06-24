package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/events"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestEventDetailRouteIsPublic(t *testing.T) {
	repo := &fakeEventRepository{}
	repo.event = events.Event{
		ID:        12,
		OwnerID:   uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8"),
		Scope:     "personal",
		Title:     "Climb",
		Type:      "climbing",
		StartAt:   time.Now().Add(time.Hour).UTC(),
		Status:    events.StatusActive,
		ShareSlug: "abc123def4",
	}
	repo.goingCount = 2
	router := NewRouter(Dependencies{
		EventService: events.NewService(repo),
	})
	req := httptest.NewRequest(nethttp.MethodGet, "/api/events/abc123def4", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status 200 without Authorization header, got %d; body: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data  events.EventDetail `json:"data"`
		Error any                `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v; body: %s", err, rec.Body.String())
	}
	if body.Data.Event.ID != repo.event.ID || body.Data.GoingCount != 2 {
		t.Fatalf("unexpected detail response: %+v", body.Data)
	}
}

type fakeEventRepository struct {
	event      events.Event
	goingCount int
}

func (r *fakeEventRepository) Create(ctx context.Context, event events.Event) (events.Event, error) {
	return event, nil
}

func (r *fakeEventRepository) FindBySlug(ctx context.Context, slug string) (events.Event, error) {
	if slug != r.event.ShareSlug {
		return events.Event{}, gorm.ErrRecordNotFound
	}
	return r.event, nil
}

func (r *fakeEventRepository) FindByID(ctx context.Context, id int64) (events.Event, error) {
	if id != r.event.ID {
		return events.Event{}, gorm.ErrRecordNotFound
	}
	return r.event, nil
}

func (r *fakeEventRepository) CountGoing(ctx context.Context, eventID int64) (int, error) {
	return r.goingCount, nil
}

func (r *fakeEventRepository) ListMine(ctx context.Context, ownerID uuid.UUID) ([]events.Event, error) {
	return []events.Event{r.event}, nil
}
