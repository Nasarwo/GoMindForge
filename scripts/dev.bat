@echo off
chcp 65001 >nul

echo Starting GoMindForge in development mode...

if not exist .env (
    echo ERROR: .env file not found!
    echo Please create .env file with required variables:
    echo    OPENROUTER_API_KEY=your-api-key-here
    echo    JWT_SECRET=your-jwt-secret
    pause
    exit /b 1
)

docker --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker is not installed or not in PATH!
    echo Please install Docker Desktop from: https://www.docker.com/products/docker-desktop
    echo Make sure Docker Desktop is running before executing this script.
    pause
    exit /b 1
)

if not exist data mkdir data
if not exist logs mkdir logs

echo Stopping existing containers...
docker compose down

echo Building and starting containers...
docker compose up --build

echo Server started at http://localhost:8080
pause
