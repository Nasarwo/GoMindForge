# GoMindForge

### Требования
- Go 1.25.3 или выше
- SQLite3

### Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/Nasarwo/GoMindForge.git
cd GoMindForge
```

2. Установите зависимости:
```bash
go mod download
```

3. Создайте файл `.env`:
```env
PORT=5000
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
LOG_LEVEL=info
LOG_FORMAT=json

# AI провайдеры (временно только эти)
DEEPSEEK_API_KEY=sk-your-deepseek-api-key-here
GROK_API_KEY=your-grok-api-key-here
```

4. Запустите миграции:
```bash
go run cmd/migrate/main.go up
```

5. Запустите сервер:
```bash
air
```

Сервер будет доступен на `http://localhost:5000`

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
  "ai_model": "deepseek-chat",
  "title": "Мой первый чат"
}
```

**Ответ:**
```json
{
  "id": 1,
  "user_id": 1,
  "ai_model": "deepseek-chat",
  "title": "Мой первый чат",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

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

### DeepSeek
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
| `PORT` | Порт сервера | `5000` |
| `JWT_SECRET` | Секретный ключ для JWT | `some-default-secret` |
| `LOG_LEVEL` | Уровень логирования (debug/info/warn/error) | `info` |
| `LOG_FORMAT` | Формат логов (json/text) | `json` |
| `DEEPSEEK_API_KEY` | API ключ DeepSeek | - |
| `GROK_API_KEY` | API ключ Grok | - |

## 📝 Примеры использования cURL

### Полный цикл работы

```bash
# 1. Регистрация
curl -X POST http://localhost:5000/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# Сохраните access_token из ответа
ACCESS_TOKEN="your_access_token_here"

# 2. Создание чата
curl -X POST http://localhost:5000/api/v1/chats \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ai_model": "deepseek-chat",
    "title": "Тестовый чат"
  }'

# 3. Отправка сообщения
curl -X POST http://localhost:5000/api/v1/chats/1/messages \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Привет! Как дела?"
  }'

# 4. Получение истории
curl -X GET http://localhost:5000/api/v1/chats/1/messages \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 5. Обновление токена
REFRESH_TOKEN="your_refresh_token_here"
curl -X POST http://localhost:5000/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }"
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
  "model": "deepseek-chat",
  "tokens": 150
}
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
├── go.mod
└── README.md
```

### Запуск в режиме разработки

Используйте Air для hot-reload:
```bash
air
```

Конфигурация Air находится в `.air.toml`

## 📄 Лицензия

MIT

## 🤝 Вклад

Pull requests приветствуются! Для крупных изменений сначала откройте issue для обсуждения.

## 📞 Поддержка

Если у вас возникли вопросы или проблемы, создайте issue в репозитории.

