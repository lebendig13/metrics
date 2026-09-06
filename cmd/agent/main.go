package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	config "github.com/lebendig13/metrics/internal/config"
	handler "github.com/lebendig13/metrics/internal/handler"
)

func main() {
	configFlags := config.ParseAgentFlags()
	baseURL := fmt.Sprintf("http://%s/update/", configFlags.ServerAddress)
	pollInterval := time.Duration(configFlags.Intervals.PollInterval) * time.Second
	reportInterval := time.Duration(configFlags.Intervals.ReportInterval) * time.Second
	log.Println("Server address: ", baseURL, "; pollInterval = ", pollInterval, "; reportInterval = ", reportInterval)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	rng := rand.New(rand.NewSource((time.Now().UnixNano())))
	agent := NewAgent()
	log.Println("Agent started")

	var timeSinceLastReport time.Duration
	for {
		time.Sleep(pollInterval)

		currentMetrics := agent.GetMetrics(rng)
		log.Println("Got all metrics. PollCount = ", agent.pollCount)

		timeSinceLastReport += pollInterval
		if timeSinceLastReport >= reportInterval {
			log.Println("Start sending metrics to server")
			err := handler.SendMetrics(client, currentMetrics, baseURL)
			if err == nil {
				agent.pollCount = 0
			} else {
				log.Println(err)
			}

			timeSinceLastReport = 0
		}
	}
}
