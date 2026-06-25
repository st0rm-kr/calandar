package schedules

import (
	"time"

	"github.com/google/uuid"
)

const (
	VisibilityPublic   = "public"
	VisibilityBusyOnly = "busy_only"
	VisibilityPrivate  = "private"

	SourceManual = "manual"
	SourceEvent  = "event"
)

type Schedule struct {
	ID         int64      `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;column:user_id" json:"user_id"`
	Title      string     `gorm:"column:title" json:"title"`
	StartAt    time.Time  `gorm:"column:start_at" json:"start_at"`
	EndAt      *time.Time `gorm:"column:end_at" json:"end_at"`
	Location   *string    `gorm:"column:location" json:"location"`
	Visibility string     `gorm:"column:visibility" json:"visibility"`
	Source     string     `gorm:"column:source" json:"source"`
	EventID    *int64     `gorm:"column:event_id" json:"event_id"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at" json:"-"`
}

func (Schedule) TableName() string {
	return "schedules"
}

type EventScheduleInput struct {
	ID       int64
	Title    string
	StartAt  time.Time
	EndAt    *time.Time
	Location *string
}

type Conflict struct {
	ID         int64      `json:"id"`
	Title      string     `json:"title"`
	StartAt    time.Time  `json:"start_at"`
	EndAt      *time.Time `json:"end_at"`
	Location   *string    `json:"location"`
	Visibility string     `json:"visibility"`
	Source     string     `json:"source"`
	EventID    *int64     `json:"event_id"`
}

type EventParticipantPreview struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
}

type EventParticipationSummary struct {
	EventID             int64                     `json:"event_id"`
	ViewerRSVP          *string                   `json:"viewer_rsvp"`
	GoingCount          int                       `json:"going_count"`
	ParticipantsPreview []EventParticipantPreview `json:"participants_preview"`
}
