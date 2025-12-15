package handler

import (
	"umrah-backend/internal/entity"
	"umrah-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type GroupHandler struct {
	svc service.GroupService
}

func NewGroupHandler(svc service.GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

// Helper to get User from Token
func getUserFromCtx(c *fiber.Ctx) (string, string) {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["user_id"].(string), claims["role"].(string)
}

func (h *GroupHandler) Create(c *fiber.Ctx) error {
	userID, role := getUserFromCtx(c)

	var req entity.CreateGroupDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	// [FIX] Tambahkan c.Context() sebagai parameter pertama
	group, err := h.svc.CreateGroup(c.Context(), userID, role, req)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(group)
}

func (h *GroupHandler) Join(c *fiber.Ctx) error {
	userID, _ := getUserFromCtx(c)

	var req entity.JoinGroupDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	// [FIX] Tambahkan c.Context()
	group, err := h.svc.JoinGroup(c.Context(), userID, req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Joined successfully", "group": group})
}

func (h *GroupHandler) GetMembers(c *fiber.Ctx) error {
	groupID := c.Params("id")

	// [FIX] Tambahkan c.Context()
	members, err := h.svc.GetGroupMembers(c.Context(), groupID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(members)
}
