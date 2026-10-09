package main

import (
	"context"
	"log"
	"net/http"

	"go.uber.org/zap"

	config "github.com/lebendig13/metrics/internal/config"
	db "github.com/lebendig13/metrics/internal/config/db"
	logger "github.com/lebendig13/metrics/internal/logger"
	models "github.com/lebendig13/metrics/internal/model"
	handler "github.com/lebendig13/metrics/internal/server"
)

func main() {
	if err := logger.Initialize("info"); err != nil {
		log.Fatal("Cannot initialize logger: ", err)
	}
	configP := config.GetServerConfig()

	sqldb, err := db.InitDBConnection(configP.DatabaseDSN)
	if err != nil {
		log.Fatal("Database error: ", err)
	}
	defer sqldb.Close()

	memStorage := models.NewMemStorage()
	server := handler.NewServer(memStorage, &configP, sqldb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler.ProcessBackup(ctx, memStorage, &configP)

	logger.Log.Info("Running server on", zap.String("address", configP.RunAddress))
	err = http.ListenAndServe(configP.RunAddress, handler.MetricsRouter(server))
	if err != nil {
		log.Fatal("Server has finished with error: ", err)
	}
}
