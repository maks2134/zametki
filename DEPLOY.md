# Как выложить Zametka в интернет бесплатно

Цель: вы и девушка открываете сайт с телефона, без локального компьютера.

Рекомендуемая связка (всё бесплатно):

| Часть | Сервис | Зачем |
|-------|--------|--------|
| База MongoDB | [MongoDB Atlas](https://www.mongodb.com/cloud/atlas) (бесплатный M0) | Хранение комнат и заметок |
| Backend (Go API + WS) | [Render](https://render.com) Free Web Service | API и realtime |
| Frontend (Next.js) | [Vercel](https://vercel.com) Hobby | Красивый сайт |

Ограничения бесплатных тарифов: Render может «засыпать» после простоя (~15 мин) — первый запрос после сна занимает 30–60 сек. Для двоих этого обычно хватает.

---

## 0. Подготовка

1. Аккаунты: GitHub, MongoDB Atlas, Render, Vercel (можно везде войти через GitHub).
2. Код должен быть в **GitHub-репозитории** (публичном или приватном).

Если репозитория ещё нет:

```bash
cd /Users/makskozlov/study/zametka
git init
git add .
git commit -m "Zametka: couple notes MVP"
# создайте пустой repo на GitHub, затем:
git remote add origin https://github.com/<ваш-логин>/zametka.git
git branch -M main
git push -u origin main
```

---

## 1. MongoDB Atlas (база)

1. Зайдите на https://www.mongodb.com/cloud/atlas → **Create** → бесплатный кластер **M0**.
2. Выберите регион ближе к вам (например Frankfurt / Netherlands).
3. **Database Access** → создайте пользователя с паролем (сохраните пароль).
4. **Network Access** → **Add IP Address** → **Allow Access from Anywhere** (`0.0.0.0/0`) — нужно, чтобы Render мог подключиться.
5. **Database** → **Connect** → **Drivers** → скопируйте URI вида:

```
mongodb+srv://USER:PASSWORD@cluster0.xxxxx.mongodb.net/?retryWrites=true&w=majority
```

Добавьте имя базы в URI:

```
mongodb+srv://USER:PASSWORD@cluster0.xxxxx.mongodb.net/zametka?retryWrites=true&w=majority
```

Пароль в URI не должен содержать спецсимволы без URL-encoding (`@`, `#` и т.п.).

---

## 2. Backend на Render

1. https://render.com → **New** → **Web Service** → подключите GitHub-репозиторий `zametka`.
2. Настройки:

| Поле | Значение |
|------|----------|
| Root Directory | `backend` |
| Runtime | Docker (или Native Go) |
| Instance | Free |

**Проще через Docker** (в репо уже есть `backend/Dockerfile`):

- **Language / Environment:** Docker
- **Root Directory:** `backend`
- **Dockerfile Path:** `./Dockerfile` (относительно Root Directory)

3. **Environment Variables** (Environment):

| Key | Value |
|-----|--------|
| `HTTP_ADDR` | `:8080` |
| `MONGO_URI` | ваш Atlas URI |
| `MONGO_DB` | `zametka` |
| `JWT_SECRET` | длинная случайная строка ≥ 32 символов |
| `JWT_TTL` | `720h` |
| `CORS_ORIGINS` | пока `http://localhost:3000` — **обновите после Vercel** |

Сгенерировать секрет:

```bash
openssl rand -base64 32
```

4. Deploy → дождитесь статуса **Live**.
5. На Render обычно приходит переменная `PORT` — backend её подхватит сам (можно не задавать `HTTP_ADDR`).
6. Скопируйте URL сервиса, например: `https://zametka-api.onrender.com`

Проверка:

```bash
curl https://zametka-api.onrender.com/api/rooms
# ожидается 405 или ошибка валидации — главное не «connection refused»
```

Создать комнату вручную:

```bash
curl -X POST https://ВАШ-BACKEND.onrender.com/api/rooms \
  -H 'Content-Type: application/json' \
  -d '{"title":"Наши заметки","name":"Макс","color":"#e85d75"}'
```

Должны вернуться `room`, `token`, `code`.

---

## 3. Frontend на Vercel

1. https://vercel.com → **Add New Project** → импорт репозитория `zametka`.
2. Настройки:

| Поле | Значение |
|------|----------|
| Root Directory | `fronted` |
| Framework | Next.js (определится сам) |

3. **Environment Variables**:

| Key | Value |
|-----|--------|
| `NEXT_PUBLIC_API_URL` | `https://ВАШ-BACKEND.onrender.com` (без слэша в конце) |

4. Deploy.
5. Скопируйте URL сайта, например: `https://zametka.vercel.app`

---

## 4. Связать CORS (важно)

Вернитесь в Render → backend → Environment:

```
CORS_ORIGINS=https://zametka.vercel.app
```

Если будете использовать несколько URL (с `www` и без):

```
CORS_ORIGINS=https://zametka.vercel.app,https://www.zametka.vercel.app
```

Сохраните → Render перезапустит сервис.

---

## 5. Как пользоваться вдвоём

1. Ты открываешь `https://zametka.vercel.app` → **Создать пространство**.
2. Жмёшь кнопку с кодом (иконка копирования) — копируется ссылка вида  
   `https://zametka.vercel.app/join/K7M4QP`
3. Отправляешь ссылку девушке (Telegram / iMessage).
4. Она вводит имя и цвет → вы оба в одной комнате.
5. Идеи и реакции появляются у обоих почти сразу (WebSocket).

На iPhone удобно: Safari → **Поделиться** → **На экран «Домой»** — будет как мини-приложение.

---

## 6. Если что-то не работает

| Симптом | Что проверить |
|---------|----------------|
| На фронте «Failed to fetch» | `NEXT_PUBLIC_API_URL`, CORS_ORIGINS, что backend Live |
| Долго грузится первый раз | Render Free «просыпается» — подождите ~1 мин и обновите |
| Join → ROOM_FULL | В комнате уже 2 человека; создайте новое пространство |
| WS offline | Откройте сайт по **https**, API тоже **https** (тогда будет `wss://`) |
| Mongo auth failed | Пароль/URI Atlas, IP Access List = `0.0.0.0/0` |

Логи backend: Render → ваш сервис → **Logs**.

---

## 7. Альтернативы (тоже бесплатно / почти)

- **Backend:** [Fly.io](https://fly.io) или [Railway](https://railway.app) (есть бесплатные кредиты).
- **Frontend:** Vercel остаётся лучшим выбором для Next.js.
- **Всё в одном:** можно держать frontend+backend на Railway, но Vercel+Render проще для старта.

---

## Чеклист «готово»

- [ ] Atlas кластер создан, URI работает
- [ ] Backend на Render отвечает по https
- [ ] Frontend на Vercel открывается
- [ ] `NEXT_PUBLIC_API_URL` = URL Render
- [ ] `CORS_ORIGINS` = URL Vercel
- [ ] Создал комнату, отправил join-ссылку, второй человек вошёл
- [ ] Новая идея появилась у партнёра без обновления страницы
