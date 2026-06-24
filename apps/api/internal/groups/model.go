package groups

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive    = "active"
	StatusDissolved = "dissolved"

	RoleOwner  = "owner"
	RoleMember = "member"

	InviteStatusPending  = "pending"
	InviteStatusAccepted = "accepted"
	InviteStatusRejected = "rejected"
)

type Group struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"column:name" json:"name"`
	Description *string    `gorm:"column:description" json:"description"`
	OwnerID     uuid.UUID  `gorm:"type:uuid;column:owner_id" json:"owner_id"`
	InviteCode  string     `gorm:"column:invite_code" json:"invite_code"`
	Status      string     `gorm:"column:status" json:"status"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (Group) TableName() string {
	return "groups"
}

type GroupMember struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	GroupID   int64      `gorm:"column:group_id" json:"group_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;column:user_id" json:"user_id"`
	Role      string     `gorm:"column:role" json:"role"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (GroupMember) TableName() string {
	return "group_members"
}

type GroupInvite struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	GroupID   int64      `gorm:"column:group_id" json:"group_id"`
	InviterID uuid.UUID  `gorm:"type:uuid;column:inviter_id" json:"inviter_id"`
	InviteeID uuid.UUID  `gorm:"type:uuid;column:invitee_id" json:"invitee_id"`
	Status    string     `gorm:"column:status" json:"status"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (GroupInvite) TableName() string {
	return "group_invites"
}

type Member struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Email       *string   `json:"email"`
	Role        string    `json:"role"`
}

type GroupDetail struct {
	Group   Group    `json:"group"`
	Members []Member `json:"members"`
}

type CreateGroupInput struct {
	Name        string
	Description *string
}

type UpdateGroupInput struct {
	Name        *string
	Description *string
}
