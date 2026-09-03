package main

import (
	"log"
	"net/http"

	handler "github.com/lebendig13/metrics/internal/handler"
	models "github.com/lebendig13/metrics/internal/model"
)

func main() {
	parseFlags()

	memStorage := models.NewMemStorage()
	server := handler.NewServer(memStorage)

	log.Println("Running server on", flagRunAddr)
	err := http.ListenAndServe(flagRunAddr, handler.MetricsRouter(server))
	if err != nil {
		log.Fatal("Server has finished with error: ", err)
	}
}
