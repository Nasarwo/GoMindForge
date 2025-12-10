package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"mindforge/internal/ai"
	"mindforge/internal/database"
	"mindforge/internal/env"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

type application struct {
	port              int
	jwtSecret         string
	db                *sql.DB
	models            database.Models
	aiProviderFactory *ai.ProviderFactory
	logger            *slog.Logger
}

func main() {
	// Настраиваем структурированное логирование
	logLevel := env.GetEnvString("LOG_LEVEL", "info")
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if env.GetEnvString("LOG_FORMAT", "json") == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Получаем параметры подключения к PostgreSQL из переменных окружения
	dbHost := env.GetEnvString("DB_HOST", "localhost")
	dbPort := env.GetEnvString("DB_PORT", "5432")
	dbUser := env.GetEnvString("DB_USER", "postgres")
	dbPassword := env.GetEnvString("DB_PASSWORD", "postgres")
	dbName := env.GetEnvString("DB_NAME", "mindforge")
	dbSSLMode := env.GetEnvString("DB_SSLMODE", "disable")

	// Формируем строку подключения к PostgreSQL
	dbDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Проверяем подключение к БД
	if err := db.Ping(); err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Настраиваем пул соединений для PostgreSQL
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("Database connection established",
		"host", dbHost,
		"port", dbPort,
		"database", dbName,
		"max_open_conns", 25,
		"max_idle_conns", 5,
	)

	models := database.NewModels(db)
	aiFactory := ai.NewProviderFactory()

	// JWT_SECRET обязателен для безопасности
	jwtSecret := env.GetEnvString("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Error("JWT_SECRET environment variable is required")
		os.Exit(1)
	}

	app := &application{
		port:              env.GetEnvInt("PORT", 8080),
		jwtSecret:         jwtSecret,
		db:                db,
		models:            models,
		aiProviderFactory: aiFactory,
		logger:            logger,
	}

	if err := app.serve(); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
