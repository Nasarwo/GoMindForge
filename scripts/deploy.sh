#!/bin/bash
# scripts/deploy.sh - Скрипт для деплоя в продакшен

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Деплой GoMindForge в продакшен${NC}"

# Проверяем наличие .env.prod файла
if [ ! -f .env.prod ]; then
    echo -e "${RED}❌ Файл .env.prod не найден!${NC}"
    echo -e "${YELLOW}📝 Создайте файл .env.prod с переменными продакшена:${NC}"
    echo "   JWT_SECRET=your-super-secret-jwt-key"
    echo "   OPENROUTER_API_KEY=your-openrouter-api-key"
    exit 1
fi

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker не установлен!${NC}"
    exit 1
fi

# Проверяем наличие docker-compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose не установлен!${NC}"
    exit 1
fi

# Создаем необходимые директории
echo -e "${YELLOW}📁 Создание директорий...${NC}"
mkdir -p ssl logs/nginx

# Останавливаем существующие контейнеры
echo -e "${YELLOW}🛑 Остановка существующих контейнеров...${NC}"
docker-compose -f docker-compose.prod.yml down

# Удаляем старые образы (опционально)
if [ "$1" = "--clean" ]; then
    echo -e "${YELLOW}🧹 Очистка старых образов...${NC}"
    docker system prune -f
fi

# Собираем и запускаем контейнеры
echo -e "${YELLOW}🔨 Сборка и запуск контейнеров...${NC}"
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

# Ждем запуска сервисов
echo -e "${YELLOW}⏳ Ожидание запуска сервисов...${NC}"
sleep 10

# Проверяем статус контейнеров
echo -e "${YELLOW}📊 Статус контейнеров:${NC}"
docker-compose -f docker-compose.prod.yml ps

# Проверяем здоровье приложения
echo -e "${YELLOW}🏥 Проверка здоровья приложения...${NC}"
if curl -f http://localhost:8080/api/v1/profile > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Приложение успешно запущено!${NC}"
    echo -e "${GREEN}🌐 API доступен по адресу: http://localhost:8080/api/v1${NC}"
else
    echo -e "${RED}❌ Приложение не отвечает!${NC}"
    echo -e "${YELLOW}📋 Логи приложения:${NC}"
    docker-compose -f docker-compose.prod.yml logs mindforge-api
    exit 1
fi

echo -e "${GREEN}🎉 Деплой завершен успешно!${NC}"
