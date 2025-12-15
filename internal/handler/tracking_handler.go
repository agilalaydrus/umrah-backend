package handler

import (
	"context" // [FIX] Import context
	"log"
	"time"
	"umrah-backend/internal/service"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type TrackingHandler struct {
	svc service.TrackingService
}

func NewTrackingHandler(svc service.TrackingService) *TrackingHandler {
	return &TrackingHandler{svc: svc}
}

type LocationPayload struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

func (h *TrackingHandler) StreamLocation(c *websocket.Conn) {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	groupID := c.Params("group_id")

	log.Printf("Start Tracking: User %s in Group %s", userID, groupID)

	defer c.Close()

	for {
		var payload LocationPayload
		if err := c.ReadJSON(&payload); err != nil {
			log.Println("WS Disconnected:", userID)
			break
		}

		locationData := service.LocationData{
			UserID:    userID,
			Latitude:  payload.Lat,
			Longitude: payload.Long,
			Timestamp: time.Now().UnixMilli(),
		}

		// [FIX] Use context.Background() because this is a WebSocket loop
		err := h.svc.UpdateLocation(context.Background(), groupID, locationData)
		if err != nil {
			log.Println("Redis Save Error:", err)
		}
	}
}

func (h *TrackingHandler) GetLocations(c *fiber.Ctx) error {
	groupID := c.Params("group_id")

	// [FIX] Use c.Context() here
	locations, err := h.svc.GetGroupLocations(c.Context(), groupID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(locations)
}
