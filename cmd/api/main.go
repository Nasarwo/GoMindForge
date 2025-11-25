package main

import (
	"database/sql"
	"log/slog"
	"mindforge/internal/ai"
	"mindforge/internal/database"
	"mindforge/internal/env"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
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

	// Получаем путь к базе данных из переменной окружения
	dbPath := env.GetEnvString("DB_PATH", "./data/data.db")

	// Создаем директорию для базы данных, если её нет
	dbDir := dbPath
	if lastSlash := strings.LastIndex(dbPath, "/"); lastSlash != -1 {
		dbDir = dbPath[:lastSlash]
	} else if lastBackslash := strings.LastIndex(dbPath, "\\"); lastBackslash != -1 {
		dbDir = dbPath[:lastBackslash]
	}

	if dbDir != dbPath {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			logger.Error("Failed to create database directory", "error", err, "path", dbDir)
			os.Exit(1)
		}
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1&_journal_mode=WAL")
	if err != nil {
		logger.Error("Failed to open database", "error", err, "path", dbPath)
		os.Exit(1)
	}
	defer db.Close()

	// Проверяем подключение к БД
	if err := db.Ping(); err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// SQLite не поддерживает множественные соединения так же, как другие БД
	// Но мы все равно устанавливаем разумные лимиты
	db.SetMaxOpenConns(1) // SQLite работает лучше с одним соединением
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Не ограничиваем время жизни соединения для SQLite

	logger.Info("Database connection established",
		"max_open_conns", 1,
		"max_idle_conns", 1,
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
