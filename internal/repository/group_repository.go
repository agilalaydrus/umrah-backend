package repository

import (
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type GroupRepository interface {
	Create(group *entity.Group) error
	FindByCode(code string) (*entity.Group, error)
	AddMember(member *entity.GroupMember) error
	IsMember(groupID, userID string) bool
	GetMembers(groupID string) ([]entity.GroupMember, error)
}

type groupRepo struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepo{db: db}
}

func (r *groupRepo) Create(group *entity.Group) error {
	return r.db.Create(group).Error
}

func (r *groupRepo) FindByCode(code string) (*entity.Group, error) {
	var group entity.Group
	if err := r.db.Where("join_code = ?", code).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepo) AddMember(member *entity.GroupMember) error {
	return r.db.Create(member).Error
}

func (r *groupRepo) IsMember(groupID, userID string) bool {
	var count int64
	r.db.Model(&entity.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count)
	return count > 0
}

func (r *groupRepo) GetMembers(groupID string) ([]entity.GroupMember, error) {
	var members []entity.GroupMember
	// Preload 'User' so we get the names of the members, not just IDs
	err := r.db.Preload("User").Where("group_id = ?", groupID).Find(&members).Error
	return members, err
}
