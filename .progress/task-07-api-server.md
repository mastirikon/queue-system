# 📋 Задание #7: API Server (Fiber + main.go)

**Дата выдачи:** 2026-01-22  
**Статус:** 🔄 В работе  
**Фаза:** API Service

---

## 🎯 Цель
Создать полноценный API сервер с Fiber, настроить роутинг, middleware и graceful shutdown.

---

## 📝 Детальные инструкции

### Часть 1: Создание API сервера

Создай файл `cmd/api/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mastirikon/queue-system/internal/config"
	"github.com/mastirikon/queue-system/internal/handler"
	"github.com/mastirikon/queue-system/internal/queue"
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

	log.Info("Starting API server",
		zap.String("env", cfg.Env),
		zap.String("host", cfg.API.Host),
		zap.Int("port", cfg.API.Port),
	)

	// Создаём Asynq Client
	queueClient := queue.NewClient(cfg.Redis.Addr, log)
	defer queueClient.Close()

	// Создаём Fiber приложение
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.API.ReadTimeout,
		WriteTimeout: cfg.API.WriteTimeout,
		ErrorHandler: customErrorHandler(log),
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Создаём handler
	taskHandler := handler.NewTaskHandler(queueClient, log)

	// Роутинг
	api := app.Group("/api/v1")
	api.Post("/tasks", taskHandler.CreateTask)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// Graceful shutdown
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
		if err := app.Listen(addr); err != nil {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server gracefully...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), cfg.API.ShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server stopped")
}

// customErrorHandler обрабатывает ошибки Fiber
func customErrorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		log.Error("Request error",
			zap.Int("status", code),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.Error(err),
		)

		return c.Status(code).JSON(fiber.Map{
			"error":   "internal_error",
			"message": err.Error(),
		})
	}
}
```

---

## ✅ Критерии выполнения

- [ ] Создан файл `cmd/api/main.go`
- [ ] Код компилируется без ошибок:
  - `go build -o bin/api ./cmd/api`
- [ ] Создана директория `bin/` для бинарников
- [ ] Результаты компиляции показаны ментору

---

## 📚 Теория: API Server Architecture

### Main Function Flow

```
1. Load Config      → config.Load()
2. Init Logger      → logger.New()
3. Init Queue       → queue.NewClient()
4. Create Fiber App → fiber.New()
5. Setup Middleware → recover, logger, cors
6. Setup Handlers   → taskHandler.CreateTask
7. Start Server     → app.Listen() (в горутине)
8. Wait for Signal  → signal.Notify()
9. Graceful Stop    → app.ShutdownWithContext()
```

### Зачем горутина для app.Listen?

```go
go func() {
    app.Listen(addr)  // Блокирующий вызов
}()

// Ожидаем сигнал завершения
<-quit
```

**Причина:**
- `app.Listen()` **блокирует** выполнение
- Без горутины программа не дойдёт до `signal.Notify()`
- С горутиной: сервер работает, main ждёт сигнал

### Graceful Shutdown

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
<-quit  // Блокируемся до получения сигнала

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

app.ShutdownWithContext(ctx)
```

**Что происходит:**
1. Получаем SIGINT (Ctrl+C) или SIGTERM
2. Перестаём принимать новые запросы
3. Ждём завершения текущих запросов (макс. 30 сек)
4. Закрываем соединения
5. Завершаем программу

**Зачем это нужно?**
- ✅ Не теряем запросы в процессе
- ✅ Корректно закрываем соединения (Redis, БД)
- ✅ Production-ready подход

### Middleware Order

```go
app.Use(recover.New())     // 1. Recover from panics
app.Use(logger.New())      // 2. Log requests
app.Use(cors.New())        // 3. CORS headers
```

**Порядок важен!**
- `recover` — первым (перехватывает панику из всех middleware)
- `logger` — вторым (логирует все запросы)
- `cors` — третьим (добавляет заголовки)

### Custom Error Handler

```go
fiber.Config{
    ErrorHandler: customErrorHandler(log),
}
```

**Зачем?**
- Централизованная обработка ошибок
- Единый формат ответов
- Логирование всех ошибок
- Скрытие внутренних деталей от клиента

### Health Check Endpoint

```go
app.Get("/health", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ok"})
})
```

**Зачем?**
- Kubernetes liveness probe
- Docker health check
- Мониторинг (Prometheus, Grafana)
- Проверка доступности сервиса

### CORS (Cross-Origin Resource Sharing)

```go
cors.New(cors.Config{
    AllowOrigins: "*",              // Разрешить все домены
    AllowMethods: "GET,POST,PUT",   // Разрешённые методы
})
```

**Зачем?**
- Позволяет frontend'у (React) обращаться к API
- Без CORS браузер заблокирует запрос
- `*` — для dev, в prod указать конкретные домены

---

## 🎓 Дополнительная информация

### Defer в Go

```go
defer log.Sync()
defer queueClient.Close()
```

**Что делает defer?**
- Выполняет функцию **при выходе** из текущей функции
- Порядок выполнения: **LIFO** (последний defer выполнится первым)

```go
defer fmt.Println("1")
defer fmt.Println("2")
defer fmt.Println("3")
// Output: 3 2 1
```

### Context с таймаутом

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

**Зачем cancel()?**
- Освобождает ресурсы context'а
- Best practice — всегда вызывать defer cancel()
- Даже если таймаут не сработал

### OS Signals

```go
os.Interrupt   // SIGINT (Ctrl+C)
syscall.SIGTERM // SIGTERM (kill команда)
```

**Типичные сигналы:**
- `SIGINT` (2) — пользователь нажал Ctrl+C
- `SIGTERM` (15) — корректное завершение (Docker, K8s)
- `SIGKILL` (9) — принудительное завершение (не перехватывается!)

### Channel для сигналов

```go
quit := make(chan os.Signal, 1)  // Буферизованный канал
signal.Notify(quit, os.Interrupt)
<-quit  // Блокируемся до получения сигнала
```

**Зачем буфер 1?**
- Чтобы не потерять сигнал, если его получили до чтения из канала
- Best practice для signal channels

### Fiber Config

```go
fiber.Config{
    ReadTimeout:  10 * time.Second,   // Таймаут чтения запроса
    WriteTimeout: 10 * time.Second,   // Таймаут записи ответа
}
```

**Зачем таймауты?**
- Защита от медленных клиентов (Slowloris attack)
- Освобождение ресурсов
- Production best practice

---

## 🧪 Проверка компиляции

После создания файла выполни:

```bash
# Создай директорию для бинарников
mkdir -p bin

# Компиляция API сервера
go build -o bin/api ./cmd/api

# Проверка, что бинарник создан
ls -lh bin/
```

Если компиляция успешна, бинарник будет в `bin/api`.

---

## 🚨 Возможные ошибки

### Ошибка: "no required module provides package"
**Решение:** Выполни `go mod tidy`

### Ошибка: "imported and not used"
**Причина:** Импортировал пакет, но не используешь
**Решение:** Удали неиспользуемый import

### Ошибка: "pkglogger redeclared in this block"
**Причина:** Конфликт имён (logger vs pkglogger)
**Решение:** Используй alias: `pkglogger "github.com/.../pkg/logger"`

---

## 🎯 Тестовый запуск (после компиляции)

После успешной компиляции, мы протестируем запуск:

```bash
# Экспортируем переменные окружения
export ENV=development
export REDIS_ADDR=localhost:6379

# Запускаем API сервер
./bin/api
```

**Ожидаемый вывод:**
```
INFO Starting API server env=development host=0.0.0.0 port=8080
INFO Listening on 0.0.0.0:8080
```

**Но это на следующем этапе!** Сейчас только компиляция.

---

## 📊 Что дальше?

После этого задания:
1. Протестируем API с Redis (нужен запущенный Redis)
2. Создадим Worker сервис
3. Протестируем полный flow: API → Redis → Worker

---

**Когда закончишь:**
1. Покажи вывод компиляции:
   - `go build -o bin/api ./cmd/api`
   - `ls -lh bin/`
2. Покажи первые 50 строк файла `cmd/api/main.go`

Я проверю и дам следующее задание! 🎯
