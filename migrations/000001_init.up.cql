CREATE TABLE IF NOT EXISTS music_streaming.listening_history (
     user_id         UUID,
     listened_at_utc TIMESTAMP,
     event_id        UUID,
     song_id         UUID,
     PRIMARY KEY (user_id, listened_at_utc, event_id)
) WITH CLUSTERING ORDER BY (listened_at_utc DESC);