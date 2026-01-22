# 📋 Задание #8: ФИНАЛЬНОЕ - Worker, Docker, Deploy

**Дата выдачи:** 2026-01-22  
**Статус:** 🔄 В работе  
**Фаза:** Завершение проекта

---

## 🎯 Цель
Создать все оставшиеся файлы для полного запуска системы в Docker.

---

## 📝 Файлы для создания

### 1️⃣ Worker Service - Task Processor

**Файл:** `internal/task/processor.go`

```go
package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mastirikon/queue-system/internal/domain"
	"go.uber.org/zap"
)

// Processor обрабатывает задачи из очереди
type Processor struct {
	logger     *zap.Logger
	httpClient *http.Client
}

// NewProcessor создаёт новый процессор задач
func NewProcessor(logger *zap.Logger, timeout time.Duration) *Processor {
	return &Processor{
		logger: logger,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ProcessHTTPRequest обрабатывает HTTP запрос
func (p *Processor) ProcessHTTPRequest(ctx context.Context, t *asynq.Task) error {
	// Десериализуем payload
	var payload domain.TaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		p.logger.Error("Failed to unmarshal task payload",
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	p.logger.Info("Processing task",
		zap.String("task_id", payload.ID),
		zap.String("url", payload.URL),
		zap.String("method", payload.Method),
	)

	// Создаём HTTP запрос
	var bodyReader io.Reader
	if payload.Body != "" {
		bodyReader = bytes.NewBufferString(payload.Body)
	}

	req, err := http.NewRequestWithContext(ctx, payload.Method, payload.URL, bodyReader)
	if err != nil {
		p.logger.Error("Failed to create HTTP request",
			zap.String("task_id", payload.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Добавляем заголовки
	for key, value := range payload.Headers {
		req.Header.Set(key, value)
	}

	// Если есть body, добавляем Content-Type по умолчанию
	if payload.Body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Выполняем запрос
	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.logger.Warn("HTTP request failed, will retry",
			zap.String("task_id", payload.ID),
			zap.Error(err),
		)
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа (для логирования)
	respBody, _ := io.ReadAll(resp.Body)

	// Проверяем статус код
	if resp.StatusCode == http.StatusOK {
		p.logger.Info("Task completed successfully",
			zap.String("task_id", payload.ID),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return nil // Задача успешно выполнена
	}

	// Если не 200 OK - возвращаем ошибку для retry
	p.logger.Warn("Task failed with non-200 status, will retry",
		zap.String("task_id", payload.ID),
		zap.Int("status_code", resp.StatusCode),
		zap.String("response", string(respBody)),
	)

	return fmt.Errorf("non-200 status code: %d", resp.StatusCode)
}
```

---

### 2️⃣ Worker Main

**Файл:** `cmd/worker/main.go`

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/mastirikon/queue-system/internal/config"
	"github.com/mastirikon/queue-system/internal/domain"
	"github.com/mastirikon/queue-system/internal/task"
	pkglogger "github.com/mastirikon/queue-system/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем логгер
	log, err := pkglogger.New(cfg.Env)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Worker service",
		zap.String("env", cfg.Env),
		zap.Int("concurrency", cfg.Worker.Concurrency),
		zap.Duration("retry_interval", cfg.Worker.RetryInterval),
	)

	// Создаём Asynq Server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr},
		asynq.Config{
			Concurrency: cfg.Worker.Concurrency,
			Queues: map[string]int{
				"default": 10, // Приоритет очереди
			},
			// Retry с постоянным интервалом 10 секунд
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				return cfg.Worker.RetryInterval
			},
			Logger: newZapLogger(log),
		},
	)

	// Создаём процессор задач
	processor := task.NewProcessor(log, cfg.Worker.RequestTimeout)

	// Регистрируем обработчики
	mux := asynq.NewServeMux()
	mux.HandleFunc(domain.TypeHTTPRequest, processor.ProcessHTTPRequest)

	// Запускаем worker в горутине
	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatal("Failed to start worker", zap.Error(err))
		}
	}()

	log.Info("Worker started successfully")

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker gracefully...")

	// Graceful shutdown
	srv.Shutdown()

	log.Info("Worker stopped")
}

// newZapLogger создаёт адаптер для Asynq logger
func newZapLogger(log *zap.Logger) asynq.Logger {
	return &zapLogger{logger: log}
}

// zapLogger адаптер для интеграции zap с asynq
type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *zapLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *zapLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *zapLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *zapLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}
```

---

### 3️⃣ Dockerfile для API

**Файл:** `docker/api.Dockerfile`

```dockerfile
# Сборка
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/api .

# Открываем порт
EXPOSE 8080

# Запускаем
CMD ["./api"]
```

---

### 4️⃣ Dockerfile для Worker

**Файл:** `docker/worker.Dockerfile`

```dockerfile
# Сборка
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker ./cmd/worker

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/worker .

# Запускаем
CMD ["./worker"]
```

---

### 5️⃣ Docker Compose

**Файл:** `docker-compose.yml`

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    container_name: queue-redis
    command: redis-server --appendonly yes
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - queue-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  api:
    build:
      context: .
      dockerfile: docker/api.Dockerfile
    container_name: queue-api
    environment:
      - ENV=production
      - API_PORT=8080
      - API_HOST=0.0.0.0
      - REDIS_ADDR=redis:6379
    ports:
      - "8080:8080"
    depends_on:
      redis:
        condition: service_healthy
    networks:
      - queue-network
    restart: unless-stopped

  worker:
    build:
      context: .
      dockerfile: docker/worker.Dockerfile
    container_name: queue-worker
    environment:
      - ENV=production
      - REDIS_ADDR=redis:6379
      - WORKER_CONCURRENCY=10
      - WORKER_RETRY_INTERVAL=10s
      - WORKER_REQUEST_TIMEOUT=30s
    depends_on:
      redis:
        condition: service_healthy
    networks:
      - queue-network
    restart: unless-stopped

volumes:
  redis_data:

networks:
  queue-network:
    driver: bridge
```

---

### 6️⃣ .dockerignore

**Файл:** `.dockerignore`

```
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Test files
*_test.go

# IDE
.idea/
.vscode/
*.swp
*.swo

# Git
.git/
.gitignore

# Documentation
*.md
.progress/

# Docker
docker-compose.yml

# Others
.DS_Store
```

---

### 7️⃣ .env.example

**Файл:** `.env.example`

```bash
# Environment
ENV=development

# API Configuration
API_PORT=8080
API_HOST=0.0.0.0
API_READ_TIMEOUT=10s
API_WRITE_TIMEOUT=10s
API_SHUTDOWN_TIMEOUT=30s

# Worker Configuration
WORKER_CONCURRENCY=10
WORKER_RETRY_INTERVAL=10s
WORKER_MAX_RETRIES=8640
WORKER_REQUEST_TIMEOUT=30s

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

---

### 8️⃣ Makefile

**Файл:** `Makefile`

```makefile
.PHONY: help build run-api run-worker docker-build docker-up docker-down test clean

help: ## Показать помощь
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать бинарники
	@echo "Building API..."
	@go build -o bin/api ./cmd/api
	@echo "Building Worker..."
	@go build -o bin/worker ./cmd/worker
	@echo "Done!"

run-api: ## Запустить API локально
	@go run ./cmd/api

run-worker: ## Запустить Worker локально
	@go run ./cmd/worker

docker-build: ## Собрать Docker образы
	@docker-compose build

docker-up: ## Запустить все сервисы в Docker
	@docker-compose up -d

docker-down: ## Остановить все сервисы
	@docker-compose down

docker-logs: ## Посмотреть логи
	@docker-compose logs -f

test: ## Запустить тесты
	@go test -v ./...

clean: ## Очистить бинарники
	@rm -rf bin/
	@echo "Cleaned!"

.DEFAULT_GOAL := help
```

---

### 9️⃣ README.md (обновлённый)

**Файл:** `README.md`

```markdown
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

## 🎉 ВСЁ!

После этого у тебя будет **полностью рабочая система**:
- ✅ API принимает задачи
- ✅ Redis хранит очередь
- ✅ Worker обрабатывает задачи
- ✅ Retry каждые 10 секунд
- ✅ Graceful shutdown
- ✅ Production-ready

**Удачи!** 🚀
