# GoMindForge

AI-чат приложение на Go с поддержкой множества провайдеров через OpenRouter API.

**Особенности:**
- 🤖 Интеграция с DeepSeek через OpenRouter (по умолчанию)
- 🔐 JWT аутентификация с refresh токенами
- 💬 Создание и управление чатами
- 📝 Переименование чатов
- 🗄️ SQLite база данных
- 📊 Структурированное логирование
- 🐳 Docker контейнеризация
- 🚀 Готов к продакшену

## 🌐 Демо

**Рабочий сервер:** `http://94.103.91.136:8080/api/v1`

### Быстрый тест API:

```bash
# Регистрация пользователя
curl -X POST http://94.103.91.136:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com", 
    "password": "password123"
  }'

# Вход в систему
curl -X POST http://94.103.91.136:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

📖 **Подробная документация демо:** [LIVE_DEMO.md](LIVE_DEMO.md)

## 🚀 Быстрый старт

### Локальная разработка

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/Nasarwo/GoMindForge.git
cd GoMindForge
```

2. **Установите зависимости:**
```bash
go mod download
```

3. **Создайте файл `.env`:**
```env
PORT=8080
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
LOG_LEVEL=info
LOG_FORMAT=json

# AI провайдеры
OPENROUTER_API_KEY=sk-or-v1-your-openrouter-api-key-here
DEEPSEEK_API_KEY=sk-your-deepseek-api-key-here
GROK_API_KEY=your-grok-api-key-here
```

4. **Запустите миграции:**
```bash
CGO_ENABLED=1 go run cmd/migrate/main.go up
```

5. **Запустите сервер:**
```bash
air
```

### Docker развертывание

1. **Настройте переменные окружения:**
```bash
cp env.prod.example .env.prod
nano .env.prod  # Отредактируйте переменные
```

2. **Запустите с Docker:**
```bash
# Включите BuildKit для быстрой сборки
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Соберите и запустите
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d
```

3. **Проверьте работу:**
```bash
curl http://localhost:8080/api/v1/profile
```

## 📋 Требования

### Для разработки:
- Go 1.25.3 или выше
- SQLite3
- Air (для hot-reload)

### Для продакшена:
- Docker 20.10+
- Docker Compose 2.0+
- Минимум 1GB RAM
- 10GB свободного места

Сервер будет доступен на `http://localhost:8080`

## 📋 API Endpoints

### Аутентификация

#### Регистрация пользователя
```http
POST /api/v1/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "xyz123abc456...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

#### Вход в систему
```http
POST /api/v1/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123"
}
```

**Ответ:** Аналогичен ответу регистрации

#### Обновление токена
```http
POST /api/v1/refresh
Content-Type: application/json

{
  "refresh_token": "xyz123abc456..."
}
```

**Ответ:**
```json
{
  "access_token": "new_access_token...",
  "refresh_token": "new_refresh_token..."
}
```

#### Выход из системы
```http
POST /api/v1/logout
Authorization: Bearer <access_token>
```

**Ответ:**
```json
{
  "message": "logged out successfully"
}
```

#### Получение профиля
```http
GET /api/v1/profile
Authorization: Bearer <access_token>
```

**Ответ:**
```json
{
  "id": 1,
  "username": "testuser",
  "email": "test@example.com",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

---

### Чаты

#### Создание чата
```http
POST /api/v1/chats
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "ai_model": "deepseek-chat"
}
```

**Ответ:**
```json
{
  "id": 1,
  "user_id": 1,
  "ai_model": "deepseek-chat",
  "title": "Новый чат",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

> **Примечание:** Все новые чаты создаются с названием "Новый чат". Название можно изменить отдельно через API переименования.

#### Получение списка чатов
```http
GET /api/v1/chats
Authorization: Bearer <access_token>
```

**Ответ:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "ai_model": "deepseek-chat",
    "title": "Мой первый чат",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:30:00Z"
  }
]
```

#### Получение конкретного чата
```http
GET /api/v1/chats/1
Authorization: Bearer <access_token>
```

#### Переименование чата
```http
PUT /api/v1/chats/1/title
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "title": "Новое название чата"
}
```

**Ответ:**
```json
{
  "message": "chat title updated successfully"
}
```

#### Удаление чата
```http
DELETE /api/v1/chats/1
Authorization: Bearer <access_token>
```

---

### Сообщения

#### Отправка сообщения в чат
```http
POST /api/v1/chats/1/messages
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": "Привет! Расскажи про Go программирование"
}
```

**Ответ:**
```json
{
  "user_message": {
    "id": 1,
    "chat_id": 1,
    "role": "user",
    "content": "Привет! Расскажи про Go программирование",
    "created_at": "2024-01-01T12:30:00Z"
  },
  "status": "processing",
  "message": "Your message has been sent. AI response will be saved automatically."
}
```

> **Примечание:** Ответ AI сохраняется автоматически в фоне. Чтобы получить его, используйте запрос на получение истории сообщений.

#### Получение истории сообщений
```http
GET /api/v1/chats/1/messages
Authorization: Bearer <access_token>
```

**Ответ:**
```json
[
  {
    "id": 1,
    "chat_id": 1,
    "role": "user",
    "content": "Привет! Расскажи про Go программирование",
    "created_at": "2024-01-01T12:30:00Z"
  },
  {
    "id": 2,
    "chat_id": 1,
    "role": "assistant",
    "content": "Go - это язык программирования...",
    "created_at": "2024-01-01T12:30:05Z"
  }
]
```

---

## 🤖 Поддерживаемые AI провайдеры

### OpenRouter (рекомендуется, используется по умолчанию)
- **Модель по умолчанию:** `deepseek/deepseek-chat`
- **Получение API ключа:** https://openrouter.ai/
- **Переменная окружения:** `OPENROUTER_API_KEY`
- **Преимущества:** Доступ к множеству моделей через единый API, включая DeepSeek
- **Статус:** Активно используется в проекте

### DeepSeek (прямое подключение)
- **Модель по умолчанию:** `deepseek-chat`
- **Получение API ключа:** https://platform.deepseek.com/
- **Переменная окружения:** `DEEPSEEK_API_KEY`

### Grok (xAI)
- **Модель по умолчанию:** `grok-beta`
- **Получение API ключа:** https://console.x.ai/
- **Переменная окружения:** `GROK_API_KEY` или `XAI_API_KEY`

## 🔧 Конфигурация

### Переменные окружения

| Переменная | Описание | По умолчанию |
|-----------|----------|--------------|
| `PORT` | Порт сервера | `8080` |
| `JWT_SECRET` | Секретный ключ для JWT | `some-default-secret` |
| `LOG_LEVEL` | Уровень логирования (debug/info/warn/error) | `info` |
| `LOG_FORMAT` | Формат логов (json/text) | `json` |
| `OPENROUTER_API_KEY` | API ключ OpenRouter (рекомендуется) | - |
| `DEEPSEEK_API_KEY` | API ключ DeepSeek | - |
| `GROK_API_KEY` | API ключ Grok | - |

## 📝 Примеры использования cURL

### Полный цикл работы

```bash
# Базовый URL (замените на ваш сервер)
BASE_URL="http://94.103.91.136:8080/api/v1"

# 1. Регистрация
curl -X POST $BASE_URL/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# Сохраните access_token из ответа
ACCESS_TOKEN="your_access_token_here"

# 2. Создание чата (автоматически создается с названием "Новый чат")
curl -X POST $BASE_URL/chats \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ai_model": "deepseek-chat"
  }'

# 3. Переименование чата
curl -X PUT $BASE_URL/chats/1/title \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Мой тестовый чат"
  }'

# 4. Отправка сообщения
curl -X POST $BASE_URL/chats/1/messages \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Привет! Как дела?"
  }'

# 5. Получение истории сообщений
curl -X GET $BASE_URL/chats/1/messages \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 6. Обновление токена
REFRESH_TOKEN="your_refresh_token_here"
curl -X POST $BASE_URL/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }"
```

### Локальная разработка

```bash
# Для локальной разработки используйте localhost
BASE_URL="http://localhost:8080/api/v1"
```

## 🗄️ База данных

Проект использует SQLite3. Миграции находятся в `cmd/migrate/migrations/`.

### Запуск миграций

```bash
# Применить миграции
go run cmd/migrate/main.go up

# Откатить миграции
go run cmd/migrate/main.go down
```

## 🔒 Безопасность

- ✅ JWT токены с коротким временем жизни (15 минут)
- ✅ Refresh tokens с ротацией
- ✅ HTTP-only cookies для refresh tokens
- ✅ Bcrypt хеширование паролей
- ✅ Валидация всех входных данных
- ✅ Проверка прав доступа к ресурсам

## 📊 Логирование

Проект использует структурированное логирование (`log/slog`):

- **JSON формат** по умолчанию (удобно для парсинга)
- **Текстовый формат** для разработки (`LOG_FORMAT=text`)
- **Настраиваемый уровень** логирования

Пример лога:
```json
{
  "time": "2024-01-01T12:00:00Z",
  "level": "INFO",
  "msg": "AI response saved",
  "chat_id": 1,
  "message_id": 42,
  "model": "deepseek/deepseek-chat",
  "tokens": 150,
  "prompt_tokens": 25,
  "completion_tokens": 125
}
```

## 🚀 Развертывание в продакшене

### Автоматическое развертывание

```bash
# Настройте переменные окружения
cp env.prod.example .env.prod
nano .env.prod  # Отредактируйте переменные

# Запустите автоматический деплой
./scripts/deploy.sh
```

### Ручное развертывание

```bash
# 1. Подготовка
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# 2. Сборка и запуск
docker-compose -f docker-compose.prod.yml --env-file .env.prod up --build -d

# 3. Проверка
docker-compose -f docker-compose.prod.yml ps
curl http://localhost:8080/api/v1/profile
```

### Мониторинг

```bash
# Проверка статуса
./scripts/monitor.sh

# Просмотр логов
docker-compose -f docker-compose.prod.yml logs -f mindforge-api

# Обновление приложения
./scripts/update.sh
```

## 🛠️ Разработка

### Структура проекта

```
GoMindForge/
├── cmd/
│   ├── api/           # API сервер
│   └── migrate/       # Миграции БД
├── internal/
│   ├── ai/            # AI провайдеры
│   ├── database/      # Модели БД
│   └── env/           # Работа с переменными окружения
├── scripts/           # Скрипты развертывания
├── docker-compose.prod.yml  # Продакшен конфигурация
├── Dockerfile         # Docker образ
└── README.md
```

### Запуск в режиме разработки

Используйте Air для hot-reload:
```bash
air
```

Конфигурация Air находится в `.air.toml`

## 📚 Документация

- **[LIVE_DEMO.md](LIVE_DEMO.md)** - Демо сервер и примеры использования
- **[DEPLOY.md](DEPLOY.md)** - Подробная инструкция по развертыванию
- **[DOCKER.md](DOCKER.md)** - Docker конфигурация и команды
- **[DOCUMENTATION.md](DOCUMENTATION.md)** - Полная техническая документация

## 📄 Лицензия

MIT

## 🤝 Вклад

Pull requests приветствуются! Для крупных изменений сначала откройте issue для обсуждения.

## 📞 Поддержка

Если у вас возникли вопросы или проблемы:
1. Проверьте [демо сервер](http://94.103.91.136:8080/api/v1)
2. Изучите документацию выше
3. Создайте issue в репозитории

