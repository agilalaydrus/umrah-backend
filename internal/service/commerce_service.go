package service

import (
	"context"
	"errors"
	"time"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"

	"github.com/google/uuid"
)

type CommerceService interface {
	// Product
	CreateProduct(ctx context.Context, req entity.Product) error
	GetCatalog(ctx context.Context) ([]entity.Product, error)

	// Order Flow
	CreateOrder(ctx context.Context, userID, productID string) (*entity.Order, error)
	UploadPaymentProof(ctx context.Context, orderID, imageURL string, userID string) error
	VerifyOrder(ctx context.Context, orderID string) error // Admin only
	GetMyOrders(ctx context.Context, userID string) ([]entity.Order, error)
	GetPendingOrders(ctx context.Context) ([]entity.Order, error)
}

type commerceService struct {
	repo repository.CommerceRepository
}

func NewCommerceService(repo repository.CommerceRepository) CommerceService {
	return &commerceService{repo: repo}
}

func (s *commerceService) CreateProduct(ctx context.Context, req entity.Product) error {
	req.ID = uuid.New()
	req.CreatedAt = time.Now()
	return s.repo.CreateProduct(ctx, &req)
}

func (s *commerceService) GetCatalog(ctx context.Context) ([]entity.Product, error) {
	return s.repo.GetActiveProducts(ctx)
}

func (s *commerceService) CreateOrder(ctx context.Context, userID, productID string) (*entity.Order, error) {
	// 1. Fetch Product to get REAL Price (Security)
	product, err := s.repo.FindProductByID(ctx, productID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if !product.IsActive {
		return nil, errors.New("product is not available")
	}

	// 2. Create Order
	order := &entity.Order{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		ProductID: product.ID,
		Amount:    product.Price, // Snapshot price
		Status:    entity.OrderPending,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *commerceService) UploadPaymentProof(ctx context.Context, orderID, imageURL string, userID string) error {
	// 1. Find Order
	order, err := s.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		return errors.New("order not found")
	}

	// 2. Verify Ownership
	if order.UserID.String() != userID {
		return errors.New("unauthorized")
	}

	// 3. Update
	order.ProofImage = imageURL
	order.Status = entity.OrderPaid // Auto-move to PAID waiting verification
	order.UpdatedAt = time.Now()

	return s.repo.UpdateOrder(ctx, order)
}

func (s *commerceService) VerifyOrder(ctx context.Context, orderID string) error {
	// Admin Action
	order, err := s.repo.FindOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	order.Status = entity.OrderCompleted
	order.UpdatedAt = time.Now()

	return s.repo.UpdateOrder(ctx, order)
}

func (s *commerceService) GetMyOrders(ctx context.Context, userID string) ([]entity.Order, error) {
	return s.repo.GetOrdersByUser(ctx, userID)
}

func (s *commerceService) GetPendingOrders(ctx context.Context) ([]entity.Order, error) {
	return s.repo.GetPendingOrders(ctx)
}
