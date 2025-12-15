package repository

import (
	"context"
	"errors" // [FIX] Wajib import errors
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type PackageRepository interface {
	CreatePackage(ctx context.Context, pkg *entity.TravelPackage) error
	GetPackages(ctx context.Context, category string) ([]entity.TravelPackage, error)
	FindPackageByID(ctx context.Context, id string) (*entity.TravelPackage, error)

	CreateBooking(ctx context.Context, booking *entity.Booking) error
	// Method ini ada di interface, jadi WAJIB diimplementasikan di bawah
	DecreaseQuota(ctx context.Context, packageID string, count int) error
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

// [FIX] INI IMPLEMENTASI YANG HILANG
func (r *packageRepo) DecreaseQuota(ctx context.Context, packageID string, count int) error {
	// Menggunakan gorm.Expr untuk Atomic Update (Thread Safe)
	// Query: UPDATE travel_packages SET available = available - count WHERE id = ? AND available >= count
	result := r.db.WithContext(ctx).
		Model(&entity.TravelPackage{}).
		Where("id = ? AND available >= ?", packageID, count).
		Update("available", gorm.Expr("available - ?", count))

	if result.Error != nil {
		return result.Error
	}

	// Jika RowsAffected == 0, artinya ID salah ATAU kuota tidak cukup (available < count)
	if result.RowsAffected == 0 {
		return errors.New("quota insufficient or package not found")
	}

	return nil
}
