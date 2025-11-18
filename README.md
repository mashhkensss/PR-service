# Backend-trainee-assignment-autumn-2025

## Структура

- `cmd/reviewer-service` — точка входа, graceful shutdown.
- `internal/app` — сборка зависимостей, wiring middleware/handlers.
- `internal/http` — chi-router, DTO, middleware (auth, rate limit, idempotency, validator) и хендлеры (`team`, `user`, `pullrequest`, `stats`, `health`).
- `internal/service` — бизнес-логика (teams/users/pr/stats, стратегия назначения, tx-runner).
- `internal/domain` — сущности, value objects и sentinels.
- `internal/persistence/postgres` — репозитории поверх `database/sql`, миграции в `migrations/`.
- `configs/.env.example` — пример `.env`.
- `openapi.yml` — контракт HTTP API.

## Запуск

### Docker Compose

1. Запустить сервисы (Postgres + миграции + приложение):
   ```bash
   docker compose up --build
   ```
    или
    ```bash
   docker-compose up
    ```
   Compose автоматически использует дефолтные переменные из `configs/.env.example`, так что дополнительных шагов после клонирования не требуется.
2. Проверить готовность сервиса:
   ```bash
   curl http://localhost:8080/health/ready
   ```
   Liveness доступен по `GET /health/live`.

### Основные переменные

| Переменная | Описание |
| --- | --- |
| `DB_DSN` | DSN Postgres (`postgres://user:pass@host:5432/db?sslmode=disable`) |
| `DB_CONNECT_RETRIES` | Количество попыток подключения к БД перед ошибкой (по умолчанию 5) |
| `DB_CONNECT_RETRY_INTERVAL` | Интервал между попытками подключения к БД (по умолчанию 1s) |
| `ADMIN_SECRET` / `USER_SECRET` | HMAC-секреты для JWT (roles: `admin`, `user`) |
| `HTTP_*` | Настройки порта/таймаутов |
| `RATE_LIMIT_*` | Токен-бакет на IP |
| `RATE_LIMIT_TRUST_FORWARD` | Доверять ли заголовкам `X-Forwarded-For`/`X-Real-IP` (true/false) |
| `IDEMPOTENCY_TTL` | TTL записей Idempotency-Key |

## Тесты

```bash
make test        # go test ./...
make cover       # сбор покрытия
```

## Observability & Health

- `/health/live` — процесс жив.
- `/health/ready` — проверяет доступность БД (`PingContext`).
- `/stats/assignments` и `/stats/summary` возвращают детальную и агрегированную статистику назначений ревьюеров.
