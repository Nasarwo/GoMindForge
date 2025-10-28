# MindForge API

## 🌐 Продакшен сервер

**API Base URL:** `https://94.103.91.136:8080/api/v1`

### Быстрый тест API:

```bash
# Регистрация пользователя
curl -X POST https://94.103.91.136:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com", 
    "password": "password123"
  }'

# Вход в систему
curl -X POST https://94.103.91.136:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

## 📱 Интеграция с React Native

### Установка зависимостей

```bash
npm install axios @react-native-async-storage/async-storage
# или
yarn add axios @react-native-async-storage/async-storage
```

### Настройка API клиента

```javascript
// api/client.js
import axios from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';

const API_BASE_URL = 'https://94.103.91.136:8080/api/v1';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
});

// Interceptor для добавления токена
apiClient.interceptors.request.use(async (config) => {
  const token = await AsyncStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Interceptor для обработки ошибок и обновления токена
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      const refreshToken = await AsyncStorage.getItem('refresh_token');
      if (refreshToken) {
        try {
          const response = await axios.post(`${API_BASE_URL}/refresh`, {
            refresh_token: refreshToken,
          });
          
          const { access_token } = response.data;
          await AsyncStorage.setItem('access_token', access_token);
          
          // Повторяем оригинальный запрос
          error.config.headers.Authorization = `Bearer ${access_token}`;
          return apiClient.request(error.config);
        } catch (refreshError) {
          // Refresh token недействителен, перенаправляем на логин
          await AsyncStorage.multiRemove(['access_token', 'refresh_token']);
          // Здесь можно добавить навигацию к экрану логина
        }
      }
    }
    return Promise.reject(error);
  }
);

export default apiClient;
```

### Сервисы для работы с API

```javascript
// services/authService.js
import apiClient from '../api/client';
import AsyncStorage from '@react-native-async-storage/async-storage';

export const authService = {
  // Регистрация
  async register(userData) {
    const response = await apiClient.post('/register', userData);
    const { access_token, refresh_token, user } = response.data;
    
    await AsyncStorage.multiSet([
      ['access_token', access_token],
      ['refresh_token', refresh_token],
    ]);
    
    return user;
  },

  // Вход
  async login(email, password) {
    const response = await apiClient.post('/login', { email, password });
    const { access_token, refresh_token, user } = response.data;
    
    await AsyncStorage.multiSet([
      ['access_token', access_token],
      ['refresh_token', refresh_token],
    ]);
    
    return user;
  },

  // Выход
  async logout() {
    try {
      await apiClient.post('/logout');
    } finally {
      await AsyncStorage.multiRemove(['access_token', 'refresh_token']);
    }
  },

  // Получение профиля
  async getProfile() {
    const response = await apiClient.get('/profile');
    return response.data.user;
  },
};
```

```javascript
// services/chatService.js
import apiClient from '../api/client';

export const chatService = {
  // Создание чата
  async createChat(aiModel = 'deepseek-chat') {
    const response = await apiClient.post('/chats', { ai_model: aiModel });
    return response.data;
  },

  // Получение списка чатов
  async getChats() {
    const response = await apiClient.get('/chats');
    return response.data;
  },

  // Получение конкретного чата
  async getChat(chatId) {
    const response = await apiClient.get(`/chats/${chatId}`);
    return response.data;
  },

  // Переименование чата
  async updateChatTitle(chatId, title) {
    const response = await apiClient.put(`/chats/${chatId}/title`, { title });
    return response.data;
  },

  // Удаление чата
  async deleteChat(chatId) {
    await apiClient.delete(`/chats/${chatId}`);
  },

  // Отправка сообщения
  async sendMessage(chatId, content) {
    const response = await apiClient.post(`/chats/${chatId}/messages`, { content });
    return response.data;
  },

  // Получение истории сообщений
  async getMessages(chatId) {
    const response = await apiClient.get(`/chats/${chatId}/messages`);
    return response.data;
  },
};
```

### Пример использования в компонентах

```javascript
// components/ChatScreen.js
import React, { useState, useEffect } from 'react';
import { View, Text, TextInput, TouchableOpacity, FlatList, Alert } from 'react-native';
import { chatService } from '../services/chatService';

const ChatScreen = ({ route }) => {
  const { chatId } = route.params;
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadMessages();
  }, []);

  const loadMessages = async () => {
    try {
      const messagesData = await chatService.getMessages(chatId);
      setMessages(messagesData);
    } catch (error) {
      Alert.alert('Ошибка', 'Не удалось загрузить сообщения');
    }
  };

  const sendMessage = async () => {
    if (!newMessage.trim()) return;

    setLoading(true);
    try {
      const response = await chatService.sendMessage(chatId, newMessage);
      
      // Добавляем сообщение пользователя
      setMessages(prev => [...prev, response.user_message]);
      setNewMessage('');
      
      // Обновляем сообщения через несколько секунд для получения ответа AI
      setTimeout(loadMessages, 2000);
    } catch (error) {
      Alert.alert('Ошибка', 'Не удалось отправить сообщение');
    } finally {
      setLoading(false);
    }
  };

  const renderMessage = ({ item }) => (
    <View style={{ 
      alignSelf: item.role === 'user' ? 'flex-end' : 'flex-start',
      backgroundColor: item.role === 'user' ? '#007AFF' : '#F0F0F0',
      padding: 10,
      margin: 5,
      borderRadius: 10,
      maxWidth: '80%'
    }}>
      <Text style={{ 
        color: item.role === 'user' ? 'white' : 'black' 
      }}>
        {item.content}
      </Text>
    </View>
  );

  return (
    <View style={{ flex: 1 }}>
      <FlatList
        data={messages}
        renderItem={renderMessage}
        keyExtractor={(item) => item.id.toString()}
        style={{ flex: 1, padding: 10 }}
      />
      
      <View style={{ 
        flexDirection: 'row', 
        padding: 10, 
        backgroundColor: 'white',
        borderTopWidth: 1,
        borderTopColor: '#E0E0E0'
      }}>
        <TextInput
          value={newMessage}
          onChangeText={setNewMessage}
          placeholder="Введите сообщение..."
          style={{ 
            flex: 1, 
            borderWidth: 1, 
            borderColor: '#E0E0E0', 
            borderRadius: 20, 
            paddingHorizontal: 15,
            paddingVertical: 10
          }}
        />
        <TouchableOpacity
          onPress={sendMessage}
          disabled={loading}
          style={{
            backgroundColor: '#007AFF',
            borderRadius: 20,
            paddingHorizontal: 20,
            paddingVertical: 10,
            marginLeft: 10
          }}
        >
          <Text style={{ color: 'white', fontWeight: 'bold' }}>
            {loading ? '...' : 'Отправить'}
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );
};

export default ChatScreen;
```

```javascript
// components/ChatListScreen.js
import React, { useState, useEffect } from 'react';
import { View, Text, FlatList, TouchableOpacity, Alert } from 'react-native';
import { chatService } from '../services/chatService';

const ChatListScreen = ({ navigation }) => {
  const [chats, setChats] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadChats();
  }, []);

  const loadChats = async () => {
    try {
      const chatsData = await chatService.getChats();
      setChats(chatsData);
    } catch (error) {
      Alert.alert('Ошибка', 'Не удалось загрузить чаты');
    } finally {
      setLoading(false);
    }
  };

  const createNewChat = async () => {
    try {
      const newChat = await chatService.createChat();
      setChats(prev => [newChat, ...prev]);
      navigation.navigate('Chat', { chatId: newChat.id });
    } catch (error) {
      Alert.alert('Ошибка', 'Не удалось создать чат');
    }
  };

  const renderChat = ({ item }) => (
    <TouchableOpacity
      onPress={() => navigation.navigate('Chat', { chatId: item.id })}
      style={{
        padding: 15,
        borderBottomWidth: 1,
        borderBottomColor: '#E0E0E0'
      }}
    >
      <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{item.title}</Text>
      <Text style={{ color: '#666', marginTop: 5 }}>
        Модель: {item.ai_model}
      </Text>
      <Text style={{ color: '#999', fontSize: 12, marginTop: 5 }}>
        Обновлен: {new Date(item.updated_at).toLocaleDateString()}
      </Text>
    </TouchableOpacity>
  );

  if (loading) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
        <Text>Загрузка...</Text>
      </View>
    );
  }

  return (
    <View style={{ flex: 1 }}>
      <FlatList
        data={chats}
        renderItem={renderChat}
        keyExtractor={(item) => item.id.toString()}
      />
      
      <TouchableOpacity
        onPress={createNewChat}
        style={{
          position: 'absolute',
          bottom: 20,
          right: 20,
          backgroundColor: '#007AFF',
          borderRadius: 30,
          width: 60,
          height: 60,
          justifyContent: 'center',
          alignItems: 'center'
        }}
      >
        <Text style={{ color: 'white', fontSize: 24, fontWeight: 'bold' }}>+</Text>
      </TouchableOpacity>
    </View>
  );
};

export default ChatListScreen;
```

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

#### Обновление токена
```http
POST /api/v1/refresh
Content-Type: application/json

{
  "refresh_token": "xyz123abc456..."
}
```

#### Получение профиля
```http
GET /api/v1/profile
Authorization: Bearer <access_token>
```

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

#### Получение списка чатов
```http
GET /api/v1/chats
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

#### Удаление чата
```http
DELETE /api/v1/chats/1
Authorization: Bearer <access_token>
```

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

#### Получение истории сообщений
```http
GET /api/v1/chats/1/messages
Authorization: Bearer <access_token>
```

## 🤖 Поддерживаемые AI провайдеры

### OpenRouter (рекомендуется, используется по умолчанию)
- **Модель по умолчанию:** `deepseek/deepseek-chat`
- **Получение API ключа:** https://openrouter.ai/
- **Переменная окружения:** `OPENROUTER_API_KEY`
- **Преимущества:** Доступ к множеству моделей через единый API, включая DeepSeek

### DeepSeek (прямое подключение)
- **Модель по умолчанию:** `deepseek-chat`
- **Получение API ключа:** https://platform.deepseek.com/
- **Переменная окружения:** `DEEPSEEK_API_KEY`

### Grok (xAI)
- **Модель по умолчанию:** `grok-beta`
- **Получение API ключа:** https://console.x.ai/
- **Переменная окружения:** `GROK_API_KEY`

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
BASE_URL="https://94.103.91.136:8080/api/v1"

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

# 2. Создание чата
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
    "title": "Мой новый чат"
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

# 6. Получение списка чатов
curl -X GET $BASE_URL/chats \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 7. Получение профиля
curl -X GET $BASE_URL/profile \
  -H "Authorization: Bearer $ACCESS_TOKEN"
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
curl https://94.103.91.136:8080/api/v1/profile
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
├── data/              # База данных
├── docker-compose.prod.yml  # Продакшен конфигурация
├── docker-compose.yml # Локальная разработка
├── Dockerfile         # Docker образ
├── env.prod.example   # Пример переменных продакшена
├── nginx.prod.conf    # Конфигурация Nginx
├── go.mod            # Go модули
├── go.sum            # Go зависимости
└── README.md         # Документация
```

### Миграции базы данных

Проект использует SQLite3. Миграции находятся в `cmd/migrate/migrations/`.

```bash
# Применить миграции
go run cmd/migrate/main.go up

# Откатить миграции
go run cmd/migrate/main.go down
```

### Логирование

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

## 🔒 Безопасность

### JWT токены
- **Access токены** - 15 минут жизни
- **Refresh токены** - 7 дней жизни
- **HTTP-only cookies** для refresh токенов
- **Secure флаг** для HTTPS

### Валидация
- Все входные данные валидируются
- Email нормализация (нижний регистр, trim)
- Пароли хешируются с bcrypt
- SQL injection защита через prepared statements

### Ошибки
- Структурированные ответы об ошибках
- Не раскрываем внутренние детали в продакшене
- Логирование всех ошибок для мониторинга

## 📊 Мониторинг и метрики

### Health Check
```bash
curl https://94.103.91.136:8080/health
```

### Логи
```bash
# Просмотр логов в реальном времени
docker-compose -f docker-compose.prod.yml logs -f mindforge-api

# Фильтрация по уровню
docker-compose -f docker-compose.prod.yml logs mindforge-api | grep ERROR
```

### Метрики
- Время ответа API
- Количество запросов
- Использование токенов AI
- Ошибки аутентификации

## 🚀 Масштабирование

### Горизонтальное масштабирование
- Stateless архитектура
- Внешняя база данных (PostgreSQL)
- Load balancer (Nginx)
- Кэширование (Redis)

### Оптимизация
- Connection pooling
- Graceful shutdown
- Timeout настройки
- Rate limiting

## 📞 Поддержка

### Устранение неполадок

**Ошибка 401 Unauthorized:**
- Проверьте токен в заголовке Authorization
- Убедитесь, что токен не истек
- Попробуйте обновить токен через /refresh

**Ошибка 500 Internal Server Error:**
- Проверьте логи сервера
- Убедитесь, что база данных доступна
- Проверьте переменные окружения

**Медленные ответы AI:**
- Проверьте API ключи провайдеров
- Убедитесь в стабильности интернет-соединения
- Проверьте лимиты API провайдеров

### Контакты
- **GitHub Issues:** https://github.com/Nasarwo/GoMindForge/issues
- **Документация:** Этот README.md
- **API Base URL:** https://94.103.91.136:8080/api/v1

## 📄 Лицензия

MIT License - см. файл LICENSE для деталей.