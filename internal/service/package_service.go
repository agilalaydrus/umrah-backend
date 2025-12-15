package service

import (
	"context"
	"errors"
	"time"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"

	"github.com/google/uuid"
)

type PackageService interface {
	CreatePackage(ctx context.Context, req entity.TravelPackage) error
	GetList(ctx context.Context, category string) ([]entity.TravelPackage, error)
	BookPackage(ctx context.Context, userID, packageID, roomType string, pax int) (*entity.Booking, error)
}

type packageService struct {
	repo repository.PackageRepository
}

func NewPackageService(repo repository.PackageRepository) PackageService {
	return &packageService{repo: repo}
}

func (s *packageService) CreatePackage(ctx context.Context, req entity.TravelPackage) error {
	req.ID = uuid.New()
	req.CreatedAt = time.Now()
	req.Available = req.Quota
	return s.repo.CreatePackage(ctx, &req)
}

func (s *packageService) GetList(ctx context.Context, category string) ([]entity.TravelPackage, error) {
	return s.repo.GetPackages(ctx, category)
}

func (s *packageService) BookPackage(ctx context.Context, userID, packageID, roomType string, pax int) (*entity.Booking, error) {
	// 1. Get Package Data (Read Only, for Pricing)
	pkg, err := s.repo.FindPackageByID(ctx, packageID)
	if err != nil {
		return nil, errors.New("package not found")
	}

	// 2. Calculate Price
	var pricePerPax float64
	switch roomType {
	case "QUAD":
		pricePerPax = pkg.PriceQuad
	case "TRIPLE":
		pricePerPax = pkg.PriceTriple
	case "DOUBLE":
		pricePerPax = pkg.PriceDouble
	default:
		return nil, errors.New("invalid room type")
	}

	totalPrice := pricePerPax * float64(pax)

	// 3. Create Booking Object
	booking := &entity.Booking{
		ID:         uuid.New(),
		UserID:     uuid.MustParse(userID),
		PackageID:  pkg.ID,
		PaxCount:   pax,
		RoomType:   roomType,
		TotalPrice: totalPrice,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}

	// 4. [ATOMIC] Deduct Quota FIRST
	// Kita kurangi dulu kuotanya menggunakan Query UPDATE atomic.
	// Jika return error (rows affected 0), berarti kuota habis.
	if err := s.repo.DecreaseQuota(ctx, pkg.ID.String(), pax); err != nil {
		return nil, errors.New("booking failed: not enough seats available")
	}

	// 5. Save Booking
	// Jika save booking gagal, idealnya kita rollback kuota (increase).
	// Tapi untuk MVP, urutan ini lebih aman daripada Overselling.
	if err := s.repo.CreateBooking(ctx, booking); err != nil {
		// TODO: Rollback quota here (IncreaseQuota)
		return nil, err
	}

	return booking, nil
}
