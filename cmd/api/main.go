package main

import (
	"database/sql"
	"log/slog"
	"mindforge/internal/ai"
	"mindforge/internal/database"
	"mindforge/internal/env"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	port              int
	jwtSecret         string
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

	db, err := sql.Open("sqlite3", "./data.db")
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

	// Устанавливаем настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Info("Database connection established",
		"max_open_conns", 25,
		"max_idle_conns", 5,
	)

	models := database.NewModels(db)
	aiFactory := ai.NewProviderFactory()

	app := &application{
		port:              env.GetEnvInt("PORT", 8080),
		jwtSecret:         env.GetEnvString("JWT_SECRET", "some-default-secret"),
		models:            models,
		aiProviderFactory: aiFactory,
		logger:            logger,
	}

	if err := app.serve(); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
