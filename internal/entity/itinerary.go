package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Feature 4: Rundown / Itinerary
type Itinerary struct {
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GroupID uuid.UUID `gorm:"type:uuid;not null;index" json:"group_id"`
	Group   *Group    `gorm:"foreignKey:GroupID" json:"group,omitempty"`

	Title       string  `gorm:"type:varchar(255);not null" json:"title"` // e.g. "City Tour Madinah"
	Description string  `gorm:"type:text" json:"description"`            // e.g. "Kumpul di Lobby"
	Location    string  `gorm:"type:varchar(255)" json:"location"`       // e.g. "Hotel Al Haram"
	Latitude    float64 `json:"latitude"`                                // For Map
	Longitude   float64 `json:"longitude"`                               // For Map

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Feature 3: Absensi / Attendance
type Attendance struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ItineraryID uuid.UUID `gorm:"type:uuid;not null;index" json:"itinerary_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Status: "PRESENT", "LATE", "ABSENT"
	Status    string    `gorm:"type:varchar(20);default:'PRESENT'" json:"status"`
	ScannedAt time.Time `json:"scanned_at"`

	// IsManual: true if Mutawwif clicked "Present" manually instead of QR Scan
	IsManual bool `gorm:"default:false" json:"is_manual"`
}
