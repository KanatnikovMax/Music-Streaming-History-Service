package domain

import (
	"time"

	"github.com/google/uuid"
)

type ListenHistoryItem struct {
	EventID       uuid.UUID
	UserID        uuid.UUID
	SongID        uuid.UUID
	ListenedAtUtc time.Time
}
