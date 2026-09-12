package main

import (
	"log"
	"net/http"

	config "github.com/lebendig13/metrics/internal/config"
	handler "github.com/lebendig13/metrics/internal/handler"
	models "github.com/lebendig13/metrics/internal/model"
)

func main() {
	configP := config.GetServerConfig()

	memStorage := models.NewMemStorage()
	server := handler.NewServer(memStorage)

	log.Println("Running server on", configP.RunAddress)
	err := http.ListenAndServe(configP.RunAddress, handler.MetricsRouter(server))
	if err != nil {
		log.Fatal("Server has finished with error: ", err)
	}
}
