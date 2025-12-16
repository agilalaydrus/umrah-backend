package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time" // [NEW] Needed for setting CreatedAt
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"
	"umrah-backend/pkg/queue"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ChatService interface {
	SendMessage(ctx context.Context, groupID, senderID, content, msgType string) (*entity.Message, error)
	GetHistory(ctx context.Context, groupID string, beforeID string) ([]entity.Message, error)
	GetRedisPubSub(groupID string) *redis.PubSub
	DeleteMessage(ctx context.Context, groupID, messageID, userID string) error
}

type chatService struct {
	repo        repository.ChatRepository
	redisClient *redis.Client
	rabbit      *queue.RabbitMQ
}

// [UPDATED] Accept RabbitMQ in constructor
func NewChatService(r repository.ChatRepository, rc *redis.Client, rabbit *queue.RabbitMQ) ChatService {
	return &chatService{
		repo:        r,
		redisClient: rc,
		rabbit:      rabbit,
	}
}

func (s *chatService) SendMessage(ctx context.Context, groupID, senderID, content, msgType string) (*entity.Message, error) {
	gUUID, _ := uuid.Parse(groupID)
	sUUID, _ := uuid.Parse(senderID)

	// [UPDATED] Prepare Message with ID and Time explicitly
	// We need to set these now because we are sending it to a queue,
	// not letting the DB generate them immediately.
	msg := &entity.Message{
		ID:        uuid.New(),
		GroupID:   gUUID,
		SenderID:  sUUID,
		Content:   content,
		Type:      entity.MessageType(msgType),
		CreatedAt: time.Now(),
	}

	// 1. [NEW] Publish to RabbitMQ (Reliability)
	// The Worker will pick this up and save it to the Database.
	// We replaced s.repo.CreateMessage(ctx, msg) with this.
	if err := s.rabbit.Publish(ctx, msg); err != nil {
		return nil, err
	}

	// 2. Publish to Redis (Real-time)
	// We keep this to ensure the frontend updates immediately (Optimistic UI),
	// even while the Worker is processing the database save in the background.
	msgJSON, _ := json.Marshal(msg)
	channel := fmt.Sprintf("chat:group:%s", groupID)
	err := s.redisClient.Publish(ctx, channel, msgJSON).Err()

	return msg, err
}

func (s *chatService) GetHistory(ctx context.Context, groupID string, beforeID string) ([]entity.Message, error) {
	return s.repo.GetMessageHistory(ctx, groupID, 50, beforeID)
}

func (s *chatService) GetRedisPubSub(groupID string) *redis.PubSub {
	channel := fmt.Sprintf("chat:group:%s", groupID)
	return s.redisClient.Subscribe(context.Background(), channel)
}

func (s *chatService) DeleteMessage(ctx context.Context, groupID, messageID, userID string) error {
	originalMsg, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return err
	}

	if originalMsg.SenderID.String() != userID {
		return errors.New("anda bukan pemilik pesan ini")
	}

	if err := s.repo.DeleteMessage(ctx, messageID, userID); err != nil {
		return err
	}

	originalMsg.Content = "🚫 Pesan ini telah dihapus"
	originalMsg.Type = entity.MsgDeleted

	msgJSON, _ := json.Marshal(originalMsg)
	channel := fmt.Sprintf("chat:group:%s", groupID)

	return s.redisClient.Publish(ctx, channel, msgJSON).Err()
}
