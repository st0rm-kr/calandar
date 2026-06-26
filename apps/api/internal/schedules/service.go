package schedules

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrInvalidScheduleTitle      = errors.New("title must be 1-80 characters")
	ErrInvalidScheduleTime       = errors.New("end_at must be after start_at")
	ErrInvalidVisibility         = errors.New("visibility must be public, busy_only, or private")
	ErrScheduleNotFound          = errors.New("schedule not found")
	ErrCannotModifyEventSchedule = errors.New("event schedule cannot be modified directly")
)

type ManualRepository interface {
	ListByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]Schedule, error)
	FindByID(ctx context.Context, userID uuid.UUID, id int64) (Schedule, error)
	CreateManual(ctx context.Context, schedule Schedule) (Schedule, error)
	UpdateManual(ctx context.Context, userID uuid.UUID, id int64, patch SchedulePatch) (Schedule, error)
	DeleteManual(ctx context.Context, userID uuid.UUID, id int64) error
	FindConflictsExcluding(ctx context.Context, userID uuid.UUID, start time.Time, end *time.Time, excludeID *int64) ([]Conflict, error)
}

type CreateManualInput struct {
	Title      string
	StartAt    time.Time
	EndAt      *time.Time
	Location   *string
	Visibility string
}

type UpdateManualInput struct {
	Title      string
	StartAt    time.Time
	EndAt      *time.Time
	Location   *string
	Visibility string
}

type SchedulePatch struct {
	Title      string
	StartAt    time.Time
	EndAt      *time.Time
	Location   *string
	Visibility string
}

type ScheduleResult struct {
	Schedule  Schedule   `json:"schedule"`
	Conflicts []Conflict `json:"conflicts"`
}

type Service struct {
	repo ManualRepository
}

func NewService(repo ManualRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]Schedule, error) {
	return s.repo.ListByRange(ctx, userID, from.UTC(), to.UTC())
}

func (s *Service) GetManual(ctx context.Context, userID uuid.UUID, scheduleID int64) (Schedule, error) {
	schedule, err := s.repo.FindByID(ctx, userID, scheduleID)
	if err != nil {
		return Schedule{}, err
	}
	if schedule.Source != SourceManual {
		return Schedule{}, ErrCannotModifyEventSchedule
	}
	return schedule, nil
}

func (s *Service) CreateManual(ctx context.Context, userID uuid.UUID, input CreateManualInput) (ScheduleResult, error) {
	title, err := normalizeTitle(input.Title)
	if err != nil {
		return ScheduleResult{}, err
	}
	if err := validateTimeRange(input.StartAt, input.EndAt); err != nil {
		return ScheduleResult{}, err
	}
	visibility, err := normalizeVisibility(input.Visibility)
	if err != nil {
		return ScheduleResult{}, err
	}

	schedule := Schedule{
		UserID:     userID,
		Title:      title,
		StartAt:    input.StartAt.UTC(),
		EndAt:      utcPtr(input.EndAt),
		Location:   normalizeLocation(input.Location),
		Visibility: visibility,
		Source:     SourceManual,
	}
	created, err := s.repo.CreateManual(ctx, schedule)
	if err != nil {
		return ScheduleResult{}, err
	}

	conflicts, err := s.repo.FindConflictsExcluding(ctx, userID, created.StartAt, created.EndAt, &created.ID)
	if err != nil {
		return ScheduleResult{}, err
	}
	return ScheduleResult{Schedule: created, Conflicts: conflicts}, nil
}

func (s *Service) UpdateManual(ctx context.Context, userID uuid.UUID, scheduleID int64, input UpdateManualInput) (ScheduleResult, error) {
	title, err := normalizeTitle(input.Title)
	if err != nil {
		return ScheduleResult{}, err
	}
	if err := validateTimeRange(input.StartAt, input.EndAt); err != nil {
		return ScheduleResult{}, err
	}
	visibility, err := normalizeVisibility(input.Visibility)
	if err != nil {
		return ScheduleResult{}, err
	}

	existing, err := s.repo.FindByID(ctx, userID, scheduleID)
	if err != nil {
		return ScheduleResult{}, err
	}
	if existing.Source != SourceManual {
		return ScheduleResult{}, ErrCannotModifyEventSchedule
	}

	updated, err := s.repo.UpdateManual(ctx, userID, scheduleID, SchedulePatch{
		Title:      title,
		StartAt:    input.StartAt.UTC(),
		EndAt:      utcPtr(input.EndAt),
		Location:   normalizeLocation(input.Location),
		Visibility: visibility,
	})
	if err != nil {
		return ScheduleResult{}, err
	}

	conflicts, err := s.repo.FindConflictsExcluding(ctx, userID, updated.StartAt, updated.EndAt, &updated.ID)
	if err != nil {
		return ScheduleResult{}, err
	}
	return ScheduleResult{Schedule: updated, Conflicts: conflicts}, nil
}

func (s *Service) DeleteManual(ctx context.Context, userID uuid.UUID, scheduleID int64) error {
	existing, err := s.repo.FindByID(ctx, userID, scheduleID)
	if err != nil {
		return err
	}
	if existing.Source != SourceManual {
		return ErrCannotModifyEventSchedule
	}
	return s.repo.DeleteManual(ctx, userID, scheduleID)
}

func (s *Service) FindConflicts(ctx context.Context, userID uuid.UUID, start time.Time, end *time.Time, excludeID *int64) ([]Conflict, error) {
	return s.repo.FindConflictsExcluding(ctx, userID, start.UTC(), utcPtr(end), excludeID)
}

func normalizeTitle(raw string) (string, error) {
	title := strings.TrimSpace(raw)
	if title == "" || utf8.RuneCountInString(title) > 80 {
		return "", ErrInvalidScheduleTitle
	}
	return title, nil
}

func validateTimeRange(start time.Time, end *time.Time) error {
	if end != nil && !end.After(start) {
		return ErrInvalidScheduleTime
	}
	return nil
}

func normalizeVisibility(raw string) (string, error) {
	if raw == "" {
		return VisibilityBusyOnly, nil
	}
	switch raw {
	case VisibilityPublic, VisibilityBusyOnly, VisibilityPrivate:
		return raw, nil
	default:
		return "", ErrInvalidVisibility
	}
}

func normalizeLocation(location *string) *string {
	if location == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*location)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func utcPtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func effectiveEnd(start time.Time, end *time.Time) time.Time {
	if end != nil {
		return *end
	}
	return start.Add(time.Minute)
}
