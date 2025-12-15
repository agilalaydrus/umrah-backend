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
	req.Available = req.Quota // Initial available = quota
	return s.repo.CreatePackage(ctx, &req)
}

func (s *packageService) GetList(ctx context.Context, category string) ([]entity.TravelPackage, error) {
	return s.repo.GetPackages(ctx, category)
}

func (s *packageService) BookPackage(ctx context.Context, userID, packageID, roomType string, pax int) (*entity.Booking, error) {
	// 1. Get Package
	pkg, err := s.repo.FindPackageByID(ctx, packageID)
	if err != nil {
		return nil, errors.New("package not found")
	}

	// 2. Check Quota
	if pkg.Available < pax {
		return nil, errors.New("not enough seats available")
	}

	// 3. Calculate Price based on Room Type
	var pricePerPax float64
	switch roomType {
	case "QUAD":
		pricePerPax = pkg.PriceQuad
	case "TRIPLE":
		pricePerPax = pkg.PriceTriple
	case "DOUBLE":
		pricePerPax = pkg.PriceDouble
	default:
		return nil, errors.New("invalid room type (QUAD/TRIPLE/DOUBLE)")
	}

	totalPrice := pricePerPax * float64(pax)

	// 4. Create Booking
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

	if err := s.repo.CreateBooking(ctx, booking); err != nil {
		return nil, err
	}

	// 5. Deduct Quota (Simple logic, add mutex/transaction for high concurrency)
	pkg.Available -= pax
	_ = s.repo.UpdatePackageQuota(ctx, pkg)

	return booking, nil
}
