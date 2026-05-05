package mocks

import (
	"MusicStreamingHistoryService/internal/domain"
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type ListeningHistoryRepositoryMock struct {
	mock.Mock
}

func (m *ListeningHistoryRepositoryMock) Save(ctx context.Context, item domain.ListeningHistoryItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *ListeningHistoryRepositoryMock) GetLastByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.ListeningHistoryItem, error) {
	args := m.Called(ctx, userID, limit)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ListeningHistoryItem), args.Error(1)
}
