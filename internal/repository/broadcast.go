package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BroadcastSelectionRepository interface {
	Toggle(userID, chatID int64) (bool, error)
	GetSelections(userID int64) ([]int64, error)
	SelectAll(userID int64, chatIDs []int64) error
	Clear(userID int64) error
}

type PostgresBroadcastSelectionRepository struct {
	db *gorm.DB
}

func NewBroadcastSelectionRepository(db *gorm.DB) BroadcastSelectionRepository {
	return &PostgresBroadcastSelectionRepository{db: db}
}

func (r *PostgresBroadcastSelectionRepository) Toggle(userID, chatID int64) (bool, error) {
	var existing BroadcastSelection
	err := r.db.Where("user_id = ? AND chat_id = ?", userID, chatID).First(&existing).Error
	if err == nil {
		if delErr := r.db.Where("user_id = ? AND chat_id = ?", userID, chatID).Delete(&BroadcastSelection{}).Error; delErr != nil {
			return false, fmt.Errorf("failed to deselect chat: %w", delErr)
		}
		return false, nil
	}
	selection := BroadcastSelection{UserID: userID, ChatID: chatID, CreatedAt: time.Now()}
	if createErr := r.db.Create(&selection).Error; createErr != nil {
		return false, fmt.Errorf("failed to select chat: %w", createErr)
	}
	return true, nil
}

func (r *PostgresBroadcastSelectionRepository) GetSelections(userID int64) ([]int64, error) {
	var selections []BroadcastSelection
	if err := r.db.Where("user_id = ?", userID).Find(&selections).Error; err != nil {
		return nil, fmt.Errorf("failed to get selections: %w", err)
	}
	chatIDs := make([]int64, len(selections))
	for i, s := range selections {
		chatIDs[i] = s.ChatID
	}
	return chatIDs, nil
}

func (r *PostgresBroadcastSelectionRepository) SelectAll(userID int64, chatIDs []int64) error {
	if len(chatIDs) == 0 {
		return nil
	}
	selections := make([]BroadcastSelection, len(chatIDs))
	for i, id := range chatIDs {
		selections[i] = BroadcastSelection{UserID: userID, ChatID: id, CreatedAt: time.Now()}
	}
	if err := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&selections).Error; err != nil {
		return fmt.Errorf("failed to select all chats: %w", err)
	}
	return nil
}

func (r *PostgresBroadcastSelectionRepository) Clear(userID int64) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&BroadcastSelection{}).Error; err != nil {
		return fmt.Errorf("failed to clear selections: %w", err)
	}
	return nil
}

type BroadcastDraftRepository interface {
	Save(userID int64, text string) error
	Get(userID int64) (string, error)
	Delete(userID int64) error
}

type PostgresBroadcastDraftRepository struct {
	db *gorm.DB
}

func NewBroadcastDraftRepository(db *gorm.DB) BroadcastDraftRepository {
	return &PostgresBroadcastDraftRepository{db: db}
}

func (r *PostgresBroadcastDraftRepository) Save(userID int64, text string) error {
	draft := BroadcastDraft{UserID: userID, Text: text, CreatedAt: time.Now()}
	if err := r.db.Save(&draft).Error; err != nil {
		return fmt.Errorf("failed to save broadcast draft: %w", err)
	}
	return nil
}

func (r *PostgresBroadcastDraftRepository) Get(userID int64) (string, error) {
	var draft BroadcastDraft
	if err := r.db.Where("user_id = ?", userID).First(&draft).Error; err != nil {
		return "", fmt.Errorf("failed to get broadcast draft: %w", err)
	}
	return draft.Text, nil
}

func (r *PostgresBroadcastDraftRepository) Delete(userID int64) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&BroadcastDraft{}).Error; err != nil {
		return fmt.Errorf("failed to delete broadcast draft: %w", err)
	}
	return nil
}
