package repository

import (
	"MusicStreamingHistoryService/internal/domain"
	"context"

	"github.com/google/uuid"
)

type ListeningHistoryRepository interface {
	Save(ctx context.Context, item domain.ListeningHistoryItem) error
	GetLastByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.ListeningHistoryItem, error)
}
