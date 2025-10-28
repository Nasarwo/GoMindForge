# Документация

## 📋 Содержание
- [Обзор API](#обзор-api)
- [Базовые настройки](#базовые-настройки)
- [Аутентификация](#аутентификация)
- [Управление чатами](#управление-чатами)
- [Работа с сообщениями](#работа-с-сообщениями)
- [Обработка ошибок](#обработка-ошибок)
- [Примеры интеграции](#примеры-интеграции)
- [Типы данных](#типы-данных)

## 🔗 Обзор API

**Базовый URL:** `http://localhost:8080/api/v1` (для разработки)  
**Протокол:** HTTP/HTTPS  
**Формат данных:** JSON  
**Аутентификация:** Bearer Token (JWT)

### Основные endpoints:
- `POST /register` - Регистрация пользователя
- `POST /login` - Вход в систему
- `POST /refresh` - Обновление токена
- `POST /logout` - Выход из системы
- `GET /profile` - Получение профиля
- `POST /chats` - Создание чата
- `GET /chats` - Получение списка чатов
- `GET /chats/:id` - Получение конкретного чата
- `PUT /chats/:id/title` - Переименование чата
- `DELETE /chats/:id` - Удаление чата
- `POST /chats/:id/messages` - Отправка сообщения
- `GET /chats/:id/messages` - Получение истории сообщений

## ⚙️ Базовые настройки

### Конфигурация для React Native

```typescript
// config/api.ts
export const API_CONFIG = {
  BASE_URL: __DEV__ 
    ? 'http://localhost:8080/api/v1'  // Для разработки
    : 'https://your-production-domain.com/api/v1',  // Для продакшена
  TIMEOUT: 30000, // 30 секунд
  RETRY_ATTEMPTS: 3,
};

// Для Android эмулятора используйте:
// BASE_URL: 'http://10.0.2.2:8080/api/v1'

// Для физического устройства используйте IP вашего компьютера:
// BASE_URL: 'http://192.168.1.100:8080/api/v1'
```

### Настройка HTTP клиента

```typescript
// services/apiClient.ts
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { API_CONFIG } from '../config/api';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_CONFIG.BASE_URL,
      timeout: API_CONFIG.TIMEOUT,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Интерцептор для добавления токена
    this.client.interceptors.request.use(
      async (config) => {
        const token = await AsyncStorage.getItem('access_token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Интерцептор для обработки ошибок аутентификации
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          // Попытка обновить токен
          const refreshed = await this.refreshToken();
          if (refreshed) {
            // Повторяем оригинальный запрос
            return this.client.request(error.config);
          } else {
            // Перенаправляем на экран входа
            await this.logout();
          }
        }
        return Promise.reject(error);
      }
    );
  }

  private async refreshToken(): Promise<boolean> {
    try {
      const refreshToken = await AsyncStorage.getItem('refresh_token');
      if (!refreshToken) return false;

      const response = await axios.post(`${API_CONFIG.BASE_URL}/refresh`, {
        refresh_token: refreshToken,
      });

      await AsyncStorage.setItem('access_token', response.data.access_token);
      await AsyncStorage.setItem('refresh_token', response.data.refresh_token);
      return true;
    } catch (error) {
      return false;
    }
  }

  private async logout(): Promise<void> {
    await AsyncStorage.multiRemove(['access_token', 'refresh_token', 'user_data']);
    // Здесь можно добавить навигацию к экрану входа
  }

  // Публичные методы для API вызовов
  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get(url, config);
    return response.data;
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post(url, data, config);
    return response.data;
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete(url, config);
    return response.data;
  }
}

export const apiClient = new ApiClient();
```

## 🔐 Аутентификация

### Регистрация пользователя

```typescript
// types/auth.ts
export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface User {
  id: number;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
}
```

```typescript
// services/authService.ts
import { apiClient } from './apiClient';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { RegisterRequest, LoginRequest, AuthResponse, User } from '../types/auth';

export class AuthService {
  async register(data: RegisterRequest): Promise<AuthResponse> {
    try {
      const response = await apiClient.post<AuthResponse>('/register', data);
      
      // Сохраняем токены и данные пользователя
      await AsyncStorage.setItem('access_token', response.access_token);
      await AsyncStorage.setItem('refresh_token', response.refresh_token);
      await AsyncStorage.setItem('user_data', JSON.stringify(response.user));
      
      return response;
    } catch (error) {
      throw this.handleAuthError(error);
    }
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    try {
      const response = await apiClient.post<AuthResponse>('/login', data);
      
      // Сохраняем токены и данные пользователя
      await AsyncStorage.setItem('access_token', response.access_token);
      await AsyncStorage.setItem('refresh_token', response.refresh_token);
      await AsyncStorage.setItem('user_data', JSON.stringify(response.user));
      
      return response;
    } catch (error) {
      throw this.handleAuthError(error);
    }
  }

  async logout(): Promise<void> {
    try {
      await apiClient.post('/logout');
    } catch (error) {
      console.warn('Logout request failed:', error);
    } finally {
      // Очищаем локальные данные в любом случае
      await AsyncStorage.multiRemove(['access_token', 'refresh_token', 'user_data']);
    }
  }

  async getProfile(): Promise<User> {
    try {
      return await apiClient.get<User>('/profile');
    } catch (error) {
      throw this.handleAuthError(error);
    }
  }

  async isAuthenticated(): Promise<boolean> {
    const token = await AsyncStorage.getItem('access_token');
    return !!token;
  }

  async getCurrentUser(): Promise<User | null> {
    try {
      const userData = await AsyncStorage.getItem('user_data');
      return userData ? JSON.parse(userData) : null;
    } catch (error) {
      return null;
    }
  }

  private handleAuthError(error: any): Error {
    if (error.response?.data?.message) {
      return new Error(error.response.data.message);
    }
    return new Error('Ошибка аутентификации');
  }
}

export const authService = new AuthService();
```

### Использование в компонентах

```typescript
// components/LoginScreen.tsx
import React, { useState } from 'react';
import { View, Text, TextInput, TouchableOpacity, Alert } from 'react-native';
import { authService } from '../services/authService';

export const LoginScreen: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  const handleLogin = async () => {
    if (!email || !password) {
      Alert.alert('Ошибка', 'Заполните все поля');
      return;
    }

    setLoading(true);
    try {
      await authService.login({ email, password });
      // Навигация к главному экрану
    } catch (error) {
      Alert.alert('Ошибка входа', error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={{ padding: 20 }}>
      <TextInput
        placeholder="Email"
        value={email}
        onChangeText={setEmail}
        keyboardType="email-address"
        autoCapitalize="none"
      />
      <TextInput
        placeholder="Пароль"
        value={password}
        onChangeText={setPassword}
        secureTextEntry
      />
      <TouchableOpacity onPress={handleLogin} disabled={loading}>
        <Text>{loading ? 'Вход...' : 'Войти'}</Text>
      </TouchableOpacity>
    </View>
  );
};
```

## 💬 Управление чатами

### Типы данных для чатов

```typescript
// types/chat.ts
export interface Chat {
  id: number;
  user_id: number;
  ai_model: string;
  title: string;
  created_at: string;
  updated_at: string;
}

export interface CreateChatRequest {
  ai_model: string;
}

export interface UpdateChatTitleRequest {
  title: string;
}
```

### Сервис для работы с чатами

```typescript
// services/chatService.ts
import { apiClient } from './apiClient';
import { Chat, CreateChatRequest, UpdateChatTitleRequest } from '../types/chat';

export class ChatService {
  async createChat(data: CreateChatRequest): Promise<Chat> {
    try {
      return await apiClient.post<Chat>('/chats', data);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async getChats(): Promise<Chat[]> {
    try {
      return await apiClient.get<Chat[]>('/chats');
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async getChat(id: number): Promise<Chat> {
    try {
      return await apiClient.get<Chat>(`/chats/${id}`);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async updateChatTitle(id: number, data: UpdateChatTitleRequest): Promise<void> {
    try {
      await apiClient.put(`/chats/${id}/title`, data);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async deleteChat(id: number): Promise<void> {
    try {
      await apiClient.delete(`/chats/${id}`);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  private handleError(error: any): Error {
    if (error.response?.data?.message) {
      return new Error(error.response.data.message);
    }
    return new Error('Ошибка при работе с чатами');
  }
}

export const chatService = new ChatService();
```

### Компонент списка чатов

```typescript
// components/ChatListScreen.tsx
import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  Alert,
  TextInput,
  Modal,
} from 'react-native';
import { chatService } from '../services/chatService';
import { Chat } from '../types/chat';

export const ChatListScreen: React.FC = () => {
  const [chats, setChats] = useState<Chat[]>([]);
  const [loading, setLoading] = useState(true);
  const [showRenameModal, setShowRenameModal] = useState(false);
  const [selectedChat, setSelectedChat] = useState<Chat | null>(null);
  const [newTitle, setNewTitle] = useState('');

  useEffect(() => {
    loadChats();
  }, []);

  const loadChats = async () => {
    try {
      const chatList = await chatService.getChats();
      setChats(chatList);
    } catch (error) {
      Alert.alert('Ошибка', 'Не удалось загрузить чаты');
    } finally {
      setLoading(false);
    }
  };

  const createNewChat = async () => {
    try {
      const newChat = await chatService.createChat({
        ai_model: 'deepseek-chat',
      });
      setChats(prev => [newChat, ...prev]);
      // Навигация к новому чату
    } catch (error) {
      Alert.alert('Ошибка', 'Не удалось создать чат');
    }
  };

  const renameChat = async () => {
    if (!selectedChat || !newTitle.trim()) return;

    try {
      await chatService.updateChatTitle(selectedChat.id, {
        title: newTitle.trim(),
      });
      
      setChats(prev =>
        prev.map(chat =>
          chat.id === selectedChat.id
            ? { ...chat, title: newTitle.trim() }
            : chat
        )
      );
      
      setShowRenameModal(false);
      setNewTitle('');
      setSelectedChat(null);
    } catch (error) {
      Alert.alert('Ошибка', 'Не удалось переименовать чат');
    }
  };

  const deleteChat = async (chat: Chat) => {
    Alert.alert(
      'Удаление чата',
      `Вы уверены, что хотите удалить чат "${chat.title}"?`,
      [
        { text: 'Отмена', style: 'cancel' },
        {
          text: 'Удалить',
          style: 'destructive',
          onPress: async () => {
            try {
              await chatService.deleteChat(chat.id);
              setChats(prev => prev.filter(c => c.id !== chat.id));
            } catch (error) {
              Alert.alert('Ошибка', 'Не удалось удалить чат');
            }
          },
        },
      ]
    );
  };

  const renderChatItem = ({ item }: { item: Chat }) => (
    <TouchableOpacity
      style={{ padding: 15, borderBottomWidth: 1 }}
      onPress={() => {
        // Навигация к чату
      }}
      onLongPress={() => {
        setSelectedChat(item);
        setNewTitle(item.title);
        setShowRenameModal(true);
      }}
    >
      <Text style={{ fontSize: 16, fontWeight: 'bold' }}>{item.title}</Text>
      <Text style={{ fontSize: 12, color: 'gray' }}>
        {new Date(item.updated_at).toLocaleDateString()}
      </Text>
    </TouchableOpacity>
  );

  if (loading) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
        <Text>Загрузка чатов...</Text>
      </View>
    );
  }

  return (
    <View style={{ flex: 1 }}>
      <TouchableOpacity
        style={{ padding: 15, backgroundColor: '#007AFF' }}
        onPress={createNewChat}
      >
        <Text style={{ color: 'white', textAlign: 'center' }}>
          Создать новый чат
        </Text>
      </TouchableOpacity>

      <FlatList
        data={chats}
        renderItem={renderChatItem}
        keyExtractor={(item) => item.id.toString()}
      />

      <Modal visible={showRenameModal} transparent animationType="slide">
        <View style={{ flex: 1, justifyContent: 'center', backgroundColor: 'rgba(0,0,0,0.5)' }}>
          <View style={{ backgroundColor: 'white', margin: 20, padding: 20 }}>
            <Text style={{ fontSize: 18, marginBottom: 15 }}>Переименовать чат</Text>
            <TextInput
              value={newTitle}
              onChangeText={setNewTitle}
              placeholder="Название чата"
              style={{ borderWidth: 1, padding: 10, marginBottom: 15 }}
            />
            <View style={{ flexDirection: 'row', justifyContent: 'space-between' }}>
              <TouchableOpacity
                onPress={() => setShowRenameModal(false)}
                style={{ padding: 10 }}
              >
                <Text>Отмена</Text>
              </TouchableOpacity>
              <TouchableOpacity
                onPress={renameChat}
                style={{ padding: 10, backgroundColor: '#007AFF' }}
              >
                <Text style={{ color: 'white' }}>Сохранить</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </View>
  );
};
```

## 📝 Работа с сообщениями

### Типы данных для сообщений

```typescript
// types/message.ts
export interface Message {
  id: number;
  chat_id: number;
  role: 'user' | 'assistant';
  content: string;
  created_at: string;
}

export interface CreateMessageRequest {
  content: string;
}

export interface CreateMessageResponse {
  user_message: Message;
  status: string;
  message: string;
}
```

### Сервис для работы с сообщениями

```typescript
// services/messageService.ts
import { apiClient } from './apiClient';
import { Message, CreateMessageRequest, CreateMessageResponse } from '../types/message';

export class MessageService {
  async sendMessage(chatId: number, data: CreateMessageRequest): Promise<CreateMessageResponse> {
    try {
      return await apiClient.post<CreateMessageResponse>(`/chats/${chatId}/messages`, data);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  async getMessages(chatId: number): Promise<Message[]> {
    try {
      return await apiClient.get<Message[]>(`/chats/${chatId}/messages`);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  private handleError(error: any): Error {
    if (error.response?.data?.message) {
      return new Error(error.response.data.message);
    }
    return new Error('Ошибка при работе с сообщениями');
  }
}

export const messageService = new MessageService();
```

### Компонент чата

```typescript
// components/ChatScreen.tsx
import React, { useState, useEffect, useRef } from 'react';
import {
  View,
  Text,
  FlatList,
  TextInput,
  TouchableOpacity,
  Alert,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { messageService } from '../services/messageService';
import { Message } from '../types/message';

interface ChatScreenProps {
  chatId: number;
}

export const ChatScreen: React.FC<ChatScreenProps> = ({ chatId }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputText, setInputText] = useState('');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const flatListRef = useRef<FlatList>(null);

  useEffect(() => {
    loadMessages();
  }, [chatId]);

  const loadMessages = async () => {
    try {
      const messageList = await messageService.getMessages(chatId);
      setMessages(messageList);
    } catch (error) {
      Alert.alert('Ошибка', 'Не удалось загрузить сообщения');
    } finally {
      setLoading(false);
    }
  };

  const sendMessage = async () => {
    if (!inputText.trim() || sending) return;

    const messageText = inputText.trim();
    setInputText('');
    setSending(true);

    try {
      // Добавляем сообщение пользователя сразу в UI
      const userMessage: Message = {
        id: Date.now(), // Временный ID
        chat_id: chatId,
        role: 'user',
        content: messageText,
        created_at: new Date().toISOString(),
      };

      setMessages(prev => [...prev, userMessage]);

      // Отправляем сообщение на сервер
      const response = await messageService.sendMessage(chatId, {
        content: messageText,
      });

      // Обновляем сообщение пользователя с реальным ID
      setMessages(prev =>
        prev.map(msg =>
          msg.id === userMessage.id ? response.user_message : msg
        )
      );

      // Ждем ответ AI и обновляем сообщения
      setTimeout(async () => {
        try {
          const updatedMessages = await messageService.getMessages(chatId);
          setMessages(updatedMessages);
        } catch (error) {
          console.error('Failed to load AI response:', error);
        }
      }, 2000);

    } catch (error) {
      Alert.alert('Ошибка', 'Не удалось отправить сообщение');
      // Удаляем сообщение пользователя из UI при ошибке
      setMessages(prev => prev.filter(msg => msg.id !== Date.now()));
    } finally {
      setSending(false);
    }
  };

  const renderMessage = ({ item }: { item: Message }) => (
    <View
      style={{
        padding: 10,
        marginVertical: 5,
        alignSelf: item.role === 'user' ? 'flex-end' : 'flex-start',
        backgroundColor: item.role === 'user' ? '#007AFF' : '#F0F0F0',
        borderRadius: 10,
        maxWidth: '80%',
      }}
    >
      <Text
        style={{
          color: item.role === 'user' ? 'white' : 'black',
        }}
      >
        {item.content}
      </Text>
      <Text
        style={{
          fontSize: 10,
          color: item.role === 'user' ? 'rgba(255,255,255,0.7)' : 'gray',
          marginTop: 5,
        }}
      >
        {new Date(item.created_at).toLocaleTimeString()}
      </Text>
    </View>
  );

  if (loading) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
        <Text>Загрузка сообщений...</Text>
      </View>
    );
  }

  return (
    <KeyboardAvoidingView
      style={{ flex: 1 }}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <FlatList
        ref={flatListRef}
        data={messages}
        renderItem={renderMessage}
        keyExtractor={(item) => item.id.toString()}
        style={{ flex: 1, padding: 10 }}
        onContentSizeChange={() => flatListRef.current?.scrollToEnd()}
      />

      <View style={{ flexDirection: 'row', padding: 10, borderTopWidth: 1 }}>
        <TextInput
          value={inputText}
          onChangeText={setInputText}
          placeholder="Введите сообщение..."
          style={{
            flex: 1,
            borderWidth: 1,
            borderRadius: 20,
            paddingHorizontal: 15,
            paddingVertical: 10,
            marginRight: 10,
          }}
          multiline
        />
        <TouchableOpacity
          onPress={sendMessage}
          disabled={!inputText.trim() || sending}
          style={{
            backgroundColor: inputText.trim() && !sending ? '#007AFF' : '#CCC',
            borderRadius: 20,
            paddingHorizontal: 20,
            paddingVertical: 10,
            justifyContent: 'center',
          }}
        >
          <Text style={{ color: 'white' }}>
            {sending ? '...' : '→'}
          </Text>
        </TouchableOpacity>
      </View>
    </KeyboardAvoidingView>
  );
};
```

## ⚠️ Обработка ошибок

### Типы ошибок API

```typescript
// types/error.ts
export interface APIError {
  status: number;
  message: string;
  code: string;
}

export class APIErrorHandler {
  static handle(error: any): string {
    if (error.response?.data) {
      const apiError = error.response.data as APIError;
      
      switch (apiError.code) {
        case 'VALIDATION_ERROR':
          return 'Ошибка валидации данных';
        case 'USER_ALREADY_EXISTS':
          return 'Пользователь с таким email уже существует';
        case 'INVALID_CREDENTIALS':
          return 'Неверный email или пароль';
        case 'UNAUTHORIZED':
          return 'Необходима авторизация';
        case 'FORBIDDEN':
          return 'Доступ запрещен';
        case 'CHAT_NOT_FOUND':
          return 'Чат не найден';
        case 'MESSAGE_NOT_FOUND':
          return 'Сообщение не найдено';
        default:
          return apiError.message || 'Произошла ошибка';
      }
    }
    
    if (error.message) {
      return error.message;
    }
    
    return 'Произошла неизвестная ошибка';
  }
}
```

### Хук для обработки ошибок

```typescript
// hooks/useErrorHandler.ts
import { useState } from 'react';
import { Alert } from 'react-native';
import { APIErrorHandler } from '../types/error';

export const useErrorHandler = () => {
  const [loading, setLoading] = useState(false);

  const handleAsync = async <T>(
    asyncFunction: () => Promise<T>,
    errorMessage?: string
  ): Promise<T | null> => {
    setLoading(true);
    try {
      const result = await asyncFunction();
      return result;
    } catch (error) {
      const message = errorMessage || APIErrorHandler.handle(error);
      Alert.alert('Ошибка', message);
      return null;
    } finally {
      setLoading(false);
    }
  };

  return { handleAsync, loading };
};
```

## 🔧 Конфигурация и переменные окружения

### Настройка для разных окружений

```typescript
// config/environment.ts
export interface Environment {
  API_BASE_URL: string;
  API_TIMEOUT: number;
  LOG_LEVEL: 'debug' | 'info' | 'warn' | 'error';
}

const development: Environment = {
  API_BASE_URL: 'http://localhost:8080/api/v1',
  API_TIMEOUT: 30000,
  LOG_LEVEL: 'debug',
};

const production: Environment = {
  API_BASE_URL: 'https://your-api-domain.com/api/v1',
  API_TIMEOUT: 10000,
  LOG_LEVEL: 'error',
};

export const environment: Environment = __DEV__ ? development : production;
```

### Настройка для Android эмулятора

```typescript
// Для Android эмулятора используйте 10.0.2.2 вместо localhost
const androidEmulator: Environment = {
  API_BASE_URL: 'http://10.0.2.2:8080/api/v1',
  API_TIMEOUT: 30000,
  LOG_LEVEL: 'debug',
};
```

### Настройка для физического устройства

```typescript
// Для физического устройства используйте IP вашего компьютера
const physicalDevice: Environment = {
  API_BASE_URL: 'http://192.168.1.100:8080/api/v1', // Замените на ваш IP
  API_TIMEOUT: 30000,
  LOG_LEVEL: 'debug',
};
```

## 📱 Полный пример интеграции

### Главный компонент приложения

```typescript
// App.tsx
import React, { useEffect, useState } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import { authService } from './services/authService';
import { LoginScreen } from './components/LoginScreen';
import { ChatListScreen } from './components/ChatListScreen';
import { ChatScreen } from './components/ChatScreen';

const Stack = createStackNavigator();

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const authenticated = await authService.isAuthenticated();
      setIsAuthenticated(authenticated);
    } catch (error) {
      setIsAuthenticated(false);
    }
  };

  if (isAuthenticated === null) {
    // Показываем загрузочный экран
    return null;
  }

  return (
    <NavigationContainer>
      <Stack.Navigator>
        {!isAuthenticated ? (
          <Stack.Screen name="Login" component={LoginScreen} />
        ) : (
          <>
            <Stack.Screen name="ChatList" component={ChatListScreen} />
            <Stack.Screen name="Chat" component={ChatScreen} />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}
```

## 📋 Типы данных

### Полный список типов

```typescript
// types/index.ts

// Аутентификация
export interface User {
  id: number;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

// Чаты
export interface Chat {
  id: number;
  user_id: number;
  ai_model: string;
  title: string;
  created_at: string;
  updated_at: string;
}

// Сообщения
export interface Message {
  id: number;
  chat_id: number;
  role: 'user' | 'assistant';
  content: string;
  created_at: string;
}

// Запросы
export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface CreateChatRequest {
  ai_model: string;
}

export interface UpdateChatTitleRequest {
  title: string;
}

export interface CreateMessageRequest {
  content: string;
}

// Ответы
export interface CreateMessageResponse {
  user_message: Message;
  status: string;
  message: string;
}

// Ошибки
export interface APIError {
  status: number;
  message: string;
  code: string;
}
```

## 🚀 Готовые команды для тестирования

### Тестирование API с помощью cURL

```bash
# 1. Регистрация
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'

# 2. Создание чата
curl -X POST http://localhost:8080/api/v1/chats \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ai_model": "deepseek-chat"
  }'

# 3. Отправка сообщения
curl -X POST http://localhost:8080/api/v1/chats/1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Привет! Как дела?"
  }'
```

**Полезные ссылки:**
- [React Native документация](https://reactnative.dev/)
- [Axios документация](https://axios-http.com/)
- [AsyncStorage документация](https://react-native-async-storage.github.io/async-storage/)

---

*Документация обновлена: 28 октября 2025*
