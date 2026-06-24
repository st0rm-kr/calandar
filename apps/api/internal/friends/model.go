package friends

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusRejected = "rejected"
)

type Friendship struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	RequesterID uuid.UUID  `gorm:"type:uuid;column:requester_id" json:"requester_id"`
	AddresseeID uuid.UUID  `gorm:"type:uuid;column:addressee_id" json:"addressee_id"`
	Status      string     `gorm:"column:status" json:"status"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (Friendship) TableName() string {
	return "friendships"
}

type Friend struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Email       *string   `json:"email"`
}

type RequestView struct {
	ID          int64     `json:"id"`
	RequesterID uuid.UUID `json:"requester_id"`
	AddresseeID uuid.UUID `json:"addressee_id"`
	Status      string    `json:"status"`
	Direction   string    `json:"direction"`
}
