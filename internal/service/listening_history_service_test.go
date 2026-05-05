package service

import (
	"MusicStreamingHistoryService/internal/domain"
	"MusicStreamingHistoryService/internal/repository/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestListeningHistoryService(repo *mocks.ListeningHistoryRepositoryMock) ListeningHistoryService {
	return NewListeningHistoryService(repo, zap.NewNop())
}

func newTestItem() domain.ListeningHistoryItem {
	return domain.ListeningHistoryItem{
		EventID:       uuid.New(),
		UserID:        uuid.New(),
		SongID:        uuid.New(),
		ListenedAtUtc: time.Now().UTC().Truncate(time.Millisecond),
	}
}

func TestRecordListening_Success(t *testing.T) {
	// Arrange
	repoMock := &mocks.ListeningHistoryRepositoryMock{}
	svc := newTestListeningHistoryService(repoMock)
	item := newTestItem()

	repoMock.On("Save", mock.Anything, item).Return(nil)

	// Act
	err := svc.RecordListening(context.Background(), item)

	// Assert
	require.NoError(t, err)
	repoMock.AssertExpectations(t)
}

func TestRecordListening_RepositoryError(t *testing.T) {
	// Arrange
	repoMock := &mocks.ListeningHistoryRepositoryMock{}
	svc := newTestListeningHistoryService(repoMock)
	item := newTestItem()
	repoErr := errors.New("cassandra unavailable")

	repoMock.On("Save", mock.Anything, item).Return(repoErr)

	// Act
	err := svc.RecordListening(context.Background(), item)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
	repoMock.AssertExpectations(t)
}

func TestGetUserHistory_Success(t *testing.T) {
	// Arrange
	repoMock := &mocks.ListeningHistoryRepositoryMock{}
	svc := newTestListeningHistoryService(repoMock)
	userID := uuid.New()

	expectedItems := []domain.ListeningHistoryItem{
		newTestItem(),
		newTestItem(),
		newTestItem(),
	}

	repoMock.On("GetLastByUser", mock.Anything, userID, 100).Return(expectedItems, nil)

	// Act
	items, err := svc.GetUserHistory(context.Background(), userID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, expectedItems, items)
	repoMock.AssertExpectations(t)
}

func TestGetUserHistory_EmptyHistory(t *testing.T) {
	// Arrange
	repoMock := &mocks.ListeningHistoryRepositoryMock{}
	svc := newTestListeningHistoryService(repoMock)
	userID := uuid.New()

	// Репозиторий вернул nil — пользователь ничего не слушал
	repoMock.On("GetLastByUser", mock.Anything, userID, 100).Return(nil, nil)

	// Act
	items, err := svc.GetUserHistory(context.Background(), userID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, items)
	repoMock.AssertExpectations(t)
}

func TestGetUserHistory_RepositoryError(t *testing.T) {
	// Arrange
	repoMock := &mocks.ListeningHistoryRepositoryMock{}
	svc := newTestListeningHistoryService(repoMock)
	userID := uuid.New()
	repoErr := errors.New("connection timeout")

	repoMock.On("GetLastByUser", mock.Anything, userID, 100).Return(nil, repoErr)

	// Act
	items, err := svc.GetUserHistory(context.Background(), userID)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, items)
	repoMock.AssertExpectations(t)
}

func TestGetUserHistory_ReturnsMaxHundredItems(t *testing.T) {
	// Arrange
	repoMock := &mocks.ListeningHistoryRepositoryMock{}
	svc := newTestListeningHistoryService(repoMock)
	userID := uuid.New()

	items := make([]domain.ListeningHistoryItem, 100)
	for i := range items {
		items[i] = newTestItem()
	}

	repoMock.On("GetLastByUser", mock.Anything, userID, 100).Return(items, nil)

	// Act
	result, err := svc.GetUserHistory(context.Background(), userID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 100)
	repoMock.AssertExpectations(t)
}
