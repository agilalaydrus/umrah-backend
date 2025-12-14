package service

import (
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
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req entity.RegisterDTO) (*entity.User, error)
	Login(req entity.LoginDTO) (string, error)
	ForgotPassword(req entity.ForgotPasswordDTO) (string, error)
	ResetPassword(req entity.ResetPasswordDTO) error
}

type authService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) AuthService {
	return &authService{repo: repo}
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

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
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
