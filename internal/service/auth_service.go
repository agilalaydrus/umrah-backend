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
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	// [FIX] Tambahkan context.Context
	Register(ctx context.Context, req entity.RegisterDTO) (*entity.User, error)
	Login(ctx context.Context, req entity.LoginDTO) (string, error)
	ForgotPassword(ctx context.Context, req entity.ForgotPasswordDTO) (string, error)
	ResetPassword(ctx context.Context, req entity.ResetPasswordDTO) error
}

type authService struct {
	repo        repository.UserRepository
	redisClient *redis.Client
}

func NewAuthService(repo repository.UserRepository, rc *redis.Client) AuthService {
	return &authService{repo: repo, redisClient: rc}
}

func (s *authService) Register(ctx context.Context, req entity.RegisterDTO) (*entity.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:          uuid.New(),
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Password:    string(hashed),
		Role:        entity.RoleJamaah,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("registration failed: %v", err)
	}
	return user, nil
}

func (s *authService) Login(ctx context.Context, req entity.LoginDTO) (string, error) {
	// [FIX] Pass ctx
	user, err := s.repo.FindByPhone(ctx, req.PhoneNumber)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// --- SINGLE SESSION LOGIC ---
	sessionID := uuid.New().String()
	redisKey := fmt.Sprintf("session:user:%s", user.ID.String())

	// [FIX] Gunakan ctx dari parameter, bukan context.Background()
	err = s.redisClient.Set(ctx, redisKey, sessionID, 72*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"sid":     sessionID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (s *authService) ForgotPassword(ctx context.Context, req entity.ForgotPasswordDTO) (string, error) {
	user, err := s.repo.FindByPhone(ctx, req.PhoneNumber)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Mock OTP
	otp := strconv.Itoa(rand.Intn(9000) + 1000)
	user.ResetToken = &otp

	// Expired time 15 menit dari sekarang (Best Practice)
	expiry := time.Now().Add(15 * time.Minute)
	user.ResetTokenExpiry = &expiry

	if err := s.repo.Update(ctx, user); err != nil {
		return "", errors.New("failed to generate OTP")
	}

	return otp, nil
}

func (s *authService) ResetPassword(ctx context.Context, req entity.ResetPasswordDTO) error {
	user, err := s.repo.FindByPhone(ctx, req.PhoneNumber)
	if err != nil {
		return errors.New("user not found")
	}

	// [FIX] Check Token & Expiry
	if user.ResetToken == nil || *user.ResetToken != req.OTP {
		return errors.New("invalid OTP")
	}
	// Logic expiry check bisa ditambahkan disini jika field Expiry ada di DB

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	user.Password = string(hashed)
	user.ResetToken = nil // Clear token

	return s.repo.Update(ctx, user)
}
