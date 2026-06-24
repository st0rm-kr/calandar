package inbox

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) ListInvitedEvents(ctx context.Context, userID uuid.UUID) ([]EventInvite, error) {
	var rows []EventInvite
	err := r.db.WithContext(ctx).
		Table("event_participants AS p").
		Select(`p.id AS participant_id, e.id AS event_id, e.title, e.start_at, e.end_at, e.location, e.status, p.created_at,
			(
				SELECT count(*) FROM schedules s
				WHERE s.user_id = p.user_id
				  AND s.deleted_at IS NULL
				  AND (s.event_id IS NULL OR s.event_id <> e.id)
				  AND s.start_at < COALESCE(e.end_at, e.start_at + interval '1 minute')
				  AND e.start_at < COALESCE(s.end_at, s.start_at + interval '1 minute')
			) AS conflict_count`).
		Joins("JOIN events e ON e.id = p.event_id AND e.deleted_at IS NULL").
		Where("p.user_id = ? AND p.rsvp = ? AND p.deleted_at IS NULL", userID, "invited").
		Order("p.created_at DESC").
		Scan(&rows).
		Error
	return rows, err
}

func (r *Repo) ListPendingFriendRequests(ctx context.Context, userID uuid.UUID) ([]FriendRequest, error) {
	var rows []FriendRequest
	err := r.db.WithContext(ctx).
		Table("friendships AS f").
		Select("f.id, f.requester_id, u.display_name AS requester_name, f.created_at").
		Joins("JOIN users u ON u.id = f.requester_id").
		Where("f.addressee_id = ? AND f.status = ? AND f.deleted_at IS NULL", userID, "pending").
		Order("f.created_at DESC").
		Scan(&rows).
		Error
	return rows, err
}

func (r *Repo) ListPendingGroupInvites(ctx context.Context, userID uuid.UUID) ([]GroupInvite, error) {
	var rows []GroupInvite
	err := r.db.WithContext(ctx).
		Table("group_invites AS gi").
		Select("gi.id, gi.group_id, g.name AS group_name, u.display_name AS inviter_name, gi.created_at").
		Joins("JOIN groups g ON g.id = gi.group_id AND g.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = gi.inviter_id").
		Where("gi.invitee_id = ? AND gi.status = ? AND gi.deleted_at IS NULL", userID, "pending").
		Order("gi.created_at DESC").
		Scan(&rows).
		Error
	return rows, err
}
