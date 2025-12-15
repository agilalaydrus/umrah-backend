package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"` // "Roaming Telkomsel 10GB"
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(15,2);not null" json:"price"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
}

type OrderStatus string

const (
	OrderPending   OrderStatus = "PENDING"   // User created order, hasn't paid
	OrderPaid      OrderStatus = "PAID"      // User uploaded proof
	OrderCompleted OrderStatus = "COMPLETED" // Admin verified proof
	OrderCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	Amount float64     `json:"amount"` // Snapshot of price at purchase time
	Status OrderStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`

	// URL to the uploaded transfer proof image
	ProofImage string `gorm:"type:text" json:"proof_image"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
