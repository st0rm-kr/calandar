package schedules

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]Schedule, error) {
	var rows []Schedule
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Where("start_at < ? AND ? < COALESCE(end_at, start_at + interval '1 minute')", to, from).
		Order("start_at ASC").
		Find(&rows).
		Error
	return rows, err
}

func (r *Repository) FindByID(ctx context.Context, userID uuid.UUID, id int64) (Schedule, error) {
	var schedule Schedule
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		First(&schedule).
		Error
	return schedule, err
}

func (r *Repository) CreateManual(ctx context.Context, schedule Schedule) (Schedule, error) {
	schedule.Source = SourceManual
	schedule.EventID = nil
	if err := r.db.WithContext(ctx).Create(&schedule).Error; err != nil {
		return Schedule{}, err
	}
	return schedule, nil
}

func (r *Repository) UpdateManual(ctx context.Context, userID uuid.UUID, id int64, patch SchedulePatch) (Schedule, error) {
	result := r.db.WithContext(ctx).
		Model(&Schedule{}).
		Where("id = ? AND user_id = ? AND source = ? AND deleted_at IS NULL", id, userID, SourceManual).
		Updates(map[string]any{
			"title":      patch.Title,
			"start_at":   patch.StartAt,
			"end_at":     patch.EndAt,
			"location":   patch.Location,
			"visibility": patch.Visibility,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return Schedule{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Schedule{}, ErrScheduleNotFound
	}
	return r.FindByID(ctx, userID, id)
}

func (r *Repository) DeleteManual(ctx context.Context, userID uuid.UUID, id int64) error {
	result := r.db.WithContext(ctx).
		Model(&Schedule{}).
		Where("id = ? AND user_id = ? AND source = ? AND deleted_at IS NULL", id, userID, SourceManual).
		Update("deleted_at", time.Now().UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrScheduleNotFound
	}
	return nil
}

func (r *Repository) FindConflictsExcluding(ctx context.Context, userID uuid.UUID, start time.Time, end *time.Time, excludeID *int64) ([]Conflict, error) {
	conflictEnd := end
	if conflictEnd == nil {
		defaultEnd := start.Add(time.Minute)
		conflictEnd = &defaultEnd
	}

	query := r.db.WithContext(ctx).
		Model(&Schedule{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Where("start_at < ? AND ? < COALESCE(end_at, start_at + interval '1 minute')", *conflictEnd, start)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var rows []Schedule
	if err := query.Order("start_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	conflicts := make([]Conflict, 0, len(rows))
	for _, row := range rows {
		conflicts = append(conflicts, Conflict{
			ID:         row.ID,
			Title:      row.Title,
			StartAt:    row.StartAt,
			EndAt:      row.EndAt,
			Location:   row.Location,
			Visibility: row.Visibility,
			Source:     row.Source,
			EventID:    row.EventID,
		})
	}
	return conflicts, nil
}
