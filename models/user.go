package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FullName    string `json:"full_name"`
	PhoneNumber string `gorm:"unique;not null" json:"phone_number"`
	Password    string `json:"-"`
	Role        string `json:"role"` // 'JAMAAH', 'MUTAWWIF'

	// New field for Reset Password logic
	ResetToken string `json:"-"`
}
