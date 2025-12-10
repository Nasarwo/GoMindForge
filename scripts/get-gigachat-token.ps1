# Скрипт для получения Access токена GigaChat через OAuth API (PowerShell)

# Конфигурация
$AUTH_KEY = "MDE5YjBhOWEtOTdiNS03MmVlLWI5NGMtYjYyN2EwMjhhNWRkOmFkOTdmNTI0LWFmYWItNDk0YS05YWYxLTI5OTM5OTY3YjEyNw=="
$CLIENT_ID = "019b0a9a-97b5-72ee-b94c-b627a028a5dd"
$SCOPE = "GIGACHAT_API_PERS"
$OAUTH_URL = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"

# Генерируем уникальный RqUID (UUID)
$RQUID = [guid]::NewGuid().ToString()

Write-Host "Получение Access токена GigaChat..." -ForegroundColor Cyan
Write-Host "RqUID: $RQUID"
Write-Host ""

# Подготавливаем заголовки
$headers = @{
    'Content-Type' = 'application/x-www-form-urlencoded'
    'Accept' = 'application/json'
    'RqUID' = $RQUID
    'Authorization' = "Basic $AUTH_KEY"
}

# Подготавливаем тело запроса
$body = @{
    scope = $SCOPE
}

try {
    # Получаем токен
    $response = Invoke-RestMethod -Uri $OAUTH_URL -Method Post -Headers $headers -Body $body -ErrorAction Stop

    if ($response.access_token) {
        Write-Host "✅ Токен успешно получен!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Access Token:" -ForegroundColor Yellow
        Write-Host $response.access_token
        Write-Host ""

        if ($response.expires_at) {
            $expiresDate = [DateTimeOffset]::FromUnixTimeSeconds($response.expires_at).LocalDateTime
            Write-Host "Истекает: $expiresDate (timestamp: $($response.expires_at))"
            Write-Host ""
        }

        Write-Host "Добавьте в .env файл:" -ForegroundColor Cyan
        Write-Host "GIGACHAT_ACCESS_TOKEN=$($response.access_token)"

        # Копируем токен в буфер обмена (опционально)
        $response.access_token | Set-Clipboard
        Write-Host ""
        Write-Host "Токен скопирован в буфер обмена!" -ForegroundColor Green
    } else {
        Write-Host "Ошибка: Токен не найден в ответе" -ForegroundColor Red
        Write-Host "Ответ сервера:" -ForegroundColor Yellow
        $response | ConvertTo-Json -Depth 10
        exit 1
    }
} catch {
    Write-Host "Ошибка: Не удалось получить токен" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "Детали ошибки:" -ForegroundColor Yellow
        Write-Host $_.ErrorDetails.Message
    }
    exit 1
}
