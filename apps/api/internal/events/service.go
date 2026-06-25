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
	ErrInvalidRSVP       = errors.New("rsvp must be invited, going, not_going, or maybe")
	ErrEventUnavailable  = errors.New("event is not available for rsvp")
	ErrCapacityFull      = errors.New("event capacity is full")
	ErrInviteForbidden   = errors.New("not authorized to invite this user to the event")
)

type EventRepository interface {
	Create(ctx context.Context, event Event) (Event, error)
	FindBySlug(ctx context.Context, slug string) (Event, error)
	FindByID(ctx context.Context, id int64) (Event, error)
	FindByIDForUpdate(ctx context.Context, id int64) (Event, error)
	CountGoing(ctx context.Context, eventID int64) (int, error)
	ListMine(ctx context.Context, ownerID uuid.UUID) ([]Event, error)
	Transaction(ctx context.Context, fn func(EventRepository) error) error
	CountGoingForUpdateExcludingUser(ctx context.Context, eventID int64, userID uuid.UUID) (int, error)
	UpsertParticipant(ctx context.Context, participant Participant) (Participant, error)
	FindParticipant(ctx context.Context, eventID int64, userID uuid.UUID) (Participant, bool, error)
	ListParticipants(ctx context.Context, eventID int64) ([]Participant, error)
	UpsertEventSchedule(ctx context.Context, userID uuid.UUID, event Event, visibility string) error
	DeleteEventSchedule(ctx context.Context, userID uuid.UUID, eventID int64) error
	DeleteEventSchedulesForEvent(ctx context.Context, eventID int64) error
	FindConflicts(ctx context.Context, userID uuid.UUID, start time.Time, end *time.Time) ([]ScheduleConflict, error)
	Cancel(ctx context.Context, userID uuid.UUID, eventID int64) (Event, error)
}

type Service struct {
	repo             EventRepository
	now              func() time.Time
	friendAuthorizer FriendAuthorizer
	groupAuthorizer  GroupAuthorizer
	notifier         Notifier
}

type FriendAuthorizer interface {
	AreFriends(ctx context.Context, a, b uuid.UUID) (bool, error)
}

type GroupAuthorizer interface {
	IsMember(ctx context.Context, groupID int64, userID uuid.UUID) (bool, error)
}

type Notifier interface {
	Notify(ctx context.Context, userID uuid.UUID, typ string, payload map[string]any) error
}

type noopNotifier struct{}

func (noopNotifier) Notify(context.Context, uuid.UUID, string, map[string]any) error {
	return nil
}

type Option func(*Service)

func WithFriendAuthorizer(authorizer FriendAuthorizer) Option {
	return func(s *Service) {
		s.friendAuthorizer = authorizer
	}
}

func WithGroupAuthorizer(authorizer GroupAuthorizer) Option {
	return func(s *Service) {
		s.groupAuthorizer = authorizer
	}
}

func WithNotifier(notifier Notifier) Option {
	return func(s *Service) {
		s.notifier = notifier
	}
}

func NewService(repo EventRepository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now, notifier: noopNotifier{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
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

func (s *Service) RSVP(ctx context.Context, userID uuid.UUID, eventID int64, input RSVPInput) (RSVPResult, error) {
	if !validRSVP(input.RSVP) {
		return RSVPResult{}, ErrInvalidRSVP
	}
	if input.Visibility == "" {
		input.Visibility = "busy_only"
	}
	if !validVisibility(input.Visibility) {
		return RSVPResult{}, ErrInvalidRSVP
	}

	var result RSVPResult
	err := s.repo.Transaction(ctx, func(repo EventRepository) error {
		event, err := repo.FindByIDForUpdate(ctx, eventID)
		if err != nil {
			return err
		}
		event = deriveExpired(event, s.now())
		if event.Status != StatusActive {
			return ErrEventUnavailable
		}

		if input.RSVP == RSVPGoing && event.Capacity != nil {
			goingCount, err := repo.CountGoingForUpdateExcludingUser(ctx, event.ID, userID)
			if err != nil {
				return err
			}
			if goingCount >= *event.Capacity {
				return ErrCapacityFull
			}
		}

		participant, err := repo.UpsertParticipant(ctx, Participant{
			EventID:       event.ID,
			UserID:        userID,
			RSVP:          input.RSVP,
			Source:        ParticipantSourceSelf,
			AddToCalendar: input.AddToCalendar,
		})
		if err != nil {
			return err
		}

		if shouldLinkEventSchedule(input) {
			if err := repo.UpsertEventSchedule(ctx, userID, event, input.Visibility); err != nil {
				return err
			}
		} else if err := repo.DeleteEventSchedule(ctx, userID, event.ID); err != nil {
			return err
		}

		conflicts, err := repo.FindConflicts(ctx, userID, event.StartAt, event.EndAt)
		if err != nil {
			return err
		}
		conflicts = filterEventSelfConflict(conflicts, event.ID)
		goingCount, err := repo.CountGoing(ctx, event.ID)
		if err != nil {
			return err
		}
		result = RSVPResult{Participant: participant, Conflicts: conflicts, GoingCount: goingCount}
		return nil
	})
	if err != nil {
		return RSVPResult{}, err
	}
	return result, nil
}

func (s *Service) Cancel(ctx context.Context, userID uuid.UUID, eventID int64) (Event, error) {
	var event Event
	var participants []Participant
	err := s.repo.Transaction(ctx, func(repo EventRepository) error {
		cancelled, err := repo.Cancel(ctx, userID, eventID)
		if err != nil {
			return err
		}
		participants, err = repo.ListParticipants(ctx, eventID)
		if err != nil {
			return err
		}
		if err := repo.DeleteEventSchedulesForEvent(ctx, eventID); err != nil {
			return err
		}
		event = cancelled
		return nil
	})
	if err != nil {
		return Event{}, err
	}
	for _, participant := range participants {
		if participant.UserID == userID {
			continue
		}
		_ = s.notifier.Notify(ctx, participant.UserID, "event_cancelled", map[string]any{
			"event_id": event.ID,
			"title":    event.Title,
		})
	}
	return event, nil
}

func (s *Service) Invite(ctx context.Context, actorID uuid.UUID, eventID int64, inviteeIDs []uuid.UUID) ([]Participant, error) {
	event, err := s.repo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	event = deriveExpired(event, s.now())
	if event.Status != StatusActive {
		return nil, ErrEventUnavailable
	}

	invited := make([]Participant, 0, len(inviteeIDs))
	for _, inviteeID := range inviteeIDs {
		if inviteeID == actorID {
			continue
		}
		allowed, err := s.canInvite(ctx, actorID, inviteeID, event)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, ErrInviteForbidden
		}

		if _, exists, err := s.repo.FindParticipant(ctx, event.ID, inviteeID); err != nil {
			return nil, err
		} else if exists {
			continue
		}

		participant, err := s.repo.UpsertParticipant(ctx, Participant{
			EventID:       event.ID,
			UserID:        inviteeID,
			RSVP:          RSVPInvited,
			Source:        ParticipantSourceInvited,
			AddToCalendar: false,
		})
		if err != nil {
			return nil, err
		}
		_ = s.notifier.Notify(ctx, inviteeID, "event_invite", map[string]any{
			"event_id": event.ID,
			"title":    event.Title,
		})
		invited = append(invited, participant)
	}
	return invited, nil
}

func (s *Service) canInvite(ctx context.Context, actorID, inviteeID uuid.UUID, event Event) (bool, error) {
	if event.Scope == "group" && event.GroupID != nil {
		if s.groupAuthorizer == nil {
			return false, nil
		}
		actorMember, err := s.groupAuthorizer.IsMember(ctx, *event.GroupID, actorID)
		if err != nil {
			return false, err
		}
		if !actorMember {
			return false, nil
		}
		return s.groupAuthorizer.IsMember(ctx, *event.GroupID, inviteeID)
	}

	if event.OwnerID != actorID {
		return false, nil
	}
	if s.friendAuthorizer == nil {
		return false, nil
	}
	return s.friendAuthorizer.AreFriends(ctx, actorID, inviteeID)
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

func shouldLinkEventSchedule(input RSVPInput) bool {
	return input.AddToCalendar && (input.RSVP == RSVPGoing || input.RSVP == RSVPMaybe)
}

func filterEventSelfConflict(conflicts []ScheduleConflict, eventID int64) []ScheduleConflict {
	filtered := conflicts[:0]
	for _, conflict := range conflicts {
		if conflict.EventID != nil && *conflict.EventID == eventID {
			continue
		}
		filtered = append(filtered, conflict)
	}
	return filtered
}

func validRSVP(rsvp string) bool {
	switch rsvp {
	case RSVPGoing, RSVPNotGoing, RSVPMaybe:
		return true
	default:
		return false
	}
}

func validVisibility(visibility string) bool {
	switch visibility {
	case "public", "busy_only", "private":
		return true
	default:
		return false
	}
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
