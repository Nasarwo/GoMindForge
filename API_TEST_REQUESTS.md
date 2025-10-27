# API Тестовые запросы

Базовый URL: `http://localhost:8080`

## 1. Регистрация пользователя

**POST** `/api/v1/register`

### Тело запроса:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

### cURL:
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Пример ответа:
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

## 2. Вход (Login)

**POST** `/api/v1/login`

### Тело запроса:
```json
{
  "email": "test@example.com",
  "password": "password123"
}
```

### cURL:
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Пример ответа:
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

## 3. Обновление токена (Refresh Token)

**POST** `/api/v1/refresh`

### Вариант 1: Использование Cookie (автоматически для браузера)
Если refresh token был установлен в cookie при логине/регистрации, браузер автоматически отправит его.

### Вариант 2: JSON Body (для мобильных приложений)

#### Тело запроса:
```json
{
  "refresh_token": "xyz123abc456..."
}
```

### cURL с cookie:
```bash
curl -X POST http://localhost:8080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -b "refresh_token=xyz123abc456..."
```

### cURL с JSON body:
```bash
curl -X POST http://localhost:8080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "xyz123abc456..."
  }'
```

### Пример ответа:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "new_refresh_token_xyz..."
}
```

---

## 4. Получение профиля пользователя

**GET** `/api/v1/profile`

### Заголовки:
```
Authorization: Bearer <access_token>
```

### cURL:
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Пример ответа:
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

## 5. Выход (Logout)

**POST** `/api/v1/logout`

### Вариант 1: С cookie (автоматически для браузера)
Браузер автоматически отправит refresh_token из cookie.

### Вариант 2: Без cookie (просто вызывает endpoint)

### cURL:
```bash
curl -X POST http://localhost:8080/api/v1/logout \
  -H "Content-Type: application/json" \
  -b "refresh_token=xyz123abc456..."
```

### Пример ответа:
```json
{
  "message": "logged out successfully"
}
```

---

## Примеры ошибок

### Ошибка валидации (400):
```json
{
  "status": 400,
  "message": "Key: 'registerRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag",
  "code": "VALIDATION_ERROR"
}
```

### Неверные учетные данные (401):
```json
{
  "status": 401,
  "message": "invalid credentials",
  "code": "INVALID_CREDENTIALS"
}
```

### Пользователь уже существует (409):
```json
{
  "status": 409,
  "message": "user with this email already exists",
  "code": "USER_ALREADY_EXISTS"
}
```

### Неверный токен (401):
```json
{
  "status": 401,
  "message": "invalid token",
  "code": "INVALID_TOKEN"
}
```

---

## Полный тестовый сценарий

### 1. Регистрация нового пользователя:
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

**Сохраните `access_token` и `refresh_token` из ответа!**

### 2. Получение профиля:
```bash
ACCESS_TOKEN="ваш_access_token_здесь"

curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 3. Обновление токена (через 15 минут, когда access token истечет):
```bash
REFRESH_TOKEN="ваш_refresh_token_здесь"

curl -X POST http://localhost:8080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }"
```

### 4. Выход:
```bash
curl -X POST http://localhost:8080/api/v1/logout \
  -H "Content-Type: application/json" \
  -b "refresh_token=$REFRESH_TOKEN"
```

---

## Примечания

- **Access Token** живет 15 минут
- **Refresh Token** живет 7 дней
- При обновлении токена старый refresh token автоматически удаляется (token rotation)
- Refresh token хранится в HTTP-only cookie для веб-приложений
- Для мобильных приложений можно использовать JSON body

