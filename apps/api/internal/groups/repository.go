package groups

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

func (r *Repo) Transaction(ctx context.Context, fn func(Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Repo{db: tx})
	})
}

func (r *Repo) CreateGroup(ctx context.Context, group Group) (Group, error) {
	if err := r.db.WithContext(ctx).Create(&group).Error; err != nil {
		return Group{}, err
	}
	return group, nil
}

func (r *Repo) FindGroupByID(ctx context.Context, id int64) (Group, error) {
	var group Group
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&group).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Group{}, ErrGroupNotFound
	}
	return group, err
}

func (r *Repo) FindGroupByInviteCode(ctx context.Context, code string) (Group, error) {
	var group Group
	err := r.db.WithContext(ctx).
		Where("invite_code = ? AND deleted_at IS NULL", code).
		First(&group).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Group{}, ErrGroupNotFound
	}
	return group, err
}

func (r *Repo) UpdateGroup(ctx context.Context, id int64, fields map[string]any) (Group, error) {
	fields["updated_at"] = time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&Group{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(fields)
	if result.Error != nil {
		return Group{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Group{}, ErrGroupNotFound
	}
	return r.FindGroupByID(ctx, id)
}

func (r *Repo) ListGroupsByMember(ctx context.Context, userID uuid.UUID) ([]Group, error) {
	var groups []Group
	err := r.db.WithContext(ctx).
		Table("groups AS g").
		Select("g.*").
		Joins("JOIN group_members m ON m.group_id = g.id AND m.deleted_at IS NULL").
		Where("m.user_id = ? AND g.deleted_at IS NULL", userID).
		Order("g.created_at DESC").
		Scan(&groups).
		Error
	return groups, err
}

func (r *Repo) FindMember(ctx context.Context, groupID int64, userID uuid.UUID) (GroupMember, bool, error) {
	var member GroupMember
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND deleted_at IS NULL", groupID, userID).
		First(&member).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return GroupMember{}, false, nil
	}
	if err != nil {
		return GroupMember{}, false, err
	}
	return member, true, nil
}

func (r *Repo) AddMember(ctx context.Context, member GroupMember) (GroupMember, error) {
	if err := r.db.WithContext(ctx).Create(&member).Error; err != nil {
		return GroupMember{}, err
	}
	return member, nil
}

func (r *Repo) RemoveMember(ctx context.Context, groupID int64, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&GroupMember{}).
		Where("group_id = ? AND user_id = ? AND deleted_at IS NULL", groupID, userID).
		Update("deleted_at", time.Now().UTC()).
		Error
}

func (r *Repo) CountMembers(ctx context.Context, groupID int64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&GroupMember{}).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Count(&count).
		Error
	return int(count), err
}

func (r *Repo) ListMembers(ctx context.Context, groupID int64) ([]Member, error) {
	var members []Member
	err := r.db.WithContext(ctx).
		Table("group_members AS m").
		Select("u.id AS user_id, u.display_name, u.avatar_url, u.email, m.role").
		Joins("JOIN users u ON u.id = m.user_id").
		Where("m.group_id = ? AND m.deleted_at IS NULL", groupID).
		Order("m.created_at ASC").
		Scan(&members).
		Error
	return members, err
}

func (r *Repo) CreateInvite(ctx context.Context, invite GroupInvite) (GroupInvite, error) {
	if err := r.db.WithContext(ctx).Create(&invite).Error; err != nil {
		return GroupInvite{}, err
	}
	return invite, nil
}

func (r *Repo) FindInviteByID(ctx context.Context, id int64) (GroupInvite, error) {
	var invite GroupInvite
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&invite).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return GroupInvite{}, ErrInviteNotFound
	}
	return invite, err
}

func (r *Repo) FindInvite(ctx context.Context, groupID int64, inviteeID uuid.UUID) (GroupInvite, bool, error) {
	var invite GroupInvite
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND invitee_id = ? AND deleted_at IS NULL", groupID, inviteeID).
		First(&invite).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return GroupInvite{}, false, nil
	}
	if err != nil {
		return GroupInvite{}, false, err
	}
	return invite, true, nil
}

func (r *Repo) UpdateInviteStatus(ctx context.Context, id int64, status string) (GroupInvite, error) {
	result := r.db.WithContext(ctx).
		Model(&GroupInvite{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now().UTC()})
	if result.Error != nil {
		return GroupInvite{}, result.Error
	}
	if result.RowsAffected == 0 {
		return GroupInvite{}, ErrInviteNotFound
	}
	return r.FindInviteByID(ctx, id)
}
