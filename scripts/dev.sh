#!/bin/bash
# scripts/dev.sh - Скрипт для локальной разработки

set -e

echo "🚀 Запуск GoMindForge в режиме разработки..."

# Проверяем наличие .env файла
if [ ! -f .env ]; then
    echo "❌ Файл .env не найден!"
    echo "📝 Создайте файл .env с необходимыми переменными:"
    echo "   OPENROUTER_API_KEY=your-api-key-here"
    echo "   JWT_SECRET=your-jwt-secret"
    exit 1
fi

# Создаем необходимые директории
mkdir -p data logs

# Останавливаем существующие контейнеры
echo "🛑 Остановка существующих контейнеров..."
docker-compose down

# Собираем и запускаем контейнеры
echo "🔨 Сборка и запуск контейнеров..."
docker-compose up --build

echo "✅ Сервер запущен на http://localhost:8080"
