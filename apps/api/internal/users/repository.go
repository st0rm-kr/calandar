package users

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type UpdateMeParams struct {
	DisplayName string
	AvatarURL   *string
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (Profile, error) {
	var profile Profile
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&profile).
		Error
	return profile, err
}

func (r *Repository) UpdateMe(ctx context.Context, id uuid.UUID, params UpdateMeParams) (Profile, error) {
	updates := map[string]any{
		"display_name": params.DisplayName,
		"avatar_url":   params.AvatarURL,
		"updated_at":   time.Now(),
	}

	if err := r.db.WithContext(ctx).
		Model(&Profile{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates).
		Error; err != nil {
		return Profile{}, err
	}

	return r.FindByID(ctx, id)
}

func (r *Repository) Search(ctx context.Context, query string) ([]Profile, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	var profiles []Profile
	err := r.db.WithContext(ctx).
		Where("status = ? AND deleted_at IS NULL", "active").
		Where("lower(display_name) LIKE ? OR lower(email) LIKE ?", pattern, pattern).
		Order("display_name ASC").
		Find(&profiles).
		Error
	return profiles, err
}
