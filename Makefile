SHELL := /bin/bash

.PHONY: dev dev-backend build-backend dev-web gen codegen bootstrap ensure-env ensure-docker-env

dev-backend:
	cd backend && go run ./cmd/server

build-backend:
	cd backend && go build -o bin/server ./cmd/server

dev-web:
	cd web && npm run dev

# Ensure example envs are present as working .env files (non-secret defaults)
ensure-env:
	@test -f backend/.env || cp backend/.env.example backend/.env
	@test -f web/.env || cp web/.env.example web/.env

# Run backend and web dev servers concurrently
dev: ensure-env
	@trap 'kill 0' SIGINT SIGTERM EXIT; \
	$(MAKE) dev-backend & \
	$(MAKE) dev-web & \
	wait

codegen:
	cd web && npm run codegen

bootstrap:
	cd web && npm install
	cd backend && go mod tidy

.PHONY: dynamodb-up dynamodb-down dynamodb-logs seed

dynamodb-up:
	cd deploy && docker compose up -d dynamodb

dynamodb-down:
	cd deploy && docker compose down

dynamodb-logs:
	cd deploy && docker compose logs -f dynamodb

seed:
	cd backend && DYNAMO_AUTO_MIGRATE=1 DYNAMODB_ENDPOINT=${DYNAMODB_ENDPOINT} go run ./cmd/seed

.PHONY: stack-up stack-down stack-logs

stack-up:
	$(MAKE) ensure-docker-env
	cd deploy && docker compose up -d --build api dynamodb web

stack-down:
	cd deploy && docker compose down

stack-logs:
	cd deploy && docker compose logs -f api dynamodb web

ensure-docker-env:
	@test -f deploy/api.env || cp deploy/api.env.example deploy/api.env
