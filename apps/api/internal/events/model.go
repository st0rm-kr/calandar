package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive    = "active"
	StatusCancelled = "cancelled"
	StatusExpired   = "expired"
)

type Event struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	OwnerID   uuid.UUID  `gorm:"type:uuid;column:owner_id" json:"owner_id"`
	Scope     string     `gorm:"column:scope" json:"scope"`
	GroupID   *int64     `gorm:"column:group_id" json:"group_id"`
	Title     string     `gorm:"column:title" json:"title"`
	Type      string     `gorm:"column:type" json:"type"`
	StartAt   time.Time  `gorm:"column:start_at" json:"start_at"`
	EndAt     *time.Time `gorm:"column:end_at" json:"end_at"`
	Location  *string    `gorm:"column:location" json:"location"`
	Capacity  *int       `gorm:"column:capacity" json:"capacity"`
	Status    string     `gorm:"column:status" json:"status"`
	ShareSlug string     `gorm:"column:share_slug" json:"share_slug"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (Event) TableName() string {
	return "events"
}

type Participant struct {
	ID            int64      `gorm:"primaryKey" json:"id"`
	EventID       int64      `gorm:"column:event_id" json:"event_id"`
	UserID        uuid.UUID  `gorm:"type:uuid;column:user_id" json:"user_id"`
	RSVP          string     `gorm:"column:rsvp" json:"rsvp"`
	Source        string     `gorm:"column:source" json:"source"`
	AddToCalendar bool       `gorm:"column:add_to_calendar" json:"add_to_calendar"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (Participant) TableName() string {
	return "event_participants"
}

type EventDetail struct {
	Event      Event `json:"event"`
	GoingCount int   `json:"going_count"`
}

type CreateEventInput struct {
	Title    string
	Type     string
	Scope    string
	GroupID  *int64
	StartAt  time.Time
	EndAt    *time.Time
	Location *string
	Capacity *int
}
