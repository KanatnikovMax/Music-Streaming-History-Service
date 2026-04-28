package cassandra

import (
	"MusicStreamingHistoryService/internal/domain"
	"MusicStreamingHistoryService/internal/repository"
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type listeningHistoryRepository struct {
	session *gocql.Session
}

func NewListeningHistoryRepository(session *gocql.Session) repository.ListeningHistoryRepository {
	return &listeningHistoryRepository{session: session}
}

func (r *listeningHistoryRepository) Save(
	ctx context.Context,
	item domain.ListenHistoryItem,
) error {
	query := `
		INSERT INTO listening_history (user_id, listened_at_utc, event_id, song_id)
		VALUES (?, ?, ?, ?)`

	err := r.session.
		Query(query, item.UserID, item.ListenedAtUtc, item.EventID, item.SongID).
		WithContext(ctx).
		Exec()

	if err != nil {
		return fmt.Errorf("failed to save listening history item: %w", err)
	}

	return nil
}

func (r *listeningHistoryRepository) GetLastByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]domain.ListenHistoryItem, error) {
	query := `
		SELECT event_id, user_id, song_id, listened_at_utc
		FROM listening_history
		WHERE user_id = ?
		LIMIT ?`

	iter := r.session.
		Query(query, userID, limit).
		WithContext(ctx).
		Iter()

	var items []domain.ListenHistoryItem
	var item domain.ListenHistoryItem
	for iter.Scan(&item.EventID, &item.UserID, &item.SongID, &item.ListenedAtUtc) {
		items = append(items, item)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get listening history items: %w", err)
	}

	return items, nil
}
