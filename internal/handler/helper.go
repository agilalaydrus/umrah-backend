package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Helper: Ambil UserID dari JWT Context
func getUserID(c *fiber.Ctx) (string, error) {
	user := c.Locals("user")
	if user == nil {
		return "", errors.New("unauthorized: missing token")
	}

	token, ok := user.(*jwt.Token)
	if !ok {
		return "", errors.New("invalid token format")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	id, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("user_id claim missing or invalid")
	}

	return id, nil
}

// Helper: Ambil Role dari JWT Context
func getUserRole(c *fiber.Ctx) (string, error) {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	role, ok := claims["role"].(string)
	if !ok {
		return "", errors.New("role claim missing")
	}
	return role, nil
}
