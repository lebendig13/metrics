package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"go.uber.org/zap"

	config "github.com/lebendig13/metrics/internal/config"
	logger "github.com/lebendig13/metrics/internal/logger"
	dbmigration "github.com/lebendig13/metrics/internal/migrate"
	repository "github.com/lebendig13/metrics/internal/repository"
	handler "github.com/lebendig13/metrics/internal/server"
)

func main() {
	if err := logger.Initialize("info"); err != nil {
		log.Fatal("Cannot initialize logger: ", err)
	}
	configP := config.GetServerConfig()
	if configP.DatabaseDSN != "" && strings.HasPrefix(configP.DatabaseDSN, "postgres://") {
		if err := dbmigration.Up(configP.DatabaseDSN); err != nil {
			log.Fatal("Failed to migrate database: ", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	storage, err := repository.NewStorage(ctx, &configP)
	if err != nil {
		log.Fatal("Cannot initialize storage: ", err)
	}
	defer func() {
		if err := storage.Close(ctx); err != nil {
			logger.Log.Error("Failed to close storage", zap.Error(err))
		} else {
			logger.Log.Info("Storage closed successfully")
		}
	}()

	server := handler.NewServer(storage, &configP)

	logger.Log.Info("Running server on", zap.String("address", configP.RunAddress))
	err = http.ListenAndServe(configP.RunAddress, handler.MetricsRouter(server))
	if err != nil {
		log.Fatal("Server has finished with error: ", err)
	}
}
