package consumer

import (
	"MusicStreamingHistoryService/internal/config"
	"MusicStreamingHistoryService/internal/domain"
	"MusicStreamingHistoryService/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const (
	batchSize    = 100
	batchTimeout = 5 * time.Second
)

type KafkaConsumer struct {
	reader *kafka.Reader
	svc    service.ListeningHistoryService
	logger *zap.Logger
}

func NewKafkaConsumer(
	cfg config.KafkaConfig,
	svc service.ListeningHistoryService,
	logger *zap.Logger,
) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MaxBytes: 10e6,
	})

	return &KafkaConsumer{
		reader: reader,
		svc:    svc,
		logger: logger,
	}
}

func (c *KafkaConsumer) Run(ctx context.Context) error {
	c.logger.Info("Consumer starting",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("group_id", c.reader.Config().GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		batch, err := c.readBatch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.logger.Error("Error reading batch", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		if len(batch) == 0 {
			continue
		}

		c.processBatch(ctx, batch)
	}
}

func (c *KafkaConsumer) readBatch(ctx context.Context) ([]kafka.Message, error) {
	batchCtx, cancel := context.WithTimeout(ctx, batchTimeout)
	defer cancel()

	var messages []kafka.Message

	for len(messages) < batchSize {
		select {
		case <-batchCtx.Done():
			return messages, nil
		default:
		}

		msg, err := c.reader.ReadMessage(batchCtx)
		if err != nil {
			if batchCtx.Err() != nil {
				return messages, nil
			}
			return messages, fmt.Errorf("failed to read message: %w", err)
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func (c *KafkaConsumer) processBatch(ctx context.Context, messages []kafka.Message) {
	c.logger.Info("processing batch", zap.Int("size", len(messages)))

	var successCount, failCount int

	for _, msg := range messages {
		item, err := c.parseMessage(msg)
		if err != nil {
			c.logger.Error("failed to parse message",
				zap.Error(err),
				zap.Int64("offset", msg.Offset),
				zap.Int("partition", msg.Partition),
				zap.String("key", string(msg.Key)),
			)
			failCount++
			continue
		}

		if err := c.svc.RecordListening(ctx, *item); err != nil {
			c.logger.Error("failed to record listening",
				zap.Error(err),
				zap.String("event_id", item.EventID.String()),
				zap.String("user_id", item.UserID.String()),
			)
			failCount++
			continue
		}

		successCount++
	}

	c.logger.Info("batch processed",
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
	)
}

func (c *KafkaConsumer) parseMessage(msg kafka.Message) (*domain.ListenHistoryItem, error) {
	var event songPlayedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	listenedAt, err := time.Parse(time.RFC3339, event.ListenedAtUtc)
	if err != nil {
		listenedAt, err = time.Parse("2006-01-02T15:04:05", event.ListenedAtUtc)
		if err != nil {
			return nil, fmt.Errorf("failed to parse listened_at_utc %q: %w", event.ListenedAtUtc, err)
		}
	}

	return &domain.ListenHistoryItem{
		EventID:       event.EventID,
		UserID:        event.UserID,
		SongID:        event.SongID,
		ListenedAtUtc: listenedAt,
	}, nil
}

func (c *KafkaConsumer) Close() error {
	c.logger.Info("Consumer shutting down")
	return c.reader.Close()
}
