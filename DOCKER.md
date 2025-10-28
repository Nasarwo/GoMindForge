# 🐳 Docker Quick Start Guide

## Быстрый старт

### 1. Локальная разработка
```bash
# Запуск в режиме разработки
./scripts/dev.sh
```

### 2. Продакшен деплой
```bash
# Настройка переменных
cp env.prod.example .env.prod
nano .env.prod  # Отредактируйте переменные

# Деплой
./scripts/deploy.sh
```

### 3. Мониторинг
```bash
# Проверка статуса
./scripts/monitor.sh

# Просмотр логов
docker-compose -f docker-compose.prod.yml logs -f mindforge-api
```

### 4. Обновление
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
