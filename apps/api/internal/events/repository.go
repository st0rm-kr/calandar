package events

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
