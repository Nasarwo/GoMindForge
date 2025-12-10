@echo off
REM Скрипт для применения миграций

echo Применение миграций PostgreSQL...
echo.

REM Установка переменных окружения по умолчанию
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=5432
if "%DB_USER%"=="" set DB_USER=postgres
if "%DB_PASSWORD%"=="" set DB_PASSWORD=postgres
if "%DB_NAME%"=="" set DB_NAME=mindforge
if "%DB_SSLMODE%"=="" set DB_SSLMODE=disable

echo Параметры подключения:
echo   DB_HOST=%DB_HOST%
echo   DB_PORT=%DB_PORT%
echo   DB_USER=%DB_USER%
echo   DB_NAME=%DB_NAME%
echo.

go run cmd/migrate/main.go %1

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ОШИБКА: Не удалось применить миграции
    echo.
    echo Убедитесь, что:
    echo   1. PostgreSQL запущен (docker-compose up -d postgres)
    echo   2. Переменные окружения настроены правильно
    echo   3. База данных создана
    echo.
    pause
    exit /b 1
)

echo.
echo Миграции применены успешно!
pause
