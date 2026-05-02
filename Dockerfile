FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/app

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/app .

COPY config.yaml .
COPY migrations/ migrations/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 50051

CMD ["./app"]