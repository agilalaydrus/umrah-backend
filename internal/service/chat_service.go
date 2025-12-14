package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/repository"

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
}

func NewChatService(r repository.ChatRepository, rc *redis.Client) ChatService {
	return &chatService{repo: r, redisClient: rc}
}

func (s *chatService) SendMessage(ctx context.Context, groupID, senderID, content, msgType string) (*entity.Message, error) {
	gUUID, _ := uuid.Parse(groupID)
	sUUID, _ := uuid.Parse(senderID)

	msg := &entity.Message{
		GroupID:  gUUID,
		SenderID: sUUID,
		Content:  content,
		Type:     entity.MessageType(msgType),
	}

	// REFACTORED: Create -> CreateMessage
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	msgJSON, _ := json.Marshal(msg)
	channel := fmt.Sprintf("chat:group:%s", groupID)
	err := s.redisClient.Publish(ctx, channel, msgJSON).Err()

	return msg, err
}

func (s *chatService) GetHistory(ctx context.Context, groupID string, beforeID string) ([]entity.Message, error) {
	// REFACTORED: GetHistory -> GetMessageHistory
	return s.repo.GetMessageHistory(ctx, groupID, 50, beforeID)
}

func (s *chatService) GetRedisPubSub(groupID string) *redis.PubSub {
	channel := fmt.Sprintf("chat:group:%s", groupID)
	return s.redisClient.Subscribe(context.Background(), channel)
}

func (s *chatService) DeleteMessage(ctx context.Context, groupID, messageID, userID string) error {
	// REFACTORED: FindByID -> FindMessageByID
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
