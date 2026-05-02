# Сервис истории прослушиваний

Микросервис для управления историей прослушивания пользователей музыкального стримингового сервиса.

## Описание

Сервис отвечает за сохранение и предоставление истории прослушивания пользователей. 
Он потребляет события о прослушивании песен из Apache Kafka и сохраняет их в Apache Cassandra. 
Другие сервисы могут получить историю прослушиваний через gRPC.

## Используемые технологии
Для реализации использовались следующие технологии:
- Go 1.26
- Apache Cassandra (gocql)
- golang-migrate
- Apache Kafka (segmentio/kafka-go)
- gRPC
- uber-go/zap для логирования
- ilyakaznacheev/cleanenv для конфигурации
- Docker

## Схема данных

### Kafka событие

Сервис батчами потребляет сообщения `SongPlayedEvent` из топика `listening-history`. 
Сообщения партиционированы по `userId`, что гарантирует порядок событий в рамках одного пользователя.

```json
{
  "EventId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "UserId":  "11111111-1111-1111-1111-111111111111",
  "SongId":  "22222222-2222-2222-2222-222222222222",
  "ListenedAtUtc": "2024-01-15T10:00:00Z"
}
```

### Таблица Cassandra

```sql
CREATE TABLE music_streaming.listening_history (
    user_id         UUID,
    listened_at_utc TIMESTAMP,
    event_id        UUID,
    song_id         UUID,
    PRIMARY KEY (user_id, listened_at_utc, event_id)
) WITH CLUSTERING ORDER BY (listened_at_utc DESC);
```

- `user_id` — partition key
- `listened_at_utc` + `event_id` – clustering key с сортировкой по убыванию, последние записи возвращаются сразу без дополнительной сортировки
- `event_id` включён в clustering key для гарантии уникальности в случае, если две песни были прослушаны одновременно

## Запуск

**Требования:** Docker, Docker Compose, Go 1.26

**Запустить инфраструктуру (Cassandra + Kafka):**
```bash
make infra-up
```

**Запустить сервис локально:**
```bash
make run
```

**Запустить всё в Docker:**
```bash
make docker-up
```