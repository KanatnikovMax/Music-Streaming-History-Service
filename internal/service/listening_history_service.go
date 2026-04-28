package service

import (
	"MusicStreamingHistoryService/internal/domain"
	"MusicStreamingHistoryService/internal/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const defaultHistoryLimit = 100

type ListeningHistoryService interface {
	RecordListening(ctx context.Context, item domain.ListenHistoryItem) error
	GetUserHistory(ctx context.Context, userID uuid.UUID) ([]domain.ListenHistoryItem, error)
}

type listeningHistoryService struct {
	repo   repository.ListeningHistoryRepository
	logger *zap.Logger
}

func NewListeningHistoryService(
	repo repository.ListeningHistoryRepository,
	logger *zap.Logger,
) ListeningHistoryService {
	return &listeningHistoryService{
		repo:   repo,
		logger: logger,
	}
}

func (s *listeningHistoryService) RecordListening(
	ctx context.Context,
	item domain.ListenHistoryItem,
) error {
	if err := s.repo.Save(ctx, item); err != nil {
		return fmt.Errorf("failed to save item: %w", err)
	}

	s.logger.Info("listening event recorded",
		zap.String("event_id", item.EventID.String()),
	)

	return nil
}

func (s *listeningHistoryService) GetUserHistory(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.ListenHistoryItem, error) {
	s.logger.Info("getting user history",
		zap.String("user_id", userID.String()),
	)

	items, err := s.repo.GetLastByUser(ctx, userID, defaultHistoryLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user history: %w", err)
	}

	s.logger.Info("got user history",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(items)),
	)

	return items, nil
}
