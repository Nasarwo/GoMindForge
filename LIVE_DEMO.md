# 🌐 GoMindForge Live Demo

## Рабочий сервер

**URL:** `http://94.103.91.136:8080/api/v1`

**Статус:** ✅ Активен и работает

**Развертывание:** Docker контейнеры в продакшене

## 🚀 Быстрые тесты

### 1. Проверка здоровья API

```bash
curl http://94.103.91.136:8080/api/v1/profile
```

**Ожидаемый ответ:** `401 Unauthorized` (это нормально, так как нужна авторизация)

### 2. Регистрация пользователя

```bash
curl -X POST http://94.103.91.136:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo_user",
    "email": "demo@example.com",
    "password": "demo123456"
  }'
```

**Ожидаемый ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "xyz123abc456...",
  "user": {
    "id": 1,
    "username": "demo_user",
    "email": "demo@example.com",
    "created_at": "2025-10-28T...",
    "updated_at": "2025-10-28T..."
  }
}
```

### 3. Вход в систему

```bash
curl -X POST http://94.103.91.136:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "demo123456"
  }'
```

### 4. Создание чата

```bash
# Замените YOUR_ACCESS_TOKEN на полученный токен
curl -X POST http://94.103.91.136:8080/api/v1/chats \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ai_model": "deepseek-chat"
  }'
```

### 5. Отправка сообщения

```bash
curl -X POST http://94.103.91.136:8080/api/v1/chats/1/messages \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Привет! Расскажи про Go программирование"
  }'
```

## 📊 Технические детали

### Серверная конфигурация:
- **ОС:** Ubuntu Server
- **Docker:** 20.10+
- **Docker Compose:** 2.0+
- **Go:** 1.25.3
- **База данных:** SQLite3

### Контейнеры:
- **mindforge-api:** Go приложение (порт 8080)
- **nginx:** Reverse proxy (порты 80, 443)

### Переменные окружения:
- **JWT_SECRET:** Настроен для продакшена
- **OPENROUTER_API_KEY:** Активный ключ для AI
- **LOG_LEVEL:** info
- **LOG_FORMAT:** json

## 🔧 Мониторинг

### Проверка статуса контейнеров:
```bash
# На сервере
docker-compose -f docker-compose.prod.yml ps
```

### Просмотр логов:
```bash
# На сервере
docker-compose -f docker-compose.prod.yml logs -f mindforge-api
```

### Использование ресурсов:
```bash
# На сервере
docker stats
```

## 🛠️ Управление

### Перезапуск сервиса:
```bash
# На сервере
docker-compose -f docker-compose.prod.yml restart mindforge-api
```

### Обновление приложения:
```bash
# На сервере
./scripts/update.sh
```

### Мониторинг:
```bash
# На сервере
./scripts/monitor.sh
```

## 📈 Производительность

### Время отклика:
- **Регистрация:** ~100-200ms
- **Вход в систему:** ~50-100ms
- **Создание чата:** ~30-50ms
- **Отправка сообщения:** ~200-500ms (зависит от AI провайдера)

### Ресурсы:
- **RAM:** ~256MB
- **CPU:** ~0.25 cores
- **Диск:** ~100MB (включая базу данных)

## 🔒 Безопасность

### Реализованные меры:
- ✅ JWT токены с коротким временем жизни (15 минут)
- ✅ Refresh tokens с ротацией
- ✅ HTTP-only cookies для refresh tokens
- ✅ Bcrypt хеширование паролей
- ✅ Валидация всех входных данных
- ✅ Проверка прав доступа к ресурсам
- ✅ Структурированное логирование

### Рекомендации для продакшена:
- 🔒 Настройте SSL сертификаты
- 🔒 Используйте файрвол
- 🔒 Настройте регулярные бэкапы
- 🔒 Мониторьте логи на подозрительную активность

## 📞 Поддержка

При возникновении проблем с демо сервером:
1. Проверьте статус: `curl http://94.103.91.136:8080/api/v1/profile`
2. Создайте issue в репозитории
3. Обратитесь к документации в `README.md`

---

**Последнее обновление:** 28 октября 2025  
**Статус:** ✅ Активен
