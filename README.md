Chorequest Monorepo (Go GraphQL + React + Tailwind)

Overview
- Backend: Go HTTP server with GraphQL (gqlgen), JWT middleware, DynamoDB client + table bootstrap.
- Frontend: React (Vite, TS), Tailwind v4, Apollo Client, Codegen config.

Getting Started
1) Backend
   - Env (optional):
     - `JWT_SECRET`: HMAC secret for parsing Bearer tokens
     - `AWS_REGION`: AWS region (default `us-east-1`)
     - `DYNAMODB_ENDPOINT`: e.g. `http://localhost:8000` for local DynamoDB
   - Launch local DynamoDB: `make dynamodb-up` (exposes `http://localhost:8000`)
   - Run: `DYNAMO_AUTO_MIGRATE=1 DYNAMODB_ENDPOINT=http://localhost:8000 make dev-backend` (GraphQL at `http://localhost:8080/query`)
   - Seed: `DYNAMODB_ENDPOINT=http://localhost:8000 make seed` (env `SEED_PARENT_ID` optional)
   - Health: `GET http://localhost:8080/healthz`

2) Frontend
   - Install deps: `make bootstrap`
   - Dev server: `make dev-web` (proxies `/query` to backend)
   - Routes:
     - Parent Dashboard: `http://localhost:5173/parent/parent-1`
     - Child View: `http://localhost:5173/child/<childId>`

3) Codegen (optional)
   - With backend running: `make codegen` to generate typed GraphQL artifacts in `web/src/gql/`.

Notes
- Tailwind v4 is configured via `@import "tailwindcss";` in `web/src/index.css`.
- Apollo points at `/query` by default. Override with `VITE_GRAPHQL_URL` if needed.
- JWT middleware is permissive if `JWT_SECRET` is unset; it only enriches context when a valid token is present.
- GraphQL Playground at `/play`.

Mobile (Capacitor)
- Install: `cd web && npm i`
- Initialize (already configured): `npm run build` then `npm run cap:sync`
- Add platforms: `npm run cap:add:ios` or `npm run cap:add:android`
- Open native projects: `npm run cap:open:ios` / `npm run cap:open:android`

Next Steps
- Extend GraphQL resolvers to persist domain (done, backed by Dynamo single-table).
- Introduce auth (roll-your-own JWT now, Auth0 later) using `web/src/lib/auth.ts` to store tokens.

GraphQL Domain (initial)
- Users (Parent/Child), Children with xp/gold, Quests with xp/gold, Assignments, Rewards.
- Key operations: `createChild`, `createQuest`, `assignQuest`, `completeAssignment`, `createReward`, `purchaseItem`, and queries for children/quests/rewards/assignments.

Capacitor Notes
- Secure token storage uses Capacitor Preferences by default; swap for a secure plugin later.
- Generate icons/splash after placing `resources/icon.png`: `npm run cap:assets`

Containers
- Full local stack: `make stack-up` (DynamoDB Local + API on :8080). Logs: `make stack-logs`. Tear down: `make stack-down`.

Auth (Dev)
- Start backend with `JWT_SECRET` set (e.g., `export JWT_SECRET=devsecret`).
- Visit `http://localhost:5173/login` to issue a dev token (backend `POST /auth/dev`) and store it locally; Apollo sends it as `Authorization: Bearer ...`.
