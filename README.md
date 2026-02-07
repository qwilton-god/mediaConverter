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

- [x] Docker Compose конфигурация всех сервисов
- [x] Nginx (API Gateway) с rate limiting и load balancing
- [x] PostgreSQL с healthcheck
- [x] Redis с healthcheck
- [x] Kafka (KRaft mode без Zookeeper)
- [x] Kafka UI для управления топиками
- [x] Prometheus для метрик
- [x] Grafana для дашбордов
- [x] Базовые Dockerfile для API и Worker сервисов
- [x] Минимальные main.go для запуска контейнеров

### В разработке

- [ ] Database миграции и repository layer
- [ ] API endpoints (POST /upload, GET /status)
- [ ] Worker с Kafka consumer
- [ ] Middleware (TraceID, Logging, Recovery, Metrics)
- [ ] Валидация файлов (magic bytes, размер, path traversal)
- [ ] Graceful shutdown
- [ ] Полное логирование и метрирование

## Запуск

```bash
docker compose up -d
```

### Доступные сервисы

| Сервис | URL | credentials |
|--------|-----|-------------|
| API Gateway | http://localhost | - |
| API Service 1 | http://localhost:8081 | - |
| API Service 2 | http://localhost:8082 | - |
| Kafka UI | http://localhost:8080 | - |
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin/admin |
| PostgreSQL | localhost:5432 | user/password |
| Redis | localhost:6379 | - |
| Kafka | localhost:9092 | - |

## API Endpoints (план)

- `POST /api/upload` - загрузка файла для обработки
- `GET /api/status/:task_id` - проверка статуса задачи
- `GET /static/:filename` - скачивание обработанного файла

## Мониторинг

### Prometheus
- http://localhost:9090

### Grafana
- http://localhost:3000
- Login: admin
- Password: admin

## Требования к качеству кода

- **Интерфейсы** вместо реализаций
- **Dependency Injection** через конструкторы
- **Graceful Shutdown** для всех сервисов
- **Context.Context** с таймаутами для всех внешних запросов
- **Connection Pooling** для БД
- **Structured Logging** с TraceID

## Безопасность

- Валидация типов файлов (magic bytes)
- Лимит размера загружаемых файлов
- Санитизация имен файлов (path traversal protection)
- Именованные параметры в SQL (защита от инъекций)
- Переменные окружения для секретов