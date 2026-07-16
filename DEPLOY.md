# Как выложить Zametka в интернет бесплатно

Цель: вы и девушка открываете сайт с телефона.

Рекомендуемая связка:

| Часть | Сервис | Зачем |
|-------|--------|--------|
| База Postgres | [Neon](https://neon.tech) Free | Надёжный SQL, без TLS-проблем Atlas |
| Backend (Go API + WS) | [Render](https://render.com) Free | API и realtime |
| Frontend (Next.js) | [Vercel](https://vercel.com) Hobby | Сайт |

> MongoDB Atlas на Render часто падает с `tls: internal error` — поэтому в проекте **PostgreSQL (Neon)**.

Ограничение: Render Free после простоя «засыпает» (~15 мин) — первый запрос может занять 30–60 сек.

---

## 0. Подготовка

1. Аккаунты: GitHub, Neon, Render, Vercel.
2. Код в GitHub: https://github.com/maks2134/zametki (или ваш fork).

---

## 1. Neon (база Postgres)

1. https://console.neon.tech → **Create Project** → регион ближе к вам (например Frankfurt).
2. После создания откройте **Connection details**.
3. Скопируйте **Connection string** (URI), вид:

```
postgresql://USER:PASSWORD@ep-xxxx.eu-central-1.aws.neon.tech/neondb?sslmode=require
```

Можно оставить имя БД `neondb` — таблицы создадутся автоматически при старте backend.

Сохраните URI — это `DATABASE_URL` для Render.

---

## 2. Backend на Render

1. https://render.com → **New** → **Web Service** → репозиторий `zametki`.
2. Настройки:

| Поле | Значение |
|------|----------|
| Root Directory | `backend` |
| Environment | **Docker** |
| Dockerfile Path | `./Dockerfile` |
| Instance | Free |

3. **Environment Variables**:

| Key | Value |
|-----|--------|
| `DATABASE_URL` | URI из Neon (целиком) |
| `JWT_SECRET` | строка ≥ 32 символов (`openssl rand -base64 32`) |
| `JWT_TTL` | `720h` |
| `CORS_ORIGINS` | пока `http://localhost:3000` — обновите после Vercel |

`PORT` Render задаёт сам — backend его подхватит.

4. Удалите старые переменные `MONGO_URI` / `MONGO_DB`, если они остались.
5. **Deploy** → дождитесь **Live**.
6. URL вида: `https://zametki-xxxx.onrender.com`

Проверка:

```bash
curl -X POST https://ВАШ-BACKEND.onrender.com/api/rooms \
  -H 'Content-Type: application/json' \
  -d '{"title":"Наши заметки","name":"Макс","color":"#e85d75"}'
```

Ожидается JSON с `room`, `token`, `code`.

---

## 3. Frontend на Vercel

1. https://vercel.com → импорт репозитория.
2. **Root Directory:** `fronted`
3. Env:

| Key | Value |
|-----|--------|
| `NEXT_PUBLIC_API_URL` | `https://ВАШ-BACKEND.onrender.com` (без `/` в конце) |

4. Deploy → URL вида `https://zametki.vercel.app`

---

## 4. CORS

В Render → Environment:

```
CORS_ORIGINS=https://zametki.vercel.app
```

Сохраните (сервис перезапустится).

---

## 5. Как пользоваться вдвоём

1. Открой сайт → **Создать пространство**.
2. Скопируй приглашение (кнопка с иконкой копирования).
3. Отправь девушке ссылку `/join/КОД`.
4. Она вводит имя и цвет — оба в одной комнате.

На iPhone: Safari → Поделиться → На экран «Домой».

---

## 6. Если не работает

| Симптом | Что проверить |
|---------|----------------|
| `postgres ping` / connection refused | `DATABASE_URL`, проект Neon не удалён, `sslmode=require` |
| Failed to fetch на фронте | `NEXT_PUBLIC_API_URL`, `CORS_ORIGINS` |
| Долго первый раз | Render Free просыпается — подождите минуту |
| ROOM_FULL | Уже 2 участника — создайте новую комнату |

Логи: Render → сервис → **Logs**.

---

## Локально

```bash
docker compose up -d postgres
cd backend && cp .env.example .env && go run ./cmd/server
cd fronted && npm run dev
```

---

## Чеклист

- [ ] Neon создан, `DATABASE_URL` скопирован
- [ ] Render Live, без `MONGO_*`
- [ ] Vercel с `NEXT_PUBLIC_API_URL`
- [ ] `CORS_ORIGINS` = URL Vercel
- [ ] Create room + join работает
- [ ] Идея появляется у партнёра без refresh
