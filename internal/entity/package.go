package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- ENUMS ---
type PackageCategory string
type AirlineClass string

const (
	// Categories
	CatUmrahReguler PackageCategory = "UMRAH_REGULER"
	CatUmrahPlus    PackageCategory = "UMRAH_PLUS" // Includes Turkey, Aqsa, etc.
	CatHajjFuroda   PackageCategory = "HAJJ_FURODA"
	CatHajjPlus     PackageCategory = "HAJJ_PLUS"
	CatRamadhan     PackageCategory = "UMRAH_RAMADHAN"

	// Airline Classes
	ClassEconomy  AirlineClass = "ECONOMY"
	ClassBusiness AirlineClass = "BUSINESS"
	ClassFirst    AirlineClass = "FIRST"
)

// --- TRAVEL PACKAGES ---
type TravelPackage struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`

	// 1. CATEGORY & TAGS
	Category    PackageCategory `gorm:"type:varchar(50);index" json:"category"`
	SubCategory string          `gorm:"type:varchar(100)" json:"sub_category"` // e.g. "Turkey", "Aqsa"
	Tags        string          `gorm:"type:varchar(255)" json:"tags"`         // Comma separated tags

	// 2. HOTEL TIERING
	HotelMakkah   string `gorm:"type:varchar(100)" json:"hotel_makkah"`
	RatingMakkah  int    `json:"rating_makkah"` // 1-5 Star
	HotelMadinah  string `gorm:"type:varchar(100)" json:"hotel_madinah"`
	RatingMadinah int    `json:"rating_madinah"` // 1-5 Star

	// 3. AIRLINE TIERING
	AirlineName  string       `gorm:"type:varchar(100)" json:"airline_name"`
	AirlineClass AirlineClass `gorm:"type:varchar(20);default:'ECONOMY'" json:"airline_class"`

	// 4. PRICING TIERS (Per Pax)
	PriceQuad   float64 `json:"price_quad"` // Cheapest (4 pax/room)
	PriceTriple float64 `json:"price_triple"`
	PriceDouble float64 `json:"price_double"` // Most Expensive (2 pax/room)

	// 5. SCHEDULE
	DurationDays  int       `json:"duration_days"`
	DepartureDate time.Time `json:"departure_date"`
	ReturnDate    time.Time `json:"return_date"`
	DepartureCity string    `gorm:"type:varchar(50)" json:"departure_city"`

	// 6. INVENTORY
	Quota     int  `json:"quota"`
	Available int  `json:"available"`
	IsActive  bool `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// --- BOOKING ---
type Booking struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	PackageID uuid.UUID      `gorm:"type:uuid;not null;index" json:"package_id"`
	Package   *TravelPackage `gorm:"foreignKey:PackageID" json:"package,omitempty"`

	PaxCount   int     `json:"pax_count"`
	RoomType   string  `json:"room_type"` // "QUAD", "TRIPLE", "DOUBLE"
	TotalPrice float64 `json:"total_price"`

	Status string `gorm:"type:varchar(20);default:'PENDING'" json:"status"` // PENDING, PAID, CONFIRMED
	Notes  string `gorm:"type:text" json:"notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
