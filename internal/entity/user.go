package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- ROLE CONSTANTS ---
const (
	RoleAdmin    = "ADMIN"
	RoleMutawwif = "MUTAWWIF"
	RoleJamaah   = "JAMAAH"
)

// DATABASE MODEL
type User struct {
	// Hapus 'default:gen_random_uuid()' agar database agnostic
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FullName    string    `gorm:"size:100;not null" json:"full_name"`
	PhoneNumber string    `gorm:"size:20;uniqueIndex;not null" json:"phone_number"`
	Password    string    `gorm:"not null" json:"-"`
	Role        string    `gorm:"size:20;default:'JAMAAH'" json:"role"`

	// Fitur Reset Password yang Aman
	ResetToken       *string    `gorm:"size:10" json:"-"`
	ResetTokenExpiry *time.Time `json:"-"` // Token harus punya expired time

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// HOOK GORM: Otomatis generate UUID sebelum data dibuat
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	// Safety net: jika role kosong, set ke JAMAAH
	if u.Role == "" {
		u.Role = RoleJamaah
	}
	return
}

// --- REQUEST DTOs ---

// 1. Public Register (Hanya untuk Jamaah)
type RegisterDTO struct {
	FullName    string `json:"full_name" validate:"required,min=3"`
	PhoneNumber string `json:"phone_number" validate:"required,min=9"`
	Password    string `json:"password" validate:"required,min=6"`
	// Role DIHAPUS dari sini agar user tidak bisa jadi Admin sembarangan
}

// 2. Admin Create User (Internal Dashboard)
type CreateUserInternalDTO struct {
	FullName    string `json:"full_name" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Role        string `json:"role" validate:"required,oneof=JAMAAH MUTAWWIF ADMIN"`
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
