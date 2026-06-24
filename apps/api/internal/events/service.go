package events

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrInvalidEventTitle = errors.New("title must be 1-80 characters")
	ErrInvalidEventType  = errors.New("type must be a supported event type")
	ErrInvalidEventScope = errors.New("scope must be personal or group")
	ErrInvalidEventTime  = errors.New("end_at must be after start_at")
	ErrInvalidCapacity   = errors.New("capacity must be greater than 0")
	ErrShareSlugConflict = errors.New("share slug conflict")
)

type EventRepository interface {
	Create(ctx context.Context, event Event) (Event, error)
	FindBySlug(ctx context.Context, slug string) (Event, error)
	FindByID(ctx context.Context, id int64) (Event, error)
	CountGoing(ctx context.Context, eventID int64) (int, error)
	ListMine(ctx context.Context, ownerID uuid.UUID) ([]Event, error)
}

type Service struct {
	repo EventRepository
	now  func() time.Time
}

func NewService(repo EventRepository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) Create(ctx context.Context, ownerID uuid.UUID, input CreateEventInput) (Event, error) {
	event, err := buildEvent(ownerID, input)
	if err != nil {
		return Event{}, err
	}

	for attempts := 0; attempts < 5; attempts++ {
		event.ShareSlug, err = generateShareSlug()
		if err != nil {
			return Event{}, err
		}
		created, err := s.repo.Create(ctx, event)
		if err == nil {
			return created, nil
		}
		if !isSlugConflict(err) {
			return Event{}, err
		}
	}
	return Event{}, ErrShareSlugConflict
}

func (s *Service) GetDetail(ctx context.Context, slug string) (EventDetail, error) {
	event, err := s.repo.FindBySlug(ctx, strings.TrimSpace(slug))
	if err != nil {
		return EventDetail{}, err
	}
	count, err := s.repo.CountGoing(ctx, event.ID)
	if err != nil {
		return EventDetail{}, err
	}
	event = deriveExpired(event, s.now())
	return EventDetail{Event: event, GoingCount: count}, nil
}

func (s *Service) ListMine(ctx context.Context, ownerID uuid.UUID) ([]Event, error) {
	events, err := s.repo.ListMine(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	for i := range events {
		events[i] = deriveExpired(events[i], now)
	}
	return events, nil
}

func buildEvent(ownerID uuid.UUID, input CreateEventInput) (Event, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" || utf8.RuneCountInString(title) > 80 {
		return Event{}, ErrInvalidEventTitle
	}
	if !validEventType(input.Type) {
		return Event{}, ErrInvalidEventType
	}
	if input.Scope == "" {
		input.Scope = "personal"
	}
	if input.Scope != "personal" && input.Scope != "group" {
		return Event{}, ErrInvalidEventScope
	}
	if input.EndAt != nil && !input.EndAt.After(input.StartAt) {
		return Event{}, ErrInvalidEventTime
	}
	if input.Capacity != nil && *input.Capacity <= 0 {
		return Event{}, ErrInvalidCapacity
	}
	if input.Location != nil {
		location := strings.TrimSpace(*input.Location)
		if location == "" {
			input.Location = nil
		} else {
			input.Location = &location
		}
	}

	return Event{
		OwnerID:  ownerID,
		Scope:    input.Scope,
		GroupID:  input.GroupID,
		Title:    title,
		Type:     input.Type,
		StartAt:  input.StartAt,
		EndAt:    input.EndAt,
		Location: input.Location,
		Capacity: input.Capacity,
		Status:   StatusActive,
	}, nil
}

func validEventType(eventType string) bool {
	switch eventType {
	case "climbing", "dining", "travel", "gaming", "other":
		return true
	default:
		return false
	}
}

func generateShareSlug() (string, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:])[:10], nil
}

func deriveExpired(event Event, now time.Time) Event {
	if event.Status != StatusActive {
		return event
	}
	expiresAt := event.StartAt
	if event.EndAt != nil {
		expiresAt = *event.EndAt
	}
	if expiresAt.Before(now) {
		event.Status = StatusExpired
	}
	return event
}

func isSlugConflict(err error) bool {
	if errors.Is(err, ErrShareSlugConflict) || errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
