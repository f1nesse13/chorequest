SHELL := /bin/bash

SQLITE_PATH ?= ./data/dev.sqlite

.PHONY: dev dev-backend build-backend dev-web gen codegen bootstrap ensure-env ensure-docker-env

# Run the backend server in development mode
dev-backend:
	cd backend && go run ./cmd/server

# Build the backend server binary
build-backend:
	cd backend && go build -o bin/server ./cmd/server

# Run the web app in development mode
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

# Generate client code artifacts for the web app
codegen:
	cd web && npm run codegen

# Install JS dependencies and tidy Go modules
bootstrap:
	cd web && npm install
	cd backend && go mod tidy

.PHONY: dynamodb-up dynamodb-down dynamodb-logs seed sqlite-db sqlite-migrate dev-backend-sqlite

# Start the local DynamoDB container
dynamodb-up:
	cd deploy && docker compose up -d dynamodb

# Tear down the DynamoDB container
dynamodb-down:
	cd deploy && docker compose down

# Tail logs from the DynamoDB container
dynamodb-logs:
	cd deploy && docker compose logs -f dynamodb

# Seed the DynamoDB instance with starter data
seed:
	cd backend && DYNAMO_AUTO_MIGRATE=1 DYNAMODB_ENDPOINT=${DYNAMODB_ENDPOINT} go run ./cmd/seed

# Ensure the SQLite data directory / file exists
sqlite-db:
	cd backend && mkdir -p $(dir $(SQLITE_PATH)) && touch $(SQLITE_PATH)

# Run the Dynamo -> SQLite migration (requires Dynamo env vars)
sqlite-migrate: sqlite-db
	cd backend && SQLITE_PATH=$(SQLITE_PATH) go run ./cmd/migrate --sqlite $(SQLITE_PATH)

# Convenience target to run backend with SQLite enabled
dev-backend-sqlite: sqlite-db
	cd backend && SQLITE_PATH=$(SQLITE_PATH) go run ./cmd/server

.PHONY: stack-up stack-down stack-logs

# Start the full local stack (API, DynamoDB, Web)
stack-up:
	$(MAKE) ensure-docker-env
	cd deploy && docker compose up -d --build api dynamodb web

# Stop the local stack containers
stack-down:
	cd deploy && docker compose down

# Tail logs across the stack services
stack-logs:
	cd deploy && docker compose logs -f api dynamodb web

# Ensure required Docker env files exist
ensure-docker-env:
	@test -f deploy/api.env || cp deploy/api.env.example deploy/api.env
