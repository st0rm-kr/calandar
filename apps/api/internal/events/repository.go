package events

import (
	"context"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/schedules"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, event Event) (Event, error) {
	if err := r.db.WithContext(ctx).Create(&event).Error; err != nil {
		return Event{}, err
	}
	return event, nil
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (Event, error) {
	var event Event
	err := r.db.WithContext(ctx).
		Where("share_slug = ? AND deleted_at IS NULL", slug).
		First(&event).
		Error
	return event, err
}

func (r *Repository) FindByID(ctx context.Context, id int64) (Event, error) {
	var event Event
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&event).
		Error
	return event, err
}

func (r *Repository) CountGoing(ctx context.Context, eventID int64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Participant{}).
		Where("event_id = ? AND rsvp = ? AND deleted_at IS NULL", eventID, "going").
		Count(&count).
		Error
	return int(count), err
}

func (r *Repository) ListMine(ctx context.Context, ownerID uuid.UUID) ([]Event, error) {
	var events []Event
	err := r.db.WithContext(ctx).
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Order("start_at DESC").
		Find(&events).
		Error
	return events, err
}

func (r *Repository) Transaction(ctx context.Context, fn func(EventRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Repository{db: tx})
	})
}

func (r *Repository) CountGoingForUpdateExcludingUser(ctx context.Context, eventID int64, userID uuid.UUID) (int, error) {
	var participants []Participant
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("event_id = ? AND user_id <> ? AND rsvp = ? AND deleted_at IS NULL", eventID, userID, RSVPGoing).
		Find(&participants).
		Error
	return len(participants), err
}

func (r *Repository) UpsertParticipant(ctx context.Context, participant Participant) (Participant, error) {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "event_id"}, {Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"rsvp":            participant.RSVP,
				"source":          participant.Source,
				"add_to_calendar": participant.AddToCalendar,
				"deleted_at":      nil,
				"updated_at":      time.Now().UTC(),
			}),
		}).
		Create(&participant).
		Error
	if err != nil {
		return Participant{}, err
	}

	var saved Participant
	err = r.db.WithContext(ctx).
		Where("event_id = ? AND user_id = ? AND deleted_at IS NULL", participant.EventID, participant.UserID).
		First(&saved).
		Error
	return saved, err
}

func (r *Repository) UpsertEventSchedule(ctx context.Context, userID uuid.UUID, event Event, visibility string) error {
	return schedules.UpsertEventSchedule(ctx, r.db, userID, schedules.EventScheduleInput{
		ID:       event.ID,
		Title:    event.Title,
		StartAt:  event.StartAt,
		EndAt:    event.EndAt,
		Location: event.Location,
	}, visibility)
}

func (r *Repository) DeleteEventSchedule(ctx context.Context, userID uuid.UUID, eventID int64) error {
	return schedules.DeleteEventSchedule(ctx, r.db, userID, eventID)
}

func (r *Repository) FindConflicts(ctx context.Context, userID uuid.UUID, start time.Time, end *time.Time) ([]ScheduleConflict, error) {
	conflicts, err := schedules.FindConflicts(ctx, r.db, userID, start, end)
	if err != nil {
		return nil, err
	}
	result := make([]ScheduleConflict, 0, len(conflicts))
	for _, conflict := range conflicts {
		result = append(result, ScheduleConflict{
			ID:         conflict.ID,
			Title:      conflict.Title,
			StartAt:    conflict.StartAt,
			EndAt:      conflict.EndAt,
			Location:   conflict.Location,
			Visibility: conflict.Visibility,
			Source:     conflict.Source,
			EventID:    conflict.EventID,
		})
	}
	return result, nil
}

func (r *Repository) Cancel(ctx context.Context, userID uuid.UUID, eventID int64) (Event, error) {
	result := r.db.WithContext(ctx).
		Model(&Event{}).
		Where("id = ? AND owner_id = ? AND deleted_at IS NULL", eventID, userID).
		Updates(map[string]any{
			"status":     StatusCancelled,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return Event{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Event{}, gorm.ErrRecordNotFound
	}
	return r.FindByID(ctx, eventID)
}
