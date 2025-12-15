package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"umrah-backend/internal/repository"

	"github.com/redis/go-redis/v9"
)

type LocationData struct {
	UserID    string  `json:"user_id"`
	FullName  string  `json:"full_name,omitempty"`
	Role      string  `json:"role,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type TrackingService interface {
	// [FIX] Add Context
	UpdateLocation(ctx context.Context, groupID string, data LocationData) error
	GetGroupLocations(ctx context.Context, groupID string) ([]LocationData, error)
}

type trackingService struct {
	redis    *redis.Client
	userRepo repository.UserRepository
}

func NewTrackingService(redis *redis.Client, userRepo repository.UserRepository) TrackingService {
	return &trackingService{
		redis:    redis,
		userRepo: userRepo,
	}
}

func (s *trackingService) UpdateLocation(ctx context.Context, groupID string, data LocationData) error {
	// [FIX] Remove context.Background(), use ctx
	key := fmt.Sprintf("group:%s:locations", groupID)

	jsonData, _ := json.Marshal(data)

	pipe := s.redis.Pipeline()
	pipe.HSet(ctx, key, data.UserID, jsonData)
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)

	return err
}

func (s *trackingService) GetGroupLocations(ctx context.Context, groupID string) ([]LocationData, error) {
	// [FIX] Use ctx
	key := fmt.Sprintf("group:%s:locations", groupID)

	result, err := s.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var locations []LocationData
	var userIDs []string

	for _, jsonStr := range result {
		var loc LocationData
		if err := json.Unmarshal([]byte(jsonStr), &loc); err == nil {
			locations = append(locations, loc)
			userIDs = append(userIDs, loc.UserID)
		}
	}

	if len(locations) == 0 {
		return locations, nil
	}

	// [FIX] Pass ctx to repo
	usersLite, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return locations, nil
	}

	userInfoMap := make(map[string]repository.UserLite)
	for _, u := range usersLite {
		userInfoMap[u.ID] = u
	}

	for i, loc := range locations {
		if user, found := userInfoMap[loc.UserID]; found {
			locations[i].FullName = user.FullName
			locations[i].Role = user.Role
		}
	}

	return locations, nil
}
