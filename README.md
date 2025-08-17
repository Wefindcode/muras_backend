## Social MVP (Golang)

Минимально жизнеспособный backend: SQLite/PostgreSQL, JWT-авторизация админа, управление постами и пользователями, парсер RSS/Atom в фоне.

### Возможности
- БД: SQLite (по умолчанию) или PostgreSQL
- Авторизация: `POST /admin/login` (JWT)
- Посты: `GET /posts`, `GET /posts/{id}`, `POST/PUT/DELETE /posts/{id}` (админ)
- Пользователи: `GET/POST /users` (админ)
- Ленты: `GET /feeds`, `POST/DELETE /feeds/{id}` (админ)
- Парсер: фоновая задача, раз в ~10 минут читает RSS/Atom из `/feeds` и создает посты

### Требования
- Go 1.22+
- (Опционально) Docker + Docker Compose

### Быстрый старт (Go, SQLite)

```bash
export DB_DRIVER=sqlite
export DATABASE_URL="file:app.db?_foreign_keys=on"
export PORT=8080
export ALLOW_CORS=true
export JWT_SECRET="dev-secret-change-me"
export DEFAULT_ADMIN_EMAIL="admin@example.com"
export DEFAULT_ADMIN_PASSWORD="admin123"

go run .
```

### Быстрый старт (Go, PostgreSQL)

```bash
# Пример DSN: postgres://user:pass@localhost:5432/dbname?sslmode=disable
export DB_DRIVER=postgres
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/social?sslmode=disable"
export PORT=8080
export ALLOW_CORS=true
export JWT_SECRET="dev-secret-change-me"
export DEFAULT_ADMIN_EMAIL="admin@example.com"
export DEFAULT_ADMIN_PASSWORD="admin123"

go run .
```

### Docker Compose (SQLite по умолчанию)

```bash
docker compose up --build -d
curl http://localhost:8080/healthz
```

### Docker Compose с PostgreSQL

Создайте `docker-compose.override.yml`:

```yaml
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: social
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
  app:
    environment:
      DB_DRIVER: postgres
      DATABASE_URL: postgres://postgres:postgres@db:5432/social?sslmode=disable
    depends_on:
      - db
volumes:
  pgdata:
```

Запуск:
```bash
docker compose -f docker-compose.yml -f docker-compose.override.yml up --build -d
```

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
- `DB_DRIVER`: `sqlite` (по умолчанию) или `postgres`
- `DATABASE_URL`: DSN
  - SQLite: `file:app.db?_foreign_keys=on` (локально) или `file:/data/app.db?_foreign_keys=on` (Docker)
  - PostgreSQL: `postgres://user:pass@host:5432/dbname?sslmode=disable`
- `PORT`, `ALLOW_CORS`, `JWT_SECRET`, `DEFAULT_ADMIN_EMAIL`, `DEFAULT_ADMIN_PASSWORD`

### Примечания
- Миграции выполняются автоматически при старте для выбранного драйвера.
- Для PostgreSQL используется автоматическая конвертация плейсхолдеров `?` -> `$1..$n`.
- Воркер парсит ленты раз в ~10 минут.
- Перед продакшеном замените `JWT_SECRET` и пароли.