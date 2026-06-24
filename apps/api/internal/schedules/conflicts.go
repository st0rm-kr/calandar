package schedules

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func FindConflicts(ctx context.Context, db *gorm.DB, userID uuid.UUID, start time.Time, end *time.Time) ([]Conflict, error) {
	conflictEnd := end
	if conflictEnd == nil {
		defaultEnd := start.Add(time.Minute)
		conflictEnd = &defaultEnd
	}

	var rows []Schedule
	err := db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Where("start_at < ? AND ? < COALESCE(end_at, start_at + interval '1 minute')", *conflictEnd, start).
		Order("start_at ASC").
		Find(&rows).
		Error
	if err != nil {
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
