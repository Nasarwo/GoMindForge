#!/bin/bash
# scripts/monitor.sh - Скрипт для мониторинга приложения

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}📊 Мониторинг GoMindForge${NC}"
echo "================================"

# Проверяем статус контейнеров
echo -e "${YELLOW}🐳 Статус контейнеров:${NC}"
docker-compose -f docker-compose.prod.yml ps

echo ""

# Проверяем использование ресурсов
echo -e "${YELLOW}💻 Использование ресурсов:${NC}"
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"

echo ""

# Проверяем здоровье приложения
echo -e "${YELLOW}🏥 Проверка здоровья API:${NC}"
if curl -f http://localhost:8080/api/v1/profile > /dev/null 2>&1; then
    echo -e "${GREEN}✅ API отвечает${NC}"
else
    echo -e "${RED}❌ API не отвечает${NC}"
fi

echo ""

# Показываем последние логи
echo -e "${YELLOW}📋 Последние логи (10 строк):${NC}"
docker-compose -f docker-compose.prod.yml logs --tail=10 mindforge-api

echo ""

# Показываем размер базы данных
if [ -f "data/data.db" ]; then
    DB_SIZE=$(du -h data/data.db | cut -f1)
    echo -e "${YELLOW}💾 Размер базы данных: $DB_SIZE${NC}"
fi

echo ""

# Показываем доступное место на диске
echo -e "${YELLOW}💽 Использование диска:${NC}"
df -h / | tail -1 | awk '{print "Использовано: " $3 " / " $2 " (" $5 ")"}'

echo ""
echo -e "${BLUE}================================"
echo -e "Для просмотра логов в реальном времени:"
echo -e "docker-compose -f docker-compose.prod.yml logs -f mindforge-api${NC}"
