package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9" // [NEW] Import Redis
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req entity.RegisterDTO) (*entity.User, error)
	Login(req entity.LoginDTO) (string, error)
	ForgotPassword(req entity.ForgotPasswordDTO) (string, error)
	ResetPassword(req entity.ResetPasswordDTO) error
}

type authService struct {
	repo        repository.UserRepository
	redisClient *redis.Client // [NEW] Inject Redis Client
}

// [UPDATED] Constructor now requires Redis Client
func NewAuthService(repo repository.UserRepository, rc *redis.Client) AuthService {
	return &authService{
		repo:        repo,
		redisClient: rc,
	}
}

func (s *authService) Register(req entity.RegisterDTO) (*entity.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:          uuid.New(),
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Password:    string(hashed),
		Role:        req.Role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("registration failed: %v", err)
	}
	return user, nil
}

func (s *authService) Login(req entity.LoginDTO) (string, error) {
	user, err := s.repo.FindByPhone(req.PhoneNumber)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// --- [NEW] SINGLE SESSION LOGIC START ---

	// 1. Generate a unique Session ID
	sessionID := uuid.New().String()

	// 2. Save to Redis (Key: "session:user:{USER_ID}")
	// This overwrites any previous session ID, effectively logging out the old device.
	redisKey := fmt.Sprintf("session:user:%s", user.ID.String())

	// Note: using context.Background() here to match your interface signature
	err = s.redisClient.Set(context.Background(), redisKey, sessionID, 72*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}

	// 3. Create Claims (Includes "sid")
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"sid":     sessionID, // [NEW] Embed Session ID in token
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	// --- SINGLE SESSION LOGIC END ---

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (s *authService) ForgotPassword(req entity.ForgotPasswordDTO) (string, error) {
	user, err := s.repo.FindByPhone(req.PhoneNumber)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Mock OTP Generation
	otp := strconv.Itoa(rand.Intn(9000) + 1000)
	user.ResetToken = &otp

	if err := s.repo.Update(user); err != nil {
		return "", errors.New("failed to generate OTP")
	}

	return otp, nil // Return OTP for dev purposes
}

func (s *authService) ResetPassword(req entity.ResetPasswordDTO) error {
	user, err := s.repo.FindByPhone(req.PhoneNumber)
	if err != nil {
		return errors.New("user not found")
	}

	if user.ResetToken == nil || *user.ResetToken != req.OTP {
		return errors.New("invalid OTP")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	user.Password = string(hashed)
	user.ResetToken = nil // Clear token

	return s.repo.Update(user)
}
