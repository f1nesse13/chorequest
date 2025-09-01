SHELL := /bin/bash

.PHONY: dev-backend build-backend dev-web gen codegen bootstrap

dev-backend:
	cd backend && go run ./cmd/server

build-backend:
	cd backend && go build -o bin/server ./cmd/server

dev-web:
	cd web && npm run dev

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
	cd deploy && docker compose up -d --build api dynamodb web

stack-down:
	cd deploy && docker compose down

stack-logs:
	cd deploy && docker compose logs -f api dynamodb web
