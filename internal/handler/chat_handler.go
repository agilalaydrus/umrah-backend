package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository" // [NEW] Needed for security check
	"umrah-backend/internal/service"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type ChatHandler struct {
	svc       service.ChatService
	groupRepo repository.GroupRepository // [NEW] Inject GroupRepo
}

// Update Factory to accept GroupRepo
func NewChatHandler(svc service.ChatService, gr repository.GroupRepository) *ChatHandler {
	return &ChatHandler{svc: svc, groupRepo: gr}
}

func (h *ChatHandler) GetHistory(c *fiber.Ctx) error {
	groupID := c.Params("group_id")
	beforeID := c.Query("before_id")

	msgs, err := h.svc.GetHistory(c.Context(), groupID, beforeID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(msgs)
}

func (h *ChatHandler) DeleteMessage(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	groupID := c.Params("group_id")
	messageID := c.Params("message_id")

	err := h.svc.DeleteMessage(c.Context(), groupID, messageID, userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Pesan berhasil ditarik"})
}

func (h *ChatHandler) StreamChat(c *websocket.Conn) {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	groupID := c.Params("group_id")

	// --- [CRITICAL SECURITY FIX] ---
	// Prevent unauthorized users from joining random groups
	if !h.groupRepo.IsMember(groupID, userID) {
		log.Printf("Security Alert: User %s tried to access Group %s", userID, groupID)
		c.WriteMessage(websocket.CloseMessage, []byte("Unauthorized"))
		c.Close()
		return
	}
	// -------------------------------

	log.Printf("Chat: User %s joined group %s", userID, groupID)

	pubsub := h.svc.GetRedisPubSub(groupID)
	defer pubsub.Close()
	defer c.Close()

	// --- [CRITICAL STABILITY FIX] ---
	// Heartbeat: Kill zombie connections if they don't respond to Ping
	c.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if err := c.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}()
	// --------------------------------

	// Redis Listener Goroutine
	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			var dbMsg entity.Message
			if err := json.Unmarshal([]byte(msg.Payload), &dbMsg); err != nil {
				continue
			}
			if err := c.WriteJSON(dbMsg); err != nil {
				break
			}
		}
	}()

	// Client Listener Loop
	for {
		var payload entity.MessagePayload
		if err := c.ReadJSON(&payload); err != nil {
			log.Println("WS Disconnected:", userID)
			break
		}
		_, err := h.svc.SendMessage(context.Background(), groupID, userID, payload.Content, string(payload.Type))
		if err != nil {
			log.Println("Chat Error:", err)
		}
	}
}
