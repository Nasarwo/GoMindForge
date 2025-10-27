#!/bin/bash

# Скрипт для тестирования API GoMindForge
# Использование: ./test.sh

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Тестирование API GoMindForge ===${NC}\n"

# 1. Регистрация
echo -e "${GREEN}1. Регистрация пользователя${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser_'$(date +%s)'",
    "email": "test_'$(date +%s)'@example.com",
    "password": "password123"
  }')

echo "$REGISTER_RESPONSE" | jq '.'
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.refresh_token')

if [ "$ACCESS_TOKEN" != "null" ] && [ "$ACCESS_TOKEN" != "" ]; then
  echo -e "${GREEN}✓ Регистрация успешна${NC}\n"
else
  echo -e "${RED}✗ Ошибка регистрации${NC}\n"
  exit 1
fi

# 2. Получение профиля
echo -e "${GREEN}2. Получение профиля пользователя${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$PROFILE_RESPONSE" | jq '.'
if echo "$PROFILE_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ Профиль получен${NC}\n"
else
  echo -e "${RED}✗ Ошибка получения профиля${NC}\n"
fi

# 3. Обновление токена
echo -e "${GREEN}3. Обновление токена${NC}"
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

echo "$REFRESH_RESPONSE" | jq '.'
NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token')
NEW_REFRESH_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.refresh_token')

if [ "$NEW_ACCESS_TOKEN" != "null" ] && [ "$NEW_ACCESS_TOKEN" != "" ]; then
  echo -e "${GREEN}✓ Токен обновлен${NC}\n"
  ACCESS_TOKEN=$NEW_ACCESS_TOKEN
  REFRESH_TOKEN=$NEW_REFRESH_TOKEN
else
  echo -e "${RED}✗ Ошибка обновления токена${NC}\n"
fi

# 4. Повторное получение профиля с новым токеном
echo -e "${GREEN}4. Получение профиля с новым токеном${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$PROFILE_RESPONSE" | jq '.'
if echo "$PROFILE_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ Профиль получен с новым токеном${NC}\n"
else
  echo -e "${RED}✗ Ошибка получения профиля${NC}\n"
fi

# 5. Выход
echo -e "${GREEN}5. Выход из системы${NC}"
LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/logout" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

echo "$LOGOUT_RESPONSE" | jq '.'
if echo "$LOGOUT_RESPONSE" | jq -e '.message' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ Выход выполнен${NC}\n"
else
  echo -e "${RED}✗ Ошибка выхода${NC}\n"
fi

# 6. Попытка использовать refresh token после выхода
echo -e "${YELLOW}6. Попытка использовать refresh token после выхода${NC}"
INVALID_REFRESH=$(curl -s -X POST "$BASE_URL/api/v1/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

echo "$INVALID_REFRESH" | jq '.'
if echo "$INVALID_REFRESH" | jq -e '.code' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ Токен правильно отклонен${NC}\n"
else
  echo -e "${RED}✗ Токен должен быть отклонен${NC}\n"
fi

echo -e "${YELLOW}=== Тестирование завершено ===${NC}"

