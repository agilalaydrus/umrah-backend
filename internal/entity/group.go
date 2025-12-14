package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 1. DATABASE MODELS
type Group struct {
	ID         uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name       string         `gorm:"size:100;not null" json:"name"`
	MutawwifID uuid.UUID      `gorm:"type:uuid;not null" json:"mutawwif_id"`         // The Leader
	Mutawwif   User           `gorm:"foreignKey:MutawwifID" json:"-"`                // Relation
	JoinCode   string         `gorm:"size:10;uniqueIndex;not null" json:"join_code"` // e.g. "UMROH-2025"
	StartDate  time.Time      `json:"start_date"`
	EndDate    time.Time      `json:"end_date"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type GroupMember struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GroupID   uuid.UUID `gorm:"type:uuid;not null" json:"group_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"` // Fetch User details
	Status    string    `gorm:"default:'ACTIVE'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// 2. REQUEST DTOs
type CreateGroupDTO struct {
	Name      string `json:"name" validate:"required,min=5"`
	JoinCode  string `json:"join_code" validate:"required,min=4,alphanum"`
	StartDate string `json:"start_date" validate:"required"` // Format: YYYY-MM-DD
	EndDate   string `json:"end_date" validate:"required"`
}

type JoinGroupDTO struct {
	JoinCode string `json:"join_code" validate:"required"`
}
