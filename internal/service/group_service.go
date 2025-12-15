package service

import (
	"context" // [FIX] Wajib import context
	"errors"
	"fmt"
	"time"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"

	"github.com/google/uuid"
)

type GroupService interface {
	// [FIX] Update Signature: Tambah parameter ctx
	CreateGroup(ctx context.Context, userID string, userRole string, req entity.CreateGroupDTO) (*entity.Group, error)
	JoinGroup(ctx context.Context, userID string, req entity.JoinGroupDTO) (*entity.Group, error)
	GetGroupMembers(ctx context.Context, groupID string) ([]entity.GroupMember, error)
}

type groupService struct {
	repo repository.GroupRepository
}

func NewGroupService(repo repository.GroupRepository) GroupService {
	return &groupService{repo: repo}
}

func (s *groupService) CreateGroup(ctx context.Context, userID string, userRole string, req entity.CreateGroupDTO) (*entity.Group, error) {
	if userRole != "MUTAWWIF" && userRole != "ADMIN" {
		return nil, errors.New("unauthorized: only mutawwif can create groups")
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("invalid start_date format (use YYYY-MM-DD)")
	}

	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("invalid end_date format (use YYYY-MM-DD)")
	}

	mutawwifUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	newGroupID := uuid.New()

	group := &entity.Group{
		ID:         newGroupID,
		Name:       req.Name,
		MutawwifID: mutawwifUUID,
		JoinCode:   req.JoinCode,
		StartDate:  start,
		EndDate:    end,
	}

	// [FIX] Pass ctx ke Repo Create
	if err := s.repo.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}

	creatorMember := &entity.GroupMember{
		ID:      uuid.New(),
		GroupID: newGroupID,
		UserID:  mutawwifUUID,
		Status:  "ACTIVE",
	}

	// [FIX] Ganti s.repo.AddMember menjadi s.repo.Join, dan pass ctx
	if err := s.repo.Join(ctx, creatorMember); err != nil {
		return nil, fmt.Errorf("group created but failed to add owner as member: %v", err)
	}

	return group, nil
}

func (s *groupService) JoinGroup(ctx context.Context, userID string, req entity.JoinGroupDTO) (*entity.Group, error) {
	// [FIX] Pass ctx ke FindByCode
	group, err := s.repo.FindByCode(ctx, req.JoinCode)
	if err != nil {
		return nil, errors.New("group not found or invalid code")
	}

	// [FIX] Update IsMember: Pass ctx dan handle return value (bool, error)
	isMember, err := s.repo.IsMember(ctx, group.ID.String(), userID)
	if err != nil {
		return nil, fmt.Errorf("error checking membership: %v", err)
	}
	if isMember {
		return nil, errors.New("already a member of this group")
	}

	userUUID, _ := uuid.Parse(userID)

	member := &entity.GroupMember{
		ID:      uuid.New(),
		GroupID: group.ID,
		UserID:  userUUID,
		Status:  "ACTIVE",
	}

	// [FIX] Ganti s.repo.AddMember menjadi s.repo.Join, pass ctx
	if err := s.repo.Join(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to join group: %v", err)
	}

	return group, nil
}

func (s *groupService) GetGroupMembers(ctx context.Context, groupID string) ([]entity.GroupMember, error) {
	// [FIX] Pass ctx
	return s.repo.GetMembers(ctx, groupID)
}
