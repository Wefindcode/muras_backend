## Social MVP (Golang)

Backend: PostgreSQL, JWT-авторизация админа, управление постами и пользователями, парсер RSS/Atom в фоне.

### Возможности
- БД: PostgreSQL
- Авторизация: `POST /admin/login` (JWT)
- Посты: `GET /posts`, `GET /posts/{id}`, `POST/PUT/DELETE /posts/{id}` (админ)
- Пользователи: `GET/POST /users` (админ)
- Ленты: `GET /feeds`, `POST/DELETE /feeds/{id}` (админ)
- Парсер: фоновая задача, раз в ~10 минут читает RSS/Atom из `/feeds` и создает посты

### Требования
- Go 1.22+
- PostgreSQL 13+
- (Опционально) Docker + Docker Compose

### Быстрый старт (Go + локальный Postgres)

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/social?sslmode=disable"
export PORT=8080
export ALLOW_CORS=true
export JWT_SECRET="dev-secret-change-me"
export DEFAULT_ADMIN_EMAIL="admin@example.com"
export DEFAULT_ADMIN_PASSWORD="admin123"

go run .
```

### Docker Compose (PostgreSQL)

```bash
docker compose up --build -d
curl http://localhost:8080/healthz
```

Сервисы:
- `db` — Postgres 16 (порт 5432), volume `pgdata`
- `app` — приложение на Go (порт 8080), использует `DATABASE_URL` c `db`

### Примеры API
```bash
# Логин
LOGIN=$(curl -s -X POST http://localhost:8080/admin/login -H "Content-Type: application/json" -d '{"email":"admin@example.com","password":"admin123"}')
TOKEN=$(echo "$LOGIN" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

# Создать пост
curl -s -X POST http://localhost:8080/posts -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"title":"Hello","content":"World"}'

# Список постов
curl -s http://localhost:8080/posts
```

### Переменные окружения
- `DATABASE_URL`: `postgres://user:pass@host:5432/dbname?sslmode=disable`
- `PORT`, `ALLOW_CORS`, `JWT_SECRET`, `DEFAULT_ADMIN_EMAIL`, `DEFAULT_ADMIN_PASSWORD`

### Примечания
- Миграции выполняются автоматически при старте.
- Воркер парсит ленты раз в ~10 минут.
- Перед продакшеном замените `JWT_SECRET` и пароли.