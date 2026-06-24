package users

import "github.com/google/uuid"

type Profile struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DisplayName string    `gorm:"column:display_name" json:"display_name"`
	AvatarURL   *string   `gorm:"column:avatar_url" json:"avatar_url"`
	Email       *string   `gorm:"column:email" json:"email"`
	Status      string    `gorm:"column:status" json:"status"`
}

func (Profile) TableName() string {
	return "users"
}
