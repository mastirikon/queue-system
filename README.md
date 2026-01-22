# Queue System (Go + Asynq + Redis)

Лёгкая очередь задач на Go, использующая Redis и библиотеку Asynq.

## 🚀 Быстрый старт

### Запуск в Docker (рекомендуется)

```bash
# Собрать и запустить все сервисы
docker-compose up -d --build

# Проверить статус
docker-compose ps

# Посмотреть логи
docker-compose logs -f
```

Сервисы будут доступны:
- API: http://localhost:8080
- Redis: localhost:6379

### Локальный запуск

```bash
# 1. Запусти Redis
docker run -d -p 6379:6379 redis:7-alpine

# 2. Собери проект
make build

# 3. Запусти API (в одном терминале)
./bin/api

# 4. Запусти Worker (в другом терминале)
./bin/worker
```

## 📡 API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Создать задачу
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://httpbin.org/post",
    "method": "POST",
    "headers": {
      "X-Custom-Header": "test"
    },
    "body": "{\"test\": \"data\"}"
  }'
```

Ответ:
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Task created successfully"
}
```

## 🏗️ Архитектура

```
┌─────────┐      ┌─────────┐      ┌─────────┐      ┌──────────┐
│ Client  │─────>│   API   │─────>│  Redis  │<─────│  Worker  │
│         │      │ (Fiber) │      │ (Asynq) │      │  (Asynq) │
└─────────┘      └─────────┘      └─────────┘      └──────────┘
                      │                                   │
                      │                                   │
                      └───────────> Logs <───────────────┘
```

1. **API** принимает HTTP POST запрос с данными задачи
2. Задача помещается в **Redis** через Asynq
3. **Worker** достаёт задачу из очереди
4. Worker выполняет HTTP запрос на указанный URL
5. Если ответ 200 OK → задача удаляется
6. Если ошибка → retry через 10 секунд (макс. 24 часа)

## 🛠️ Makefile команды

```bash
make help          # Показать все команды
make build         # Собрать бинарники
make run-api       # Запустить API
make run-worker    # Запустить Worker
make docker-build  # Собрать Docker образы
make docker-up     # Запустить в Docker
make docker-down   # Остановить Docker
make docker-logs   # Посмотреть логи
make test          # Запустить тесты
make clean         # Очистить бинарники
```

## 📦 Структура проекта

```
queue-system/
├── cmd/
│   ├── api/          # API сервис
│   └── worker/       # Worker сервис
├── internal/
│   ├── config/       # Конфигурация
│   ├── domain/       # Модели данных
│   ├── handler/      # HTTP handlers
│   ├── queue/        # Asynq client
│   └── task/         # Task processor
├── pkg/
│   └── logger/       # Логгер
├── docker/
│   ├── api.Dockerfile
│   └── worker.Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## ⚙️ Конфигурация

Переменные окружения (смотри `.env.example`):

| Переменная | Описание | По умолчанию |
|-----------|----------|--------------|
| `ENV` | Окружение (development/production) | development |
| `API_PORT` | Порт API сервера | 8080 |
| `REDIS_ADDR` | Адрес Redis | localhost:6379 |
| `WORKER_CONCURRENCY` | Количество worker'ов | 10 |
| `WORKER_RETRY_INTERVAL` | Интервал retry | 10s |
| `WORKER_REQUEST_TIMEOUT` | Таймаут HTTP запроса | 30s |

## 🔧 Требования

- Go 1.25+
- Docker & Docker Compose (для контейнеризации)
- Redis (если запуск локальный)

## 📊 Мониторинг

Логи структурированы в JSON (production) или цветной вывод (development).

Примеры логов:
```json
{
  "level": "info",
  "timestamp": "2026-01-22T21:44:00Z",
  "msg": "Task enqueued successfully",
  "task_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## 🐛 Troubleshooting

### API не запускается
```bash
# Проверь, что порт 8080 свободен
lsof -i :8080

# Проверь логи
docker-compose logs api
```

### Worker не обрабатывает задачи
```bash
# Проверь подключение к Redis
docker-compose logs redis

# Проверь логи worker
docker-compose logs worker
```

### Задачи не удаляются
- Проверь, что целевой URL возвращает 200 OK
- Проверь логи worker для деталей ошибки

## 📝 Лицензия

MIT

## 👨‍💻 Автор

Queue System - Production-ready task queue на Go
```

---

## ✅ Порядок действий

### Шаг 1: Создай файлы
```bash
# Task processor
touch internal/task/processor.go

# Worker main
touch cmd/worker/main.go

# Docker
mkdir -p docker
touch docker/api.Dockerfile
touch docker/worker.Dockerfile

# Конфиги
touch docker-compose.yml
touch .dockerignore
touch .env.example
touch Makefile
```

### Шаг 2: Скопируй код
Открой каждый файл и скопируй код из этого задания.

### Шаг 3: Собери всё
```bash
# Собери Worker
go build -o bin/worker ./cmd/worker

# Проверь, что оба бинарника есть
ls -lh bin/
```

### Шаг 4: Запусти в Docker
```bash
# Собери образы
docker-compose build

# Запусти все сервисы
docker-compose up -d

# Проверь статус
docker-compose ps

# Посмотри логи
docker-compose logs -f
```

### Шаг 5: Тестируй API
```bash
# Health check
curl http://localhost:8080/health

# Создай задачу
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://httpbin.org/post",
    "method": "POST",
    "body": "{\"test\": \"hello\"}"
  }'
```

---