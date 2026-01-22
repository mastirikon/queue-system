# 📋 Задание #6: HTTP Handler для приёма задач

**Дата выдачи:** 2026-01-22  
**Статус:** 🔄 В работе  
**Фаза:** API Service

---

## 🎯 Цель
Создать HTTP handler для приёма задач через POST запрос и добавления их в очередь Asynq.

---

## 📝 Детальные инструкции

### Часть 1: Request/Response структуры

Создай файл `internal/handler/request.go`:

```go
package handler

// CreateTaskRequest — запрос на создание задачи
type CreateTaskRequest struct {
	URL     string            `json:"url" validate:"required,url"`
	Method  string            `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
```

Создай файл `internal/handler/response.go`:

```go
package handler

// ErrorResponse — стандартный ответ с ошибкой
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// CreateTaskResponse — ответ на создание задачи
type CreateTaskResponse struct {
	TaskID  string `json:"task_id"`
	Message string `json:"message"`
}
```

---

### Часть 2: HTTP Handler

Создай файл `internal/handler/task_handler.go`:

```go
package handler

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mastirikon/queue-system/internal/domain"
	"github.com/mastirikon/queue-system/internal/queue"
	"go.uber.org/zap"
)

// TaskHandler обрабатывает HTTP запросы для задач
type TaskHandler struct {
	queueClient *queue.Client
	logger      *zap.Logger
	validator   *validator.Validate
}

// NewTaskHandler создаёт новый TaskHandler
func NewTaskHandler(queueClient *queue.Client, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		queueClient: queueClient,
		logger:      logger,
		validator:   validator.New(),
	}
}

// CreateTask обрабатывает POST /tasks
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	// Парсим JSON из body
	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("Failed to parse request body",
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid JSON format",
		})
	}

	// Валидация
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Warn("Request validation failed",
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
	}

	// Создаём задачу
	task := &domain.Task{
		ID:        uuid.New().String(),
		URL:       req.URL,
		Method:    req.Method,
		Headers:   req.Headers,
		Body:      req.Body,
		CreatedAt: time.Now(),
	}

	// Отправляем в очередь
	if err := h.queueClient.EnqueueTask(c.Context(), task); err != nil {
		h.logger.Error("Failed to enqueue task",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "enqueue_failed",
			Message: "Failed to enqueue task",
		})
	}

	// Успешный ответ
	return c.Status(fiber.StatusCreated).JSON(CreateTaskResponse{
		TaskID:  task.ID,
		Message: "Task created successfully",
	})
}
```

---

## ✅ Критерии выполнения

- [ ] Создан файл `internal/handler/request.go`
- [ ] Создан файл `internal/handler/response.go`
- [ ] Создан файл `internal/handler/task_handler.go`
- [ ] Код компилируется без ошибок:
  - `go build ./internal/handler`
- [ ] Результаты компиляции показаны ментору

---

## 📚 Теория: HTTP Handlers в Go

### Fiber vs net/http

**Почему Fiber?**
- ✅ Быстрее net/http (основан на fasthttp)
- ✅ Express-like API (знакомо для Node.js разработчиков)
- ✅ Встроенные middleware (logger, recover, cors)
- ✅ Легковесный и простой

**Пример net/http (для сравнения):**
```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Много boilerplate кода
}
```

**Fiber:**
```go
func handler(c *fiber.Ctx) error {
    return c.JSON(data)  // Просто!
}
```

### HTTP Status Codes

```go
fiber.StatusBadRequest          // 400 - неправильный запрос
fiber.StatusCreated             // 201 - ресурс создан
fiber.StatusInternalServerError // 500 - внутренняя ошибка
```

**Правило:**
- 2xx — успех (200 OK, 201 Created)
- 4xx — ошибка клиента (400 Bad Request, 404 Not Found)
- 5xx — ошибка сервера (500 Internal Server Error)

### Валидация с validator/v10

```go
type Request struct {
    URL    string `validate:"required,url"`
    Method string `validate:"required,oneof=GET POST"`
}

validator.New().Struct(&req)  // Валидация
```

**Теги валидации:**
- `required` — обязательное поле
- `url` — валидный URL
- `oneof=GET POST` — одно из значений
- `email` — валидный email
- `min=1,max=100` — минимум/максимум

### UUID для ID задач

```go
uuid.New().String()  // "550e8400-e29b-41d4-a716-446655440000"
```

**Зачем UUID?**
- ✅ Уникальность гарантирована
- ✅ Распределённая генерация (без координации)
- ✅ Нет коллизий
- ✅ Стандарт в индустрии

**Альтернативы:**
- Auto-increment ID — требует БД
- Timestamp — может быть коллизия
- Random string — сложнее генерировать

### Context в Fiber

```go
c.Context()  // context.Context для передачи в другие функции
```

**Зачем?**
- Timeout'ы и cancellation
- Передача request ID для трейсинга
- Best practice в Go

### Structured Logging

```go
h.logger.Warn("Failed to parse request body",
    zap.Error(err),
)
```

**Уровни логов:**
- `Debug` — детальная информация для отладки
- `Info` — обычные события (task created)
- `Warn` — предупреждения (invalid request)
- `Error` — ошибки, требующие внимания

### Error Handling

```go
if err := c.BodyParser(&req); err != nil {
    return c.Status(400).JSON(ErrorResponse{...})
}
```

**Правило:**
1. Проверяй ошибку сразу после вызова
2. Логируй ошибку
3. Верни понятный ответ клиенту
4. Не показывай клиенту внутренние детали

---

## 🎓 Дополнительная информация

### REST API Best Practices

**Эндпоинт:**
```
POST /tasks  — создать задачу
GET  /tasks  — получить список (будет в будущем)
GET  /tasks/:id  — получить одну задачу (будет в будущем)
```

**Ответы:**
```json
// Успех (201 Created)
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Task created successfully"
}

// Ошибка (400 Bad Request)
{
  "error": "validation_error",
  "message": "URL is required"
}
```

### JSON Tags в Go

```go
type Task struct {
    ID   string `json:"id"`           // Имя поля в JSON
    URL  string `json:"url"`
}
```

**Правила:**
- `json:"id"` — lowercase в JSON (стандарт)
- `json:"task_id"` — snake_case
- `json:"-"` — игнорировать поле
- `json:",omitempty"` — не показывать, если пусто

### Validator Tags

```go
URL string `validate:"required,url"`
```

**Часто используемые:**
- `required` — не может быть пустым
- `url` — валидный URL
- `email` — валидный email
- `min=1,max=100` — длина строки
- `gte=0,lte=100` — число в диапазоне
- `oneof=GET POST` — enum

---

## 🧪 Проверка компиляции

После создания файлов выполни:

```bash
# Проверка handler пакета
go build ./internal/handler

# Обновление зависимостей (если нужно)
go mod tidy
```

Если ошибок нет — отлично! ✅

---

## 🚨 Возможные ошибки

### Ошибка: "cannot find package github.com/gofiber/fiber/v2"
**Решение:** Выполни `go get github.com/gofiber/fiber/v2`

### Ошибка: "cannot find package github.com/go-playground/validator/v10"
**Решение:** Выполни `go get github.com/go-playground/validator/v10`

### Ошибка: "imported and not used"
**Причина:** Импортировал пакет, но не используешь
**Решение:** Удали неиспользуемый import

---

## 📊 Что дальше?

После этого задания мы создадим:
1. Fiber сервер с роутингом
2. Middleware (logger, recover)
3. Graceful shutdown
4. Всё соберём в `cmd/api/main.go`

---

**Когда закончишь:**
1. Покажи вывод компиляции:
   - `go build ./internal/handler`
2. Покажи первые 30 строк файла `internal/handler/task_handler.go`

Я проверю и дам следующее задание! 🎯
