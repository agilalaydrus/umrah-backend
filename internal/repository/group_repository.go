package repository

import (
	"context"
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type GroupRepository interface {
	Create(ctx context.Context, group *entity.Group) error
	Join(ctx context.Context, member *entity.GroupMember) error
	FindByCode(ctx context.Context, code string) (*entity.Group, error)
	GetByID(ctx context.Context, id string) (*entity.Group, error)
	GetMembers(ctx context.Context, groupID string) ([]entity.GroupMember, error)
	GetAllGroups(ctx context.Context) ([]entity.Group, error)
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
}

type groupRepo struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepo{db: db}
}

func (r *groupRepo) Create(ctx context.Context, group *entity.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// [FIX] Implementasi FindByCode (Sesuai Interface)
func (r *groupRepo) FindByCode(ctx context.Context, code string) (*entity.Group, error) {
	var group entity.Group
	if err := r.db.WithContext(ctx).Where("join_code = ?", code).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// [FIX] Ganti nama dari AddMember jadi Join (Sesuai Interface)
func (r *groupRepo) Join(ctx context.Context, member *entity.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// [FIX] Implementasi GetByID (Yang sebelumnya hilang)
func (r *groupRepo) GetByID(ctx context.Context, id string) (*entity.Group, error) {
	var group entity.Group
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error
	return &group, err
}

func (r *groupRepo) GetMembers(ctx context.Context, groupID string) ([]entity.GroupMember, error) {
	var members []entity.GroupMember
	// Preload 'User' agar kita dapat nama membernya
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("group_id = ?", groupID).
		Find(&members).Error
	return members, err
}

func (r *groupRepo) GetAllGroups(ctx context.Context) ([]entity.Group, error) {
	var groups []entity.Group

	// Preload data Mutawwif untuk admin dashboard
	err := r.db.WithContext(ctx).
		Preload("Mutawwif").
		Find(&groups).Error

	return groups, err
}

func (r *groupRepo) IsMember(ctx context.Context, groupID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error

	// Jika count > 0 berarti dia member
	return count > 0, err
}
