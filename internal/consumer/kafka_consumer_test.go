package consumer

import (
	servicemocks "MusicStreamingHistoryService/internal/service/mocks"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestConsumer(svc *servicemocks.ListeningHistoryServiceMock) *KafkaConsumer {
	return &KafkaConsumer{
		reader: nil,
		svc:    svc,
		logger: zap.NewNop(),
	}
}

func newKafkaMessage(t *testing.T, event songPlayedEvent) kafka.Message {
	t.Helper()

	body, err := json.Marshal(event)
	require.NoError(t, err)

	return kafka.Message{
		Key:   []byte(event.UserID.String()),
		Value: body,
	}
}

func newValidEvent() songPlayedEvent {
	return songPlayedEvent{
		EventID:       uuid.New(),
		UserID:        uuid.New(),
		SongID:        uuid.New(),
		ListenedAtUtc: "2024-01-15T10:00:00Z",
	}
}

func TestParseMessage_ValidEvent(t *testing.T) {
	// Arrange
	consumer := newTestConsumer(&servicemocks.ListeningHistoryServiceMock{})
	event := newValidEvent()
	msg := newKafkaMessage(t, event)

	// Act
	item, err := consumer.parseMessage(msg)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, event.EventID, item.EventID)
	assert.Equal(t, event.UserID, item.UserID)
	assert.Equal(t, event.SongID, item.SongID)

	expectedTime, _ := time.Parse(time.RFC3339, event.ListenedAtUtc)
	assert.Equal(t, expectedTime.UTC(), item.ListenedAtUtc.UTC())
}

func TestParseMessage_InvalidJSON(t *testing.T) {
	// Arrange
	consumer := newTestConsumer(&servicemocks.ListeningHistoryServiceMock{})
	msg := kafka.Message{
		Key:   []byte("user-id"),
		Value: []byte(`{"EventId": "not-closed`),
	}

	// Act
	item, err := consumer.parseMessage(msg)

	// Assert
	require.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "failed to unmarshal event")
}

func TestParseMessage_InvalidEventID(t *testing.T) {
	// Arrange
	consumer := newTestConsumer(&servicemocks.ListeningHistoryServiceMock{})

	msg := kafka.Message{
		Value: []byte(`{
			"EventId": "not-a-uuid",
			"UserId": "11111111-1111-1111-1111-111111111111",
			"SongId": "22222222-2222-2222-2222-222222222222",
			"ListenedAtUtc": "2024-01-15T10:00:00Z"
		}`),
	}

	// Act
	item, err := consumer.parseMessage(msg)

	// Assert
	require.Error(t, err)
	assert.Nil(t, item)
}

func TestParseMessage_InvalidDateFormat(t *testing.T) {
	// Arrange
	consumer := newTestConsumer(&servicemocks.ListeningHistoryServiceMock{})
	event := newValidEvent()
	event.ListenedAtUtc = "15-01-2024 10:00:00"

	msg := newKafkaMessage(t, event)

	// Act
	item, err := consumer.parseMessage(msg)

	// Assert
	require.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "failed to parse listened_at_utc")
}

func TestParseMessage_AlternativeDateFormat(t *testing.T) {
	// Arrange
	consumer := newTestConsumer(&servicemocks.ListeningHistoryServiceMock{})
	event := newValidEvent()
	event.ListenedAtUtc = "2024-01-15T10:00:00"

	msg := newKafkaMessage(t, event)

	// Act
	item, err := consumer.parseMessage(msg)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, 2024, item.ListenedAtUtc.Year())
	assert.Equal(t, time.January, item.ListenedAtUtc.Month())
	assert.Equal(t, 15, item.ListenedAtUtc.Day())
}

func TestParseMessage_EmptyBody(t *testing.T) {
	// Arrange
	consumer := newTestConsumer(&servicemocks.ListeningHistoryServiceMock{})
	msg := kafka.Message{
		Key:   []byte("user-id"),
		Value: []byte{},
	}

	// Act
	item, err := consumer.parseMessage(msg)

	// Assert
	require.Error(t, err)
	assert.Nil(t, item)
}

func TestProcessBatch_AllMessagesSucceed(t *testing.T) {
	// Arrange
	svcMock := &servicemocks.ListeningHistoryServiceMock{}
	consumer := newTestConsumer(svcMock)

	messages := []kafka.Message{
		newKafkaMessage(t, newValidEvent()),
		newKafkaMessage(t, newValidEvent()),
		newKafkaMessage(t, newValidEvent()),
	}

	svcMock.On("RecordListening", mock.Anything, mock.Anything).
		Return(nil).
		Times(3)

	// Act
	consumer.processBatch(context.Background(), messages)

	// Assert
	svcMock.AssertExpectations(t)
	svcMock.AssertNumberOfCalls(t, "RecordListening", 3)
}

func TestProcessBatch_InvalidMessageSkipped(t *testing.T) {
	// Arrange
	svcMock := &servicemocks.ListeningHistoryServiceMock{}
	consumer := newTestConsumer(svcMock)

	messages := []kafka.Message{
		newKafkaMessage(t, newValidEvent()),
		{Value: []byte(`invalid json`)},
		newKafkaMessage(t, newValidEvent()),
	}

	svcMock.On("RecordListening", mock.Anything, mock.Anything).
		Return(nil).
		Times(2)

	// Act
	consumer.processBatch(context.Background(), messages)

	// Assert
	svcMock.AssertExpectations(t)
	svcMock.AssertNumberOfCalls(t, "RecordListening", 2)
}

func TestProcessBatch_ServiceErrorSkipsMessage(t *testing.T) {
	// Arrange
	svcMock := &servicemocks.ListeningHistoryServiceMock{}
	consumer := newTestConsumer(svcMock)

	event1 := newValidEvent()
	event2 := newValidEvent()

	messages := []kafka.Message{
		newKafkaMessage(t, event1),
		newKafkaMessage(t, event2),
	}

	svcMock.On("RecordListening", mock.Anything, mock.Anything).
		Return(assert.AnError).
		Once()
	svcMock.On("RecordListening", mock.Anything, mock.Anything).
		Return(nil).
		Once()

	// Act
	consumer.processBatch(context.Background(), messages)

	// Assert
	svcMock.AssertNumberOfCalls(t, "RecordListening", 2)
}

func TestProcessBatch_EmptyBatch(t *testing.T) {
	// Arrange
	svcMock := &servicemocks.ListeningHistoryServiceMock{}
	consumer := newTestConsumer(svcMock)

	// Act
	consumer.processBatch(context.Background(), []kafka.Message{})

	// Assert
	svcMock.AssertNotCalled(t, "RecordListening")
}
