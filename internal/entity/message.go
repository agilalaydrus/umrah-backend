package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageType string

const (
	MsgText    MessageType = "TEXT"
	MsgSOS     MessageType = "SOS"
	MsgInfo    MessageType = "INFO"
	MsgDeleted MessageType = "DELETED" // [BARU] Tambahkan ini
)

type Message struct {
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GroupID uuid.UUID `gorm:"type:uuid;not null;index" json:"group_id"`

	// --- PERBAIKAN PENTING ---
	// 1. Pastikan SenderID tipe datanya SAMA PERSIS dengan ID di entity User (uuid.UUID)
	SenderID uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`

	// 2. Ubah jadi Pointer (*User) dan tambahkan references:ID
	// Menggunakan pointer memungkinkan nilai null dan lebih efisien memori
	Sender *User `gorm:"foreignKey:SenderID;references:ID" json:"sender"`

	Content   string         `gorm:"type:text;not null" json:"content"`
	Type      MessageType    `gorm:"type:varchar(10);default:'TEXT'" json:"type"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// DTO for incoming WebSocket payload
type MessagePayload struct {
	Content string      `json:"content"`
	Type    MessageType `json:"type"`
}
