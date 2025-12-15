package repository

import (
	"context"
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type UserLite struct {
	ID       string
	FullName string
	Role     string
}

type UserRepository interface {
	// [FIX] Tambahkan context.Context di parameter pertama semua method
	Create(ctx context.Context, user *entity.User) error
	FindByPhone(ctx context.Context, phone string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	FindByIDs(ctx context.Context, ids []string) ([]UserLite, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) FindByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var user entity.User
	// Gunakan WithContext(ctx)
	if err := r.db.WithContext(ctx).Where("phone_number = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepo) FindByIDs(ctx context.Context, ids []string) ([]UserLite, error) {
	var users []entity.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}

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
