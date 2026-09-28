package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	config "github.com/lebendig13/metrics/internal/config"
	handler "github.com/lebendig13/metrics/internal/handler"
	logger "github.com/lebendig13/metrics/internal/logger"
	models "github.com/lebendig13/metrics/internal/model"
)

func main() {
	if err := logger.Initialize("info"); err != nil {
		log.Fatal("Cannot initialize logger: ", err)
	}
	configP := config.GetServerConfig()

	memStorage := models.NewMemStorage()
	server := handler.NewServer(memStorage, &configP)

	// Для переписывания этой части кода использовала ИИ
	// Создаем контекст, который отменится при получении сигналов от ОС (Ctrl+C, kill)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	handler.ProcessBackup(ctx, memStorage, &configP)

	httpServer := &http.Server{
		Addr:    configP.RunAddress,
		Handler: handler.MetricsRouter(server),
	}

	go func() {
		logger.Log.Info("Running server on", zap.String("address", configP.RunAddress))
		err := httpServer.ListenAndServe()
		if err != nil {
			logger.Log.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Ждем сигнала завершения (код заблокируется тут, пока вы не нажмете Ctrl+C)
	<-ctx.Done()
	logger.Log.Info("Shutting down server gracefully...")

	// Даем серверу 5 секунд на завершение текущих активных сетевых запросов клиентов
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
	}

	// Небольшая пауза, чтобы горутина бэкапа гарантированно успела записать файл на диск
	time.Sleep(100 * time.Millisecond)
	logger.Log.Info("Server successfully stopped")
}
