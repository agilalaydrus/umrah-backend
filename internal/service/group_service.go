package service

import (
	"errors"
	"fmt"
	"time"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"

	"github.com/google/uuid"
)

type GroupService interface {
	CreateGroup(userID string, userRole string, req entity.CreateGroupDTO) (*entity.Group, error)
	JoinGroup(userID string, req entity.JoinGroupDTO) (*entity.Group, error)
	GetGroupMembers(groupID string) ([]entity.GroupMember, error)
}

type groupService struct {
	repo repository.GroupRepository
}

func NewGroupService(repo repository.GroupRepository) GroupService {
	return &groupService{repo: repo}
}

func (s *groupService) CreateGroup(userID string, userRole string, req entity.CreateGroupDTO) (*entity.Group, error) {
	// 1. Authorization: Only Mutawwif/Admin
	if userRole != "MUTAWWIF" && userRole != "ADMIN" {
		return nil, errors.New("unauthorized: only mutawwif can create groups")
	}

	// 2. Parse Dates
	// Menambahkan error handling agar tidak panic jika format tanggal salah
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

	// 3. Setup Group Object
	// PENTING: Kita generate ID di sini (uuid.New()) agar tidak kosong saat masuk DB
	newGroupID := uuid.New()

	group := &entity.Group{
		ID:         newGroupID, // <--- FIX: Generate UUID baru
		Name:       req.Name,
		MutawwifID: mutawwifUUID,
		JoinCode:   req.JoinCode,
		StartDate:  start,
		EndDate:    end,
	}

	// 4. Save Group to Database
	if err := s.repo.Create(group); err != nil {
		// Return error asli supaya kita tahu jika JoinCode sudah terpakai
		return nil, fmt.Errorf("database error: %v", err)
	}

	// 5. Automatically Add Creator (Mutawwif) as a Member
	creatorMember := &entity.GroupMember{
		ID:      uuid.New(), // <--- FIX: Generate UUID untuk Member juga
		GroupID: newGroupID,
		UserID:  mutawwifUUID,
		Status:  "ACTIVE",
	}

	// Simpan Mutawwif sebagai member
	if err := s.repo.AddMember(creatorMember); err != nil {
		return nil, fmt.Errorf("group created but failed to add owner as member: %v", err)
	}

	return group, nil
}

func (s *groupService) JoinGroup(userID string, req entity.JoinGroupDTO) (*entity.Group, error) {
	// 1. Find Group
	group, err := s.repo.FindByCode(req.JoinCode)
	if err != nil {
		return nil, errors.New("group not found or invalid code")
	}

	// 2. Check if already joined
	if s.repo.IsMember(group.ID.String(), userID) {
		return nil, errors.New("already a member of this group")
	}

	// 3. Add to Group
	userUUID, _ := uuid.Parse(userID)

	member := &entity.GroupMember{
		ID:      uuid.New(), // <--- FIX: Generate UUID member baru
		GroupID: group.ID,
		UserID:  userUUID,
		Status:  "ACTIVE",
	}

	if err := s.repo.AddMember(member); err != nil {
		return nil, fmt.Errorf("failed to join group: %v", err)
	}

	return group, nil
}

func (s *groupService) GetGroupMembers(groupID string) ([]entity.GroupMember, error) {
	return s.repo.GetMembers(groupID)
}
