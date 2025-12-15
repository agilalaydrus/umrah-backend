package entity

import (
	"time"

	"github.com/google/uuid"
)

type ManasikType string

const (
	TypeGuide ManasikType = "GUIDE" // Long text/Fiqih
	TypeDua   ManasikType = "DUA"   // Prayer with Audio
)

type ManasikCategory string

const (
	CatUmrah   ManasikCategory = "UMRAH"
	CatHajj    ManasikCategory = "HAJJ"
	CatGeneral ManasikCategory = "GENERAL"
)

type Manasik struct {
	ID       uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title    string          `gorm:"type:varchar(255);not null" json:"title"`
	Type     ManasikType     `gorm:"type:varchar(20)" json:"type"`
	Category ManasikCategory `gorm:"type:varchar(20);index" json:"category"`

	ArabicText  string `gorm:"type:text" json:"arabic_text,omitempty"`
	LatinText   string `gorm:"type:text" json:"latin_text,omitempty"`
	Translation string `gorm:"type:text" json:"translation,omitempty"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	AudioURL      string `gorm:"type:varchar(255)" json:"audio_url,omitempty"`
	ImageURL      string `gorm:"type:varchar(255)" json:"image_url,omitempty"`
	SequenceOrder int    `json:"sequence_order"` // For sorting steps (1,2,3)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
