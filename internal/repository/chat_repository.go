package repository

import (
	"context"
	"errors"
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type ChatRepository interface {
	// REFACTORED: 'Create' -> 'CreateMessage' for clarity
	CreateMessage(ctx context.Context, msg *entity.Message) error
	GetMessageHistory(ctx context.Context, groupID string, limit int, beforeID string) ([]entity.Message, error)
	FindMessageByID(ctx context.Context, messageID string) (*entity.Message, error)
	DeleteMessage(ctx context.Context, messageID string, userID string) error
}

type chatRepo struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepo{db: db}
}

func (r *chatRepo) CreateMessage(ctx context.Context, msg *entity.Message) error {
	if err := r.db.WithContext(ctx).Create(msg).Error; err != nil {
		return err
	}
	// Reload data to ensure Sender relationship is populated
	return r.db.WithContext(ctx).Preload("Sender").First(msg, msg.ID).Error
}

func (r *chatRepo) GetMessageHistory(ctx context.Context, groupID string, limit int, beforeID string) ([]entity.Message, error) {
	var messages []entity.Message

	query := r.db.WithContext(ctx).
		Preload("Sender").
		Where("group_id = ?", groupID).
		Order("created_at desc").
		Limit(limit)

	if beforeID != "" {
		subQuery := r.db.Table("messages").Select("created_at").Where("id = ?", beforeID)
		query = query.Where("created_at < (?)", subQuery)
	}

	err := query.Find(&messages).Error
	return messages, err
}

func (r *chatRepo) FindMessageByID(ctx context.Context, messageID string) (*entity.Message, error) {
	var msg entity.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		First(&msg, "id = ?", messageID).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *chatRepo) DeleteMessage(ctx context.Context, messageID string, userID string) error {
	result := r.db.WithContext(ctx).
		Model(&entity.Message{}).
		Where("id = ? AND sender_id = ?", messageID, userID).
		Updates(map[string]interface{}{
			"content": "🚫 Pesan ini telah dihapus",
			"type":    "DELETED",
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("pesan tidak ditemukan atau anda bukan pengirimnya")
	}
	return nil
}
