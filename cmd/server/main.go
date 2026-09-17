package main

import (
	"log"
	"net/http"

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
	server := handler.NewServer(memStorage)

	logger.Log.Info("Running server on", zap.String("address", configP.RunAddress))
	err := http.ListenAndServe(configP.RunAddress, handler.MetricsRouter(server))
	if err != nil {
		log.Fatal("Server has finished with error: ", err)
	}
}
