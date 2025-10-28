# 🐳 Docker Quick Start Guide

## 🌐 Демо сервер

**Рабочий сервер:** `http://94.103.91.136:8080/api/v1`

### Быстрый тест:
```bash
curl http://94.103.91.136:8080/api/v1/profile
```

## Быстрый старт

### 1. Локальная разработка
```bash
# Запуск в режиме разработки
./scripts/dev.sh
```

### 2. Продакшен деплой (оптимизированный)
```bash
# Включите BuildKit для быстрой сборки
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Настройка переменных
cp env.prod.example .env.prod
nano .env.prod  # Отредактируйте переменные

# Автоматический деплой
./scripts/deploy.sh
```

### 3. Ручной деплой
```bash
# Включите BuildKit
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Создайте необходимые директории
mkdir -p ssl logs/nginx

# Запустите миграции локально
CGO_ENABLED=1 go run cmd/migrate/main.go up

# Соберите и запустите
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

# Скопируйте базу данных в Docker volume
docker run --rm -v gomindforge_mindforge_data:/app/data -v $(pwd):/workspace alpine sh -c "cp /workspace/data.db /app/data/data.db && chown 1001:1001 /app/data/data.db"
```

### 4. Мониторинг
```bash
# Проверка статуса
./scripts/monitor.sh

# Просмотр логов
docker-compose -f docker-compose.prod.yml logs -f mindforge-api
```

### 5. Обновление
```bash
# Обновление приложения
./scripts/update.sh
```

## Структура файлов

```
GoMindForge/
├── Dockerfile                 # Образ приложения
├── docker-compose.yml         # Локальная разработка
├── docker-compose.prod.yml    # Продакшен
├── nginx.prod.conf           # Конфигурация Nginx
├── env.prod.example          # Пример переменных
├── scripts/
│   ├── dev.sh               # Разработка
│   ├── deploy.sh            # Деплой
│   ├── update.sh            # Обновление
│   └── monitor.sh           # Мониторинг
└── DEPLOY.md                # Подробная инструкция
```

## Переменные окружения

### Обязательные:
- `JWT_SECRET` - секрет для JWT токенов
- `OPENROUTER_API_KEY` - API ключ OpenRouter

### Опциональные:
- `PORT` - порт сервера (по умолчанию: 8080)
- `LOG_LEVEL` - уровень логирования (info/debug/warn/error)
- `LOG_FORMAT` - формат логов (json/text)

## Полезные команды

```bash
# Проверка статуса
docker-compose -f docker-compose.prod.yml ps

# Перезапуск сервиса
docker-compose -f docker-compose.prod.yml restart mindforge-api

# Очистка неиспользуемых образов
docker system prune -f

# Просмотр использования ресурсов
docker stats

# Быстрая сборка с BuildKit
export DOCKER_BUILDKIT=1
docker-compose -f docker-compose.prod.yml build --no-cache --parallel

# Проверка логов в реальном времени
docker-compose -f docker-compose.prod.yml logs -f mindforge-api

# Вход в контейнер для отладки
docker exec -it gomindforge-mindforge-api-1 sh

# Проверка размера образов
docker images | grep gomindforge
```

## Безопасность

1. **Обязательно измените JWT_SECRET** на случайную строку минимум 32 символа
2. **Настройте файрвол** - разрешите только порты 80, 443, 22
3. **Используйте HTTPS** - настройте SSL сертификаты
4. **Регулярные бэкапы** - настройте автоматические бэкапы базы данных

## Поддержка

При возникновении проблем:
1. Проверьте логи: `docker-compose -f docker-compose.prod.yml logs`
2. Проверьте статус контейнеров: `docker ps`
3. Проверьте переменные окружения: `docker-compose -f docker-compose.prod.yml config`
4. Обратитесь к подробной инструкции в `DEPLOY.md`
