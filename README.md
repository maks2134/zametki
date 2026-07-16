# Zametka

Мини-сервис парных заметок для двоих: общая «комната» по секретному коду, идеи с категориями, реакции и real-time через WebSocket.

## Стек

- **Backend:** Go, Fiber, PostgreSQL (Neon), JWT, WebSocket
- **Frontend:** Next.js (App Router), Tailwind, shadcn/ui, framer-motion, zustand

## Быстрый старт

### 1. Postgres + backend

```bash
# из корня репозитория
docker compose up -d postgres

cd backend
cp .env.example .env
go run ./cmd/server
```

API: `http://localhost:8080`

### 2. Frontend

```bash
cd fronted
cp .env.local.example .env.local
npm install
npm run dev
```

UI: `http://localhost:3000`

## Выложить в интернет (бесплатно)

Пошагово: **[DEPLOY.md](DEPLOY.md)** — Neon (Postgres) + Render + Vercel.  
MongoDB больше не используется (на Render часто ломается TLS к Atlas).

## Как пользоваться

1. Один партнёр создаёт пространство (имя + цвет) и копирует код/ссылку.
2. Второй открывает `/join/<CODE>` или вводит код на лендинге.
3. Оба оставляют идеи, ставят реакции — изменения приходят мгновенно по WebSocket.

## API (кратко)

| Метод | Путь | Auth |
|-------|------|------|
| POST | `/api/rooms` | — |
| POST | `/api/rooms/join` | — |
| GET | `/api/rooms/me` | Bearer |
| GET/POST | `/api/notes` | Bearer |
| PATCH/DELETE | `/api/notes/:id` | Bearer (только автор) |
| POST/DELETE | `/api/notes/:id/reactions` | Bearer |
| GET | `/ws?token=...` | query JWT |

## Тесты backend

```bash
cd backend
go test -race ./internal/domain/... ./internal/auth/... ./internal/service/... ./internal/transport/ws/...
```

## Структура

```
backend/          # Go API (module zametka)
fronted/          # Next.js UI
docker-compose.yml
```
# zametki
