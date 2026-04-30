package consumer

import "github.com/google/uuid"

type songPlayedEvent struct {
	EventID       uuid.UUID `json:"EventId"`
	UserID        uuid.UUID `json:"UserId"`
	SongID        uuid.UUID `json:"SongId"`
	ListenedAtUtc string    `json:"ListenedAtUtc"`
}
