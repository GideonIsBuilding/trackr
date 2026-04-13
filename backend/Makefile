DB_URL ?= postgres://jobtracker_user:devops247@localhost:5432/job_tracker?sslmode=disable

.PHONY: run build migrate-up migrate-down migrate-create tidy

run:
	go run ./cmd

build:
	go build -o bin/api ./cmd

tidy:
	go mod tidy

# Run all pending migrations
migrate-up:
	migrate -path ../migrations -database "$(DB_URL)" up

# Roll back the most recent migration
migrate-down:
	migrate -path ../migrations -database "$(DB_URL)" down 1

# Usage: make migrate-create name=add_indexes
migrate-create:
	migrate create -ext sql -dir ../migrations -seq $(name)
