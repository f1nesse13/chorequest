Chorequest Monorepo (Go GraphQL + React + Tailwind)

Overview
- Backend: Go HTTP server with GraphQL (gqlgen), JWT middleware, and SQLite persistence.
- Frontend: React (Vite, TS), Tailwind v4, Apollo Client, Codegen config.

Getting Started
1) Backend
   - Env (optional):
     - `JWT_SECRET`: HMAC secret for parsing Bearer tokens
     - `SQLITE_PATH`: path to a local SQLite database (`./data/dev.sqlite` for local dev)
   - Local SQLite quick start:
     1. `export SQLITE_PATH=./data/dev.sqlite`
     2. `make migrate`
     3. `make seed`
     4. `make dev-backend`
   - Health: `GET http://localhost:8080/healthz`

2) Frontend
   - Install deps: `make bootstrap`
   - Dev server: `make dev-web` (proxies `/query` to backend)
   - Routes:
     - Family Dashboard: `http://localhost:5173/family/<familyId>`
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
- Full local stack: `make stack-up` (API + Web). Logs: `make stack-logs`. Tear down: `make stack-down`.

Auth (Dev)
- Start backend with `JWT_SECRET` set (e.g., `export JWT_SECRET=devsecret`).
- Visit `http://localhost:5173/login` to issue a dev token (backend `POST /auth/dev`) and store it locally; Apollo sends it as `Authorization: Bearer ...`.
