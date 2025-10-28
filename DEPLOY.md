# 🚀 Инструкция по деплою GoMindForge в продакшен

## 🌐 Демо сервер

**Рабочий сервер:** `http://94.103.91.136:8080/api/v1`

### Быстрый тест:
```bash
curl http://94.103.91.136:8080/api/v1/profile
```

## 📋 Предварительные требования

### На сервере должно быть установлено:
- **Docker** (версия 20.10+)
- **Docker Compose** (версия 2.0+)
- **Git** (для клонирования репозитория)
- **curl** (для проверки здоровья)
- **Go 1.25.3+** (для миграций)

### Проверка установки:
```bash
docker --version
docker-compose --version
git --version
curl --version
go version
```

## 🔧 Подготовка к деплою

### 1. Подготовка сервера

```bash
# Обновляем систему
sudo apt update && sudo apt upgrade -y

# Устанавливаем Docker (если не установлен)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Устанавливаем Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Устанавливаем Go 1.25.3
wget https://go.dev/dl/go1.25.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Устанавливаем зависимости для компиляции
sudo apt install -y gcc sqlite3 libsqlite3-dev

# Перезагружаемся для применения изменений группы
sudo reboot
```

### 2. Клонирование репозитория

```bash
# Клонируем репозиторий
git clone https://github.com/Nasarwo/GoMindForge.git
cd GoMindForge

# Делаем скрипты исполняемыми
chmod +x scripts/*.sh
```

### 3. Настройка переменных окружения

```bash
# Создаем файл с переменными продакшена
cp env.prod.example .env.prod

# Редактируем файл
nano .env.prod
```

**Обязательно измените в .env.prod:**
- `JWT_SECRET` - случайная строка минимум 32 символа
- `OPENROUTER_API_KEY` - ваш реальный API ключ OpenRouter

### 4. Настройка домена и SSL (опционально)

Если у вас есть домен и SSL сертификаты:

```bash
# Создаем директорию для SSL
mkdir -p ssl

# Копируем ваши SSL сертификаты
cp your-cert.pem ssl/cert.pem
cp your-key.pem ssl/key.pem

# Редактируем nginx конфигурацию
nano nginx.prod.conf
```

В файле `nginx.prod.conf` замените:
- `your-domain.com` на ваш реальный домен
- Пути к SSL сертификатам

## 🚀 Деплой

### Вариант 1: Автоматический деплой (рекомендуется)

```bash
# Включаем BuildKit для быстрой сборки
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Запускаем автоматический деплой
./scripts/deploy.sh

# Проверяем статус
docker-compose -f docker-compose.prod.yml ps
```

### Вариант 2: Ручной деплой

```bash
# Включаем BuildKit
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Создаем необходимые директории
mkdir -p ssl logs/nginx

# Запускаем миграции локально
CGO_ENABLED=1 go run cmd/migrate/main.go up

# Собираем и запускаем контейнеры
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

# Копируем базу данных в Docker volume
docker run --rm -v gomindforge_mindforge_data:/app/data -v $(pwd):/workspace alpine sh -c "cp /workspace/data.db /app/data/data.db && chown 1001:1001 /app/data/data.db"

# Проверяем статус
docker-compose -f docker-compose.prod.yml ps
```

### Вариант 3: Деплой с Nginx (для продакшена)

```bash
# Настройте SSL сертификаты в директории ssl/
# Отредактируйте nginx.prod.conf с вашим доменом

# Запускаем с Nginx
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

# Проверяем статус
docker-compose -f docker-compose.prod.yml ps
```

## 🔍 Проверка работы

### 1. Проверка контейнеров
```bash
docker-compose -f docker-compose.prod.yml ps
```

Должны быть запущены:
- `mindforge-api` (статус: Up)
- `nginx` (если используется)

### 2. Проверка API
```bash
# Проверяем здоровье API
curl http://localhost:8080/api/v1/profile

# Тест регистрации пользователя
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# Тест входа в систему
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Или через Nginx (если настроен)
curl https://your-domain.com/api/v1/profile
```

### 3. Проверка логов
```bash
# Логи приложения
docker-compose -f docker-compose.prod.yml logs mindforge-api

# Логи Nginx
docker-compose -f docker-compose.prod.yml logs nginx

# Логи в реальном времени
docker-compose -f docker-compose.prod.yml logs -f mindforge-api
```

## 📊 Мониторинг

### Использование скрипта мониторинга
```bash
./scripts/monitor.sh
```

### Ручная проверка ресурсов
```bash
# Использование ресурсов
docker stats

# Место на диске
df -h

# Размер базы данных
du -h data/data.db
```

## 🔄 Обновление приложения

### Автоматическое обновление
```bash
./scripts/update.sh
```

### Ручное обновление
```bash
# Останавливаем контейнеры
docker-compose -f docker-compose.prod.yml down

# Обновляем код
git pull origin main

# Пересобираем и запускаем
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d
```

## 🛠️ Устранение неполадок

### Проблема: Контейнер не запускается
```bash
# Проверяем логи
docker-compose -f docker-compose.prod.yml logs mindforge-api

# Проверяем переменные окружения
docker-compose -f docker-compose.prod.yml config
```

### Проблема: API не отвечает
```bash
# Проверяем, что контейнер запущен
docker ps

# Проверяем порты
netstat -tlnp | grep 8080

# Проверяем здоровье
curl -v http://localhost:8080/api/v1/profile
```

### Проблема: Ошибки базы данных
```bash
# Проверяем права доступа к файлу БД
ls -la data/

# Проверяем логи приложения
docker-compose -f docker-compose.prod.yml logs mindforge-api | grep -i error
```

## 🔒 Безопасность

### 1. Настройка файрвола
```bash
# Разрешаем только необходимые порты
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw enable
```

### 2. Регулярные бэкапы
```bash
# Создаем скрипт бэкапа
cat > backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp -r data "$BACKUP_DIR/"
echo "Backup created: $BACKUP_DIR"
EOF

chmod +x backup.sh

# Добавляем в cron для ежедневных бэкапов
echo "0 2 * * * /path/to/GoMindForge/backup.sh" | crontab -
```

### 3. Мониторинг безопасности
```bash
# Проверяем логи на подозрительную активность
docker-compose -f docker-compose.prod.yml logs nginx | grep -E "(40[0-9]|50[0-9])"

# Мониторим использование ресурсов
watch -n 5 'docker stats --no-stream'
```

## 📈 Масштабирование

### Горизонтальное масштабирование
```yaml
# В docker-compose.prod.yml
services:
  mindforge-api:
    # ... существующая конфигурация
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 256M
          cpus: '0.25'
```

### Вертикальное масштабирование
```yaml
# Увеличиваем лимиты ресурсов
services:
  mindforge-api:
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '1.0'
```

## 🆘 Поддержка

### Полезные команды
```bash
# Перезапуск сервиса
docker-compose -f docker-compose.prod.yml restart mindforge-api

# Очистка неиспользуемых образов
docker system prune -f

# Просмотр всех контейнеров
docker ps -a

# Вход в контейнер для отладки
docker exec -it gomindforge_mindforge-api_1 sh
```

### Логи и отладка
```bash
# Все логи
docker-compose -f docker-compose.prod.yml logs

# Логи конкретного сервиса
docker-compose -f docker-compose.prod.yml logs mindforge-api

# Логи с временными метками
docker-compose -f docker-compose.prod.yml logs -t mindforge-api
```

---

## ✅ Чек-лист деплоя

- [ ] Docker и Docker Compose установлены
- [ ] Репозиторий склонирован
- [ ] Файл `.env.prod` создан и настроен
- [ ] SSL сертификаты настроены (если нужны)
- [ ] Скрипты сделаны исполняемыми
- [ ] Контейнеры запущены
- [ ] API отвечает на запросы
- [ ] Логи не содержат ошибок
- [ ] Мониторинг настроен
- [ ] Бэкапы настроены

**🎉 Поздравляем! Ваш GoMindForge успешно развернут в продакшене!**
