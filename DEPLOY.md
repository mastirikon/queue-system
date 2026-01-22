# 🚀 Deployment Guide

## Быстрый деплой

После внесения изменений в код:

```bash
make deploy
```

**Вот и всё!** Система автоматически:
1. Соберёт бинарники для Linux
2. Загрузит их на vdska
3. Пересоберёт Docker образы
4. Перезапустит сервисы

---

## Детальный процесс

### 1️⃣ Внесение изменений

Редактируешь код в Cursor:
- `internal/handler/` - HTTP handlers
- `internal/task/` - обработка задач
- `cmd/api/` или `cmd/worker/` - точки входа
- `docker-compose-simple.yml` - конфигурация Docker

### 2️⃣ Локальное тестирование (опционально)

```bash
# Запусти Redis
docker run -d -p 6379:6379 redis:7-alpine

# Запусти API
go run ./cmd/api

# В другом терминале - Worker
go run ./cmd/worker

# Протестируй
curl http://localhost:8080/health
```

### 3️⃣ Настройка переменных окружения (первый раз)

Перед первым деплоем отредактируй `.env.production`:

```bash
# Открой файл
nano .env.production

# Главное - проверь WORKER_TARGET_URL:
WORKER_TARGET_URL=https://твой-домен.com/endpoint

# Сохрани (Ctrl+O, Enter, Ctrl+X)
```

### 4️⃣ Деплой

**Только код:**
```bash
make deploy
```

**Код + конфиги + .env:**
```bash
make deploy-full
```

**Вручную (пошагово):**
```bash
# Собрать
make build-linux

# Загрузить
scp bin/api-linux root@vdska:/home/finance-system/queue-system/bin/
scp bin/worker-linux root@vdska:/home/finance-system/queue-system/bin/

# Перезапустить на сервере
ssh root@vdska
cd /home/finance-system/queue-system
docker compose -f docker-compose-simple.yml up -d --build
```

---

## 📊 Мониторинг

### Посмотреть статус
```bash
make status-remote
```

Или вручную:
```bash
ssh root@vdska "docker compose -f docker-compose-simple.yml ps"
```

### Посмотреть логи
```bash
make logs-remote
```

Или вручную:
```bash
ssh root@vdska "docker compose -f docker-compose-simple.yml logs -f"
```

### Asynq Web UI
Открой в браузере:
```
http://your-vdska-ip:8081
```

---

## 🔧 Полезные команды

### Перезапуск отдельного сервиса
```bash
ssh root@vdska
cd /home/finance-system/queue-system

# Перезапустить только API
docker compose -f docker-compose-simple.yml restart api

# Перезапустить только Worker
docker compose -f docker-compose-simple.yml restart worker
```

### Остановить всё
```bash
ssh root@vdska
cd /home/finance-system/queue-system
docker compose -f docker-compose-simple.yml down
```

### Запустить заново
```bash
ssh root@vdska
cd /home/finance-system/queue-system
docker compose -f docker-compose-simple.yml up -d
```

### Посмотреть логи конкретного сервиса
```bash
ssh root@vdska
cd /home/finance-system/queue-system

docker compose -f docker-compose-simple.yml logs api
docker compose -f docker-compose-simple.yml logs worker
docker compose -f docker-compose-simple.yml logs redis
```

---

## 🐛 Troubleshooting

### Сервис не запускается
```bash
# Посмотри логи
make logs-remote

# Проверь статус
make status-remote

# Пересобери с нуля
ssh root@vdska
cd /home/finance-system/queue-system
docker compose -f docker-compose-simple.yml down
docker compose -f docker-compose-simple.yml up -d --build --force-recreate
```

### Изменения не применились
```bash
# Убедись что загрузил новые бинарники
make build-linux
scp bin/api-linux root@vdska:/home/finance-system/queue-system/bin/
scp bin/worker-linux root@vdska:/home/finance-system/queue-system/bin/

# Пересобери образы с нуля
ssh root@vdska
cd /home/finance-system/queue-system
docker compose -f docker-compose-simple.yml up -d --build --force-recreate
```

### Порты заняты
```bash
ssh root@vdska

# Проверь что слушает порты
lsof -i :8080
lsof -i :8081
lsof -i :6379

# Останови старые контейнеры
docker ps -a
docker stop <container_id>
docker rm <container_id>
```

---

## 📝 Чеклист перед деплоем

- [ ] Код работает локально
- [ ] Все тесты проходят
- [ ] Закоммитил изменения в Git
- [ ] Собрал бинарники для Linux (`make build-linux`)
- [ ] Задеплоил (`make deploy`)
- [ ] Проверил логи (`make logs-remote`)
- [ ] Проверил статус (`make status-remote`)
- [ ] Протестировал API (`curl http://vdska:8080/health`)
- [ ] Проверил Asynq Monitor (http://vdska:8081)

---

## 🎯 Быстрая справка

| Команда | Описание |
|---------|----------|
| `make deploy` | Деплой кода на vdska |
| `make deploy-full` | Деплой всего (код + конфиги) |
| `make logs-remote` | Логи с vdska |
| `make status-remote` | Статус сервисов |
| `make build-linux` | Собрать бинарники для Linux |

---

**Pro tip:** Добавь алиас в `~/.zshrc`:
```bash
alias qs-deploy="cd /Users/anton/DEV/myProjects/go-finance-system/queue-system && make deploy"
alias qs-logs="cd /Users/anton/DEV/myProjects/go-finance-system/queue-system && make logs-remote"
```

Тогда можешь деплоить из любой директории:
```bash
qs-deploy
qs-logs
```
