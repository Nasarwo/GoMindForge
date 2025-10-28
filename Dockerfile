# Dockerfile для GoMindForge
# Многоэтапная сборка для оптимизации размера образа

# Этап 1: Сборка приложения
FROM golang:1.25.3-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Этап 2: Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты для runtime
RUN apk --no-cache add ca-certificates tzdata sqlite

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/main .

# Копируем миграции
COPY --from=builder /app/cmd/migrate/migrations ./migrations

# Создаем директорию для базы данных
RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app

# Создаем символическую ссылку для базы данных
RUN ln -sf /app/data/data.db /app/data.db

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
