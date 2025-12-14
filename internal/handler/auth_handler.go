package handler

import (
	"umrah-backend/internal/entity"
	"umrah-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	svc       service.AuthService
	validator *validator.Validate
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc, validator: validator.New()}
}

// Validation Helper
func (h *AuthHandler) validate(req interface{}) string {
	if err := h.validator.Struct(req); err != nil {
		return err.Error()
	}
	return ""
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req entity.RegisterDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}
	if errStr := h.validate(req); errStr != "" {
		return c.Status(400).JSON(fiber.Map{"error": errStr})
	}

	user, err := h.svc.Register(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(user)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req entity.LoginDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}
	if errStr := h.validate(req); errStr != "" {
		return c.Status(400).JSON(fiber.Map{"error": errStr})
	}

	token, err := h.svc.Login(req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"token": token})
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req entity.ForgotPasswordDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	otp, err := h.svc.ForgotPassword(req)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "OTP Sent", "debug_otp": otp})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req entity.ResetPasswordDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}
	if errStr := h.validate(req); errStr != "" {
		return c.Status(400).JSON(fiber.Map{"error": errStr})
	}

	if err := h.svc.ResetPassword(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Password updated successfully"})
}
