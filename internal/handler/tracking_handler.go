package handler

import (
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

// Struct untuk membaca data JSON dari HP Jamaah
type LocationPayload struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

// WEBSOCKET: Stream Location (Write)
func (h *TrackingHandler) StreamLocation(c *websocket.Conn) {
	// 1. Ambil User info dari Token JWT
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	// 2. Ambil Group ID dari URL
	groupID := c.Params("group_id")

	log.Printf("Start Tracking: User %s in Group %s", userID, groupID)

	defer c.Close()

	for {
		var payload LocationPayload

		// 3. Baca JSON dari Client (HP)
		if err := c.ReadJSON(&payload); err != nil {
			log.Println("WS Disconnected:", userID)
			break
		}

		// 4. Siapkan Data untuk Service [FIX: Bungkus jadi struct]
		locationData := service.LocationData{
			UserID:    userID,
			Latitude:  payload.Lat,
			Longitude: payload.Long,
			Timestamp: time.Now().UnixMilli(), // Tambahkan waktu server
		}

		// 5. Panggil Service UpdateLocation [FIX: Hapus context, kirim struct]
		err := h.svc.UpdateLocation(groupID, locationData)
		if err != nil {
			log.Println("Redis Save Error:", err)
		}
	}
}

// HTTP: Get Map Markers (Read)
func (h *TrackingHandler) GetLocations(c *fiber.Ctx) error {
	groupID := c.Params("group_id")

	// Panggil Service GetGroupLocations [FIX: Hapus c.Context()]
	locations, err := h.svc.GetGroupLocations(groupID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(locations)
}
