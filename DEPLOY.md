# Инструкция по развертыванию в продакшене

## Предварительные требования

- Docker и Docker Compose установлены на сервере
- Доступ к серверу с правами sudo
- Домен настроен (если используете SSL)
- SSL сертификаты (если используете HTTPS)

## Шаг 1: Подготовка сервера

```bash
# Обновление системы (Ubuntu/Debian)
sudo apt update && sudo apt upgrade -y

# Установка Docker (если не установлен)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Установка Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER
```

## Шаг 2: Клонирование проекта

```bash
# Клонируйте репозиторий
git clone <repository-url>
cd GoMindForge

# Или загрузите проект на сервер другим способом
```

## Шаг 3: Настройка переменных окружения

```bash
# Создайте файл с переменными окружения
cp env.prod.example .env.prod

# Отредактируйте файл
nano .env.prod
```

**Важно:** Установите следующие переменные:
- `JWT_SECRET` - случайная строка минимум 32 символа
- `DB_PASSWORD` - надежный пароль для PostgreSQL
- `DEEPSEEK_API_KEY` - ваш API ключ DeepSeek
- `GIGACHAT_AUTH_KEY` - ваш GigaChat authorization key
- `GIGACHAT_CLIENT_ID` - ваш GigaChat client ID
- `QWEN_API_KEY` - ваш Qwen API ключ

## Шаг 4: Настройка SSL (опционально, но рекомендуется)

```bash
# Создайте директорию для SSL сертификатов
mkdir -p ssl

# Поместите ваши SSL сертификаты:
# - ssl/cert.pem (SSL сертификат)
# - ssl/key.pem (Приватный ключ)

# Если используете Let's Encrypt:
sudo certbot certonly --standalone -d your-domain.com
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem ssl/cert.pem
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem ssl/key.pem
sudo chmod 644 ssl/cert.pem
sudo chmod 600 ssl/key.pem
```

**Обновите `nginx.prod.conf`:**
- Замените `your-domain.com` на ваш домен
- Убедитесь, что пути к сертификатам правильные

## Шаг 5: Применение миграций

```bash
# Запустите миграции через бинарник migrate в контейнере
docker-compose -f docker-compose.prod.yml --env-file .env.prod run --rm mindforge-api ./migrate up

# Или откат миграций (если нужно)
docker-compose -f docker-compose.prod.yml --env-file .env.prod run --rm mindforge-api ./migrate down
```

## Шаг 6: Запуск приложения

```bash
# Сборка и запуск
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

# Проверка статуса
docker-compose -f docker-compose.prod.yml ps

# Просмотр логов
docker-compose -f docker-compose.prod.yml logs -f
```

## Шаг 7: Проверка работоспособности

```bash
# Проверка health check
curl http://localhost:8080/health

# Или через Nginx (если настроен)
curl https://your-domain.com/health

# Проверка API
curl https://your-domain.com/api/v1/health
```

## Управление приложением

### Просмотр логов

```bash
# Все сервисы
docker-compose -f docker-compose.prod.yml logs -f

# Только API
docker-compose -f docker-compose.prod.yml logs -f mindforge-api

# Только PostgreSQL
docker-compose -f docker-compose.prod.yml logs -f postgres

# Только Nginx
docker-compose -f docker-compose.prod.yml logs -f nginx
```

### Перезапуск сервисов

```bash
# Перезапуск всех сервисов
docker-compose -f docker-compose.prod.yml restart

# Перезапуск конкретного сервиса
docker-compose -f docker-compose.prod.yml restart mindforge-api
```

### Остановка и запуск

```bash
# Остановка
docker-compose -f docker-compose.prod.yml stop

# Запуск
docker-compose -f docker-compose.prod.yml start

# Остановка и удаление контейнеров
docker-compose -f docker-compose.prod.yml down

# Остановка и удаление контейнеров + volumes (ОСТОРОЖНО!)
docker-compose -f docker-compose.prod.yml down -v
```

### Обновление приложения

```bash
# 1. Получите последние изменения
git pull

# 2. Пересоберите и перезапустите
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

# 3. Примените новые миграции (если есть)
docker-compose -f docker-compose.prod.yml run --rm mindforge-api \
  go run cmd/migrate/main.go up
```

## Резервное копирование базы данных

```bash
# Создание бэкапа
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U postgres mindforge > backup_$(date +%Y%m%d_%H%M%S).sql

# Восстановление из бэкапа
docker-compose -f docker-compose.prod.yml exec -T postgres psql -U postgres mindforge < backup_YYYYMMDD_HHMMSS.sql
```

## Мониторинг

### Проверка использования ресурсов

```bash
# Статус контейнеров
docker-compose -f docker-compose.prod.yml ps

# Использование ресурсов
docker stats

# Проверка health check
docker inspect --format='{{.State.Health.Status}}' <container_name>
```

### Настройка systemd (опционально)

Создайте файл `/etc/systemd/system/mindforge.service`:

```ini
[Unit]
Description=MindForge API Docker Compose Service
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/path/to/GoMindForge
ExecStart=/usr/local/bin/docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose.prod.yml down
User=your-user
Group=your-group

[Install]
WantedBy=multi-user.target
```

Активируйте сервис:

```bash
sudo systemctl enable mindforge
sudo systemctl start mindforge
```

## Устранение неполадок

### Проблемы с подключением к БД

```bash
# Проверка статуса PostgreSQL
docker-compose -f docker-compose.prod.yml exec postgres pg_isready -U postgres

# Просмотр логов PostgreSQL
docker-compose -f docker-compose.prod.yml logs postgres
```

### Проблемы с API

```bash
# Проверка логов API
docker-compose -f docker-compose.prod.yml logs mindforge-api

# Проверка health check
curl http://localhost:8080/health

# Вход в контейнер для отладки
docker-compose -f docker-compose.prod.yml exec mindforge-api sh
```

### Проблемы с Nginx

```bash
# Проверка конфигурации
docker-compose -f docker-compose.prod.yml exec nginx nginx -t

# Просмотр логов
docker-compose -f docker-compose.prod.yml logs nginx
```

## Безопасность

1. **Никогда не коммитьте `.env.prod` в Git**
2. **Используйте сильные пароли для БД и JWT_SECRET**
3. **Регулярно обновляйте Docker образы**
4. **Настройте firewall для ограничения доступа**
5. **Используйте SSL/TLS в продакшене**
6. **Регулярно создавайте резервные копии БД**

## Полезные команды

```bash
# Очистка неиспользуемых образов
docker system prune -a

# Просмотр использования дискового пространства
docker system df

# Просмотр сетей
docker network ls

# Просмотр volumes
docker volume ls
```
