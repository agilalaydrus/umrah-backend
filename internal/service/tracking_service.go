package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"umrah-backend/internal/repository"

	"github.com/redis/go-redis/v9"
)

// Struct response untuk Frontend
type LocationData struct {
	UserID    string  `json:"user_id"`
	FullName  string  `json:"full_name,omitempty"`
	Role      string  `json:"role,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type TrackingService interface {
	UpdateLocation(groupID string, data LocationData) error
	GetGroupLocations(groupID string) ([]LocationData, error)
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

func (s *trackingService) UpdateLocation(groupID string, data LocationData) error {
	ctx := context.Background()
	key := fmt.Sprintf("group:%s:locations", groupID)

	// Simpan data lokasi ke Redis Hash
	jsonData, _ := json.Marshal(data)

	pipe := s.redis.Pipeline()
	pipe.HSet(ctx, key, data.UserID, jsonData)
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)

	return err
}

func (s *trackingService) GetGroupLocations(groupID string) ([]LocationData, error) {
	ctx := context.Background()
	key := fmt.Sprintf("group:%s:locations", groupID)

	// 1. Ambil data mentah dari Redis
	result, err := s.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var locations []LocationData
	var userIDs []string

	// 2. Parse JSON dari Redis & Kumpulkan User ID
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

	// 3. Ambil Nama User dari Postgres (Pakai struct UserLite)
	usersLite, err := s.userRepo.FindByIDs(userIDs)
	if err != nil {
		return locations, nil // Return lokasi saja jika DB error
	}

	// 4. Buat Map untuk pencarian cepat (ID -> UserLite)
	// Ini menggantikan variabel 'userMap' yang tadi error
	userInfoMap := make(map[string]repository.UserLite)
	for _, u := range usersLite {
		userInfoMap[u.ID] = u
	}

	// 5. Gabungkan Data (Lokasi + Nama + Role)
	for i, loc := range locations {
		if user, found := userInfoMap[loc.UserID]; found {
			locations[i].FullName = user.FullName
			locations[i].Role = user.Role
		}
	}

	return locations, nil
}
