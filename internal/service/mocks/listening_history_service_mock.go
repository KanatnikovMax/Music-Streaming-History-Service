package mocks

import (
	"MusicStreamingHistoryService/internal/domain"
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type ListeningHistoryServiceMock struct {
	mock.Mock
}

func (m *ListeningHistoryServiceMock) RecordListening(ctx context.Context, item domain.ListeningHistoryItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *ListeningHistoryServiceMock) GetUserHistory(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.ListeningHistoryItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ListeningHistoryItem), args.Error(1)
}
