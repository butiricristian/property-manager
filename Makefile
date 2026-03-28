.PHONY: run build migrate-up migrate-down migrate-create sqlc frontend build-frontend

DATABASE_URL ?= postgres://cristianbutiri@localhost:5432/property_manager?sslmode=disable

run:
	go run ./cmd/server

build-frontend:
	cd web && npm ci && npm run build

build: build-frontend
	go build -o bin/server ./cmd/server

migrate-up:
	migrate -database "$(DATABASE_URL)" -path internal/database/migrations up

migrate-down:
	migrate -database "$(DATABASE_URL)" -path internal/database/migrations down 1

migrate-create:
	migrate create -ext sql -dir internal/database/migrations -seq $(name)

sqlc:
	sqlc generate

dev-frontend:
	cd web && npm run dev
