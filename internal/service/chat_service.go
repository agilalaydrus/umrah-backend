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

	// Update Signature: Tambah ctx dan beforeID
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

	// [FIX] Pass 'ctx' ke repository
	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, err
	}

	msgJSON, _ := json.Marshal(msg)
	channel := fmt.Sprintf("chat:group:%s", groupID)
	err := s.redisClient.Publish(ctx, channel, msgJSON).Err()

	return msg, err
}

func (s *chatService) GetHistory(ctx context.Context, groupID string, beforeID string) ([]entity.Message, error) {
	// [FIX] Pass 'ctx' dan 'beforeID'
	// Kita set limit hardcode 50 pesan per load
	return s.repo.GetHistory(ctx, groupID, 50, beforeID)
}

func (s *chatService) GetRedisPubSub(groupID string) *redis.PubSub {
	channel := fmt.Sprintf("chat:group:%s", groupID)
	return s.redisClient.Subscribe(context.Background(), channel)
}

// Implementasi "Smart Backend" untuk Delete
func (s *chatService) DeleteMessage(ctx context.Context, groupID, messageID, userID string) error {
	// 1. AMBIL DATA ASLI (LENGKAP)
	// Agar Frontend menerima object utuh (Nama, Foto, Timestamp asli)
	originalMsg, err := s.repo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Validasi kepemilikan
	if originalMsg.SenderID.String() != userID {
		return errors.New("anda bukan pemilik pesan ini")
	}

	// 2. HAPUS DI DATABASE (Soft Delete)
	// Status di DB berubah jadi DELETED, content jadi "Pesan dihapus"
	if err := s.repo.DeleteMessage(ctx, messageID, userID); err != nil {
		return err
	}

	// 3. MANIPULASI OBJECT UNTUK BROADCAST
	// Kita pakai object asli yg datanya lengkap tadi, tapi kita ubah isinya
	// sebelum dikirim ke Redis/Frontend.
	originalMsg.Content = "🚫 Pesan ini telah dihapus"
	originalMsg.Type = entity.MsgDeleted

	// 4. BROADCAST KE REDIS
	// Frontend akan terima JSON lengkap. ID sama, Sender ada, tapi isinya "Pesan dihapus".
	msgJSON, _ := json.Marshal(originalMsg)
	channel := fmt.Sprintf("chat:group:%s", groupID)

	return s.redisClient.Publish(ctx, channel, msgJSON).Err()
}
