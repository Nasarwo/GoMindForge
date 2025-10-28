#!/bin/bash
# scripts/update.sh - Скрипт для обновления приложения

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔄 Обновление GoMindForge${NC}"

# Проверяем наличие .env.prod файла
if [ ! -f .env.prod ]; then
    echo -e "${RED}❌ Файл .env.prod не найден!${NC}"
    exit 1
fi

# Создаем бэкап базы данных
echo -e "${YELLOW}💾 Создание бэкапа базы данных...${NC}"
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

if [ -d "data" ]; then
    cp -r data "$BACKUP_DIR/"
    echo -e "${GREEN}✅ Бэкап создан в $BACKUP_DIR${NC}"
fi

# Останавливаем контейнеры
echo -e "${YELLOW}🛑 Остановка контейнеров...${NC}"
docker-compose -f docker-compose.prod.yml down

# Обновляем код (если используется git)
if [ -d ".git" ]; then
    echo -e "${YELLOW}📥 Обновление кода из git...${NC}"
    git pull origin main
fi

# Пересобираем и запускаем контейнеры
echo -e "${YELLOW}🔨 Пересборка и запуск контейнеров...${NC}"
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

# Ждем запуска
echo -e "${YELLOW}⏳ Ожидание запуска...${NC}"
sleep 10

# Проверяем здоровье
echo -e "${YELLOW}🏥 Проверка здоровья...${NC}"
if curl -f http://localhost:8080/api/v1/profile > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Обновление завершено успешно!${NC}"
else
    echo -e "${RED}❌ Ошибка при обновлении!${NC}"
    echo -e "${YELLOW}📋 Логи:${NC}"
    docker-compose -f docker-compose.prod.yml logs mindforge-api
    exit 1
fi

echo -e "${GREEN}🎉 Приложение обновлено!${NC}"
