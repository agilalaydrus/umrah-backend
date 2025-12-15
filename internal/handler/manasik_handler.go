package handler

import (
	"umrah-backend/internal/entity"
	"umrah-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ManasikHandler struct {
	svc service.ManasikService
}

func NewManasikHandler(svc service.ManasikService) *ManasikHandler {
	return &ManasikHandler{svc: svc}
}

// POST /manasik (Admin)
func (h *ManasikHandler) Create(c *fiber.Ctx) error {
	var req entity.Manasik
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	if err := h.svc.AddContent(c.Context(), req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "Content added"})
}

// GET /manasik?category=UMRAH
func (h *ManasikHandler) GetList(c *fiber.Ctx) error {
	cat := c.Query("category")
	data, err := h.svc.GetGuide(c.Context(), cat)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}
