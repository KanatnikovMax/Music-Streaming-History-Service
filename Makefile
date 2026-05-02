.PHONY: run build infra-up infra-down infra-logs proto docker-up docker-down

run:
	go run cmd/app/main.go

build:
	go build -o bin/app cmd/app/main.go

infra-up:
	docker-compose up -d cassandra_db kafka kafka-ui kafka-init

infra-down:
	docker-compose down

infra-logs:
	docker-compose logs -f

proto:
	protoc \
		--proto_path=pkg/proto/listening_history \
		--go_out=pkg/proto/listening_history \
		--go_opt=paths=source_relative \
		--go-grpc_out=pkg/proto/listening_history \
		--go-grpc_opt=paths=source_relative \
		pkg/proto/listening_history/listening_history.proto

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down -v