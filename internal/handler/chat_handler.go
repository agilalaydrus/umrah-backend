package handler

import (
	"context"
	"encoding/json"
	"log"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/service"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type ChatHandler struct {
	svc service.ChatService
}

func NewChatHandler(svc service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

// 1. HTTP: Get Chat History
// Di internal/handler/chat_handler.go
func (h *ChatHandler) GetHistory(c *fiber.Ctx) error {
	groupID := c.Params("group_id")

	// Ambil query param ?before_id=... (bisa kosong)
	beforeID := c.Query("before_id")

	// Pass context dan beforeID
	msgs, err := h.svc.GetHistory(c.Context(), groupID, beforeID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(msgs)
}

// 2. [BARU] HTTP: Delete Message
func (h *ChatHandler) DeleteMessage(c *fiber.Ctx) error {
	// Ambil User ID dari Token (Security: hanya boleh hapus punya sendiri)
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	// Ambil Parameter URL
	groupID := c.Params("group_id")
	messageID := c.Params("message_id")

	// Panggil Service
	err := h.svc.DeleteMessage(c.Context(), groupID, messageID, userID)
	if err != nil {
		// Asumsi error karena data tidak ketemu / bukan pemilik
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Pesan berhasil ditarik",
	})
}

// 3. WebSocket: Live Chat (Tetap Sama)
func (h *ChatHandler) StreamChat(c *websocket.Conn) {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	groupID := c.Params("group_id")

	log.Printf("Chat: User %s joined group %s", userID, groupID)

	pubsub := h.svc.GetRedisPubSub(groupID)
	defer pubsub.Close()
	defer c.Close()

	// Listener Redis (Akan menerima event pesan baru DAN pesan dihapus)
	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			var dbMsg entity.Message
			if err := json.Unmarshal([]byte(msg.Payload), &dbMsg); err != nil {
				log.Println("JSON Error:", err)
				continue
			}

			if err := c.WriteJSON(dbMsg); err != nil {
				log.Println("WS Write Error:", err)
				break
			}
		}
	}()

	// Listener Client
	for {
		var payload entity.MessagePayload
		if err := c.ReadJSON(&payload); err != nil {
			log.Println("WS Disconnected")
			break
		}
		_, err := h.svc.SendMessage(context.Background(), groupID, userID, payload.Content, string(payload.Type))
		if err != nil {
			log.Println("Chat Error:", err)
		}
	}
}
