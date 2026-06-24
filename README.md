# ParkWatch — Parking Violation Portal

Fullstack technical assignment (Tan Digital 2026). A single web app for two roles — **Officer** and
**Member** — backed by Go microservices behind a single API Gateway. Fine rules are versioned, and
every violation keeps an immutable snapshot of how its fine was calculated, so updating the rules
never changes fines that were already issued.

➡️ **Architecture, data-flow diagram, and ERD are in [DESIGN.md](DESIGN.md).**

## The five flows

1. An officer submits a violation (plate, type, location, timestamp, photo).
2. The system calculates the fine from the **active** rule version.
3. An officer publishes a new rule version — past fines are unaffected.
4. A member pays a fine (the payment provider is mocked; choose a `success`/`failed` scenario).
5. A transaction history shows each violation, its fine, and **which rule version applied at issue
   time**.

## Stack

- **Backend:** Go (chi, pgx/v5, amqp091, go-redis, minio-go, golang-jwt) — services `gateway`,
  `identity`, `rules`, `violation`, `payment`, `notification`.
- **Frontend:** Next.js (App Router) + TypeScript, Tailwind v4 + shadcn/ui, light/dark theme,
  TanStack Query.
- **Infra:** PostgreSQL, RabbitMQ, Redis, MinIO — via Docker Compose.

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (WSL2 backend on Windows) — the
  only thing you need to run the whole stack.
- [Node.js](https://nodejs.org/) 20+ — only if you want to run the frontend in dev mode.

## Run it locally

```bash
cp .env.example .env
docker compose up -d --build
```

That's it — this builds and starts everything: Postgres, RabbitMQ, Redis, MinIO, the six Go
services, and the **web app**. Then open **http://localhost:3000**.

On first start each service creates its schema and seeds demo data (the two users, the day-one
ruleset, and the demo plate). The web app proxies `/api/*` to the gateway, so the auth cookie stays
first-party (no CORS in the normal flow). The first build takes a few minutes (it compiles the Go
services and the Next.js app).

> **Frontend dev mode (hot reload):** stop the web container and run the dev server instead —
> `docker compose stop web` then `cd web && npm install && npm run dev`.

### Demo accounts

| Role    | Email                      | Password      |
|---------|----------------------------|---------------|
| Officer | `officer@parkwatch.test`   | `password123` |
| Member  | `member@parkwatch.test`    | `password123` |

The seeded plate **`B1234ABC`** belongs to the member, so a violation an officer files against it
shows up in the member's "My violations" and "Pay a fine" views.

### A quick walkthrough

1. Sign in as the **officer** → *Submit violation* (plate `B1234ABC`) → see the computed fine.
2. *Fine rules* → change an amount and *Publish* → the new version is active; history keeps the old
   one intact.
3. Sign in as the **member** → *Pay a fine* → choose **success** or **failed** and pay. On success
   the violation flips to paid and a notification appears in the bell.

### Beyond the five flows

A few extras built on top of the core slice:

- **Violation detail page** — click any row in the violations list to see the full fine snapshot
  (base × time × repeat, prior-unpaid count, applied rule version) and the photo.
- **Notifications** are clickable (deep-link to the related violation), and the officer who issued
  a violation is also notified when it gets paid.
- **Printable receipt** — paid invoices link to a print-friendly receipt (Print / Save as PDF).
- **Light / dark theme** toggle.

### Service URLs

| Surface       | URL                          |
|---------------|------------------------------|
| Web app       | http://localhost:3000        |
| API Gateway   | http://localhost:8080        |
| RabbitMQ UI   | http://localhost:15672 (parkwatch / parkwatch) |
| MinIO Console | http://localhost:9001 (parkwatch / parkwatch123) |

## Tests

```bash
go test ./...
```

The fine-calculation logic (`pkg/fine`) and JWT handling (`pkg/jwt`) are unit-tested, including the
time/repeat multiplier boundaries and rounding.

## Assumptions & trade-offs

- **Schema bootstrap, not migrations.** Each service runs `CREATE TABLE IF NOT EXISTS` and seeds on
  startup to keep "one command to run it" simple. Production would use versioned migrations.
- **Plate ↔ member** ownership is a small registry table (`plates`), seeded for the demo. There is no
  self-service plate management UI.
- **Trusted gateway headers.** Downstream services trust `X-User-*` from the gateway and assume they
  are only reachable through it on the internal network.
- **Event publishing is best-effort.** A production system would use a transactional outbox and
  idempotent consumers; here the consumers are idempotent on natural keys (e.g. one invoice per
  violation) but there is no outbox.
- **Violation timestamps are treated as UTC** end to end so the shown time matches the applied time
  multiplier.

See [DESIGN.md](DESIGN.md) for the full rationale and **What I'd do with more time**.
