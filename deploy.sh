#!/bin/bash

# 🚀 Queue System - Deploy Script
# Автоматический деплой на vdska

set -e  # Останавливаемся при ошибке

# Цвета для вывода
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  🚀 Queue System Deploy${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Проверка наличия .env.production
if [ ! -f ".env.production" ]; then
    echo -e "${RED}❌ Ошибка: файл .env.production не найден${NC}"
    echo "Создайте файл .env.production перед деплоем"
    exit 1
fi

# Шаг 1: Сборка бинарников для Linux
echo -e "${YELLOW}📦 Шаг 1/5: Сборка бинарников для Linux...${NC}"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/api-linux ./cmd/api
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/worker-linux ./cmd/worker
echo -e "${GREEN}✅ Бинарники собраны${NC}"
echo ""

# Шаг 2: Загрузка бинарников
echo -e "${YELLOW}📤 Шаг 2/5: Загрузка бинарников на vdska...${NC}"
scp bin/api-linux root@vdska:/home/finance-system/queue-system/bin/
scp bin/worker-linux root@vdska:/home/finance-system/queue-system/bin/
echo -e "${GREEN}✅ Бинарники загружены${NC}"
echo ""

# Шаг 3: Загрузка конфигураций
echo -e "${YELLOW}📤 Шаг 3/5: Загрузка конфигураций...${NC}"
scp docker-compose.yml root@vdska:/home/finance-system/queue-system/
scp .env.production root@vdska:/home/finance-system/queue-system/.env
echo -e "${GREEN}✅ Конфигурации загружены${NC}"
echo ""

# Шаг 4: Перезапуск сервисов
echo -e "${YELLOW}🔄 Шаг 4/5: Перезапуск сервисов на vdska...${NC}"
ssh root@vdska "cd /home/finance-system/queue-system && docker compose up -d --build"
echo -e "${GREEN}✅ Сервисы перезапущены${NC}"
echo ""

# Шаг 5: Проверка статуса
echo -e "${YELLOW}🔍 Шаг 5/5: Проверка статуса...${NC}"
ssh root@vdska "cd /home/finance-system/queue-system && docker compose ps"
echo ""

# Готово
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ✅ Деплой завершён успешно!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "📊 Полезные команды:"
echo -e "  ${BLUE}Логи:${NC}       ssh root@vdska 'cd /home/finance-system/queue-system && docker compose logs -f'"
echo -e "  ${BLUE}Статус:${NC}     ssh root@vdska 'cd /home/finance-system/queue-system && docker compose ps'"
echo -e "  ${BLUE}Restart:${NC}    ssh root@vdska 'cd /home/finance-system/queue-system && docker compose restart api worker'"
echo -e "  ${BLUE}API test:${NC}   curl http://81.85.72.23:8080/health"
echo -e "  ${BLUE}Monitor:${NC}    http://81.85.72.23:8081"
echo ""
