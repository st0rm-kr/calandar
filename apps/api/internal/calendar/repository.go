package calendar

import (
	"context"
	"time"

	"github.com/bytedance/calandar/apps/api/internal/friends"
	"github.com/bytedance/calandar/apps/api/internal/schedules"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]schedules.Schedule, error) {
	return schedules.NewRepository(r.db).ListByRange(ctx, userID, from, to)
}

func (r *Repository) ListParticipantEventsByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]schedules.Schedule, error) {
	return schedules.NewRepository(r.db).ListParticipantEventsByRange(ctx, userID, from, to)
}

func (r *Repository) ListEventParticipationSummaries(ctx context.Context, userID uuid.UUID, eventIDs []int64) (map[int64]schedules.EventParticipationSummary, error) {
	return schedules.NewRepository(r.db).ListEventParticipationSummaries(ctx, userID, eventIDs)
}

func (r *Repository) ListCalendarFriends(ctx context.Context, userID uuid.UUID) ([]CalendarFriend, error) {
	var rows []CalendarFriend
	err := r.db.WithContext(ctx).
		Table("friendships AS f").
		Select("u.id AS user_id, u.display_name").
		Joins("JOIN users u ON u.id = CASE WHEN f.requester_id = ? THEN f.addressee_id ELSE f.requester_id END", userID).
		Where("f.status = ? AND f.deleted_at IS NULL", friends.StatusAccepted).
		Where("f.requester_id = ? OR f.addressee_id = ?", userID, userID).
		Order("u.display_name ASC, u.id ASC").
		Scan(&rows).
		Error
	return rows, err
}

func (r *Repository) ListCalendarGroups(ctx context.Context, userID uuid.UUID) ([]CalendarGroup, error) {
	var rows []CalendarGroup
	err := r.db.WithContext(ctx).
		Table("groups AS g").
		Select("g.id, g.name").
		Joins("JOIN group_members m ON m.group_id = g.id AND m.deleted_at IS NULL").
		Where("m.user_id = ? AND g.deleted_at IS NULL AND g.status = ?", userID, "active").
		Order("g.name ASC, g.id ASC").
		Scan(&rows).
		Error
	return rows, err
}

func (r *Repository) ListFriendSchedulesByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]FriendSchedule, error) {
	var rows []struct {
		ID         int64
		UserID     uuid.UUID
		Title      string
		StartAt    time.Time
		EndAt      *time.Time
		Location   *string
		Visibility string
		Source     string
		EventID    *int64
		FriendID   uuid.UUID
		FriendName string
	}
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT
				s.id,
				s.user_id,
				s.title,
				s.start_at,
				s.end_at,
				s.location,
				s.visibility,
				s.source,
				s.event_id,
				u.id AS friend_id,
				u.display_name AS friend_name
			FROM schedules s
			JOIN users u ON u.id = s.user_id
			JOIN friendships f ON (
				(f.requester_id = ? AND f.addressee_id = s.user_id) OR
				(f.addressee_id = ? AND f.requester_id = s.user_id)
			)
			WHERE s.user_id <> ?
				AND s.deleted_at IS NULL
				AND s.visibility <> ?
				AND f.status = ?
				AND f.deleted_at IS NULL
				AND s.start_at < ?
				AND ? < COALESCE(s.end_at, s.start_at + interval '1 minute')
			ORDER BY s.start_at ASC
		`, userID, userID, userID, schedules.VisibilityPrivate, friends.StatusAccepted, to, from).
		Scan(&rows).
		Error
	if err != nil {
		return nil, err
	}

	result := make([]FriendSchedule, 0, len(rows))
	for _, row := range rows {
		result = append(result, FriendSchedule{
			Schedule: schedules.Schedule{
				ID:         row.ID,
				UserID:     row.UserID,
				Title:      row.Title,
				StartAt:    row.StartAt,
				EndAt:      row.EndAt,
				Location:   row.Location,
				Visibility: row.Visibility,
				Source:     row.Source,
				EventID:    row.EventID,
			},
			FriendID:   row.FriendID,
			FriendName: row.FriendName,
		})
	}
	return result, nil
}

func (r *Repository) ListGroupEventsByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]GroupEvent, error) {
	var rows []struct {
		ID        int64
		OwnerID   uuid.UUID
		Title     string
		StartAt   time.Time
		EndAt     *time.Time
		Location  *string
		GroupID   int64
		GroupName string
	}
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT
				e.id,
				e.owner_id,
				e.title,
				e.start_at,
				e.end_at,
				e.location,
				g.id AS group_id,
				g.name AS group_name
			FROM events e
			JOIN groups g ON g.id = e.group_id
			JOIN group_members m ON m.group_id = g.id AND m.deleted_at IS NULL
			WHERE m.user_id = ?
				AND e.scope = ?
				AND e.group_id IS NOT NULL
				AND e.status = ?
				AND e.deleted_at IS NULL
				AND g.status = ?
				AND g.deleted_at IS NULL
				AND e.start_at < ?
				AND ? < COALESCE(e.end_at, e.start_at + interval '1 minute')
			ORDER BY e.start_at ASC
		`, userID, "group", "active", "active", to, from).
		Scan(&rows).
		Error
	if err != nil {
		return nil, err
	}

	result := make([]GroupEvent, 0, len(rows))
	for _, row := range rows {
		eventID := row.ID
		result = append(result, GroupEvent{
			Schedule: schedules.Schedule{
				ID:         row.ID,
				UserID:     row.OwnerID,
				Title:      row.Title,
				StartAt:    row.StartAt,
				EndAt:      row.EndAt,
				Location:   row.Location,
				Visibility: schedules.VisibilityPublic,
				Source:     schedules.SourceEvent,
				EventID:    &eventID,
			},
			GroupID:   row.GroupID,
			GroupName: row.GroupName,
		})
	}
	return result, nil
}
