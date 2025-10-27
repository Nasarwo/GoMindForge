# Примеры JSON запросов для API GoMindForge

## 🔐 Аутентификация

### 1. Регистрация
**POST** `/api/v1/register`

```json
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

---

### 2. Вход (Login)
**POST** `/api/v1/login`

```json
{
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

---

### 3. Обновление токена
**POST** `/api/v1/refresh`

```json
{
  "refresh_token": "xyz123abc456..."
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "new_refresh_token_xyz..."
}
```

---

### 4. Выход (Logout)
**POST** `/api/v1/logout`

Тело запроса не требуется (может быть пустым `{}`)

**Ответ:**
```json
{
  "message": "logged out successfully"
}
```

---

## 💬 Чаты

### 5. Создание чата
**POST** `/api/v1/chats`  
**Headers:** `Authorization: Bearer YOUR_ACCESS_TOKEN`

```json
{
  "ai_model": "gpt-4",
  "title": "Мой первый чат"
}
```

**Минимальный запрос (без title):**
```json
{
  "ai_model": "gpt-4"
}
```

**Ответ:**
```json
{
  "id": 1,
  "user_id": 1,
  "ai_model": "gpt-4",
  "title": "Мой первый чат",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

---

### 6. Получение списка чатов
**GET** `/api/v1/chats`  
**Headers:** `Authorization: Bearer YOUR_ACCESS_TOKEN`

Тело запроса не требуется

**Ответ:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "ai_model": "gpt-4",
    "title": "Мой первый чат",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:30:00Z"
  },
  {
    "id": 2,
    "user_id": 1,
    "ai_model": "claude-3",
    "title": "Чат с Claude",
    "created_at": "2024-01-01T13:00:00Z",
    "updated_at": "2024-01-01T13:15:00Z"
  }
]
```

---

### 7. Получение конкретного чата
**GET** `/api/v1/chats/1`  
**Headers:** `Authorization: Bearer YOUR_ACCESS_TOKEN`

Тело запроса не требуется

**Ответ:**
```json
{
  "id": 1,
  "user_id": 1,
  "ai_model": "gpt-4",
  "title": "Мой первый чат",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:30:00Z"
}
```

---

### 8. Удаление чата
**DELETE** `/api/v1/chats/1`  
**Headers:** `Authorization: Bearer YOUR_ACCESS_TOKEN`

Тело запроса не требуется

**Ответ:**
```json
{
  "message": "chat deleted successfully"
}
```

---

## 📨 Сообщения

### 9. Отправка сообщения в чат
**POST** `/api/v1/chats/1/messages`  
**Headers:** `Authorization: Bearer YOUR_ACCESS_TOKEN`

```json
{
  "content": "Привет! Как дела?"
}
```

**Ответ:**
```json
{
  "id": 1,
  "chat_id": 1,
  "role": "user",
  "content": "Привет! Как дела?",
  "created_at": "2024-01-01T12:30:00Z"
}
```

---

### 10. Получение истории сообщений чата
**GET** `/api/v1/chats/1/messages`  
**Headers:** `Authorization: Bearer YOUR_ACCESS_TOKEN`

Тело запроса не требуется

**Ответ:**
```json
[
  {
    "id": 1,
    "chat_id": 1,
    "role": "user",
    "content": "Привет! Как дела?",
    "created_at": "2024-01-01T12:30:00Z"
  },
  {
    "id": 2,
    "chat_id": 1,
    "role": "assistant",
    "content": "Привет! У меня всё отлично, спасибо! Чем могу помочь?",
    "created_at": "2024-01-01T12:30:05Z"
  },
  {
    "id": 3,
    "chat_id": 1,
    "role": "user",
    "content": "Расскажи про Go",
    "created_at": "2024-01-01T12:31:00Z"
  }
]
```

---

## ❌ Примеры ошибок

### Ошибка валидации (400)
```json
{
  "status": 400,
  "message": "Key: 'createChatRequest.AIModel' Error:Field validation for 'AIModel' failed on the 'required' tag",
  "code": "VALIDATION_ERROR"
}
```

### Неверные учетные данные (401)
```json
{
  "status": 401,
  "message": "invalid credentials",
  "code": "INVALID_CREDENTIALS"
}
```

### Неверный токен (401)
```json
{
  "status": 401,
  "message": "invalid token",
  "code": "INVALID_TOKEN"
}
```

### Чат не найден (404)
```json
{
  "status": 404,
  "message": "chat not found",
  "code": "CHAT_NOT_FOUND"
}
```

### Доступ запрещен (403)
```json
{
  "status": 403,
  "message": "forbidden",
  "code": "FORBIDDEN"
}
```

---

## 📋 Полный пример использования

### Шаг 1: Регистрация
```json
POST /api/v1/register
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securepass123"
}
```

### Шаг 2: Создание чата
```json
POST /api/v1/chats
Authorization: Bearer YOUR_ACCESS_TOKEN
{
  "ai_model": "gpt-4",
  "title": "Мой чат с GPT-4"
}
```

### Шаг 3: Отправка сообщения
```json
POST /api/v1/chats/1/messages
Authorization: Bearer YOUR_ACCESS_TOKEN
{
  "content": "Расскажи про программирование на Go"
}
```

### Шаг 4: Получение истории
```json
GET /api/v1/chats/1/messages
Authorization: Bearer YOUR_ACCESS_TOKEN
```

### Шаг 5: Получение всех чатов
```json
GET /api/v1/chats
Authorization: Bearer YOUR_ACCESS_TOKEN
```

