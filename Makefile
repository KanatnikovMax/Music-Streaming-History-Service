.PHONY: run build infra-up infra-down infra-logs

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