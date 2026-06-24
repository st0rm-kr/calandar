package friends

import (
	"context"
	"errors"
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

func (r *Repo) FindBetween(ctx context.Context, a, b uuid.UUID) (Friendship, bool, error) {
	var friendship Friendship
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where(
			"(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
			a, b, b, a,
		).
		First(&friendship).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Friendship{}, false, nil
	}
	if err != nil {
		return Friendship{}, false, err
	}
	return friendship, true, nil
}

func (r *Repo) Create(ctx context.Context, friendship Friendship) (Friendship, error) {
	if err := r.db.WithContext(ctx).Create(&friendship).Error; err != nil {
		return Friendship{}, err
	}
	return friendship, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, id int64, status string) (Friendship, error) {
	result := r.db.WithContext(ctx).
		Model(&Friendship{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now().UTC()})
	if result.Error != nil {
		return Friendship{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Friendship{}, ErrRequestNotFound
	}
	return r.FindByID(ctx, id)
}

func (r *Repo) FindByID(ctx context.Context, id int64) (Friendship, error) {
	var friendship Friendship
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&friendship).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Friendship{}, ErrRequestNotFound
	}
	return friendship, err
}

func (r *Repo) ListAccepted(ctx context.Context, userID uuid.UUID) ([]Friend, error) {
	var friends []Friend
	err := r.db.WithContext(ctx).
		Table("friendships AS f").
		Select(`u.id AS user_id, u.display_name, u.avatar_url, u.email`).
		Joins(`JOIN users u ON u.id = CASE WHEN f.requester_id = ? THEN f.addressee_id ELSE f.requester_id END`, userID).
		Where("f.status = ? AND f.deleted_at IS NULL", StatusAccepted).
		Where("f.requester_id = ? OR f.addressee_id = ?", userID, userID).
		Scan(&friends).
		Error
	return friends, err
}

func (r *Repo) ListRequests(ctx context.Context, userID uuid.UUID) ([]Friendship, error) {
	var rows []Friendship
	err := r.db.WithContext(ctx).
		Where("status = ? AND deleted_at IS NULL", StatusPending).
		Where("requester_id = ? OR addressee_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&rows).
		Error
	return rows, err
}

func (r *Repo) DeleteBetween(ctx context.Context, a, b uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&Friendship{}).
		Where("deleted_at IS NULL").
		Where(
			"(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
			a, b, b, a,
		).
		Update("deleted_at", time.Now().UTC()).
		Error
}
