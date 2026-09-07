package main

import (
	"log"
	"net/http"

	config "github.com/lebendig13/metrics/internal/config"
	handler "github.com/lebendig13/metrics/internal/handler"
	models "github.com/lebendig13/metrics/internal/model"
)

func main() {
	configFlags := config.ParseServerFlags()

	memStorage := models.NewMemStorage()
	server := handler.NewServer(memStorage)

	log.Println("Running server on", configFlags.RunAddress)
	err := http.ListenAndServe(configFlags.RunAddress, handler.MetricsRouter(server))
	if err != nil {
		log.Fatal("Server has finished with error: ", err)
	}
}
