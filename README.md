# Media Converter

Распределенная система асинхронной обработки медиа-файлов.

## Стек технологий

- **Backend:** Go 1.25.5
- **Broker:** Apache Kafka (KRaft mode)
- **Database:** PostgreSQL 15, Redis 7
- **Gateway:** Nginx (rate limiting, load balancing)
- **Monitoring:** Prometheus, Grafana
- **Logging:** Uber Zap (structured JSON logging)
- **Containers:** Docker, Docker Compose

## Архитектура

```mermaid
graph TB
    Client[Client] -->|HTTP| Nginx[Nginx API Gateway]
    Nginx -->|Load Balancing| API1[API Service 1]
    Nginx -->|Load Balancing| API2[API Service 2]
    Nginx -->|Static Files| Static[Static Files]

    API1 -->|Publish Task| Kafka[Kafka Broker]
    API2 -->|Publish Task| Kafka

    API1 -->|Write| PG[(PostgreSQL)]
    API1 -->|Cache/Read| Redis[(Redis)]
    API2 -->|Write| PG
    API2 -->|Cache/Read| Redis

    Kafka -->|Consume Task| Worker[Worker Service]
    Worker -->|Update Status| PG
    Worker -->|Update Cache| Redis
    Worker -->|Process Image| Uploads["/uploads"]

    subgraph Monitoring
        Prometheus[Prometheus]
        Grafana[Grafana]
    end

    API1 -->|Metrics| Prometheus
    API2 -->|Metrics| Prometheus
    Worker -->|Metrics| Prometheus
    Prometheus --> Grafana
```

## Текущий статус

### Реализовано

**Инфраструктура:**
- [x] Docker Compose конфигурация всех сервисов
- [x] Nginx (API Gateway) с rate limiting и load balancing
- [x] PostgreSQL 15 с connection pooling и healthcheck
- [x] Redis 7 с healthcheck
- [x] Kafka (KRaft mode) с Kafka UI
- [x] Prometheus, Grafana

**API Service:**
- [x] Database миграции
- [x] Repository pattern с интерфейсами
- [x] Redis cache для статусов задач
- [x] POST /upload - загрузка файлов (валидация размера, типа)
- [x] GET /status/:id - проверка статуса
- [x] Kafka Producer
- [x] Middleware: TraceID, Logging, Recovery
- [x] Graceful shutdown
- [x] Статический фронтенд для тестирования

**Worker Service:**
- [x] Kafka Consumer
- [x] Processor с обновлением статуса в БД и Redis
- [x] Graceful shutdown
- [x] Flow: pending → processing → completed
- [x] Worker pool
- [x] Prometheus метрики
- [x] Magic bytes проверка файлов
- [x] Полноценное тестирование
- [x] Документация API
- [ ] Rate limiting per user

## Запуск

```bash
# Клонировать и запустить все сервисы
docker compose up -d

# Применить миграции БД
make migrate-up

```

### Доступные сервисы

| Сервис | URL | Описание |
|--------|-----|----------|
| API Gateway | http://localhost | Фронтенд + API |
| API Service 1 | http://localhost:8081 | Инстанс 1 |
| API Service 2 | http://localhost:8082 | Инстанс 2 (load balancing) |
| Kafka UI | http://localhost:8080 | Управление Kafka |
| Prometheus | http://localhost:9090 | Метрики |
| Grafana | http://localhost:3000 | Дашборды (admin/admin) |
| PostgreSQL | localhost:5432 | База данных (user/password) |
| Redis | localhost:6379 | Кэш |
| Kafka | localhost:9092 | Message broker |

## API Endpoints

### POST /upload - Загрузка файла

Загружает файл для обработки и возвращает ID задачи.

**Параметры формы:**
- `file` (обязательно): Файл для обработки (JPEG, PNG, GIF, PDF, MP4)
- `output_format` (опциональ): Формат вывода (jpg, png)
- `target_width` (опциональ): Целевая ширина в пикселях
- `target_height` (опциональ): Целевая высота в пикселях
- `crop` (опциональ): Обрезка по центру (true/false)

**Примеры:**

Базовая загрузка:
```bash
curl -X POST http://localhost/upload \
  -F "file=@image.jpg" \
  -v
```

С параметрами конвертации:
```bash
curl -X POST http://localhost/upload \
  -F "file=@image.jpg" \
  -F "output_format=png" \
  -F "target_width=800" \
  -F "target_height=600" \
  -F "crop=true" \
  -v
```

**Успешный ответ (201):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "trace_id": "dca9a48326bd1a644a51239432795179",
  "status": "pending",
  "original_filename": "image.jpg",
  "output_format": "png",
  "target_width": 800,
  "target_height": 600,
  "crop": true,
  "created_at": "2026-02-07T18:00:00Z"
}
```

**Ошибка (400/500):**
```json
{
  "error": "Invalid file",
  "trace_id": "dca9a48326bd1a644a51239432795179"
}
```

### GET /status/:id - Проверка статуса

Возвращает текущий статус задачи обработки.

**Пример:**
```bash
curl http://localhost/status/550e8400-e29b-41d4-a716-446655440000
```

**Ответ при статусе pending (200):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "trace_id": "dca9a48326bd1a644a51239432795179",
  "status": "pending",
  "original_filename": "image.jpg",
  "created_at": "2026-02-07T18:00:00Z"
}
```

**Ответ при статусе completed (200):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "trace_id": "dca9a48326bd1a644a51239432795179",
  "status": "completed",
  "original_filename": "image.jpg",
  "output_filename": "550e8400-e29b-41d4-a716-446655440000.png",
  "created_at": "2026-02-07T18:00:00Z",
  "completed_at": "2026-02-07T18:00:03Z"
}
```

**Ответ при ошибке (404):**
```json
{
  "error": "Task not found",
  "trace_id": "dca9a48326bd1a644a51239432795179"
}
```

**Жизненный цикл задачи:**
`pending` → `processing` → `completed` / `failed`

### GET /download/:filename - Скачивание обработанного файла

Скачивает обработанный файл.

**Пример:**
```bash
curl -O http://localhost/download/550e8400-e29b-41d4-a716-446655440000.png
```

### GET /health - Health check

Проверка здоровья сервиса.

**Пример:**
```bash
curl http://localhost/health
```

**Ответ (200):**
```json
{
  "status": "healthy"
}
```

## Миграции базы данных

```bash
# Применить миграции
make migrate-up

# Откатить миграции
make migrate-down
```

Или вручную:
```bash
docker run --rm -v $(PWD)/api/database/migrations:/migrations \
  --network mediaconverter_default \
  migrate/migrate -path /migrations \
  -database "postgres://user:password@postgres:5432/mediadb?sslmode=disable" up
```

## Мониторинг

### Prometheus
http://localhost:9090

### Grafana
http://localhost:3000
- Login: admin
- Password: admin

### Kafka UI
http://localhost:8080
- Топик: media_tasks
- Consumer Group: worker-group

## Логирование

Все сервисы используют структурированное JSON логирование с Uber Zap:

```json
{
  "level": "info",
  "ts": 1770487798.4077508,
  "caller": "worker/main.go:59",
  "msg": "Processing task",
  "task_id": "16f26fa1-deb9-417c-b446-7ca7e7dd425b",
  "trace_id": "dca9a48326bd1a644a51239432795179"
}
```

**TraceID propagation:**
1. Nginx генерирует request_id
2. Передается через X-Trace-ID header
3. Проходит через весь flow: API → Kafka → Worker

## Принципы разработки

- Интерфейсы вместо реализаций
- Dependency Injection через конструкторы
- Graceful Shutdown для всех сервисов
- Context.Context с таймаутами
- Connection Pooling (PostgreSQL, Redis)
- Structured Logging с TraceID
- Repository pattern
- Middleware chain

## Безопасность

- [x] Валидация размера файла (100MB max)
- [x] Валидация типа по расширению (.jpg, .png, .gif, .pdf, .mp4)
- [x] Санитизация имен файлов (filepath.Base)
- [x] Именованные параметры в SQL (pgx)
- [x] Переменные окружения для секретов
- [X] Magic bytes проверка

## Graceful Shutdown

Все сервисы корректно обрабатывают SIGINT/SIGTERM:

1. Остановить прием новых запросов
2. Дождаться завершения текущих (30s timeout)
3. Закрыть соединения с БД, Redis, Kafka
4. Завершить процесс

```bash
docker compose stop api-service  # Graceful shutdown за 30s
```

## Тесты

Проект содержит unit тесты для основных компонентов:

```bash
# Запустить тесты worker service
cd worker && go test ./... -v

# Запустить тесты API service
cd api && go test ./... -v

# Запустить тесты с покрытием
cd api && go test ./... -cover

# Запустить конкретный пакет
cd worker && go test ./converter/... -v
```

**Тестовое покрытие:**
- `worker/converter` - тесты конвертера изображений (resize, crop, format conversion)
- `api/handlers` - тесты HTTP обработчиков (upload, status)


### Healthcheck
```bash
curl http://localhost/health
# Или напрямую к API
curl http://localhost:8081/health
```
