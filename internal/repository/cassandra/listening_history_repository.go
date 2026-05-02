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

func toGoCqlUUID(id uuid.UUID) gocql.UUID {
	return gocql.UUID(id)
}

func toGoogleUUID(id gocql.UUID) uuid.UUID {
	return uuid.UUID(id)
}

func (r *listeningHistoryRepository) Save(
	ctx context.Context,
	item domain.ListenHistoryItem,
) error {
	query := `
		INSERT INTO listening_history (user_id, listened_at_utc, event_id, song_id)
		VALUES (?, ?, ?, ?)`

	err := r.session.
		Query(
			query,
			toGoCqlUUID(item.UserID),
			item.ListenedAtUtc,
			toGoCqlUUID(item.EventID),
			toGoCqlUUID(item.SongID),
		).
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
		Query(query, toGoCqlUUID(userID), limit).
		WithContext(ctx).
		Iter()

	var items []domain.ListenHistoryItem
	var item domain.ListenHistoryItem
	var eventID, rowUserID, songID gocql.UUID
	for iter.Scan(&eventID, &rowUserID, &songID, &item.ListenedAtUtc) {
		item.EventID = toGoogleUUID(eventID)
		item.UserID = toGoogleUUID(rowUserID)
		item.SongID = toGoogleUUID(songID)
		items = append(items, item)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get listening history items: %w", err)
	}

	return items, nil
}
