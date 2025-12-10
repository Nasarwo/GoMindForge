#!/bin/bash
# Скрипт для получения Access токена GigaChat через OAuth API

# Конфигурация
AUTH_KEY="MDE5YjBhOWEtOTdiNS03MmVlLWI5NGMtYjYyN2EwMjhhNWRkOmFkOTdmNTI0LWFmYWItNDk0YS05YWYxLTI5OTM5OTY3YjEyNw=="
CLIENT_ID="019b0a9a-97b5-72ee-b94c-b627a028a5dd"
SCOPE="GIGACHAT_API_PERS"
OAUTH_URL="https://ngw.devices.sberbank.ru:9443/api/v2/oauth"

# Генерируем уникальный RqUID
if command -v uuidgen &> /dev/null; then
    RQUID=$(uuidgen)
elif command -v uuid &> /dev/null; then
    RQUID=$(uuid)
else
    # Простой UUID v4 генератор (если нет uuidgen/uuid)
    RQUID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "57b0f0b7-3e77-4ba2-b812-80a0ac399cc0")
fi

echo "Получение Access токена GigaChat..."
echo "RqUID: $RQUID"
echo ""

# Получаем токен
RESPONSE=$(curl -s -L -X POST "$OAUTH_URL" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -H 'Accept: application/json' \
  -H "RqUID: $RQUID" \
  -H "Authorization: Basic $AUTH_KEY" \
  --data-urlencode "scope=$SCOPE")

# Проверяем ответ
if [ $? -ne 0 ]; then
    echo "Ошибка: Не удалось выполнить запрос"
    exit 1
fi

# Извлекаем токен
ACCESS_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
EXPIRES_AT=$(echo "$RESPONSE" | grep -o '"expires_at":[0-9]*' | cut -d':' -f2)

if [ -z "$ACCESS_TOKEN" ]; then
    echo "Ошибка: Не удалось получить токен"
    echo "Ответ сервера:"
    echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
    exit 1
fi

echo "✅ Токен успешно получен!"
echo ""
echo "Access Token:"
echo "$ACCESS_TOKEN"
echo ""
if [ -n "$EXPIRES_AT" ]; then
    EXPIRES_DATE=$(date -d "@$EXPIRES_AT" 2>/dev/null || date -r "$EXPIRES_AT" 2>/dev/null || echo "N/A")
    echo "Истекает: $EXPIRES_DATE (timestamp: $EXPIRES_AT)"
    echo ""
fi

echo "Добавьте в .env файл:"
echo "GIGACHAT_ACCESS_TOKEN=$ACCESS_TOKEN"
