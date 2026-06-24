package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/config"
	"github.com/bytedance/calandar/apps/api/internal/events"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const testJWTSecret = "test-secret"

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

func TestRSVPCapacityFullMapsToConflict(t *testing.T) {
	capacity := 1
	repo := &fakeEventRepository{}
	repo.event = events.Event{
		ID:        12,
		OwnerID:   uuid.MustParse("d3f7e3f1-b2a9-46f0-9a4b-41a198e624c8"),
		Scope:     "personal",
		Title:     "Climb",
		Type:      "climbing",
		StartAt:   time.Now().Add(time.Hour).UTC(),
		Capacity:  &capacity,
		Status:    events.StatusActive,
		ShareSlug: "abc123def4",
	}
	repo.goingCount = 1
	router := NewRouter(Dependencies{
		Config:       configWithJWTSecret(),
		EventService: events.NewService(repo),
	})
	userID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	req := httptest.NewRequest(nethttp.MethodPost, "/api/events/12/rsvp", strings.NewReader(`{"rsvp":"going","add_to_calendar":true}`))
	req.Header.Set("Authorization", "Bearer "+signHS256JWT(t, userID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusConflict {
		t.Fatalf("expected status 409, got %d; body: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v; body: %s", err, rec.Body.String())
	}
	if body.Error.Code != "capacity_full" {
		t.Fatalf("expected capacity_full error, got %q", body.Error.Code)
	}
}

func configWithJWTSecret() config.Config {
	return config.Config{SupabaseJWTSecret: testJWTSecret}
}

func signHS256JWT(t *testing.T, sub uuid.UUID) string {
	t.Helper()

	headerJSON, err := json.Marshal(map[string]any{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	claimsJSON, err := json.Marshal(map[string]any{
		"sub": sub.String(),
		"aud": "authenticated",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	mac := hmac.New(sha256.New, []byte(testJWTSecret))
	if _, err := mac.Write([]byte(unsigned)); err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
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

func (r *fakeEventRepository) FindByIDForUpdate(ctx context.Context, id int64) (events.Event, error) {
	return r.FindByID(ctx, id)
}

func (r *fakeEventRepository) CountGoing(ctx context.Context, eventID int64) (int, error) {
	return r.goingCount, nil
}

func (r *fakeEventRepository) ListMine(ctx context.Context, ownerID uuid.UUID) ([]events.Event, error) {
	return []events.Event{r.event}, nil
}

func (r *fakeEventRepository) Transaction(ctx context.Context, fn func(events.EventRepository) error) error {
	return fn(r)
}

func (r *fakeEventRepository) CountGoingForUpdateExcludingUser(ctx context.Context, eventID int64, userID uuid.UUID) (int, error) {
	return r.goingCount, nil
}

func (r *fakeEventRepository) UpsertParticipant(ctx context.Context, participant events.Participant) (events.Participant, error) {
	return participant, nil
}

func (r *fakeEventRepository) FindParticipant(ctx context.Context, eventID int64, userID uuid.UUID) (events.Participant, bool, error) {
	return events.Participant{}, false, nil
}

func (r *fakeEventRepository) ListParticipants(ctx context.Context, eventID int64) ([]events.Participant, error) {
	return nil, nil
}

func (r *fakeEventRepository) UpsertEventSchedule(ctx context.Context, userID uuid.UUID, event events.Event, visibility string) error {
	return nil
}

func (r *fakeEventRepository) DeleteEventSchedule(ctx context.Context, userID uuid.UUID, eventID int64) error {
	return nil
}

func (r *fakeEventRepository) DeleteEventSchedulesForEvent(ctx context.Context, eventID int64) error {
	return nil
}

func (r *fakeEventRepository) FindConflicts(ctx context.Context, userID uuid.UUID, start time.Time, end *time.Time) ([]events.ScheduleConflict, error) {
	return nil, nil
}

func (r *fakeEventRepository) Cancel(ctx context.Context, userID uuid.UUID, eventID int64) (events.Event, error) {
	return r.event, nil
}
