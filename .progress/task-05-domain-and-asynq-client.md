# 📋 Задание #5: Domain модели и Asynq Client

**Дата выдачи:** 2026-01-08  
**Статус:** 🔄 В работе  
**Фаза:** API Service

---

## 🎯 Цель
Создать domain модели для задач и настроить Asynq Client для отправки задач в очередь Redis.

---

## 📝 Детальные инструкции

### Часть 1: Domain модели

Создай файл `internal/domain/task.go`:

```go
package domain

import (
	"encoding/json"
	"time"
)

// Task представляет задачу для обработки
type Task struct {
	ID          string    `json:"id"`           // Уникальный ID задачи (UUID)
	URL         string    `json:"url"`          // URL для HTTP запроса
	Method      string    `json:"method"`       // HTTP метод (POST, GET и т.д.)
	Headers     Headers   `json:"headers"`      // HTTP заголовки
	Body        string    `json:"body"`         // Тело запроса (если есть)
	CreatedAt   time.Time `json:"created_at"`   // Время создания задачи
}

// Headers представляет HTTP заголовки
type Headers map[string]string

// TaskPayload — это payload для Asynq задачи (что отправляем в Redis)
type TaskPayload struct {
	ID      string  `json:"id"`
	URL     string  `json:"url"`
	Method  string  `json:"method"`
	Headers Headers `json:"headers"`
	Body    string  `json:"body"`
}

// ToPayload конвертирует Task в TaskPayload для Asynq
func (t *Task) ToPayload() ([]byte, error) {
	payload := TaskPayload{
		ID:      t.ID,
		URL:     t.URL,
		Method:  t.Method,
		Headers: t.Headers,
		Body:    t.Body,
	}
	return json.Marshal(payload)
}

// TaskFromPayload создаёт Task из payload
func TaskFromPayload(data []byte) (*TaskPayload, error) {
	var payload TaskPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
```

---

### Часть 2: Константы для типов задач

Создай файл `internal/domain/task_types.go`:

```go
package domain

// Типы задач в системе
const (
	// TypeHTTPRequest — задача HTTP запроса
	TypeHTTPRequest = "http:request"
)
```

---

### Часть 3: Asynq Client обёртка

Создай файл `internal/queue/client.go`:

```go
package queue

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mastirikon/queue-system/internal/domain"
	"go.uber.org/zap"
)

// Client — обёртка над Asynq Client
type Client struct {
	client *asynq.Client
	logger *zap.Logger
}

// NewClient создаёт новый queue client
func NewClient(redisAddr string, logger *zap.Logger) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})

	return &Client{
		client: client,
		logger: logger,
	}
}

// EnqueueTask отправляет задачу в очередь
func (c *Client) EnqueueTask(ctx context.Context, task *domain.Task) error {
	// Конвертируем Task в payload
	payload, err := task.ToPayload()
	if err != nil {
		c.logger.Error("Failed to marshal task payload",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return err
	}

	// Создаём Asynq задачу
	asynqTask := asynq.NewTask(domain.TypeHTTPRequest, payload)

	// Опции задачи
	opts := []asynq.Option{
		asynq.MaxRetry(8640),                    // 24 часа при 10 сек интервале
		asynq.Timeout(30 * time.Second),         // Таймаут выполнения задачи
		asynq.Retention(24 * time.Hour),         // Хранить 24 часа после завершения
		asynq.TaskID(task.ID),                   // Устанавливаем ID задачи
	}

	// Отправляем задачу
	info, err := c.client.EnqueueContext(ctx, asynqTask, opts...)
	if err != nil {
		c.logger.Error("Failed to enqueue task",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return err
	}

	c.logger.Info("Task enqueued successfully",
		zap.String("task_id", task.ID),
		zap.String("queue", info.Queue),
		zap.Time("next_process_at", info.NextProcessAt),
	)

	return nil
}

// Close закрывает соединение с Redis
func (c *Client) Close() error {
	return c.client.Close()
}
```

---

## ✅ Критерии выполнения

- [ ] Создан файл `internal/domain/task.go` с моделями
- [ ] Создан файл `internal/domain/task_types.go` с константами
- [ ] Создан файл `internal/queue/client.go` с Asynq Client
- [ ] Код компилируется без ошибок:
  - `go build ./internal/domain`
  - `go build ./internal/queue`
- [ ] Результаты компиляции показаны ментору

---

## 📚 Теория: Domain-Driven Design

### Зачем отдельный пакет domain?

```
internal/domain/  — бизнес-логика, модели данных
internal/queue/   — инфраструктура (работа с очередями)
internal/handler/ — HTTP слой (будет позже)
```

**Принцип разделения ответственности:**
- `domain` — что такое Task (модель)
- `queue` — как отправить Task в очередь (инфраструктура)
- `handler` — как принять Task от клиента (HTTP)

**Преимущества:**
- Легко тестировать (мокируем queue)
- Легко менять инфраструктуру (Redis → RabbitMQ)
- Чистая архитектура

### Почему TaskPayload отдельно от Task?

```go
type Task struct {
    ID        string
    URL       string
    CreatedAt time.Time  // ← Не нужно в Redis!
}

type TaskPayload struct {
    ID     string
    URL    string
    // Без CreatedAt — экономим место в Redis
}
```

**Причины:**
- Task — полная модель (для API, логов, БД в будущем)
- TaskPayload — минимальные данные для очереди
- Экономия памяти в Redis (критично для 750MB!)

### Asynq Options

#### MaxRetry(8640)
```go
asynq.MaxRetry(8640)  // 24 часа × 6 попыток/мин
```

**Что это делает:**
- Задача будет пытаться выполниться максимум 8640 раз
- После этого → переходит в "dead" очередь
- Можно вручную перезапустить из dead

#### Timeout(30s)
```go
asynq.Timeout(30 * time.Second)
```

**Таймаут на выполнение одной попытки:**
- Если задача не завершилась за 30 сек → retry
- Защита от зависших задач

#### Retention(24h)
```go
asynq.Retention(24 * time.Hour)
```

**Хранение после завершения:**
- Успешные задачи хранятся 24 часа
- Можно посмотреть историю в Asynq Web UI
- Автоматически удаляются через 24 часа

#### TaskID(id)
```go
asynq.TaskID(task.ID)
```

**Уникальный ID:**
- Позволяет отслеживать задачу
- Идемпотентность (не добавим дубликат с тем же ID)
- Удобно для логов и дебага

### Context в EnqueueTask

```go
func (c *Client) EnqueueTask(ctx context.Context, task *domain.Task) error
```

**Зачем context.Context?**
- Можно отменить операцию (timeout, cancellation)
- Передать trace ID (для распределённого трейсинга)
- Best practice в Go для IO операций

**Пример использования:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := client.EnqueueTask(ctx, task)
```

### Структурированное логирование

```go
c.logger.Info("Task enqueued successfully",
    zap.String("task_id", task.ID),
    zap.String("queue", info.Queue),
    zap.Time("next_process_at", info.NextProcessAt),
)
```

**Преимущества:**
- Легко фильтровать по task_id
- Легко агрегировать метрики
- JSON формат в production → ELK/Grafana

**Плохой пример (не делай так):**
```go
log.Printf("Task %s enqueued in queue %s", task.ID, info.Queue)
// ❌ Сложно парсить, сложно фильтровать
```

---

## 🎓 Дополнительная информация

### Asynq vs другие очереди

**Почему Asynq?**
- ✅ Redis-based (уже используем Redis)
- ✅ Встроенный retry с exponential backoff
- ✅ Web UI для мониторинга из коробки
- ✅ Scheduled tasks (можно отложить выполнение)
- ✅ Уникальность задач (TaskID)
- ✅ Легковесный (~750MB RAM хватит)

**Альтернативы:**
- RabbitMQ — тяжелее, нужен отдельный сервер
- Kafka — overkill для 20-30 задач/мин
- AWS SQS — cloud-only, платно

### JSON Marshal/Unmarshal

```go
payload, err := json.Marshal(task)
```

**Почему JSON?**
- Читаемый формат (легко дебажить в Redis CLI)
- Универсальный (можно обработать из любого языка)
- Не самый быстрый, но для 20-30 задач/мин — достаточно

**Альтернативы:**
- Protocol Buffers — быстрее, но сложнее
- MessagePack — компактнее, но менее читаем
- Gob — только Go

Для наших нагрузок JSON — оптимальный выбор.

---

## 🧪 Проверка компиляции

После создания файлов выполни:

```bash
# Проверка domain пакета
go build ./internal/domain

# Проверка queue пакета
go build ./internal/queue

# Обновление зависимостей (если нужно)
go mod tidy
```

Если ошибок нет — отлично! ✅

---

## 🚨 Возможные ошибки

### Ошибка: "cannot find package"
**Решение:** Выполни `go mod tidy`

### Ошибка: "imported and not used"
**Причина:** Импортировал пакет, но не используешь
**Решение:** Удали неиспользуемый import или используй

### Ошибка: "undefined: asynq.RedisClientOpt"
**Причина:** Неправильная версия asynq
**Решение:** Проверь `go.mod`, должно быть `github.com/hibiken/asynq v0.25.1`

---

## 📊 Что дальше?

После этого задания мы создадим:
1. HTTP handler для приёма задач
2. Fiber сервер
3. Graceful shutdown
4. Всё соберём в `cmd/api/main.go`

---

**Когда закончишь:**
1. Покажи вывод компиляции:
   - `go build ./internal/domain`
   - `go build ./internal/queue`
2. Покажи первые 20 строк файла `internal/queue/client.go`

Я проверю и дам следующее задание! 🎯

