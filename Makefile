.PHONY: run build infra-up infra-down infra-logs proto

run:
	go run cmd/app/main.go

build:
	go build -o bin/app cmd/app/main.go

infra-up:
	docker-compose up -d

infra-down:
	docker-compose down

infra-logs:
	docker-compose logs -f cassandra

proto:
	protoc \
		--proto_path=pkg/proto/listening_history \
		--go_out=pkg/proto/listening_history \
		--go_opt=paths=source_relative \
		--go-grpc_out=pkg/proto/listening_history \
		--go-grpc_opt=paths=source_relative \
		pkg/proto/listening_history/listening_history.proto