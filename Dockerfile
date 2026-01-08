# Dockerfile для GoMindForge
# Многоэтапная сборка для оптимизации размера образа

# Этап 1: Сборка приложения
FROM golang:alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение с оптимизациями
RUN GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o main ./cmd/api

# Собираем бинарник для миграций
RUN GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o migrate ./cmd/migrate

# Этап 2: Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты для runtime (включая wget для healthcheck)
RUN apk --no-cache add ca-certificates tzdata wget

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарные файлы из этапа сборки
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate

# Копируем миграции
COPY --from=builder /app/cmd/migrate/migrations ./migrations

# Создаем директорию для логов
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порт
EXPOSE 8080

# Устанавливаем переменные окружения по умолчанию
ENV PORT=8080
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json
ENV GIN_MODE=release

# Команда запуска
CMD ["./main"]
