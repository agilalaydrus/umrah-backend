package worker

import (
	"context"
	"encoding/json"
	"log"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"
	"umrah-backend/pkg/notification" // [NEW]
	"umrah-backend/pkg/queue"

	"github.com/google/uuid"
)

type ChatWorker struct {
	rabbit    *queue.RabbitMQ
	chatRepo  repository.ChatRepository
	groupRepo repository.GroupRepository // [NEW] Needed to find members
	fcm       *notification.FCMService   // [NEW]
}

// Updated Constructor
func NewChatWorker(
	r *queue.RabbitMQ,
	cRepo repository.ChatRepository,
	gRepo repository.GroupRepository,
	fcm *notification.FCMService,
) *ChatWorker {
	return &ChatWorker{rabbit: r, chatRepo: cRepo, groupRepo: gRepo, fcm: fcm}
}

func (w *ChatWorker) Start() {
	msgs, err := w.rabbit.Channel.Consume(
		w.rabbit.Queue.Name, "", false, false, false, false, nil,
	)
	if err != nil {
		log.Fatal("Failed to register consumer:", err)
	}

	go func() {
		log.Println("👷 Chat Worker Started")
		for d := range msgs {
			var msg entity.Message
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Error decoding: %v", err)
				d.Nack(false, false)
				continue
			}

			// 1. Save to DB
			// Use context.Background() instead of nil
			if err := w.chatRepo.CreateMessage(context.Background(), &msg); err != nil {
				log.Printf("DB Error: %v", err)
				d.Nack(false, true)
				continue
			}

			// 2. [NEW] Send Push Notification
			go w.sendNotification(&msg)

			d.Ack(false)
		}
	}()
}

func (w *ChatWorker) sendNotification(msg *entity.Message) {
	// [FIX 1] Add context.Background() as the first argument
	members, err := w.groupRepo.GetMembers(context.Background(), msg.GroupID.String())
	if err != nil {
		log.Printf("Failed to fetch members: %v", err)
		return
	}

	var tokens []string
	for _, member := range members {
		// [FIX] Replace 'member.User != nil' with 'member.User.ID != uuid.Nil'
		// We check if the ID is valid to ensure the user data was loaded.
		if member.User.ID != uuid.Nil && member.User.ID != msg.SenderID && member.User.FCMToken != "" {
			tokens = append(tokens, member.User.FCMToken)
		}
	}

	if len(tokens) > 0 {
		w.fcm.SendPush(tokens, "New Message", msg.Content, map[string]string{
			"type":     "CHAT",
			"group_id": msg.GroupID.String(),
		})
	}
}
