package repository

import (
	"context"
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type ManasikRepository interface {
	Create(ctx context.Context, data *entity.Manasik) error
	GetByCategory(ctx context.Context, category string) ([]entity.Manasik, error)
}

type manasikRepo struct {
	db *gorm.DB
}

func NewManasikRepository(db *gorm.DB) ManasikRepository {
	return &manasikRepo{db: db}
}

func (r *manasikRepo) Create(ctx context.Context, data *entity.Manasik) error {
	return r.db.WithContext(ctx).Create(data).Error
}

func (r *manasikRepo) GetByCategory(ctx context.Context, category string) ([]entity.Manasik, error) {
	var contents []entity.Manasik
	err := r.db.WithContext(ctx).
		Where("category = ?", category).
		Order("sequence_order asc").
		Find(&contents).Error
	return contents, err
}
