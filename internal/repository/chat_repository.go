package repository

import (
	"context"
	"errors"
	"umrah-backend/internal/entity"

	"gorm.io/gorm"
)

type ChatRepository interface {
	Create(ctx context.Context, msg *entity.Message) error
	GetHistory(ctx context.Context, groupID string, limit int, beforeID string) ([]entity.Message, error)
	DeleteMessage(ctx context.Context, messageID string, userID string) error
	FindByID(ctx context.Context, messageID string) (*entity.Message, error)
}

type chatRepo struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepo{db: db}
}

// 1. Create Message
func (r *chatRepo) Create(ctx context.Context, msg *entity.Message) error {
	// Gunakan WithContext(ctx)
	if err := r.db.WithContext(ctx).Create(msg).Error; err != nil {
		return err
	}
	// Reload data agar Sender terisi
	return r.db.WithContext(ctx).Preload("Sender").First(msg, msg.ID).Error
}

// 2. Get History (Support Pagination)
func (r *chatRepo) GetHistory(ctx context.Context, groupID string, limit int, beforeID string) ([]entity.Message, error) {
	var messages []entity.Message

	// Query Dasar
	query := r.db.WithContext(ctx).
		Preload("Sender").
		Where("group_id = ?", groupID).
		Order("created_at desc").
		Limit(limit)

	// Logic Pagination: Jika user scroll ke atas (Load More)
	if beforeID != "" {
		// Cari pesan yang waktunya LEBIH LAMA (<) dari pesan dengan ID 'beforeID'
		subQuery := r.db.Table("messages").Select("created_at").Where("id = ?", beforeID)
		query = query.Where("created_at < (?)", subQuery)
	}

	err := query.Find(&messages).Error
	return messages, err
}

// 3. [BARU] FindByID (Untuk Service mengambil data lengkap sebelum delete)
func (r *chatRepo) FindByID(ctx context.Context, messageID string) (*entity.Message, error) {
	var msg entity.Message

	// PENTING: Preload("Sender") agar data user terbawa lengkap
	err := r.db.WithContext(ctx).
		Preload("Sender").
		First(&msg, "id = ?", messageID).Error

	if err != nil {
		return nil, err
	}

	return &msg, nil
}

// 4. Soft Delete Message
func (r *chatRepo) DeleteMessage(ctx context.Context, messageID string, userID string) error {
	query := r.db.WithContext(ctx).
		Model(&entity.Message{}).
		Where("id = ? AND sender_id = ?", messageID, userID)

	// Update Content & Type
	result := query.Updates(map[string]interface{}{
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
