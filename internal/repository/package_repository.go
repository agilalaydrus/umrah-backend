package repository

import (
	"context"
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type PackageRepository interface {
	CreatePackage(ctx context.Context, pkg *entity.TravelPackage) error
	GetPackages(ctx context.Context, category string) ([]entity.TravelPackage, error)
	FindPackageByID(ctx context.Context, id string) (*entity.TravelPackage, error)

	CreateBooking(ctx context.Context, booking *entity.Booking) error
	UpdatePackageQuota(ctx context.Context, pkg *entity.TravelPackage) error
}

type packageRepo struct {
	db *gorm.DB
}

func NewPackageRepository(db *gorm.DB) PackageRepository {
	return &packageRepo{db: db}
}

func (r *packageRepo) CreatePackage(ctx context.Context, pkg *entity.TravelPackage) error {
	return r.db.WithContext(ctx).Create(pkg).Error
}

func (r *packageRepo) GetPackages(ctx context.Context, category string) ([]entity.TravelPackage, error) {
	var packages []entity.TravelPackage
	query := r.db.WithContext(ctx).Where("is_active = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Order("departure_date asc").Find(&packages).Error
	return packages, err
}

func (r *packageRepo) FindPackageByID(ctx context.Context, id string) (*entity.TravelPackage, error) {
	var pkg entity.TravelPackage
	err := r.db.WithContext(ctx).First(&pkg, "id = ?", id).Error
	return &pkg, err
}

func (r *packageRepo) CreateBooking(ctx context.Context, booking *entity.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

func (r *packageRepo) UpdatePackageQuota(ctx context.Context, pkg *entity.TravelPackage) error {
	return r.db.WithContext(ctx).Save(pkg).Error
}
