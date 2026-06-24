package schedules

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UpsertEventSchedule(ctx context.Context, db *gorm.DB, userID uuid.UUID, event EventScheduleInput, visibility string) error {
	schedule := Schedule{
		UserID:     userID,
		Title:      event.Title,
		StartAt:    event.StartAt,
		EndAt:      event.EndAt,
		Location:   event.Location,
		Visibility: visibility,
		Source:     SourceEvent,
		EventID:    &event.ID,
	}

	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "event_id"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: "source", Value: SourceEvent},
				clause.Eq{Column: "deleted_at", Value: nil},
			}},
			DoUpdates: clause.Assignments(map[string]any{
				"title":      event.Title,
				"start_at":   event.StartAt,
				"end_at":     event.EndAt,
				"location":   event.Location,
				"deleted_at": nil,
				"updated_at": time.Now().UTC(),
			}),
		}).
		Create(&schedule).
		Error
}

func DeleteEventSchedule(ctx context.Context, db *gorm.DB, userID uuid.UUID, eventID int64) error {
	return db.WithContext(ctx).
		Model(&Schedule{}).
		Where("user_id = ? AND event_id = ? AND source = ? AND deleted_at IS NULL", userID, eventID, SourceEvent).
		Update("deleted_at", time.Now().UTC()).
		Error
}

func DeleteEventSchedulesForEvent(ctx context.Context, db *gorm.DB, eventID int64) error {
	return db.WithContext(ctx).
		Model(&Schedule{}).
		Where("event_id = ? AND source = ? AND deleted_at IS NULL", eventID, SourceEvent).
		Update("deleted_at", time.Now().UTC()).
		Error
}
