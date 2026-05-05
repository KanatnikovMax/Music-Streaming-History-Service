package domain

import (
	"time"

	"github.com/google/uuid"
)

type ListeningHistoryItem struct {
	EventID       uuid.UUID
	UserID        uuid.UUID
	SongID        uuid.UUID
	ListenedAtUtc time.Time
}
