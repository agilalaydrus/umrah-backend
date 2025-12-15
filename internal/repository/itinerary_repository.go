package repository

import (
	"context"
	"errors"
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type ItineraryRepository interface {
	// Itinerary CRUD for Admin/Mutawwif
	CreateItinerary(ctx context.Context, itinerary *entity.Itinerary) error
	GetItineraryByGroup(ctx context.Context, groupID string) ([]entity.Itinerary, error)
	FindItineraryByID(ctx context.Context, id string) (*entity.Itinerary, error)
	UpdateItinerary(ctx context.Context, itinerary *entity.Itinerary) error
	DeleteItinerary(ctx context.Context, id string) error

	// Attendance Management
	CreateAttendance(ctx context.Context, attendance *entity.Attendance) error
	GetAttendanceByItinerary(ctx context.Context, itineraryID string) ([]entity.Attendance, error)
	GetAttendanceByUserAndItinerary(ctx context.Context, userID, itineraryID string) (*entity.Attendance, error)
}

type itineraryRepo struct {
	db *gorm.DB
}

func NewItineraryRepository(db *gorm.DB) ItineraryRepository {
	return &itineraryRepo{db: db}
}

// -----------------------------------------------------------
// Implementation: Itinerary
// -----------------------------------------------------------

func (r *itineraryRepo) CreateItinerary(ctx context.Context, itinerary *entity.Itinerary) error {
	return r.db.WithContext(ctx).Create(itinerary).Error
}

func (r *itineraryRepo) GetItineraryByGroup(ctx context.Context, groupID string) ([]entity.Itinerary, error) {
	var itineraries []entity.Itinerary
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("start_time asc").
		Find(&itineraries).Error
	return itineraries, err
}

func (r *itineraryRepo) FindItineraryByID(ctx context.Context, id string) (*entity.Itinerary, error) {
	var itinerary entity.Itinerary
	err := r.db.WithContext(ctx).First(&itinerary, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &itinerary, nil
}

func (r *itineraryRepo) UpdateItinerary(ctx context.Context, itinerary *entity.Itinerary) error {
	return r.db.WithContext(ctx).Save(itinerary).Error
}

func (r *itineraryRepo) DeleteItinerary(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Itinerary{}, "id = ?", id).Error
}

// -----------------------------------------------------------
// Implementation: Attendance
// -----------------------------------------------------------

func (r *itineraryRepo) CreateAttendance(ctx context.Context, attendance *entity.Attendance) error {
	return r.db.WithContext(ctx).Create(attendance).Error
}

func (r *itineraryRepo) GetAttendanceByItinerary(ctx context.Context, itineraryID string) ([]entity.Attendance, error) {
	var attendances []entity.Attendance
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("itinerary_id = ?", itineraryID).
		Find(&attendances).Error
	return attendances, err
}

func (r *itineraryRepo) GetAttendanceByUserAndItinerary(ctx context.Context, userID, itineraryID string) (*entity.Attendance, error) {
	var attendance entity.Attendance
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND itinerary_id = ?", userID, itineraryID).
		First(&attendance).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if not found, not an error
		}
		return nil, err
	}
	return &attendance, nil
}
