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

```
┌─────────┐     ┌────────────┐     ┌─────────┐
│ Client  │───▶│   Nginx    │───▶│  API    │
└─────────┘     └────────────┘     └────┬────┘
                     │                  │
                     │                  ▼
                     │            ┌───────────┐
                     │            │  Kafka    │
                     │            └─────┬─────┘
                     │                  │
                     │            ┌─────▼─────┐
                     │            │  Worker   │
                     │            └─────┬─────┘
                     │                  │
                     ▼            ┌─────▼─────┐
                ┌────────┐        │PostgreSQL │
                │ Static │        └───────────┘
                └────────┘             ▲
                                     ┌─┴───┐
                                     │Redis│
                                     └─────┘
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
- [x] Flow: pending → processing → completed (на данный момент стоит заглушка)

### Не реализовано
- [ ] Worker pool
- [ ] Prometheus метрики
- [ ] Magic bytes проверка файлов
- [ ] Retry policies
- [ ] Полноценное тестирование
- [ ] Документация API
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

### Загрузка файла

```bash
curl -X POST http://localhost/upload \
  -F "file=@image.jpg" \
  -v
```

**Ответ:**
```json
{
  "id": "uuid-task-id",
  "trace_id": "uuid-trace-id",
  "status": "pending",
  "original_filename": "image.jpg",
  "created_at": "2026-02-07T18:00:00Z"
}
```

### Проверка статуса

```bash
curl http://localhost/status/<task_id>
```

**Ответ:**
```json
{
  "id": "uuid-task-id",
  "trace_id": "uuid-trace-id",
  "status": "completed",
  "original_filename": "image.jpg",
  "created_at": "2026-02-07T18:00:00Z",
  "completed_at": "2026-02-07T18:00:03Z"
}
```

**Статусы:** `pending` → `processing` → `completed` / `failed`

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
- [ ] Magic bytes проверка (TODO)

## Graceful Shutdown

Все сервисы корректно обрабатывают SIGINT/SIGTERM:

1. Остановить прием новых запросов
2. Дождаться завершения текущих (30s timeout)
3. Закрыть соединения с БД, Redis, Kafka
4. Завершить процесс

```bash
docker compose stop api-service  # Graceful shutdown за 30s
```

## Troubleshooting

### Диск заполнен
```bash
# Очистить неиспользуемые Docker ресурсы
docker system prune -a --volumes -f
```

### Kafka не запускается
```bash
# Пересоздать Kafka volume
docker compose down kafka
docker volume rm mediaconverter_kafka_data
docker compose up -d kafka
```

### Проверить логи
```bash
# Все сервисы
docker compose logs -f

# Конкретный сервис
docker compose logs -f api-service
docker compose logs -f worker-service
```

### Healthcheck
```bash
curl http://localhost/health
# Или напрямую к API
curl http://localhost:8081/health
```

## Production TODO

- [ ] Prometheus метрики (request duration, error rate, queue size)
- [ ] Worker pool с ограничением concurrency
- [ ] Retry policies для Kafka consumer
- [ ] Magic bytes проверка файлов
- [ ] Graceful degradation при падении Kafka
- [ ] Индексы в PostgreSQL
- [ ] OpenAPI/Swagger документация
- [ ] Unit и integration тесты
- [ ] CI/CD pipeline
- [ ] Configuration management
- [ ] Distributed tracing (Jaeger/Zipkin)
