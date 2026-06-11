.PHONY: test test-go test-web lint build up down

test: test-go test-web

test-go:
	GOTOOLCHAIN=local go test ./apps/api ./apps/worker ./internal/...

test-web:
	cd apps/web && npm test

lint:
	GOTOOLCHAIN=local go vet ./apps/api ./apps/worker ./internal/...
	cd apps/web && npm run lint

build:
	GOTOOLCHAIN=local go build ./apps/api ./apps/worker
	cd apps/web && npm run build

up:
	docker compose up --build

down:
	docker compose down --remove-orphans
