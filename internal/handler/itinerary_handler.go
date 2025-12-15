package handler

import (
	"umrah-backend/internal/entity"
	"umrah-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type ItineraryHandler struct {
	svc service.ItineraryService
}

func NewItineraryHandler(svc service.ItineraryService) *ItineraryHandler {
	return &ItineraryHandler{svc: svc}
}

// POST /itineraries (Admin/Mutawwif)
func (h *ItineraryHandler) Create(c *fiber.Ctx) error {
	var req entity.Itinerary
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Basic Role Check
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	role := claims["role"].(string)

	if role == entity.RoleJamaah {
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if err := h.svc.CreateItinerary(c.Context(), req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Itinerary created"})
}

// GET /groups/:group_id/rundown (All)
func (h *ItineraryHandler) GetRundown(c *fiber.Ctx) error {
	groupID := c.Params("group_id")
	data, err := h.svc.GetRundown(c.Context(), groupID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

// POST /attendance/scan (Jamaah scans QR)
func (h *ItineraryHandler) Scan(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	var req struct {
		ItineraryID string `json:"itinerary_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid QR Data"})
	}

	if err := h.svc.ScanAttendance(c.Context(), userID, req.ItineraryID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Attendance recorded successfully"})
}

// GET /itineraries/:id/attendance (Mutawwif Monitor)
func (h *ItineraryHandler) GetReport(c *fiber.Ctx) error {
	id := c.Params("id")
	data, err := h.svc.GetAttendanceReport(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}
