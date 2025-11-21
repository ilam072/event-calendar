up:
	docker-compose up -d
	go run cmd/app/main.go
down:
	docker-compose down

test:
	go test ./...
