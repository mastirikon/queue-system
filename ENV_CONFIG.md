# ⚙️ Конфигурация через переменные окружения

Все настройки системы управляются через файл `.env` на сервере.

---

## 📝 Структура конфигурации

### Локальная разработка
Файл: `.env` (создай сам на основе `.env.production`)

### Production (vdska)
Файл: `.env.production` → копируется как `.env` на сервер при деплое

---

## 🔧 Настройки

### Environment
```bash
ENV=production                    # Режим: development или production
```

### API Server
```bash
API_PORT=8080                     # Порт API сервера
API_HOST=0.0.0.0                  # Host для прослушивания
API_READ_TIMEOUT=10s              # Таймаут чтения запроса
API_WRITE_TIMEOUT=10s             # Таймаут записи ответа
API_SHUTDOWN_TIMEOUT=30s          # Таймаут graceful shutdown
```

### Worker
```bash
WORKER_CONCURRENCY=10             # Количество одновременных задач
WORKER_RETRY_INTERVAL=10s         # Интервал между retry
WORKER_MAX_RETRIES=8640           # Макс. попыток (24 часа при 10s)
WORKER_REQUEST_TIMEOUT=30s        # Таймаут HTTP запроса
```

### Target URL (главное!)
```bash
WORKER_TARGET_URL=https://tasker-google-sheets.ku-34.netcraze.pro/notify
```

**Это URL, на который Worker будет отправлять все уведомления!**

---

## 🚀 Изменение конфигурации

### Способ 1: Редактируем локально и деплоим

```bash
# 1. Редактируешь .env.production на Mac
nano .env.production

# 2. Деплоишь
make deploy-full
```

### Способ 2: Редактируем прямо на сервере

```bash
# 1. Подключаешься к серверу
ssh root@vdska

# 2. Редактируешь .env
cd /home/finance-system/queue-system
nano .env

# 3. Перезапускаешь сервисы
docker compose -f docker-compose-simple.yml restart api worker
```

---

## 📊 Примеры конфигураций

### Высокая производительность
```bash
WORKER_CONCURRENCY=50
WORKER_RETRY_INTERVAL=5s
WORKER_REQUEST_TIMEOUT=60s
```

### Экономия ресурсов
```bash
WORKER_CONCURRENCY=5
WORKER_RETRY_INTERVAL=30s
WORKER_REQUEST_TIMEOUT=20s
```

### Быстрый retry
```bash
WORKER_RETRY_INTERVAL=5s
WORKER_MAX_RETRIES=17280    # 24 часа при 5s интервале
```

---

## 🔒 Безопасность

### Если нужна авторизация

Добавь в `.env.production`:
```bash
WORKER_AUTH_TOKEN=your-secret-token-here
```

Потом обнови код handler'а чтобы добавлять заголовок:
```go
Headers: map[string]string{
    "Content-Type": "application/json",
    "Authorization": "Bearer " + os.Getenv("WORKER_AUTH_TOKEN"),
}
```

---

## 🎯 Разные URL для разных окружений

### Development (.env)
```bash
WORKER_TARGET_URL=http://localhost:8000/test
```

### Staging (.env.staging)
```bash
WORKER_TARGET_URL=https://staging.tasker-google-sheets.ku-34.netcraze.pro/notify
```

### Production (.env.production)
```bash
WORKER_TARGET_URL=https://tasker-google-sheets.ku-34.netcraze.pro/notify
```

---

## 🛠️ Проверка текущей конфигурации

На сервере:
```bash
cd /home/finance-system/queue-system

# Посмотреть .env
cat .env

# Посмотреть какие переменные использует контейнер
docker exec queue-api env | grep WORKER
docker exec queue-worker env | grep WORKER
```

---

## 📝 Шаблон .env для быстрого старта

Скопируй и отредактируй:

```bash
# === ОСНОВНЫЕ НАСТРОЙКИ ===
ENV=production

# === API ===
API_PORT=8080
API_HOST=0.0.0.0

# === WORKER ===
WORKER_CONCURRENCY=10
WORKER_RETRY_INTERVAL=10s
WORKER_REQUEST_TIMEOUT=30s

# === ГЛАВНОЕ: КУДА ОТПРАВЛЯТЬ ДАННЫЕ ===
WORKER_TARGET_URL=https://your-domain.com/endpoint

# === ДОПОЛНИТЕЛЬНО (если нужно) ===
# WORKER_AUTH_TOKEN=secret-token-123
# WORKER_CUSTOM_HEADER_1=value1
# WORKER_CUSTOM_HEADER_2=value2
```

---

## 🔄 После изменения .env

**Всегда перезапускай сервисы:**

```bash
# Перезапуск без пересборки (быстро)
docker compose -f docker-compose-simple.yml restart api worker

# Или полная перезагрузка
docker compose -f docker-compose-simple.yml down
docker compose -f docker-compose-simple.yml up -d
```

---

## 💡 Pro Tips

1. **Не коммить .env в Git** (уже в .gitignore)
2. **Используй разные файлы** для разных окружений (.env.development, .env.production)
3. **Храни бэкап** важных настроек
4. **Документируй** нестандартные значения

---

## 🆘 Troubleshooting

### Изменения не применились?
```bash
# Проверь что .env загружен
docker exec queue-worker env | grep TARGET_URL

# Если пусто - перезапусти с пересборкой
docker compose -f docker-compose-simple.yml up -d --force-recreate
```

### Неправильный URL?
```bash
# Проверь логи при старте
docker compose -f docker-compose-simple.yml logs api | grep TARGET
```

### Синтаксическая ошибка в .env?
```bash
# Проверь формат (без пробелов вокруг =)
# Правильно:
WORKER_TARGET_URL=https://example.com
# Неправильно:
WORKER_TARGET_URL = https://example.com
```
