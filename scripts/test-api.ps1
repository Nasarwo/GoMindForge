# Скрипт для тестирования всех API эндпоинтов
$ErrorActionPreference = "Continue"

$BASE_URL = "http://localhost:8080"
$API_URL = "$BASE_URL/api/v1"

Write-Host ""
Write-Host "=========================================="
Write-Host "  TESTING API GoMindForge"
Write-Host "=========================================="
Write-Host "Base URL: $BASE_URL"
Write-Host "API URL: $API_URL"
Write-Host ""

# Test 1: Health Check
Write-Host "[1/12] Health Check..."
try {
    $health = Invoke-RestMethod -Uri "$BASE_URL/health" -Method GET -TimeoutSec 2
    Write-Host "OK: $($health | ConvertTo-Json -Compress)"
} catch {
    Write-Host "FAILED: $_"
    Write-Host "Server is not running! Start it with: go run ./cmd/api"
    exit
}

# Test 2: Register
Write-Host "`n[2/12] Register..."
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$registerBody = @{
    username = "testuser_$timestamp"
    email = "test_$timestamp@example.com"
    password = "testpass123"
} | ConvertTo-Json

try {
    $register = Invoke-RestMethod -Uri "$API_URL/register" -Method POST -Body $registerBody -ContentType "application/json"
    $accessToken = $register.access_token
    $refreshToken = $register.refresh_token
    $userId = $register.user.id
    $userEmail = $register.user.email
    Write-Host "OK: User ID = $userId, Email = $userEmail"
} catch {
    Write-Host "FAILED: $_"
    exit
}

# Test 3: Get Profile
Write-Host "`n[3/12] Get Profile..."
$headers = @{ "Authorization" = "Bearer $accessToken" }
try {
    $profile = Invoke-RestMethod -Uri "$API_URL/profile" -Method GET -Headers $headers
    Write-Host "OK: Profile retrieved - $($profile.email)"
} catch {
    Write-Host "FAILED: $_"
}

# Test 4: Create Chat
Write-Host "`n[4/12] Create Chat..."
$chatBody = @{ ai_model = "deepseek-chat" } | ConvertTo-Json
try {
    $chat = Invoke-RestMethod -Uri "$API_URL/chats" -Method POST -Body $chatBody -ContentType "application/json" -Headers $headers
    $chatId = $chat.id
    Write-Host "OK: Chat ID = $chatId"
} catch {
    Write-Host "FAILED: $_"
    $chatId = $null
}

# Test 5: Get Chats List
Write-Host "`n[5/12] Get Chats List..."
try {
    $chats = Invoke-RestMethod -Uri "$API_URL/chats" -Method GET -Headers $headers
    $count = if ($chats -is [Array]) { $chats.Count } elseif ($chats.chats) { $chats.chats.Count } else { 0 }
    Write-Host "OK: Total chats = $count"
} catch {
    Write-Host "FAILED: $_"
}

# Test 6: Get Chat
if ($chatId) {
    Write-Host "`n[6/12] Get Chat..."
    try {
        $chat = Invoke-RestMethod -Uri "$API_URL/chats/$chatId" -Method GET -Headers $headers
        Write-Host "OK: Chat retrieved"
    } catch {
        Write-Host "FAILED: $_"
    }
} else {
    Write-Host "`n[6/12] Get Chat... SKIPPED (no chat ID)"
}

# Test 7: Update Chat Title
if ($chatId) {
    Write-Host "`n[7/12] Update Chat Title..."
    $titleBody = @{ title = "Test Chat $(Get-Date -Format 'HH:mm:ss')" } | ConvertTo-Json
    try {
        Invoke-RestMethod -Uri "$API_URL/chats/$chatId/title" -Method PUT -Body $titleBody -ContentType "application/json" -Headers $headers | Out-Null
        Write-Host "OK: Title updated"
    } catch {
        Write-Host "FAILED: $_"
    }
} else {
    Write-Host "`n[7/12] Update Chat Title... SKIPPED (no chat ID)"
}

# Test 8: Create Message
if ($chatId) {
    Write-Host "`n[8/12] Create Message..."
    $messageBody = @{ content = "Hello! Tell me about Go programming." } | ConvertTo-Json
    try {
        $message = Invoke-RestMethod -Uri "$API_URL/chats/$chatId/messages" -Method POST -Body $messageBody -ContentType "application/json" -Headers $headers
        $msgId = if ($message.user_message) { $message.user_message.id } else { $message.id }
        Write-Host "OK: Message sent, ID = $msgId, Status = $($message.status)"
        if ($message.ai_response) {
            $preview = $message.ai_response.Substring(0, [Math]::Min(100, $message.ai_response.Length))
            Write-Host "  AI Response preview: $preview..."
        }
    } catch {
        Write-Host "FAILED: $_"
    }
} else {
    Write-Host "`n[8/12] Create Message... SKIPPED (no chat ID)"
}

# Test 9: Get Messages
if ($chatId) {
    Write-Host "`n[9/12] Get Messages..."
    Start-Sleep -Seconds 3
    try {
        $messages = Invoke-RestMethod -Uri "$API_URL/chats/$chatId/messages" -Method GET -Headers $headers
        $msgCount = if ($messages -is [Array]) { $messages.Count } elseif ($messages.messages) { $messages.messages.Count } else { 0 }
        Write-Host "OK: Total messages = $msgCount"
    } catch {
        Write-Host "FAILED: $_"
    }
} else {
    Write-Host "`n[9/12] Get Messages... SKIPPED (no chat ID)"
}

# Test 10: Refresh Token
Write-Host "`n[10/12] Refresh Token..."
$refreshBody = @{ refresh_token = $refreshToken } | ConvertTo-Json
try {
    $refresh = Invoke-RestMethod -Uri "$API_URL/refresh" -Method POST -Body $refreshBody -ContentType "application/json"
    $accessToken = $refresh.access_token
    $refreshToken = $refresh.refresh_token
    $headers = @{ "Authorization" = "Bearer $accessToken" }
    Write-Host "OK: Token refreshed"
} catch {
    Write-Host "FAILED: $_"
}

# Test 11: Login
Write-Host "`n[11/12] Login..."
$loginBody = @{ email = $userEmail; password = "testpass123" } | ConvertTo-Json
try {
    $login = Invoke-RestMethod -Uri "$API_URL/login" -Method POST -Body $loginBody -ContentType "application/json"
    Write-Host "OK: Login successful"
} catch {
    Write-Host "FAILED: $_"
}

# Test 12: Logout
Write-Host "`n[12/12] Logout..."
try {
    Invoke-RestMethod -Uri "$API_URL/logout" -Method POST -Headers $headers | Out-Null
    Write-Host "OK: Logout successful"
} catch {
    Write-Host "FAILED: $_"
}

Write-Host "`n=========================================="
Write-Host "  TESTING COMPLETED"
Write-Host "=========================================="
