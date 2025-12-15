package service

import (
	"context"
	"errors"
	"time"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"

	"github.com/google/uuid"
)

type ItineraryService interface {
	CreateItinerary(ctx context.Context, req entity.Itinerary) error
	GetRundown(ctx context.Context, groupID string) ([]entity.Itinerary, error)

	// Core Logic: Attendance
	ScanAttendance(ctx context.Context, userID, itineraryID string) error
	GetAttendanceReport(ctx context.Context, itineraryID string) ([]entity.Attendance, error)
}

type itineraryService struct {
	repo repository.ItineraryRepository
}

func NewItineraryService(repo repository.ItineraryRepository) ItineraryService {
	return &itineraryService{repo: repo}
}

func (s *itineraryService) CreateItinerary(ctx context.Context, req entity.Itinerary) error {
	// Simple passthrough, validation can be added here
	return s.repo.CreateItinerary(ctx, &req)
}

func (s *itineraryService) GetRundown(ctx context.Context, groupID string) ([]entity.Itinerary, error) {
	return s.repo.GetItineraryByGroup(ctx, groupID)
}

func (s *itineraryService) ScanAttendance(ctx context.Context, userID, itineraryID string) error {
	// 1. Check if user already scanned
	existing, err := s.repo.GetAttendanceByUserAndItinerary(ctx, userID, itineraryID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("already scanned") // Or just return nil to be idempotent
	}

	// 2. Verify Itinerary Exists
	itinerary, err := s.repo.FindItineraryByID(ctx, itineraryID)
	if err != nil {
		return errors.New("invalid QR code: itinerary not found")
	}

	// 3. Create Attendance Record
	attendance := &entity.Attendance{
		ID:          uuid.New(),
		UserID:      uuid.MustParse(userID),
		ItineraryID: itinerary.ID,
		Status:      "PRESENT",
		ScannedAt:   time.Now(),
		IsManual:    false,
	}

	return s.repo.CreateAttendance(ctx, attendance)
}

func (s *itineraryService) GetAttendanceReport(ctx context.Context, itineraryID string) ([]entity.Attendance, error) {
	return s.repo.GetAttendanceByItinerary(ctx, itineraryID)
}
