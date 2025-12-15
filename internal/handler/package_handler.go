package handler

import (
	"umrah-backend/internal/entity"
	"umrah-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type PackageHandler struct {
	svc service.PackageService
}

func NewPackageHandler(svc service.PackageService) *PackageHandler {
	return &PackageHandler{svc: svc}
}

// POST /packages (Admin)
func (h *PackageHandler) Create(c *fiber.Ctx) error {
	var req entity.TravelPackage
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	if err := h.svc.CreatePackage(c.Context(), req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "Package created"})
}

// GET /packages?category=UMRAH_PLUS
func (h *PackageHandler) GetList(c *fiber.Ctx) error {
	cat := c.Query("category") // Optional filter
	data, err := h.svc.GetList(c.Context(), cat)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

// POST /bookings
func (h *PackageHandler) Book(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	var req struct {
		PackageID string `json:"package_id"`
		RoomType  string `json:"room_type"` // QUAD, TRIPLE, DOUBLE
		PaxCount  int    `json:"pax_count"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	booking, err := h.svc.BookPackage(c.Context(), userID, req.PackageID, req.RoomType, req.PaxCount)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(booking)
}
