# DelayedNotifier

`DelayedNotifier` — сервис отложенных уведомлений с HTTP API.
Он принимает запрос на отправку сообщения "в будущее", сохраняет запись, публикует задачу в RabbitMQ и затем отправляет уведомление по нужному каналу (`mail` / `tg`) в заданное время.

Проект сделан как учебный/практический backend-проект с упором на Go, инфраструктуру и интеграцию нескольких сервисов:
`PostgreSQL + Redis + RabbitMQ + HTTP API + senders`.

## Цель проекта

- Принять уведомление через HTTP API.
- Сохранить/закешировать его состояние.
- Отложенно доставить задачу через RabbitMQ.
- Отправить сообщение через выбранный канал.
- Дать API для получения статуса и удаления уведомления.

## Как это работает

Базовый поток обработки:

1. Клиент отправляет `POST /notify`.
2. HTTP-слой валидирует payload и собирает `models.Record`.
3. `notify.Notifier` сохраняет запись и публикует отложенную задачу в RabbitMQ.
4. Worker читает задачи из очереди и передает их в слой отправителей (`sender`).
5. `sender` выбирает нужный канал (`mail` или `tg`) и отправляет сообщение.
6. Статус можно запросить через `GET /notify/{id}/`.

## Технологии

- Go `1.24`
- HTTP router: `chi`
- PostgreSQL: `pgx/v5`
- Redis: `go-redis/v9`
- RabbitMQ: `amqp091-go`
- Telegram Bot API: `telegram-bot-api/v5`
- SMTP: `gomail.v2`
- Config loading: `cleanenv`
- Validation: `go-playground/validator`
- Retry: `wb-go/wbf/retry`
- Docker / Docker Compose

## Архитектура (по папкам)

- `cmd/` — точка входа приложения
- `config/` — конфигурация и валидация (`cleanenv`)
- `internal/app/` — сборка приложения, DI, lifecycle, graceful shutdown
- `internal/web/` — HTTP сервер, handlers, static UI
- `internal/models/` — доменные модели
- `internal/service/notify/` — usecase/orchestrator уведомлений
- `internal/service/rabbit/` — RabbitMQ клиент и работа с очередями
- `internal/service/sender/` — маршрутизация отправки по каналам
- `internal/service/sender/mailsender/` — отправка email
- `internal/service/sender/tgsender/` — отправка Telegram
- `internal/repository/postgres/` — PostgreSQL repository
- `internal/cache/` — Redis cache-слой
- `docker/` — инфраструктурные файлы (например, init для Postgres)

## Возможности (текущее API)

- `POST /notify` — создать отложенное уведомление
- `GET /notify/{id}/` — получить статус уведомления
- `DELETE /notify/{id}/` — удалить уведомление

UI со статикой отдается с того же HTTP сервера (по корневому пути `/`).

## Пример запроса

`POST /notify`

```json
{
  "message": "Напомнить о дедлайне",
  "dateTime": "2026-02-25T15:30:00Z",
  "sendChan": "mail",
  "from": "no-reply@example.com",
  "to": ["user@example.com"]
}
```

`sendChan` поддерживает:

- `mail`
- `tg`

## Конфигурация

Приложение читает конфигурацию из `.env` (через `cleanenv`).
Шаблон находится в `.env.example`.

Основные группы переменных:

- HTTP (`httpHost`, `httpPort`, `staticFilesPath`)
- SMTP (`EMAIL_SMTP_*`)
- Telegram (`TG_SENDER_BOT_TOKEN`)
- RabbitMQ (`AMQP_*`)
- Retry (`PUBLISH_RETRY_*`, `SEND_RETRY_*`)
- PostgreSQL (`pg*`)
- Redis (`REDIS_*`)
- MailHog (только для тестов/локальных утилит)

## Быстрый старт (Docker Compose)

### 1. Подготовить `.env`

PowerShell:

```powershell
Copy-Item .env.example .env
```

Заполни минимум:

- `EMAIL_SMTP_HOST`
- `EMAIL_SMTP_PORT`
- `EMAIL_SMTP_USER`
- `EMAIL_SMTP_PASSWORD`
- `TG_SENDER_BOT_TOKEN` (можно оставить пустым, если нужен только email)

### 2. Запустить весь стек

```bash
docker compose up --build
```

Что поднимется:

- `app` — приложение (`http://localhost:8080`)
- `postgres` — PostgreSQL (`localhost:5432`)
- `redis` — Redis (`localhost:6379`)
- `rabbitmq` — RabbitMQ (`localhost:5672`)
- RabbitMQ Management UI — `http://localhost:15672` (`guest/guest`)

## Локальный запуск (без контейнера приложения)

Если хочешь запускать Go-приложение локально, но инфраструктуру оставить в Docker:

```bash
docker compose up -d postgres redis rabbitmq
go run ./cmd
```

## Тесты

В проекте разделены обычные тесты и интеграционные тесты через build tag `integration`.

### Обычные тесты (без интеграции)

```bash
go test ./...
```

Или через `go generate`:

```bash
go generate -run testunit ./...
```

### Интеграционные тесты (MailHog)

Интеграционный тест для `mailsender` вынесен под тег `integration`:

- `internal/service/sender/mailsender/mail_sender_integration_test.go`

Этот тест:

- проверяет доступность MailHog API
- отправляет реальные письма через SMTP MailHog
- читает сообщения из MailHog API и проверяет доставку

#### Запуск MailHog

Вариант 1 (через `make`):

```bash
make mailhog
```

Вариант 2 (docker напрямую):

```bash
docker run --rm --name mailhog -p 1025:1025 -p 8025:8025 mailhog/mailhog
```

MailHog UI будет доступен на `http://localhost:8025`.

#### Запуск интеграционных тестов

```bash
go test -tags=integration ./internal/service/sender/mailsender
```

Или через `go generate`:

```bash
go generate -run testintegration ./...
```

## Полезные команды

Запуск интеграционных тестов c автоподнятием/остановкой MailHog (через `make`):

```bash
make integration-test
```

## Точка входа приложения

Основная точка входа:

- `cmd/main.go`

Последовательность запуска:

1. `config.LoadConfig()`
2. `app.New(cfg)`
3. `application.Run(context.Background())`

## Текущий статус проекта

Проект в активной разработке. Основной каркас и ключевые интеграции уже есть:

- HTTP API
- конфигурация через `.env`
- интеграция с PostgreSQL / Redis / RabbitMQ
- email sender
- Telegram sender (опционально, если задан токен)
- интеграционные тесты для email через MailHog

## Что можно улучшить дальше

- CI (GitHub Actions) с отдельными job для `unit` и `integration`
- миграции БД (если планируется production lifecycle)
- единый формат API-ответов (JSON для всех endpoints)
- observability (метрики, tracing, structured logs)
- swagger/openapi для HTTP API
