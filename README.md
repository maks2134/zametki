# Zametka

Мини-сервис парных заметок для двоих: общая «комната» по секретному коду, идеи с категориями, реакции и real-time через WebSocket.

## Стек

- **Backend:** Go, Fiber, MongoDB, JWT, WebSocket
- **Frontend:** Next.js (App Router), Tailwind, shadcn/ui, framer-motion, zustand

## Быстрый старт

### 1. MongoDB + backend

```bash
# из корня репозитория
docker compose up -d mongo

cd backend
cp .env.example .env   # JWT_SECRET уже ≥32 символов в примере
go run ./cmd/server
```

API: `http://localhost:8080`

Или весь backend в Docker:

```bash
docker compose up --build
```

### 2. Frontend

```bash
cd fronted
cp .env.local.example .env.local
npm install
npm run dev
```

UI: `http://localhost:3000`

## Выложить в интернет (бесплатно)

Пошаговая инструкция для вас двоих: **[DEPLOY.md](DEPLOY.md)**  
(MongoDB Atlas + Render + Vercel).

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
