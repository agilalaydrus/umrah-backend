package repository

import (
	"context"
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type CommerceRepository interface {
	// Product Management (Admin View)
	CreateProduct(ctx context.Context, product *entity.Product) error
	GetActiveProducts(ctx context.Context) ([]entity.Product, error)
	FindProductByID(ctx context.Context, id string) (*entity.Product, error)
	// UpdateProduct, DeleteProduct, etc. would go here, but omitted for now.

	// Order Management (User/Admin View)
	CreateOrder(ctx context.Context, order *entity.Order) error
	FindOrderByID(ctx context.Context, id string) (*entity.Order, error)
	GetOrdersByUser(ctx context.Context, userID string) ([]entity.Order, error)
	GetPendingOrders(ctx context.Context) ([]entity.Order, error)
	UpdateOrder(ctx context.Context, order *entity.Order) error
}

type commerceRepo struct {
	db *gorm.DB
}

func NewCommerceRepository(db *gorm.DB) CommerceRepository {
	return &commerceRepo{db: db}
}

// -----------------------------------------------------------
// Implementation: Product
// -----------------------------------------------------------

func (r *commerceRepo) CreateProduct(ctx context.Context, product *entity.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *commerceRepo) GetActiveProducts(ctx context.Context) ([]entity.Product, error) {
	var products []entity.Product
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&products).Error
	return products, err
}

func (r *commerceRepo) FindProductByID(ctx context.Context, id string) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// -----------------------------------------------------------
// Implementation: Order
// -----------------------------------------------------------

func (r *commerceRepo) CreateOrder(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *commerceRepo) FindOrderByID(ctx context.Context, id string) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("User").
		First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *commerceRepo) GetOrdersByUser(ctx context.Context, userID string) ([]entity.Order, error) {
	var orders []entity.Order
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&orders).Error
	return orders, err
}

func (r *commerceRepo) GetPendingOrders(ctx context.Context) ([]entity.Order, error) {
	var orders []entity.Order
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("User").
		Where("status = ?", entity.OrderPaid).
		Order("created_at asc").
		Find(&orders).Error
	return orders, err
}

func (r *commerceRepo) UpdateOrder(ctx context.Context, order *entity.Order) error {
	// Saves all fields, including status and proof image URL
	return r.db.WithContext(ctx).Save(order).Error
}
