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

func (r *Repository) ListParticipantEventsByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]Schedule, error) {
	var rows []Schedule
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT
				ep.id AS id,
				ep.user_id AS user_id,
				e.title AS title,
				e.start_at AS start_at,
				e.end_at AS end_at,
				e.location AS location,
				? AS visibility,
				? AS source,
				e.id AS event_id
			FROM event_participants ep
			JOIN events e ON e.id = ep.event_id
			WHERE ep.user_id = ?
				AND ep.deleted_at IS NULL
				AND ep.rsvp <> ?
				AND e.deleted_at IS NULL
				AND e.status = ?
				AND e.start_at < ?
				AND ? < COALESCE(e.end_at, e.start_at + interval '1 minute')
			ORDER BY e.start_at ASC
		`, VisibilityPublic, SourceEvent, userID, "going", "active", to, from).
		Scan(&rows).
		Error
	return rows, err
}

func (r *Repository) ListEventParticipationSummaries(ctx context.Context, userID uuid.UUID, eventIDs []int64) (map[int64]EventParticipationSummary, error) {
	summaries := make(map[int64]EventParticipationSummary, len(eventIDs))
	if len(eventIDs) == 0 {
		return summaries, nil
	}
	for _, eventID := range eventIDs {
		summaries[eventID] = EventParticipationSummary{EventID: eventID}
	}

	var countRows []struct {
		EventID int64
		Count   int
	}
	if err := r.db.WithContext(ctx).
		Table("event_participants").
		Select("event_id, COUNT(*) AS count").
		Where("event_id IN ? AND rsvp = ? AND deleted_at IS NULL", eventIDs, "going").
		Group("event_id").
		Scan(&countRows).
		Error; err != nil {
		return nil, err
	}
	for _, row := range countRows {
		summary := summaries[row.EventID]
		summary.GoingCount = row.Count
		summaries[row.EventID] = summary
	}

	var viewerRows []struct {
		EventID int64
		RSVP    string
	}
	if err := r.db.WithContext(ctx).
		Table("event_participants").
		Select("event_id, rsvp").
		Where("event_id IN ? AND user_id = ? AND deleted_at IS NULL", eventIDs, userID).
		Scan(&viewerRows).
		Error; err != nil {
		return nil, err
	}
	for _, row := range viewerRows {
		summary := summaries[row.EventID]
		rsvp := row.RSVP
		summary.ViewerRSVP = &rsvp
		summaries[row.EventID] = summary
	}

	var previewRows []struct {
		EventID     int64
		UserID      uuid.UUID
		DisplayName string
		AvatarURL   *string
	}
	if err := r.db.WithContext(ctx).
		Raw(`
			SELECT event_id, user_id, display_name, avatar_url
			FROM (
				SELECT
					ep.event_id,
					u.id AS user_id,
					u.display_name,
					u.avatar_url,
					ROW_NUMBER() OVER (PARTITION BY ep.event_id ORDER BY ep.created_at ASC, ep.id ASC) AS rn
				FROM event_participants ep
				JOIN users u ON u.id = ep.user_id
				WHERE ep.event_id IN ?
					AND ep.rsvp = ?
					AND ep.deleted_at IS NULL
			) ranked
			WHERE rn <= 4
			ORDER BY event_id ASC, rn ASC
		`, eventIDs, "going").
		Scan(&previewRows).
		Error; err != nil {
		return nil, err
	}
	for _, row := range previewRows {
		summary := summaries[row.EventID]
		summary.ParticipantsPreview = append(summary.ParticipantsPreview, EventParticipantPreview{
			ID:          row.UserID,
			DisplayName: row.DisplayName,
			AvatarURL:   row.AvatarURL,
		})
		summaries[row.EventID] = summary
	}

	return summaries, nil
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
