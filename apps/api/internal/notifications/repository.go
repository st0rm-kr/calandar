package notifications

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, notification Notification) error {
	return r.db.WithContext(ctx).Create(&notification).Error
}

func (r *Repo) List(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	var rows []Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&rows).
		Error
	return rows, err
}

func (r *Repo) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND read_at IS NULL AND deleted_at IS NULL", userID).
		Count(&count).
		Error
	return int(count), err
}

func (r *Repo) MarkRead(ctx context.Context, userID uuid.UUID, id int64) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL AND deleted_at IS NULL", id, userID).
		Updates(map[string]any{"read_at": time.Now().UTC(), "updated_at": time.Now().UTC()})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND read_at IS NULL AND deleted_at IS NULL", userID).
		Updates(map[string]any{"read_at": time.Now().UTC(), "updated_at": time.Now().UTC()}).
		Error
}
