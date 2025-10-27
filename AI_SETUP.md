# Настройка AI провайдеров

## 🔑 Получение API ключей

### DeepSeek (бесплатный)
1. Перейдите на https://platform.deepseek.com/
2. Зарегистрируйтесь или войдите
3. Перейдите в раздел API Keys
4. Создайте новый API ключ
5. **DeepSeek предоставляет бесплатные лимиты!**

### Grok (xAI)
1. Перейдите на https://console.x.ai/
2. Зарегистрируйтесь или войдите
3. Перейдите в раздел API Keys
4. Создайте новый API ключ
5. **Grok может иметь бесплатный пробный период**

## ⚙️ Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```env
# JWT секрет
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Порт сервера
PORT=8080

# DeepSeek API ключ
DEEPSEEK_API_KEY=sk-your-deepseek-api-key-here

# Grok API ключ (xAI)
GROK_API_KEY=your-grok-api-key-here
# или
XAI_API_KEY=your-grok-api-key-here
```

## 📝 Примеры использования

### Создание чата с DeepSeek
```json
POST /api/v1/chats
{
  "ai_model": "deepseek-chat",
  "title": "Чат с DeepSeek"
}
```

### Создание чата с Grok
```json
POST /api/v1/chats
{
  "ai_model": "grok-beta",
  "title": "Чат с Grok"
}
```

### Отправка сообщения
```json
POST /api/v1/chats/1/messages
{
  "content": "Привет! Расскажи про Go программирование"
}
```

**Ответ будет содержать:**
```json
{
  "user_message": {
    "id": 1,
    "chat_id": 1,
    "role": "user",
    "content": "Привет! Расскажи про Go программирование",
    "created_at": "2024-01-01T12:00:00Z"
  },
  "assistant_message": {
    "id": 2,
    "chat_id": 1,
    "role": "assistant",
    "content": "Go - это язык программирования...",
    "created_at": "2024-01-01T12:00:05Z"
  },
  "model": "deepseek-chat",
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 150,
    "total_tokens": 160
  }
}
```

## 🎯 Поддерживаемые модели

### DeepSeek
- `deepseek-chat` (по умолчанию)
- `deepseek-coder`

### Grok
- `grok-beta` (по умолчанию)
- `grok-2`

## 💡 Маппинг моделей

Система автоматически определяет провайдера по названию модели:
- Если в `ai_model` содержится "deepseek" → используется DeepSeek
- Если в `ai_model` содержится "grok" → используется Grok
- По умолчанию → DeepSeek

## 🔒 Безопасность

⚠️ **Важно:** Никогда не коммитьте файл `.env` в git!
Убедитесь, что `.env` добавлен в `.gitignore`.

## 🐛 Отладка

Если возникают проблемы с API:
1. Проверьте, что API ключи установлены в `.env`
2. Проверьте логи сервера на наличие ошибок
3. Убедитесь, что интернет-соединение стабильное
4. Проверьте лимиты API у провайдера

