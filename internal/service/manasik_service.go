package service

import (
	"context"
	"time"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"

	"github.com/google/uuid"
)

type ManasikService interface {
	AddContent(ctx context.Context, req entity.Manasik) error
	GetGuide(ctx context.Context, category string) ([]entity.Manasik, error)
}

type manasikService struct {
	repo repository.ManasikRepository
}

func NewManasikService(repo repository.ManasikRepository) ManasikService {
	return &manasikService{repo: repo}
}

func (s *manasikService) AddContent(ctx context.Context, req entity.Manasik) error {
	req.ID = uuid.New()
	req.CreatedAt = time.Now()
	return s.repo.Create(ctx, &req)
}

func (s *manasikService) GetGuide(ctx context.Context, category string) ([]entity.Manasik, error) {
	return s.repo.GetByCategory(ctx, category)
}
