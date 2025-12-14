package repository

import (
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

// 1. Struct UserLite (Untuk Tracking Service)
type UserLite struct {
	ID       string
	FullName string
	Role     string
}

// 2. Interface UserRepository (LENGKAP)
type UserRepository interface {
	Create(user *entity.User) error
	FindByPhone(phone string) (*entity.User, error)

	// Method ini WAJIB ADA karena dipakai AuthService (Forgot/Reset Password)
	Update(user *entity.User) error

	// Method ini WAJIB ADA karena dipakai TrackingService
	FindByIDs(ids []string) ([]UserLite, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// --- Implementasi Method ---

func (r *userRepo) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *userRepo) FindByPhone(phone string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Where("phone_number = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// [PENTING] Ini dikembalikan agar AuthService tidak error
func (r *userRepo) Update(user *entity.User) error {
	return r.db.Save(user).Error
}

// [PENTING] Ini tetap ada untuk TrackingService
func (r *userRepo) FindByIDs(ids []string) ([]UserLite, error) {
	var users []entity.User

	// Query database: SELECT * FROM users WHERE id IN (...)
	if err := r.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}

	// Mapping ke UserLite
	var result []UserLite
	for _, u := range users {
		result = append(result, UserLite{
			ID:       u.ID.String(),
			FullName: u.FullName,
			Role:     u.Role,
		})
	}
	return result, nil
}
