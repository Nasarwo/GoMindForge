package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.port),
		Handler:      app.initializeRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Канал для получения сигналов от ОС
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		app.logger.Info("Starting server", "port", app.port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Ждем сигнала для graceful shutdown
	<-quit
	app.logger.Info("Shutting down server")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := srv.Shutdown(ctx); err != nil {
		app.logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	app.logger.Info("Server exited gracefully")
	return nil
}
