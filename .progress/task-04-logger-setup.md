# 📋 Задание #4: Настройка базового логгера

**Дата выдачи:** 2025-12-16  
**Статус:** 🔄 В работе  
**Фаза:** Setup & Project Structure

---

## 🎯 Цель
Создать переиспользуемый логгер на базе `zap` с поддержкой двух режимов: development (читаемый) и production (JSON).

---

## 📝 Детальные инструкции

### Создай файл `pkg/logger/logger.go`:

```go
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New создаёт новый логгер в зависимости от окружения
// env может быть "development" или "production"
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		return newProduction()
	}
	return newDevelopment()
}

// newProduction создаёт production логгер (JSON, INFO+)
func newProduction() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	
	// Настройка формата времени
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	// Уровень логирования
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	
	return config.Build(
		zap.AddCaller(),      // Добавляет информацию о месте вызова
		zap.AddStacktrace(zap.ErrorLevel), // Stacktrace только для ERROR+
	)
}

// newDevelopment создаёт development логгер (консоль, DEBUG+)
func newDevelopment() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	
	// Красивый цветной вывод для консоли
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	// Уровень логирования
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	
	return config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
}

// NewNop создаёт no-op логгер (для тестов)
func NewNop() *zap.Logger {
	return zap.NewNop()
}
```

---

## ✅ Критерии выполнения

- [ ] Создан файл `pkg/logger/logger.go` с корректным содержимым
- [ ] Код компилируется без ошибок: `go build ./pkg/logger`
- [ ] Тест production логгера выполнен успешно
- [ ] Тест development логгера выполнен успешно
- [ ] Результаты тестов показаны ментору

---

## 📚 Теория: Структура логгера

### Почему два режима?

**Development:**
- Цветной вывод в консоль
- Удобно читать во время разработки
- Уровень DEBUG (все логи)
- Читаемый формат

**Production:**
- JSON формат
- Легко парсится системами мониторинга (ELK, Grafana Loki)
- Уровень INFO (без debug)
- Компактный формат

### Зачем AddCaller()?

```go
zap.AddCaller()
```

Добавляет в лог информацию о файле и строке, где был вызван логгер:

```json
{"level":"info","timestamp":"...","caller":"main.go:42","msg":"..."}
```

**Плюсы:**
- Легко найти место в коде
- Быстрый дебаг

**Минусы:**
- Небольшой overhead (несколько наносекунд)
- Но для наших нагрузок (20-30 задач/мин) — незаметно

### Зачем AddStacktrace()?

```go
zap.AddStacktrace(zap.ErrorLevel)
```

Добавляет полный stacktrace только для ERROR и выше.

**Зачем:**
- При критических ошибках видно весь путь вызовов
- Для INFO/DEBUG не нужен (засоряет логи)

### Почему отдельная функция NewNop()?

```go
func NewNop() *zap.Logger {
	return zap.NewNop()
}
```

No-op (no operation) логгер — ничего не делает.

**Использование:**
- Юнит-тесты (не засоряем вывод)
- Бенчмарки (убираем overhead логирования)

### Формат времени ISO8601

```go
config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
```

Формат: `2025-12-16T10:30:45.123Z`

**Преимущества:**
- Международный стандарт
- Легко парсится
- Включает таймзону
- Сортируется лексикографически

---

## 🧪 Тесты для проверки

### Тест 1: Production логгер (JSON)

```bash
cd /Users/anton/DEV/myProjects/go-finance-system/queue-system

cat > test_logger_prod.go << 'EOF'
package main

import (
	"github.com/mastirikon/queue-system/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	log, _ := logger.New("production")
	defer log.Sync()
	
	log.Info("Production logger test",
		zap.String("env", "production"),
		zap.Int("port", 8080),
	)
	
	log.Error("Test error message",
		zap.String("error", "connection refused"),
		zap.Int("retry", 3),
	)
}
EOF

go run test_logger_prod.go
rm test_logger_prod.go
```

**Ожидаемый вывод (JSON):**
```json
{"level":"info","timestamp":"2025-12-16T...","caller":"test_logger_prod.go:12","msg":"Production logger test","env":"production","port":8080}
{"level":"error","timestamp":"2025-12-16T...","caller":"test_logger_prod.go:17","msg":"Test error message","error":"connection refused","retry":3,"stacktrace":"..."}
```

---

### Тест 2: Development логгер (цветной консоль)

```bash
cd /Users/anton/DEV/myProjects/go-finance-system/queue-system

cat > test_logger_dev.go << 'EOF'
package main

import (
	"github.com/mastirikon/queue-system/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	log, _ := logger.New("development")
	defer log.Sync()
	
	log.Debug("Debug message (только в dev)")
	
	log.Info("Development logger test",
		zap.String("env", "development"),
		zap.Int("port", 8080),
	)
	
	log.Warn("Warning message",
		zap.String("reason", "high memory usage"),
	)
	
	log.Error("Error message",
		zap.String("error", "timeout"),
	)
}
EOF

go run test_logger_dev.go
rm test_logger_dev.go
```

**Ожидаемый вывод (цветной, читаемый):**
```
2025-12-16T10:30:45.123+0300  DEBUG  test_logger_dev.go:13  Debug message (только в dev)
2025-12-16T10:30:45.124+0300  INFO   test_logger_dev.go:15  Development logger test  {"env": "development", "port": 8080}
2025-12-16T10:30:45.125+0300  WARN   test_logger_dev.go:20  Warning message  {"reason": "high memory usage"}
2025-12-16T10:30:45.126+0300  ERROR  test_logger_dev.go:24  Error message  {"error": "timeout"}
test_logger_dev.go:24
main.main
...stacktrace...
```

---

## 🎓 Дополнительная информация

### Уровни логирования в zap:

1. **DEBUG** — подробная информация для отладки (только dev)
2. **INFO** — обычная работа приложения (события, старт/стоп)
3. **WARN** — предупреждения (не критично, но внимание требуется)
4. **ERROR** — ошибки (что-то пошло не так)
5. **FATAL** — фатальные ошибки (приложение не может продолжить, os.Exit(1))
6. **PANIC** — panic (приложение паникует)

**В production:** INFO и выше  
**В development:** DEBUG и выше

### Почему pkg/, а не internal/?

```
pkg/logger/     ✅ — можно переиспользовать в других проектах
internal/logger/ ⚠️ — только внутри этого модуля
```

Логгер — универсальный компонент, его можно использовать где угодно, поэтому `pkg/`.

---

## 🚨 Возможные ошибки

### Ошибка: "cannot find package"
**Решение:** Выполни `go mod tidy` для обновления зависимостей.

### Ошибка: "Sync /dev/stdout: invalid argument"
**Решение:** Это нормально на Windows/некоторых терминалах. Можно игнорировать или обернуть:
```go
_ = log.Sync() // Игнорируем ошибку
```

---

**Когда закончишь:**
1. Покажи содержимое `pkg/logger/logger.go`
2. Выполни оба теста
3. Покажи вывод тестов (production и development)

Я проверю и дам следующее задание! 🎯

