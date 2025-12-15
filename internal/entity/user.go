package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- [NEW] ROLE CONSTANTS ---
// Using constants prevents typos in your code logic
const (
	RoleAdmin    = "ADMIN"    // Travel Agent Staff
	RoleMutawwif = "MUTAWWIF" // Tour Leader
	RoleJamaah   = "JAMAAH"   // Pilgrim
)

// DATABASE MODEL
type User struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	FullName    string         `gorm:"size:100;not null" json:"full_name"`
	PhoneNumber string         `gorm:"size:20;uniqueIndex;not null" json:"phone_number"`
	Password    string         `gorm:"not null" json:"-"`
	Role        string         `gorm:"size:20;default:'JAMAAH'" json:"role"`
	ResetToken  *string        `gorm:"size:10" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// REQUEST DTOs (With Validation Tags)
type RegisterDTO struct {
	FullName    string `json:"full_name" validate:"required,min=3"`
	PhoneNumber string `json:"phone_number" validate:"required,min=9"`
	Password    string `json:"password" validate:"required,min=6"`

	// [UPDATE] Added ADMIN to the allowed roles list
	Role string `json:"role" validate:"required,oneof=JAMAAH MUTAWWIF ADMIN"`
}

type LoginDTO struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Password    string `json:"password" validate:"required"`
}

type ForgotPasswordDTO struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
}

type ResetPasswordDTO struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	OTP         string `json:"otp" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}
