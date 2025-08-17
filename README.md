## Social MVP (Golang)

Минимально жизнеспособный backend: SQLite, JWT-авторизация админа, управление постами и пользователями, парсер RSS/Atom в фоне.

### Возможности
- БД: SQLite (файл `app.db` локально, `/data/app.db` в контейнере)
- Авторизация: `POST /admin/login` (JWT)
- Посты: `GET /posts`, `GET /posts/{id}`, `POST/PUT/DELETE /posts/{id}` (админ)
- Пользователи: `GET/POST /users` (админ)
- Ленты: `GET /feeds`, `POST/DELETE /feeds/{id}` (админ)
- Парсер: фоновая задача, раз в ~10 минут читает RSS/Atom из `/feeds` и создает посты

### Требования
- Go 1.22+
- (Опционально) Docker + Docker Compose

### Быстрый старт (Go)

```bash
# В корне проекта
export PORT=8080
export ALLOW_CORS=true
export DATABASE_URL="file:app.db?_foreign_keys=on"
export JWT_SECRET="dev-secret-change-me"
export DEFAULT_ADMIN_EMAIL="admin@example.com"
export DEFAULT_ADMIN_PASSWORD="admin123"

go run .
# сервер слушает :8080
```

Проверка:
```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```

Авторизация и вызовы API:
```bash
# Вход админа (см. env по умолчанию)
LOGIN=$(curl -s -X POST http://localhost:8080/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}')
TOKEN=$(echo "$LOGIN" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

# Создать пост (админ)
curl -s -X POST http://localhost:8080/posts \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"title":"Hello","content":"World"}'

# Список постов (публично)
curl -s http://localhost:8080/posts

# Создать пользователя (админ)
curl -s -X POST http://localhost:8080/users \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"email":"user1@example.com","password":"pass123","is_admin":false}'

# Добавить ленту (админ)
curl -s -X POST http://localhost:8080/feeds \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"url":"https://news.ycombinator.com/rss"}'
```

### Быстрый старт (Docker Compose)

```bash
docker compose up --build -d
curl http://localhost:8080/healthz
```

Переменные окружения можно переопределять в `docker-compose.yml`.
Данные SQLite сохраняются в volume `appdata:/data`.

### Быстрый старт (Docker без Compose)

```bash
docker build -t social-mvp:latest .
docker run --rm -p 8080:8080 -v social_appdata:/data \
  -e JWT_SECRET="change-me-in-prod" \
  -e DEFAULT_ADMIN_EMAIL="admin@example.com" \
  -e DEFAULT_ADMIN_PASSWORD="admin123" \
  --name social-mvp social-mvp:latest
```

### Переменные окружения
- `PORT`: порт HTTP (по умолчанию `8080`)
- `ALLOW_CORS`: `true/false` (по умолчанию `true`)
- `DATABASE_URL`: SQLite DSN
  - локально: `file:app.db?_foreign_keys=on`
  - в Docker: `file:/data/app.db?_foreign_keys=on`
- `JWT_SECRET`: секрет для подписи JWT (обязательно смените в проде)
- `DEFAULT_ADMIN_EMAIL`, `DEFAULT_ADMIN_PASSWORD`: создаются при первом запуске, если ещё нет админа

### Примечания
- Миграции выполняются автоматически при старте.
- Воркер парсит ленты раз в ~10 минут.
- CORS включен по умолчанию для удобства разработки.
- Безопасность: замените `JWT_SECRET` и пароль админа перед публикацией.