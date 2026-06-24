package notifications

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	TypeFriendRequest  = "friend_request"
	TypeFriendAccepted = "friend_accepted"
	TypeEventInvite    = "event_invite"
	TypeEventChanged   = "event_changed"
	TypeEventCancelled = "event_cancelled"
	TypeGroupInvite    = "group_invite"
)

type Payload map[string]any

func (p Payload) Value() (driver.Value, error) {
	if p == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(p)
}

func (p *Payload) Scan(value any) error {
	if value == nil {
		*p = Payload{}
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unsupported payload type")
	}
	if len(data) == 0 {
		*p = Payload{}
		return nil
	}
	return json.Unmarshal(data, p)
}

type Notification struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;column:user_id" json:"user_id"`
	Type      string     `gorm:"column:type" json:"type"`
	Payload   Payload    `gorm:"column:payload;type:jsonb" json:"payload"`
	ReadAt    *time.Time `gorm:"column:read_at" json:"read_at"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}

type ListResult struct {
	UnreadCount int            `json:"unread_count"`
	Items       []Notification `json:"items"`
}
